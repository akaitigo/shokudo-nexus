import type { CreateFoodItemInput, CreateFusionRequestInput, FoodCategory, FoodUnit } from "./domain";
import { FOOD_CATEGORIES, FOOD_UNITS } from "./domain";

/** 食品登録バリデーションエラー。 */
export interface FoodItemErrors {
	name?: string;
	category?: string;
	expiryDate?: string;
	quantity?: string;
	unit?: string;
}

/** 融通リクエストバリデーションエラー。 */
export interface FusionRequestErrors {
	requesterShokudoId?: string;
	desiredCategory?: string;
	desiredQuantity?: string;
	unit?: string;
}

/** 食品登録フォームのバリデーション結果。 */
export interface FoodItemValidationResult {
	readonly valid: boolean;
	readonly errors: FoodItemErrors;
}

/** 融通リクエストバリデーション結果。 */
export interface FusionRequestValidationResult {
	readonly valid: boolean;
	readonly errors: FusionRequestErrors;
}

/** 食品名のバリデーション定数。 */
const FOOD_NAME_MAX_LENGTH = 200;
const QUANTITY_MIN = 1;
const QUANTITY_MAX = 10000;

function isFoodCategory(value: string): value is FoodCategory {
	return (FOOD_CATEGORIES as readonly string[]).includes(value);
}

function isFoodUnit(value: string): value is FoodUnit {
	return (FOOD_UNITS as readonly string[]).includes(value);
}

/** 食品登録フォームのバリデーション。 */
export function validateFoodItemInput(input: CreateFoodItemInput): FoodItemValidationResult {
	const errors: FoodItemErrors = {};

	if (!input.name.trim()) {
		errors.name = "食品名は必須です";
	} else if (input.name.length > FOOD_NAME_MAX_LENGTH) {
		errors.name = `食品名は${String(FOOD_NAME_MAX_LENGTH)}文字以内で入力してください`;
	}

	if (!input.category) {
		errors.category = "カテゴリを選択してください";
	} else if (!isFoodCategory(input.category)) {
		errors.category = "無効なカテゴリです";
	}

	if (!input.expiryDate) {
		errors.expiryDate = "消費期限は必須です";
	}

	if (input.quantity === "") {
		errors.quantity = "数量は必須です";
	} else if (input.quantity < QUANTITY_MIN || input.quantity > QUANTITY_MAX) {
		errors.quantity = `数量は${String(QUANTITY_MIN)}〜${String(QUANTITY_MAX)}の範囲で入力してください`;
	}

	if (!input.unit) {
		errors.unit = "単位を選択してください";
	} else if (!isFoodUnit(input.unit)) {
		errors.unit = "無効な単位です";
	}

	return {
		valid: Object.keys(errors).length === 0,
		errors,
	};
}

/** 消費期限が過去かどうかを判定。 */
export function isExpiryDatePast(dateStr: string): boolean {
	const today = new Date();
	today.setHours(0, 0, 0, 0);
	const expiry = new Date(dateStr);
	return expiry < today;
}

/** 消費期限の危険度レベル。 */
export type ExpiryLevel = "danger" | "warning" | "safe";

/** 消費期限の危険度を判定。3日以内=danger, 7日以内=warning, それ以降=safe。 */
export function getExpiryLevel(dateStr: string): ExpiryLevel {
	const today = new Date();
	today.setHours(0, 0, 0, 0);
	const expiry = new Date(dateStr);
	const diffMs = expiry.getTime() - today.getTime();
	const diffDays = Math.ceil(diffMs / (1000 * 60 * 60 * 24));

	if (diffDays <= 3) {
		return "danger";
	}
	if (diffDays <= 7) {
		return "warning";
	}
	return "safe";
}

/** 融通リクエスト作成フォームのバリデーション。 */
export function validateFusionRequestInput(input: CreateFusionRequestInput): FusionRequestValidationResult {
	const errors: FusionRequestErrors = {};

	if (!input.requesterShokudoId.trim()) {
		errors.requesterShokudoId = "リクエスト元食堂IDは必須です";
	}

	if (!input.desiredCategory) {
		errors.desiredCategory = "カテゴリを選択してください";
	} else if (!isFoodCategory(input.desiredCategory)) {
		errors.desiredCategory = "無効なカテゴリです";
	}

	if (input.desiredQuantity === "") {
		errors.desiredQuantity = "数量は必須です";
	} else if (input.desiredQuantity < QUANTITY_MIN || input.desiredQuantity > QUANTITY_MAX) {
		errors.desiredQuantity = `数量は${String(QUANTITY_MIN)}〜${String(QUANTITY_MAX)}の範囲で入力してください`;
	}

	if (!input.unit) {
		errors.unit = "単位を選択してください";
	} else if (!isFoodUnit(input.unit)) {
		errors.unit = "無効な単位です";
	}

	return {
		valid: Object.keys(errors).length === 0,
		errors,
	};
}
