import type {
	CreateFoodItemInput,
	CreateFusionRequestInput,
	FoodCategory,
	FoodItem,
	FusionRequest,
	PaginationInfo,
} from "@/types/domain";

/** gRPC エラーコード。 */
type GrpcErrorCode =
	| "INVALID_ARGUMENT"
	| "NOT_FOUND"
	| "ALREADY_EXISTS"
	| "PERMISSION_DENIED"
	| "UNAUTHENTICATED"
	| "UNAVAILABLE"
	| "INTERNAL"
	| "DEADLINE_EXCEEDED"
	| "UNKNOWN";

/** gRPC エラーコードに応じたユーザーフレンドリーなメッセージ。 */
const GRPC_ERROR_MESSAGES: Record<GrpcErrorCode, string> = {
	INVALID_ARGUMENT: "入力内容に問題があります。内容を確認してください。",
	NOT_FOUND: "指定されたリソースが見つかりません。",
	ALREADY_EXISTS: "同じリソースが既に存在します。",
	PERMISSION_DENIED: "この操作を行う権限がありません。",
	UNAUTHENTICATED: "認証が必要です。ログインしてください。",
	UNAVAILABLE: "サーバーに接続できません。しばらく待ってから再度お試しください。",
	INTERNAL: "サーバーでエラーが発生しました。",
	DEADLINE_EXCEEDED: "リクエストがタイムアウトしました。",
	UNKNOWN: "予期しないエラーが発生しました。",
};

function isGrpcErrorCode(code: string): code is GrpcErrorCode {
	return code in GRPC_ERROR_MESSAGES;
}

const DEFAULT_ERROR_MESSAGE = "予期しないエラーが発生しました。";

/** API エラー。 */
export class ApiError extends Error {
	readonly code: string;
	readonly userMessage: string;

	constructor(code: string, message: string) {
		super(message);
		this.name = "ApiError";
		this.code = code;
		this.userMessage = isGrpcErrorCode(code) ? GRPC_ERROR_MESSAGES[code] : DEFAULT_ERROR_MESSAGE;
	}
}

/** 食品一覧のレスポンス。 */
export interface ListFoodItemsResult {
	readonly items: readonly FoodItem[];
	readonly pagination: PaginationInfo;
}

/** 融通リクエスト一覧のレスポンス。 */
export interface ListFusionRequestsResult {
	readonly items: readonly FusionRequest[];
	readonly pagination: PaginationInfo;
}

/** API クライアントインターフェース。 */
export interface ApiClient {
	createFoodItem(input: CreateFoodItemInput): Promise<FoodItem>;
	listFoodItems(pageSize: number, pageToken: string, categoryFilter: FoodCategory | ""): Promise<ListFoodItemsResult>;
	deleteFoodItem(id: string): Promise<void>;
	createFusionRequest(input: CreateFusionRequestInput): Promise<FusionRequest>;
	listFusionRequests(pageSize: number, pageToken: string, statusFilter: string): Promise<ListFusionRequestsResult>;
	respondToFusionRequest(id: string, response: "APPROVED" | "REJECTED", foodItemId: string): Promise<FusionRequest>;
}
