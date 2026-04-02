import { useCallback, useEffect, useState } from "react";
import { Link } from "react-router-dom";
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

function formatDate(dateStr: string): string {
	try {
		const date = new Date(dateStr);
		return date.toLocaleDateString("ja-JP");
	} catch {
		return dateStr;
	}
}

/** 融通リクエスト一覧コンポーネント。 */
export function FusionRequestList(): React.ReactElement {
	const { items, totalCount, nextPageToken, loading, error, fetchRequests, respondToRequest } = useFusionRequests();
	const [statusFilter, setStatusFilter] = useState("");
	const [responding, setResponding] = useState<string | null>(null);

	useEffect(() => {
		void fetchRequests(statusFilter);
	}, [statusFilter, fetchRequests]);

	const handleLoadMore = useCallback(() => {
		if (nextPageToken) {
			void fetchRequests(statusFilter, nextPageToken);
		}
	}, [statusFilter, nextPageToken, fetchRequests]);

	const handleRespond = useCallback(
		async (request: FusionRequest, response: "APPROVED" | "REJECTED") => {
			let foodItemId = "";
			if (response === "APPROVED") {
				const input = window.prompt("提供する食品アイテムIDを入力してください:");
				if (input === null) return;
				foodItemId = input.trim();
				if (!foodItemId) {
					return;
				}
			}
			const label = response === "APPROVED" ? "承認" : "拒否";
			if (!window.confirm(`このリクエストを${label}しますか？`)) {
				return;
			}
			setResponding(request.id);
			try {
				await respondToRequest(request.id, response, foodItemId);
			} catch {
				// エラーはフックで処理済み
			} finally {
				setResponding(null);
			}
		},
		[respondToRequest],
	);

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
								{request.status === "pending" ? (
									<div className="data-item-actions">
										<button
											type="button"
											className="btn btn-success btn-sm"
											onClick={() => void handleRespond(request, "APPROVED")}
											disabled={responding === request.id}
										>
											承認
										</button>
										<button
											type="button"
											className="btn btn-danger btn-sm"
											onClick={() => void handleRespond(request, "REJECTED")}
											disabled={responding === request.id}
										>
											拒否
										</button>
									</div>
								) : null}
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
