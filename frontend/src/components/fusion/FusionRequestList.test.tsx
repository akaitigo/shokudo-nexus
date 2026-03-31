import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { FusionRequestList } from "@/components/fusion/FusionRequestList";
import type { ApiClient } from "@/lib/api-client";
import { ApiClientProvider } from "@/lib/api-context";
import type { FusionRequest } from "@/types/domain";

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

function createMockClient(items: FusionRequest[] = mockRequests): ApiClient {
	return {
		createFoodItem: vi.fn(),
		listFoodItems: vi.fn(),
		deleteFoodItem: vi.fn(),
		createFusionRequest: vi.fn(),
		listFusionRequests: vi.fn().mockResolvedValue({
			items,
			pagination: { nextPageToken: "", totalCount: items.length },
		}),
		respondToFusionRequest: vi.fn().mockResolvedValue({
			...mockRequests[0],
			status: "approved",
		}),
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
		const approvedOnly: FusionRequest[] = [{ ...secondRequest }];
		renderList(createMockClient(approvedOnly));

		await waitFor(() => {
			expect(screen.getByText(/shokudo-002/)).toBeInTheDocument();
		});
		expect(screen.queryByRole("button", { name: "承認" })).not.toBeInTheDocument();
	});

	it("リクエストがない場合にempty stateが表示される", async () => {
		renderList(createMockClient([]));

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
});
