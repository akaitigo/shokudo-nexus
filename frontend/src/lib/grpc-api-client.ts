import type { ServiceType } from "@bufbuild/protobuf";
import { createPromiseClient } from "@connectrpc/connect";
import { createGrpcWebTransport } from "@connectrpc/connect-web";
import type { ApiClient, ListFoodItemsResult, ListFusionRequestsResult } from "@/lib/api-client";
import { ApiError } from "@/lib/api-client";
import type {
	CreateFoodItemInput,
	CreateFusionRequestInput,
	FoodCategory,
	FoodItem,
	FusionRequest,
} from "@/types/domain";

/** gRPC レスポンスの食品アイテムフィールド。 */
interface FoodItemFields {
	id?: unknown;
	name?: unknown;
	category?: unknown;
	expiryDate?: unknown;
	quantity?: unknown;
	unit?: unknown;
	donorId?: unknown;
	status?: unknown;
	createdAt?: unknown;
	updatedAt?: unknown;
}

/** gRPC レスポンスの融通リクエストフィールド。 */
interface FusionRequestFields {
	id?: unknown;
	requesterShokudoId?: unknown;
	desiredCategory?: unknown;
	desiredQuantity?: unknown;
	unit?: unknown;
	message?: unknown;
	status?: unknown;
	responderShokudoId?: unknown;
	foodItemId?: unknown;
	createdAt?: unknown;
	updatedAt?: unknown;
}

/** gRPC レスポンスの一覧フィールド。 */
interface ListResponseFields {
	items?: unknown[];
	nextPageToken?: unknown;
	totalCount?: unknown;
}

/** サービス定義の注入パラメータ。gen ファイルへの直接依存を回避する。 */
export interface GrpcServiceDefs {
	readonly foodInventoryService: ServiceType;
	readonly fusionService: ServiceType;
}

/**
 * RPC メソッドを安全に呼び出すヘルパー。
 * createPromiseClient が返す Client<ServiceType> の各メソッドは
 * index signature 由来で undefined になりうるため、
 * 存在チェック付きで呼び出す。
 */
type RpcCaller = Record<string, ((request: unknown) => Promise<unknown>) | undefined>;

function callRpc(client: RpcCaller, method: string, request: unknown): Promise<unknown> {
	const fn = client[method];
	if (fn === undefined) {
		throw new ApiError("INTERNAL", `RPC method "${method}" not found on client`);
	}
	return fn(request);
}

/**
 * Connect gRPC-Web トランスポートを使用する実 API クライアントを作成する。
 *
 * @param baseUrl - gRPC バックエンドの URL（例: "http://localhost:8080"）
 * @param services - 生成されたサービス定義（FoodInventoryService, FusionService）
 */
export function createGrpcApiClient(baseUrl: string, services: GrpcServiceDefs): ApiClient {
	const transport = createGrpcWebTransport({ baseUrl });
	const foodClient = createPromiseClient(services.foodInventoryService, transport) as unknown as RpcCaller;
	const fusionClient = createPromiseClient(services.fusionService, transport) as unknown as RpcCaller;

	return {
		async createFoodItem(input: CreateFoodItemInput): Promise<FoodItem> {
			try {
				const response = await callRpc(foodClient, "createFoodItem", {
					name: input.name,
					category: input.category || undefined,
					expiryDate: input.expiryDate,
					quantity: typeof input.quantity === "number" ? input.quantity : 0,
					unit: input.unit || undefined,
					donorId: input.donorId,
				});
				return mapFoodItemResponse(response as FoodItemFields);
			} catch (error) {
				throw toApiError(error);
			}
		},

		async listFoodItems(
			pageSize: number,
			pageToken: string,
			categoryFilter: FoodCategory | "",
		): Promise<ListFoodItemsResult> {
			try {
				const response = await callRpc(foodClient, "listFoodItems", {
					pageSize,
					pageToken,
					categoryFilter: categoryFilter || undefined,
				});
				const fields = response as ListResponseFields;
				const rawItems = fields.items ?? [];
				return {
					items: rawItems.map((item) => mapFoodItemResponse(item as FoodItemFields)),
					pagination: {
						nextPageToken: String(fields.nextPageToken ?? ""),
						totalCount: Number(fields.totalCount ?? 0),
					},
				};
			} catch (error) {
				throw toApiError(error);
			}
		},

		async deleteFoodItem(id: string): Promise<void> {
			try {
				await callRpc(foodClient, "deleteFoodItem", { id });
			} catch (error) {
				throw toApiError(error);
			}
		},

		async createFusionRequest(input: CreateFusionRequestInput): Promise<FusionRequest> {
			try {
				const response = await callRpc(fusionClient, "createFusionRequest", {
					requesterShokudoId: input.requesterShokudoId,
					desiredCategory: input.desiredCategory || undefined,
					desiredQuantity: typeof input.desiredQuantity === "number" ? input.desiredQuantity : 0,
					unit: input.unit || undefined,
					message: input.message,
				});
				return mapFusionRequestResponse(response as FusionRequestFields);
			} catch (error) {
				throw toApiError(error);
			}
		},

		async listFusionRequests(
			pageSize: number,
			pageToken: string,
			statusFilter: string,
		): Promise<ListFusionRequestsResult> {
			try {
				const response = await callRpc(fusionClient, "listFusionRequests", {
					pageSize,
					pageToken,
					statusFilter: statusFilter || undefined,
				});
				const fields = response as ListResponseFields;
				const rawItems = fields.items ?? [];
				return {
					items: rawItems.map((item) => mapFusionRequestResponse(item as FusionRequestFields)),
					pagination: {
						nextPageToken: String(fields.nextPageToken ?? ""),
						totalCount: Number(fields.totalCount ?? 0),
					},
				};
			} catch (error) {
				throw toApiError(error);
			}
		},

		async respondToFusionRequest(
			id: string,
			response: "APPROVED" | "REJECTED",
			foodItemId: string,
		): Promise<FusionRequest> {
			try {
				const res = await callRpc(fusionClient, "respondToFusionRequest", {
					id,
					response,
					foodItemId,
				});
				return mapFusionRequestResponse(res as FusionRequestFields);
			} catch (error) {
				throw toApiError(error);
			}
		},
	};
}

/**
 * gRPC レスポンスから FoodItem ドメインモデルに変換する。
 */
function mapFoodItemResponse(fields: FoodItemFields): FoodItem {
	return {
		id: String(fields.id ?? ""),
		name: String(fields.name ?? ""),
		category: String(fields.category ?? "") as FoodItem["category"],
		expiryDate: String(fields.expiryDate ?? ""),
		quantity: Number(fields.quantity ?? 0),
		unit: String(fields.unit ?? "") as FoodItem["unit"],
		donorId: String(fields.donorId ?? ""),
		status: String(fields.status ?? "available") as FoodItem["status"],
		createdAt: String(fields.createdAt ?? ""),
		updatedAt: String(fields.updatedAt ?? ""),
	};
}

/** gRPC レスポンスから FusionRequest ドメインモデルに変換する。 */
function mapFusionRequestResponse(fields: FusionRequestFields): FusionRequest {
	return {
		id: String(fields.id ?? ""),
		requesterShokudoId: String(fields.requesterShokudoId ?? ""),
		desiredCategory: String(fields.desiredCategory ?? "") as FusionRequest["desiredCategory"],
		desiredQuantity: Number(fields.desiredQuantity ?? 0),
		unit: String(fields.unit ?? "") as FusionRequest["unit"],
		message: String(fields.message ?? ""),
		status: String(fields.status ?? "pending") as FusionRequest["status"],
		responderShokudoId: String(fields.responderShokudoId ?? ""),
		foodItemId: String(fields.foodItemId ?? ""),
		createdAt: String(fields.createdAt ?? ""),
		updatedAt: String(fields.updatedAt ?? ""),
	};
}

/** Connect エラーを ApiError に変換する。 */
function toApiError(error: unknown): ApiError {
	if (error instanceof ApiError) {
		return error;
	}
	if (error instanceof Error && "code" in error) {
		const connectError = error as Error & { code: unknown };
		return new ApiError(String(connectError.code), error.message);
	}
	if (error instanceof Error) {
		return new ApiError("UNKNOWN", error.message);
	}
	return new ApiError("UNKNOWN", String(error));
}
