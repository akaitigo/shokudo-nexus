import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { FusionRequestList } from "@/components/fusion/FusionRequestList";
import type { ApiClient } from "@/lib/api-client";
import { ApiClientProvider } from "@/lib/api-context";
import type { FoodItem, FusionRequest } from "@/types/domain";

const mockRequests: FusionRequest[] = [
	{
		id: "fusion-0001",
		requesterShokudoId: "shokudo-001",
		desiredCategory: "野菜",
		desiredQuantity: 5,
		unit: "kg",
		message: "にんじんが欲しいです",
		status: "pending",
		responderShokudoId: "",
		foodItemId: "",
		createdAt: "2026-03-30T00:00:00Z",
		updatedAt: "2026-03-30T00:00:00Z",
	},
	{
		id: "fusion-0002",
		requesterShokudoId: "shokudo-002",
		desiredCategory: "肉",
		desiredQuantity: 3,
		unit: "パック",
		message: "",
		status: "approved",
		responderShokudoId: "shokudo-003",
		foodItemId: "food-0001",
		createdAt: "2026-03-29T00:00:00Z",
		updatedAt: "2026-03-30T00:00:00Z",
	},
];

// にんじん（野菜/kg/10）は fusion-0001（野菜/kg/5）に一致する。
// レタス（数量3<5）と鶏肉（カテゴリ違い）は一致しない。
const mockFoodItems: FoodItem[] = [
	{
		id: "food-veg-1",
		name: "にんじん",
		category: "野菜",
		expiryDate: "2026-05-01",
		quantity: 10,
		unit: "kg",
		donorId: "donor-1",
		status: "available",
		createdAt: "2026-03-30T00:00:00Z",
		updatedAt: "2026-03-30T00:00:00Z",
	},
	{
		id: "food-veg-2",
		name: "レタス",
		category: "野菜",
		expiryDate: "2026-05-01",
		quantity: 3,
		unit: "kg",
		donorId: "donor-1",
		status: "available",
		createdAt: "2026-03-30T00:00:00Z",
		updatedAt: "2026-03-30T00:00:00Z",
	},
	{
		id: "food-meat-1",
		name: "鶏肉",
		category: "肉",
		expiryDate: "2026-05-01",
		quantity: 10,
		unit: "パック",
		donorId: "donor-2",
		status: "available",
		createdAt: "2026-03-30T00:00:00Z",
		updatedAt: "2026-03-30T00:00:00Z",
	},
];

function createMockClient(overrides: Partial<ApiClient> = {}, requests: FusionRequest[] = mockRequests): ApiClient {
	return {
		createFoodItem: vi.fn(),
		listFoodItems: vi.fn().mockResolvedValue({
			items: mockFoodItems,
			pagination: { nextPageToken: "", totalCount: mockFoodItems.length },
		}),
		deleteFoodItem: vi.fn(),
		createFusionRequest: vi.fn(),
		listFusionRequests: vi.fn().mockResolvedValue({
			items: requests,
			pagination: { nextPageToken: "", totalCount: requests.length },
		}),
		respondToFusionRequest: vi.fn().mockResolvedValue({
			...mockRequests[0],
			status: "approved",
		}),
		...overrides,
	};
}

function renderList(client?: ApiClient) {
	const apiClient = client ?? createMockClient();
	return render(
		<ApiClientProvider client={apiClient}>
			<MemoryRouter>
				<FusionRequestList />
			</MemoryRouter>
		</ApiClientProvider>,
	);
}

describe("FusionRequestList", () => {
	it("ページ見出しが表示される", () => {
		renderList();
		expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent("融通リクエスト一覧");
	});

	it("融通リクエストが一覧に表示される", async () => {
		renderList();
		await waitFor(() => {
			expect(screen.getByText(/shokudo-001/)).toBeInTheDocument();
			expect(screen.getByText(/shokudo-002/)).toBeInTheDocument();
		});
	});

	it("ステータスフィルタのセレクトが表示される", () => {
		renderList();
		expect(screen.getByLabelText("ステータスフィルタ")).toBeInTheDocument();
	});

	it("ステータスフィルタを変更するとリストが更新される", async () => {
		const user = userEvent.setup();
		const client = createMockClient();
		renderList(client);

		await waitFor(() => {
			expect(client.listFusionRequests).toHaveBeenCalledTimes(1);
		});

		await user.selectOptions(screen.getByLabelText("ステータスフィルタ"), "pending");

		await waitFor(() => {
			expect(client.listFusionRequests).toHaveBeenCalledWith(20, "", "pending");
		});
	});

	it("pending ステータスのリクエストに承認/拒否ボタンが表示される", async () => {
		renderList();
		await waitFor(() => {
			expect(screen.getByRole("button", { name: "承認" })).toBeInTheDocument();
			expect(screen.getByRole("button", { name: "拒否" })).toBeInTheDocument();
		});
	});

	it("approved ステータスのリクエストには承認/拒否ボタンが表示されない", async () => {
		const secondRequest = mockRequests[1];
		if (!secondRequest) {
			throw new Error("mockRequests[1] is undefined");
		}
		renderList(createMockClient({}, [{ ...secondRequest }]));

		await waitFor(() => {
			expect(screen.getByText(/shokudo-002/)).toBeInTheDocument();
		});
		expect(screen.queryByRole("button", { name: "承認" })).not.toBeInTheDocument();
	});

	it("リクエストがない場合にempty stateが表示される", async () => {
		renderList(createMockClient({}, []));

		await waitFor(() => {
			expect(screen.getByText("融通リクエストはありません")).toBeInTheDocument();
		});
	});

	it("ステータスバッジが表示される", async () => {
		renderList();
		await waitFor(() => {
			expect(screen.getByText("承認待ち")).toBeInTheDocument();
			expect(screen.getByText("承認済み")).toBeInTheDocument();
		});
	});

	it("新規リクエストリンクが表示される", () => {
		renderList();
		expect(screen.getByRole("link", { name: "新規リクエスト" })).toBeInTheDocument();
	});

	it("承認ボタンを押すと一致する食品アイテムの選択UIが表示される（window.promptは使わない）", async () => {
		const user = userEvent.setup();
		const promptSpy = vi.spyOn(window, "prompt");
		const client = createMockClient();
		renderList(client);

		await waitFor(() => {
			expect(screen.getByRole("button", { name: "承認" })).toBeInTheDocument();
		});
		await user.click(screen.getByRole("button", { name: "承認" }));

		// カテゴリでフィルタした食品一覧を取得する。
		await waitFor(() => {
			expect(client.listFoodItems).toHaveBeenCalledWith(20, "", "野菜");
		});

		const select = await screen.findByLabelText("提供する食品アイテム");
		expect(select).toBeInTheDocument();
		// 一致するアイテム（にんじん）のみが選択肢に出る。
		expect(screen.getByRole("option", { name: /にんじん/ })).toBeInTheDocument();
		expect(screen.queryByRole("option", { name: /レタス/ })).not.toBeInTheDocument();
		expect(screen.queryByRole("option", { name: /鶏肉/ })).not.toBeInTheDocument();

		// window.prompt は一切呼ばれない。
		expect(promptSpy).not.toHaveBeenCalled();
		promptSpy.mockRestore();
	});

	it("食品を選択して確定すると APPROVED で応答する", async () => {
		const user = userEvent.setup();
		const client = createMockClient();
		renderList(client);

		await waitFor(() => {
			expect(screen.getByRole("button", { name: "承認" })).toBeInTheDocument();
		});
		await user.click(screen.getByRole("button", { name: "承認" }));

		const select = await screen.findByLabelText("提供する食品アイテム");
		await user.selectOptions(select, "food-veg-1");
		await user.click(screen.getByRole("button", { name: "確定" }));

		await waitFor(() => {
			expect(client.respondToFusionRequest).toHaveBeenCalledWith("fusion-0001", "APPROVED", "food-veg-1");
		});
	});

	it("一致する食品がない場合はメッセージを表示し確定できない", async () => {
		const user = userEvent.setup();
		const client = createMockClient({
			listFoodItems: vi.fn().mockResolvedValue({
				items: [],
				pagination: { nextPageToken: "", totalCount: 0 },
			}),
		});
		renderList(client);

		await waitFor(() => {
			expect(screen.getByRole("button", { name: "承認" })).toBeInTheDocument();
		});
		await user.click(screen.getByRole("button", { name: "承認" }));

		await waitFor(() => {
			expect(screen.getByText("承認可能な食品アイテムがありません")).toBeInTheDocument();
		});
		expect(screen.getByRole("button", { name: "確定" })).toBeDisabled();
	});

	it("拒否ボタンを押して拒否すると REJECTED で応答する", async () => {
		const user = userEvent.setup();
		const client = createMockClient();
		renderList(client);

		await waitFor(() => {
			expect(screen.getByRole("button", { name: "拒否" })).toBeInTheDocument();
		});
		await user.click(screen.getByRole("button", { name: "拒否" }));

		expect(screen.getByText("このリクエストを拒否しますか？")).toBeInTheDocument();
		await user.click(screen.getByRole("button", { name: "拒否する" }));

		await waitFor(() => {
			expect(client.respondToFusionRequest).toHaveBeenCalledWith("fusion-0001", "REJECTED", "");
		});
	});

	it("キャンセルすると操作パネルが閉じる", async () => {
		const user = userEvent.setup();
		const client = createMockClient();
		renderList(client);

		await waitFor(() => {
			expect(screen.getByRole("button", { name: "拒否" })).toBeInTheDocument();
		});
		await user.click(screen.getByRole("button", { name: "拒否" }));
		expect(screen.getByText("このリクエストを拒否しますか？")).toBeInTheDocument();

		await user.click(screen.getByRole("button", { name: "キャンセル" }));

		expect(screen.queryByText("このリクエストを拒否しますか？")).not.toBeInTheDocument();
		expect(screen.getByRole("button", { name: "拒否" })).toBeInTheDocument();
		expect(client.respondToFusionRequest).not.toHaveBeenCalled();
	});
});
