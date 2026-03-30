import type { CreateFoodItemInput, CreateFusionRequestInput } from "@/types/domain";
import {
	getExpiryLevel,
	isExpiryDatePast,
	validateFoodItemInput,
	validateFusionRequestInput,
} from "@/types/validation";
import { describe, expect, it } from "vitest";

describe("validateFoodItemInput", () => {
	const validInput: CreateFoodItemInput = {
		name: "にんじん",
		category: "野菜",
		expiryDate: "2026-04-15",
		quantity: 10,
		unit: "kg",
		donorId: "donor-001",
	};

	it("有効な入力で valid=true を返す", () => {
		const result = validateFoodItemInput(validInput);
		expect(result.valid).toBe(true);
		expect(Object.keys(result.errors)).toHaveLength(0);
	});

	it("食品名が空の場合にエラーを返す", () => {
		const result = validateFoodItemInput({ ...validInput, name: "" });
		expect(result.valid).toBe(false);
		expect(result.errors.name).toBe("食品名は必須です");
	});

	it("食品名がスペースのみの場合にエラーを返す", () => {
		const result = validateFoodItemInput({ ...validInput, name: "   " });
		expect(result.valid).toBe(false);
		expect(result.errors.name).toBe("食品名は必須です");
	});

	it("食品名が200文字を超える場合にエラーを返す", () => {
		const longName = "あ".repeat(201);
		const result = validateFoodItemInput({ ...validInput, name: longName });
		expect(result.valid).toBe(false);
		expect(result.errors.name).toContain("200文字以内");
	});

	it("食品名がちょうど200文字の場合は有効", () => {
		const exactName = "あ".repeat(200);
		const result = validateFoodItemInput({ ...validInput, name: exactName });
		expect(result.errors.name).toBeUndefined();
	});

	it("カテゴリが空の場合にエラーを返す", () => {
		const result = validateFoodItemInput({ ...validInput, category: "" });
		expect(result.valid).toBe(false);
		expect(result.errors.category).toBe("カテゴリを選択してください");
	});

	it("消費期限が空の場合にエラーを返す", () => {
		const result = validateFoodItemInput({ ...validInput, expiryDate: "" });
		expect(result.valid).toBe(false);
		expect(result.errors.expiryDate).toBe("消費期限は必須です");
	});

	it("数量が空の場合にエラーを返す", () => {
		const result = validateFoodItemInput({ ...validInput, quantity: "" });
		expect(result.valid).toBe(false);
		expect(result.errors.quantity).toBe("数量は必須です");
	});

	it("数量が0の場合にエラーを返す", () => {
		const result = validateFoodItemInput({ ...validInput, quantity: 0 });
		expect(result.valid).toBe(false);
		expect(result.errors.quantity).toContain("1〜10000");
	});

	it("数量が10001の場合にエラーを返す", () => {
		const result = validateFoodItemInput({ ...validInput, quantity: 10001 });
		expect(result.valid).toBe(false);
		expect(result.errors.quantity).toContain("1〜10000");
	});

	it("数量が1の場合は有効", () => {
		const result = validateFoodItemInput({ ...validInput, quantity: 1 });
		expect(result.errors.quantity).toBeUndefined();
	});

	it("数量が10000の場合は有効", () => {
		const result = validateFoodItemInput({ ...validInput, quantity: 10000 });
		expect(result.errors.quantity).toBeUndefined();
	});

	it("単位が空の場合にエラーを返す", () => {
		const result = validateFoodItemInput({ ...validInput, unit: "" });
		expect(result.valid).toBe(false);
		expect(result.errors.unit).toBe("単位を選択してください");
	});

	it("複数フィールドが空の場合、全てのエラーを返す", () => {
		const result = validateFoodItemInput({
			name: "",
			category: "",
			expiryDate: "",
			quantity: "",
			unit: "",
			donorId: "",
		});
		expect(result.valid).toBe(false);
		expect(result.errors.name).toBeDefined();
		expect(result.errors.category).toBeDefined();
		expect(result.errors.expiryDate).toBeDefined();
		expect(result.errors.quantity).toBeDefined();
		expect(result.errors.unit).toBeDefined();
	});
});

describe("validateFusionRequestInput", () => {
	const validInput: CreateFusionRequestInput = {
		requesterShokudoId: "shokudo-001",
		desiredCategory: "野菜",
		desiredQuantity: 5,
		unit: "kg",
		message: "テストメッセージ",
	};

	it("有効な入力で valid=true を返す", () => {
		const result = validateFusionRequestInput(validInput);
		expect(result.valid).toBe(true);
	});

	it("食堂IDが空の場合にエラーを返す", () => {
		const result = validateFusionRequestInput({ ...validInput, requesterShokudoId: "" });
		expect(result.valid).toBe(false);
		expect(result.errors.requesterShokudoId).toBe("リクエスト元食堂IDは必須です");
	});

	it("カテゴリが空の場合にエラーを返す", () => {
		const result = validateFusionRequestInput({ ...validInput, desiredCategory: "" });
		expect(result.valid).toBe(false);
		expect(result.errors.desiredCategory).toBe("カテゴリを選択してください");
	});

	it("数量が空の場合にエラーを返す", () => {
		const result = validateFusionRequestInput({ ...validInput, desiredQuantity: "" });
		expect(result.valid).toBe(false);
		expect(result.errors.desiredQuantity).toBe("数量は必須です");
	});

	it("単位が空の場合にエラーを返す", () => {
		const result = validateFusionRequestInput({ ...validInput, unit: "" });
		expect(result.valid).toBe(false);
		expect(result.errors.unit).toBe("単位を選択してください");
	});

	it("メッセージは空でもエラーにならない", () => {
		const result = validateFusionRequestInput({ ...validInput, message: "" });
		expect(result.valid).toBe(true);
	});
});

describe("isExpiryDatePast", () => {
	it("過去の日付で true を返す", () => {
		expect(isExpiryDatePast("2020-01-01")).toBe(true);
	});

	it("未来の日付で false を返す", () => {
		expect(isExpiryDatePast("2030-12-31")).toBe(false);
	});
});

describe("getExpiryLevel", () => {
	it("3日以内の場合 danger を返す", () => {
		const tomorrow = new Date();
		tomorrow.setDate(tomorrow.getDate() + 1);
		const dateStr = tomorrow.toISOString().split("T")[0];
		if (dateStr) {
			expect(getExpiryLevel(dateStr)).toBe("danger");
		}
	});

	it("5日後の場合 warning を返す", () => {
		const fiveDays = new Date();
		fiveDays.setDate(fiveDays.getDate() + 5);
		const dateStr = fiveDays.toISOString().split("T")[0];
		if (dateStr) {
			expect(getExpiryLevel(dateStr)).toBe("warning");
		}
	});

	it("10日後の場合 safe を返す", () => {
		const tenDays = new Date();
		tenDays.setDate(tenDays.getDate() + 10);
		const dateStr = tenDays.toISOString().split("T")[0];
		if (dateStr) {
			expect(getExpiryLevel(dateStr)).toBe("safe");
		}
	});

	it("過去の日付の場合 danger を返す", () => {
		expect(getExpiryLevel("2020-01-01")).toBe("danger");
	});
});
