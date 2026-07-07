import { useCallback, useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useFoodItems } from "@/hooks/useFoodItems";
import { useFusionRequests } from "@/hooks/useFusionRequests";
import type { FusionRequest, FusionRequestStatus } from "@/types/domain";

const STATUS_LABELS: Record<FusionRequestStatus, string> = {
	pending: "承認待ち",
	approved: "承認済み",
	rejected: "拒否済み",
	completed: "完了",
};

const STATUS_BADGE_CLASS: Record<FusionRequestStatus, string> = {
	pending: "badge badge-pending",
	approved: "badge badge-approved",
	rejected: "badge badge-rejected",
	completed: "badge badge-completed",
};

const STATUS_FILTERS: readonly { value: string; label: string }[] = [
	{ value: "", label: "すべてのステータス" },
	{ value: "pending", label: "承認待ち" },
	{ value: "approved", label: "承認済み" },
	{ value: "rejected", label: "拒否済み" },
	{ value: "completed", label: "完了" },
];

interface FoodItemOption {
	readonly id: string;
	readonly label: string;
}

function formatDate(dateStr: string): string {
	try {
		const date = new Date(dateStr);
		return date.toLocaleDateString("ja-JP");
	} catch {
		return dateStr;
	}
}

/** 承認・拒否操作の対象と種別。 */
interface PendingAction {
	readonly request: FusionRequest;
	readonly type: "APPROVED" | "REJECTED";
}

/** 融通リクエスト一覧コンポーネント。 */
export function FusionRequestList(): React.ReactElement {
	const { items, totalCount, nextPageToken, loading, error, fetchRequests, respondToRequest } = useFusionRequests();
	const { items: foodItems, loading: foodLoading, fetchItems: fetchFoodItems } = useFoodItems();
	const [statusFilter, setStatusFilter] = useState("");
	const [action, setAction] = useState<PendingAction | null>(null);
	const [selectedFoodItemId, setSelectedFoodItemId] = useState("");
	const [submitting, setSubmitting] = useState(false);

	useEffect(() => {
		void fetchRequests(statusFilter);
	}, [statusFilter, fetchRequests]);

	const handleLoadMore = useCallback(() => {
		if (nextPageToken) {
			void fetchRequests(statusFilter, nextPageToken);
		}
	}, [statusFilter, nextPageToken, fetchRequests]);

	const closeAction = useCallback(() => {
		setAction(null);
		setSelectedFoodItemId("");
	}, []);

	// 承認操作を開始し、希望カテゴリに一致する食品アイテムを読み込む。
	const openApprove = useCallback(
		(request: FusionRequest) => {
			setAction({ request, type: "APPROVED" });
			setSelectedFoodItemId("");
			void fetchFoodItems(request.desiredCategory);
		},
		[fetchFoodItems],
	);

	const openReject = useCallback((request: FusionRequest) => {
		setAction({ request, type: "REJECTED" });
	}, []);

	// 承認可能な食品アイテム: 在庫あり・希望カテゴリ/単位一致・数量充足のもの。
	const foodItemOptions = useMemo<readonly FoodItemOption[]>(() => {
		if (action?.type !== "APPROVED") {
			return [];
		}
		const req = action.request;
		return foodItems
			.filter(
				(item) =>
					item.status === "available" &&
					item.category === req.desiredCategory &&
					item.unit === req.unit &&
					item.quantity >= req.desiredQuantity,
			)
			.map((item) => ({ id: item.id, label: `${item.name}（${String(item.quantity)}${item.unit}）` }));
	}, [foodItems, action]);

	const confirmApprove = useCallback(async () => {
		if (action === null || !selectedFoodItemId) {
			return;
		}
		setSubmitting(true);
		try {
			await respondToRequest(action.request.id, "APPROVED", selectedFoodItemId);
			closeAction();
		} catch {
			// エラーはフックで処理済み
		} finally {
			setSubmitting(false);
		}
	}, [action, selectedFoodItemId, respondToRequest, closeAction]);

	const confirmReject = useCallback(async () => {
		if (action === null) {
			return;
		}
		setSubmitting(true);
		try {
			await respondToRequest(action.request.id, "REJECTED", "");
			closeAction();
		} catch {
			// エラーはフックで処理済み
		} finally {
			setSubmitting(false);
		}
	}, [action, respondToRequest, closeAction]);

	return (
		<div>
			<div className="page-header">
				<h1 className="page-title">融通リクエスト一覧</h1>
				<p className="page-description">食材融通リクエストの一覧（全{String(totalCount)}件）</p>
			</div>

			{error ? <div className="alert alert-error">{error}</div> : null}

			<div className="filter-bar">
				<select
					className="form-select"
					value={statusFilter}
					onChange={(e) => setStatusFilter(e.target.value)}
					aria-label="ステータスフィルタ"
				>
					{STATUS_FILTERS.map((f) => (
						<option key={f.value} value={f.value}>
							{f.label}
						</option>
					))}
				</select>
				<Link to="/fusion/new" className="btn btn-primary">
					新規リクエスト
				</Link>
			</div>

			{loading && items.length === 0 ? (
				<div className="loading">読み込み中...</div>
			) : items.length === 0 ? (
				<div className="empty-state">
					<p>融通リクエストはありません</p>
					<Link to="/fusion/new" className="btn btn-primary">
						リクエストを作成する
					</Link>
				</div>
			) : (
				<>
					<div className="data-list">
						{items.map((request) => (
							<div key={request.id} className="data-item">
								<div className="data-item-info">
									<div className="data-item-name">
										{request.desiredCategory} - {String(request.desiredQuantity)} {request.unit}
									</div>
									<div className="data-item-meta">
										<span>食堂: {request.requesterShokudoId}</span>
										{request.message ? <span>{request.message}</span> : null}
										<span>作成: {formatDate(request.createdAt)}</span>
										<span className={STATUS_BADGE_CLASS[request.status]}>{STATUS_LABELS[request.status]}</span>
									</div>
								</div>
								<RequestActions
									request={request}
									isActive={action?.request.id === request.id}
									actionType={action?.type ?? null}
									foodLoading={foodLoading}
									foodItemOptions={foodItemOptions}
									selectedFoodItemId={selectedFoodItemId}
									submitting={submitting}
									onOpenApprove={openApprove}
									onOpenReject={openReject}
									onSelectFood={setSelectedFoodItemId}
									onConfirmApprove={() => void confirmApprove()}
									onConfirmReject={() => void confirmReject()}
									onCancel={closeAction}
								/>
							</div>
						))}
					</div>

					{nextPageToken ? (
						<div className="pagination">
							<button type="button" className="btn btn-outline" onClick={handleLoadMore} disabled={loading}>
								{loading ? "読み込み中..." : "もっと見る"}
							</button>
						</div>
					) : null}
				</>
			)}
		</div>
	);
}

/** pending リクエストの承認/拒否操作。開いていれば操作パネル、そうでなければ承認・拒否ボタンを表示する。 */
function RequestActions({
	request,
	isActive,
	actionType,
	foodLoading,
	foodItemOptions,
	selectedFoodItemId,
	submitting,
	onOpenApprove,
	onOpenReject,
	onSelectFood,
	onConfirmApprove,
	onConfirmReject,
	onCancel,
}: {
	readonly request: FusionRequest;
	readonly isActive: boolean;
	readonly actionType: "APPROVED" | "REJECTED" | null;
	readonly foodLoading: boolean;
	readonly foodItemOptions: readonly FoodItemOption[];
	readonly selectedFoodItemId: string;
	readonly submitting: boolean;
	readonly onOpenApprove: (request: FusionRequest) => void;
	readonly onOpenReject: (request: FusionRequest) => void;
	readonly onSelectFood: (id: string) => void;
	readonly onConfirmApprove: () => void;
	readonly onConfirmReject: () => void;
	readonly onCancel: () => void;
}): React.ReactElement | null {
	if (request.status !== "pending") {
		return null;
	}

	if (!isActive) {
		return (
			<div className="data-item-actions">
				<button type="button" className="btn btn-success btn-sm" onClick={() => onOpenApprove(request)}>
					承認
				</button>
				<button type="button" className="btn btn-danger btn-sm" onClick={() => onOpenReject(request)}>
					拒否
				</button>
			</div>
		);
	}

	return (
		<div className="data-item-actions">
			{actionType === "APPROVED" ? (
				<ApprovePanel
					foodLoading={foodLoading}
					options={foodItemOptions}
					selectedFoodItemId={selectedFoodItemId}
					submitting={submitting}
					onSelect={onSelectFood}
					onConfirm={onConfirmApprove}
					onCancel={onCancel}
				/>
			) : (
				<RejectPanel submitting={submitting} onConfirm={onConfirmReject} onCancel={onCancel} />
			)}
		</div>
	);
}

/** 承認時に提供する食品アイテムを選択するパネル。 */
function ApprovePanel({
	foodLoading,
	options,
	selectedFoodItemId,
	submitting,
	onSelect,
	onConfirm,
	onCancel,
}: {
	readonly foodLoading: boolean;
	readonly options: readonly FoodItemOption[];
	readonly selectedFoodItemId: string;
	readonly submitting: boolean;
	readonly onSelect: (id: string) => void;
	readonly onConfirm: () => void;
	readonly onCancel: () => void;
}): React.ReactElement {
	return (
		<>
			{foodLoading ? (
				<span className="loading">読み込み中...</span>
			) : options.length === 0 ? (
				<span className="form-error">承認可能な食品アイテムがありません</span>
			) : (
				<select
					className="form-select"
					aria-label="提供する食品アイテム"
					value={selectedFoodItemId}
					onChange={(e) => onSelect(e.target.value)}
				>
					<option value="">選択してください</option>
					{options.map((opt) => (
						<option key={opt.id} value={opt.id}>
							{opt.label}
						</option>
					))}
				</select>
			)}
			<button
				type="button"
				className="btn btn-success btn-sm"
				onClick={onConfirm}
				disabled={submitting || !selectedFoodItemId}
			>
				確定
			</button>
			<button type="button" className="btn btn-outline btn-sm" onClick={onCancel} disabled={submitting}>
				キャンセル
			</button>
		</>
	);
}

/** 拒否の確認パネル。 */
function RejectPanel({
	submitting,
	onConfirm,
	onCancel,
}: {
	readonly submitting: boolean;
	readonly onConfirm: () => void;
	readonly onCancel: () => void;
}): React.ReactElement {
	return (
		<>
			<span>このリクエストを拒否しますか？</span>
			<button type="button" className="btn btn-danger btn-sm" onClick={onConfirm} disabled={submitting}>
				拒否する
			</button>
			<button type="button" className="btn btn-outline btn-sm" onClick={onCancel} disabled={submitting}>
				キャンセル
			</button>
		</>
	);
}
