import { useApiClient } from "@/lib/api-context";
import type { CreateFusionRequestInput, FusionRequest } from "@/types/domain";
import { useCallback, useState } from "react";

interface UseFusionRequestsReturn {
	readonly items: readonly FusionRequest[];
	readonly totalCount: number;
	readonly nextPageToken: string;
	readonly loading: boolean;
	readonly error: string | null;
	readonly fetchRequests: (statusFilter?: string, pageToken?: string) => Promise<void>;
	readonly createRequest: (input: CreateFusionRequestInput) => Promise<FusionRequest>;
	readonly respondToRequest: (id: string, response: "APPROVED" | "REJECTED", foodItemId?: string) => Promise<void>;
}

const PAGE_SIZE = 20;

/** 融通リクエストの操作フック。 */
export function useFusionRequests(): UseFusionRequestsReturn {
	const api = useApiClient();
	const [items, setItems] = useState<readonly FusionRequest[]>([]);
	const [totalCount, setTotalCount] = useState(0);
	const [nextPageToken, setNextPageToken] = useState("");
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const fetchRequests = useCallback(
		async (statusFilter = "", pageToken = "") => {
			setLoading(true);
			setError(null);
			try {
				const result = await api.listFusionRequests(PAGE_SIZE, pageToken, statusFilter);
				if (pageToken) {
					setItems((prev) => [...prev, ...result.items]);
				} else {
					setItems(result.items);
				}
				setTotalCount(result.pagination.totalCount);
				setNextPageToken(result.pagination.nextPageToken);
			} catch (e) {
				const message = e instanceof Error ? e.message : "融通リクエスト一覧の取得に失敗しました";
				setError(message);
			} finally {
				setLoading(false);
			}
		},
		[api],
	);

	const createRequest = useCallback(
		async (input: CreateFusionRequestInput): Promise<FusionRequest> => {
			setError(null);
			try {
				const created = await api.createFusionRequest(input);
				return created;
			} catch (e) {
				const message = e instanceof Error ? e.message : "融通リクエストの作成に失敗しました";
				setError(message);
				throw e;
			}
		},
		[api],
	);

	const respondToRequest = useCallback(
		async (id: string, response: "APPROVED" | "REJECTED", foodItemId = "") => {
			setError(null);
			try {
				const updated = await api.respondToFusionRequest(id, response, foodItemId);
				setItems((prev) => prev.map((item) => (item.id === id ? updated : item)));
			} catch (e) {
				const message = e instanceof Error ? e.message : "融通リクエストへの応答に失敗しました";
				setError(message);
				throw e;
			}
		},
		[api],
	);

	return { items, totalCount, nextPageToken, loading, error, fetchRequests, createRequest, respondToRequest };
}
