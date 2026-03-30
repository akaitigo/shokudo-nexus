import { useFoodItems } from "@/hooks/useFoodItems";
import { FOOD_CATEGORIES } from "@/types/domain";
import type { FoodCategory, FoodItem } from "@/types/domain";
import { getExpiryLevel } from "@/types/validation";
import type { ExpiryLevel } from "@/types/validation";
import { useCallback, useEffect, useState } from "react";
import { Link } from "react-router-dom";

function expiryBadgeClass(level: ExpiryLevel): string {
	switch (level) {
		case "danger":
			return "badge badge-danger";
		case "warning":
			return "badge badge-warning";
		case "safe":
			return "badge badge-safe";
	}
}

function expiryLabel(level: ExpiryLevel): string {
	switch (level) {
		case "danger":
			return "3日以内";
		case "warning":
			return "7日以内";
		case "safe":
			return "余裕あり";
	}
}

function formatDate(dateStr: string): string {
	try {
		const date = new Date(dateStr);
		return date.toLocaleDateString("ja-JP");
	} catch {
		return dateStr;
	}
}

/** 食品在庫一覧コンポーネント。 */
export function FoodItemList(): React.ReactElement {
	const { items, totalCount, nextPageToken, loading, error, fetchItems, deleteItem } = useFoodItems();
	const [categoryFilter, setCategoryFilter] = useState<FoodCategory | "">("");
	const [deleting, setDeleting] = useState<string | null>(null);

	useEffect(() => {
		void fetchItems(categoryFilter);
	}, [categoryFilter, fetchItems]);

	const handleLoadMore = useCallback(() => {
		if (nextPageToken) {
			void fetchItems(categoryFilter, nextPageToken);
		}
	}, [categoryFilter, nextPageToken, fetchItems]);

	const handleDelete = useCallback(
		async (item: FoodItem) => {
			if (!window.confirm(`「${item.name}」を削除しますか？`)) {
				return;
			}
			setDeleting(item.id);
			try {
				await deleteItem(item.id);
			} catch {
				// エラーはフックで処理済み
			} finally {
				setDeleting(null);
			}
		},
		[deleteItem],
	);

	return (
		<div>
			<div className="page-header">
				<h1 className="page-title">在庫一覧</h1>
				<p className="page-description">登録されている食品の一覧（全{String(totalCount)}件）</p>
			</div>

			{error ? <div className="alert alert-error">{error}</div> : null}

			<div className="filter-bar">
				<select
					className="form-select"
					value={categoryFilter}
					onChange={(e) => setCategoryFilter(e.target.value as FoodCategory | "")}
					aria-label="カテゴリフィルタ"
				>
					<option value="">すべてのカテゴリ</option>
					{FOOD_CATEGORIES.map((cat) => (
						<option key={cat} value={cat}>
							{cat}
						</option>
					))}
				</select>
				<Link to="/food/new" className="btn btn-primary">
					食品を登録
				</Link>
			</div>

			{loading && items.length === 0 ? (
				<div className="loading">読み込み中...</div>
			) : items.length === 0 ? (
				<div className="empty-state">
					<p>登録されている食品はありません</p>
					<Link to="/food/new" className="btn btn-primary">
						食品を登録する
					</Link>
				</div>
			) : (
				<>
					<div className="data-list">
						{items.map((item) => {
							const level = getExpiryLevel(item.expiryDate);
							return (
								<div key={item.id} className="data-item">
									<div className="data-item-info">
										<div className="data-item-name">{item.name}</div>
										<div className="data-item-meta">
											<span>{item.category}</span>
											<span>
												{String(item.quantity)} {item.unit}
											</span>
											<span>期限: {formatDate(item.expiryDate)}</span>
											<span className={expiryBadgeClass(level)}>{expiryLabel(level)}</span>
										</div>
									</div>
									<div className="data-item-actions">
										<button
											type="button"
											className="btn btn-danger btn-sm"
											onClick={() => void handleDelete(item)}
											disabled={deleting === item.id}
										>
											{deleting === item.id ? "削除中..." : "削除"}
										</button>
									</div>
								</div>
							);
						})}
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
