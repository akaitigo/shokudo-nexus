import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { FusionRequestForm } from "@/components/fusion/FusionRequestForm";
import type { ApiClient } from "@/lib/api-client";
import { ApiClientProvider } from "@/lib/api-context";
import type { FusionRequest } from "@/types/domain";

const mockFusionRequest: FusionRequest = {
	id: "fusion-0001",
	requesterShokudoId: "shokudo-001",
	desiredCategory: "野菜",
	desiredQuantity: 5,
	unit: "kg",
	message: "テスト",
	status: "pending",
	responderShokudoId: "",
	foodItemId: "",
	createdAt: "2026-03-30T00:00:00Z",
	updatedAt: "2026-03-30T00:00:00Z",
};

function createMockClient(overrides: Partial<ApiClient> = {}): ApiClient {
	return {
		createFoodItem: vi.fn(),
		listFoodItems: vi.fn(),
		deleteFoodItem: vi.fn(),
		createFusionRequest: vi.fn().mockResolvedValue(mockFusionRequest),
		listFusionRequests: vi.fn().mockResolvedValue({ items: [], pagination: { nextPageToken: "", totalCount: 0 } }),
		respondToFusionRequest: vi.fn(),
		...overrides,
	};
}

function renderForm(client?: ApiClient) {
	const apiClient = client ?? createMockClient();
	return render(
		<ApiClientProvider client={apiClient}>
			<MemoryRouter>
				<FusionRequestForm />
			</MemoryRouter>
		</ApiClientProvider>,
	);
}

describe("FusionRequestForm", () => {
	it("フォームの見出しが表示される", () => {
		renderForm();
		expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent("融通リクエスト作成");
	});

	it("必須フィールドのラベルが表示される", () => {
		renderForm();
		expect(screen.getByLabelText(/食堂ID/)).toBeInTheDocument();
		expect(screen.getByLabelText(/希望カテゴリ/)).toBeInTheDocument();
		expect(screen.getByLabelText(/希望数量/)).toBeInTheDocument();
		expect(screen.getByLabelText(/単位/)).toBeInTheDocument();
		expect(screen.getByLabelText(/メッセージ/)).toBeInTheDocument();
	});

	it("空のフォームを送信するとバリデーションエラーが表示される", async () => {
		const user = userEvent.setup();
		renderForm();

		await user.click(screen.getByRole("button", { name: "リクエストを作成" }));

		expect(screen.getByText("リクエスト元食堂IDは必須です")).toBeInTheDocument();
		expect(screen.getByText("カテゴリを選択してください")).toBeInTheDocument();
		expect(screen.getByText("数量は必須です")).toBeInTheDocument();
		expect(screen.getByText("単位を選択してください")).toBeInTheDocument();
	});

	it("有効な入力でフォームを送信すると API を呼び出す", async () => {
		const user = userEvent.setup();
		const client = createMockClient();
		renderForm(client);

		await user.type(screen.getByLabelText(/食堂ID/), "shokudo-001");
		await user.selectOptions(screen.getByLabelText(/希望カテゴリ/), "野菜");
		await user.type(screen.getByLabelText(/希望数量/), "5");
		await user.selectOptions(screen.getByLabelText(/単位/), "kg");

		await user.click(screen.getByRole("button", { name: "リクエストを作成" }));

		await waitFor(() => {
			expect(client.createFusionRequest).toHaveBeenCalledTimes(1);
		});
	});

	it("キャンセルボタンが表示される", () => {
		renderForm();
		expect(screen.getByRole("button", { name: "キャンセル" })).toBeInTheDocument();
	});
});
