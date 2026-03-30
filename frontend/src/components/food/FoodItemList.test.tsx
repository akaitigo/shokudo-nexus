import { FoodItemList } from "@/components/food/FoodItemList";
import type { ApiClient } from "@/lib/api-client";
import { ApiClientProvider } from "@/lib/api-context";
import type { FoodItem } from "@/types/domain";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";

const mockItems: FoodItem[] = [
	{
		id: "food-0001",
		name: "にんじん",
		category: "野菜",
		expiryDate: "2026-04-01",
		quantity: 10,
		unit: "kg",
		donorId: "donor-001",
		status: "available",
		createdAt: "2026-03-30T00:00:00Z",
		updatedAt: "2026-03-30T00:00:00Z",
	},
	{
		id: "food-0002",
		name: "鶏肉",
		category: "肉",
		expiryDate: "2026-04-10",
		quantity: 5,
		unit: "パック",
		donorId: "donor-002",
		status: "available",
		createdAt: "2026-03-30T00:00:00Z",
		updatedAt: "2026-03-30T00:00:00Z",
	},
];

function createMockClient(items: FoodItem[] = mockItems): ApiClient {
	return {
		createFoodItem: vi.fn(),
		listFoodItems: vi.fn().mockResolvedValue({
			items,
			pagination: { nextPageToken: "", totalCount: items.length },
		}),
		deleteFoodItem: vi.fn().mockResolvedValue(undefined),
		createFusionRequest: vi.fn(),
		listFusionRequests: vi.fn(),
		respondToFusionRequest: vi.fn(),
	};
}

function renderList(client?: ApiClient) {
	const apiClient = client ?? createMockClient();
	return render(
		<ApiClientProvider client={apiClient}>
			<MemoryRouter>
				<FoodItemList />
			</MemoryRouter>
		</ApiClientProvider>,
	);
}

describe("FoodItemList", () => {
	it("ページ見出しが表示される", () => {
		renderList();
		expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent("在庫一覧");
	});

	it("食品アイテムが一覧に表示される", async () => {
		renderList();
		await waitFor(() => {
			expect(screen.getByText("にんじん")).toBeInTheDocument();
			expect(screen.getByText("鶏肉")).toBeInTheDocument();
		});
	});

	it("カテゴリフィルタのセレクトが表示される", () => {
		renderList();
		expect(screen.getByLabelText("カテゴリフィルタ")).toBeInTheDocument();
	});

	it("カテゴリフィルタを変更するとリストが更新される", async () => {
		const user = userEvent.setup();
		const client = createMockClient();
		renderList(client);

		await waitFor(() => {
			expect(client.listFoodItems).toHaveBeenCalledTimes(1);
		});

		await user.selectOptions(screen.getByLabelText("カテゴリフィルタ"), "野菜");

		await waitFor(() => {
			expect(client.listFoodItems).toHaveBeenCalledWith(20, "", "野菜");
		});
	});

	it("食品がない場合にempty stateが表示される", async () => {
		const client = createMockClient([]);
		renderList(client);

		await waitFor(() => {
			expect(screen.getByText("登録されている食品はありません")).toBeInTheDocument();
		});
	});

	it("食品登録リンクが表示される", () => {
		renderList();
		expect(screen.getByRole("link", { name: "食品を登録" })).toBeInTheDocument();
	});

	it("各食品に削除ボタンが表示される", async () => {
		renderList();
		await waitFor(() => {
			const deleteButtons = screen.getAllByRole("button", { name: "削除" });
			expect(deleteButtons).toHaveLength(2);
		});
	});

	it("消費期限の色分けバッジが表示される", async () => {
		renderList();
		await waitFor(() => {
			// いずれかのバッジが表示されている
			const badges = document.querySelectorAll(".badge");
			expect(badges.length).toBeGreaterThan(0);
		});
	});
});
