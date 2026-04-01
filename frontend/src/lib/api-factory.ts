import type { ApiClient } from "@/lib/api-client";
import { mockApiClient } from "@/lib/mock-api-client";

/**
 * 環境に応じた ApiClient を作成する。
 *
 * - `VITE_API_URL` が設定されている場合: gRPC-Web クライアントを返す
 * - 未設定の場合: モック API クライアントを返す（開発用デフォルト）
 *
 * gRPC クライアントは動的 import で遅延ロードする。
 * これにより service_pb.js が未生成の開発環境でもモックで動作する。
 */
export function createApiClient(): ApiClient {
	const apiUrl = import.meta.env.VITE_API_URL;
	if (!apiUrl) {
		return mockApiClient;
	}
	return createLazyGrpcClient(apiUrl);
}

/**
 * gRPC クライアントを遅延初期化するプロキシを返す。
 * 各メソッド呼び出し時に初めて grpc-api-client モジュールをロードする。
 */
function createLazyGrpcClient(apiUrl: string): ApiClient {
	let resolvedClient: ApiClient | null = null;

	async function getClient(): Promise<ApiClient> {
		if (resolvedClient === null) {
			const { createGrpcApiClient } = await import("@/lib/grpc-api-client");
			resolvedClient = createGrpcApiClient(apiUrl);
		}
		return resolvedClient;
	}

	return {
		async createFoodItem(input) {
			const client = await getClient();
			return client.createFoodItem(input);
		},
		async listFoodItems(pageSize, pageToken, categoryFilter) {
			const client = await getClient();
			return client.listFoodItems(pageSize, pageToken, categoryFilter);
		},
		async deleteFoodItem(id) {
			const client = await getClient();
			return client.deleteFoodItem(id);
		},
		async createFusionRequest(input) {
			const client = await getClient();
			return client.createFusionRequest(input);
		},
		async listFusionRequests(pageSize, pageToken, statusFilter) {
			const client = await getClient();
			return client.listFusionRequests(pageSize, pageToken, statusFilter);
		},
		async respondToFusionRequest(id, response, foodItemId) {
			const client = await getClient();
			return client.respondToFusionRequest(id, response, foodItemId);
		},
	};
}
