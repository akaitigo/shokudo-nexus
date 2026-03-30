import { useApiClient } from "@/lib/api-context";
import type { CreateFoodItemInput, FoodCategory, FoodItem } from "@/types/domain";
import { useCallback, useState } from "react";

interface UseFoodItemsReturn {
	readonly items: readonly FoodItem[];
	readonly totalCount: number;
	readonly nextPageToken: string;
	readonly loading: boolean;
	readonly error: string | null;
	readonly fetchItems: (categoryFilter: FoodCategory | "", pageToken?: string) => Promise<void>;
	readonly createItem: (input: CreateFoodItemInput) => Promise<FoodItem>;
	readonly deleteItem: (id: string) => Promise<void>;
}

const PAGE_SIZE = 20;

/** 食品在庫の操作フック。 */
export function useFoodItems(): UseFoodItemsReturn {
	const api = useApiClient();
	const [items, setItems] = useState<readonly FoodItem[]>([]);
	const [totalCount, setTotalCount] = useState(0);
	const [nextPageToken, setNextPageToken] = useState("");
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const fetchItems = useCallback(
		async (categoryFilter: FoodCategory | "", pageToken = "") => {
			setLoading(true);
			setError(null);
			try {
				const result = await api.listFoodItems(PAGE_SIZE, pageToken, categoryFilter);
				if (pageToken) {
					setItems((prev) => [...prev, ...result.items]);
				} else {
					setItems(result.items);
				}
				setTotalCount(result.pagination.totalCount);
				setNextPageToken(result.pagination.nextPageToken);
			} catch (e) {
				const message = e instanceof Error ? e.message : "食品一覧の取得に失敗しました";
				setError(message);
			} finally {
				setLoading(false);
			}
		},
		[api],
	);

	const createItem = useCallback(
		async (input: CreateFoodItemInput): Promise<FoodItem> => {
			setError(null);
			try {
				const created = await api.createFoodItem(input);
				return created;
			} catch (e) {
				const message = e instanceof Error ? e.message : "食品の登録に失敗しました";
				setError(message);
				throw e;
			}
		},
		[api],
	);

	const deleteItem = useCallback(
		async (id: string) => {
			setError(null);
			try {
				await api.deleteFoodItem(id);
				setItems((prev) => prev.filter((item) => item.id !== id));
			} catch (e) {
				const message = e instanceof Error ? e.message : "食品の削除に失敗しました";
				setError(message);
				throw e;
			}
		},
		[api],
	);

	return { items, totalCount, nextPageToken, loading, error, fetchItems, createItem, deleteItem };
}
