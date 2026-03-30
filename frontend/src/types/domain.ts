/** 食品カテゴリの定数。 */
export const FOOD_CATEGORIES = ["野菜", "肉", "魚", "乳製品", "穀物", "その他"] as const;
export type FoodCategory = (typeof FOOD_CATEGORIES)[number];

/** 食品の単位の定数。 */
export const FOOD_UNITS = ["kg", "個", "パック", "本", "袋", "箱"] as const;
export type FoodUnit = (typeof FOOD_UNITS)[number];

/** 食品アイテムのステータス。 */
export type FoodItemStatus = "available" | "reserved" | "consumed" | "expired";

/** 食品アイテムのドメインモデル。 */
export interface FoodItem {
	readonly id: string;
	readonly name: string;
	readonly category: FoodCategory;
	readonly expiryDate: string;
	readonly quantity: number;
	readonly unit: FoodUnit;
	readonly donorId: string;
	readonly status: FoodItemStatus;
	readonly createdAt: string;
	readonly updatedAt: string;
}

/** 融通リクエストのステータス。 */
export type FusionRequestStatus = "pending" | "approved" | "rejected" | "completed";

/** 融通リクエストのドメインモデル。 */
export interface FusionRequest {
	readonly id: string;
	readonly requesterShokudoId: string;
	readonly desiredCategory: FoodCategory;
	readonly desiredQuantity: number;
	readonly unit: FoodUnit;
	readonly message: string;
	readonly status: FusionRequestStatus;
	readonly responderShokudoId: string;
	readonly foodItemId: string;
	readonly createdAt: string;
	readonly updatedAt: string;
}

/** 食品登録フォームの入力値。 */
export interface CreateFoodItemInput {
	readonly name: string;
	readonly category: FoodCategory | "";
	readonly expiryDate: string;
	readonly quantity: number | "";
	readonly unit: FoodUnit | "";
	readonly donorId: string;
}

/** 融通リクエスト作成フォームの入力値。 */
export interface CreateFusionRequestInput {
	readonly requesterShokudoId: string;
	readonly desiredCategory: FoodCategory | "";
	readonly desiredQuantity: number | "";
	readonly unit: FoodUnit | "";
	readonly message: string;
}

/** ページネーション情報。 */
export interface PaginationInfo {
	readonly nextPageToken: string;
	readonly totalCount: number;
}
