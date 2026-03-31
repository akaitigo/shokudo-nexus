import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { FoodItemForm } from "@/components/food/FoodItemForm";
import type { ApiClient } from "@/lib/api-client";
import { ApiClientProvider } from "@/lib/api-context";
import type { FoodItem } from "@/types/domain";

const mockFoodItem: FoodItem = {
	id: "food-0001",
	name: "にんじん",
	category: "野菜",
	expiryDate: "2026-04-15",
	quantity: 10,
	unit: "kg",
	donorId: "donor-default",
	status: "available",
	createdAt: "2026-03-30T00:00:00Z",
	updatedAt: "2026-03-30T00:00:00Z",
};

function createMockClient(overrides: Partial<ApiClient> = {}): ApiClient {
	return {
		createFoodItem: vi.fn().mockResolvedValue(mockFoodItem),
		listFoodItems: vi.fn().mockResolvedValue({ items: [], pagination: { nextPageToken: "", totalCount: 0 } }),
		deleteFoodItem: vi.fn().mockResolvedValue(undefined),
		createFusionRequest: vi.fn(),
		listFusionRequests: vi.fn(),
		respondToFusionRequest: vi.fn(),
		...overrides,
	};
}

function renderForm(client?: ApiClient) {
	const apiClient = client ?? createMockClient();
	return render(
		<ApiClientProvider client={apiClient}>
			<MemoryRouter>
				<FoodItemForm />
			</MemoryRouter>
		</ApiClientProvider>,
	);
}

describe("FoodItemForm", () => {
	it("フォームの見出しが表示される", () => {
		renderForm();
		expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent("食品登録");
	});

	it("必須フィールドのラベルが表示される", () => {
		renderForm();
		expect(screen.getByLabelText(/食品名/)).toBeInTheDocument();
		expect(screen.getByLabelText(/カテゴリ/)).toBeInTheDocument();
		expect(screen.getByLabelText(/消費期限/)).toBeInTheDocument();
		expect(screen.getByLabelText(/数量/)).toBeInTheDocument();
		expect(screen.getByLabelText(/単位/)).toBeInTheDocument();
	});

	it("空のフォームを送信するとバリデーションエラーが表示される", async () => {
		const user = userEvent.setup();
		renderForm();

		await user.click(screen.getByRole("button", { name: "登録する" }));

		expect(screen.getByText("食品名は必須です")).toBeInTheDocument();
		expect(screen.getByText("カテゴリを選択してください")).toBeInTheDocument();
		expect(screen.getByText("消費期限は必須です")).toBeInTheDocument();
		expect(screen.getByText("数量は必須です")).toBeInTheDocument();
		expect(screen.getByText("単位を選択してください")).toBeInTheDocument();
	});

	it("有効な入力でフォームを送信すると API を呼び出す", async () => {
		const user = userEvent.setup();
		const client = createMockClient();
		renderForm(client);

		await user.type(screen.getByLabelText(/食品名/), "にんじん");
		await user.selectOptions(screen.getByLabelText(/カテゴリ/), "野菜");
		await user.type(screen.getByLabelText(/消費期限/), "2026-04-15");
		await user.type(screen.getByLabelText(/数量/), "10");
		await user.selectOptions(screen.getByLabelText(/単位/), "kg");

		await user.click(screen.getByRole("button", { name: "登録する" }));

		await waitFor(() => {
			expect(client.createFoodItem).toHaveBeenCalledTimes(1);
		});
	});

	it("登録ボタンとキャンセルボタンが表示される", () => {
		renderForm();
		expect(screen.getByRole("button", { name: "登録する" })).toBeInTheDocument();
		expect(screen.getByRole("button", { name: "キャンセル" })).toBeInTheDocument();
	});

	it("カテゴリのドロップダウンに全カテゴリが含まれる", () => {
		renderForm();
		const select = screen.getByLabelText(/カテゴリ/);
		expect(select).toBeInTheDocument();
		const options = select.querySelectorAll("option");
		// 「選択してください」+ 6カテゴリ = 7
		expect(options).toHaveLength(7);
	});

	it("単位のドロップダウンに全単位が含まれる", () => {
		renderForm();
		const select = screen.getByLabelText(/単位/);
		const options = select.querySelectorAll("option");
		// 「選択してください」+ 6単位 = 7
		expect(options).toHaveLength(7);
	});
});
