import type { ApiClient, ListFoodItemsResult, ListFusionRequestsResult } from "@/lib/api-client";
import type {
	CreateFoodItemInput,
	CreateFusionRequestInput,
	FoodCategory,
	FoodItem,
	FoodUnit,
	FusionRequest,
	FusionRequestStatus,
} from "@/types/domain";

/** テスト・開発用のモック API クライアント。 */

let foodItems: FoodItem[] = [];
let fusionRequests: FusionRequest[] = [];
let nextFoodId = 1;
let nextFusionId = 1;

function generateId(prefix: string, counter: number): string {
	return `${prefix}-${String(counter).padStart(4, "0")}`;
}

function nowIso(): string {
	return new Date().toISOString();
}

/** モックデータをリセットする（テスト用）。 */
export function resetMockData(): void {
	foodItems = [];
	fusionRequests = [];
	nextFoodId = 1;
	nextFusionId = 1;
}

/** 遅延をシミュレートする。 */
function delay(ms: number): Promise<void> {
	return new Promise((resolve) => {
		setTimeout(resolve, ms);
	});
}

export const mockApiClient: ApiClient = {
	async createFoodItem(input: CreateFoodItemInput): Promise<FoodItem> {
		await delay(200);
		const id = generateId("food", nextFoodId);
		nextFoodId += 1;
		const now = nowIso();
		const item: FoodItem = {
			id,
			name: input.name,
			category: input.category as FoodCategory,
			expiryDate: input.expiryDate,
			quantity: input.quantity as number,
			unit: input.unit as FoodUnit,
			donorId: input.donorId || "donor-default",
			status: "available",
			createdAt: now,
			updatedAt: now,
		};
		foodItems.push(item);
		return item;
	},

	async listFoodItems(
		pageSize: number,
		pageToken: string,
		categoryFilter: FoodCategory | "",
	): Promise<ListFoodItemsResult> {
		await delay(150);
		let filtered = [...foodItems];
		if (categoryFilter) {
			filtered = filtered.filter((item) => item.category === categoryFilter);
		}

		const startIndex = pageToken ? Number.parseInt(pageToken, 10) : 0;
		const endIndex = startIndex + pageSize;
		const pageItems = filtered.slice(startIndex, endIndex);
		const hasNext = endIndex < filtered.length;

		return {
			items: pageItems,
			pagination: {
				nextPageToken: hasNext ? String(endIndex) : "",
				totalCount: filtered.length,
			},
		};
	},

	async deleteFoodItem(id: string): Promise<void> {
		await delay(150);
		const index = foodItems.findIndex((item) => item.id === id);
		if (index === -1) {
			throw new Error("食品が見つかりません");
		}
		foodItems.splice(index, 1);
	},

	async createFusionRequest(input: CreateFusionRequestInput): Promise<FusionRequest> {
		await delay(200);
		const id = generateId("fusion", nextFusionId);
		nextFusionId += 1;
		const now = nowIso();
		const request: FusionRequest = {
			id,
			requesterShokudoId: input.requesterShokudoId,
			desiredCategory: input.desiredCategory as FoodCategory,
			desiredQuantity: input.desiredQuantity as number,
			unit: input.unit as FoodUnit,
			message: input.message,
			status: "pending",
			responderShokudoId: "",
			foodItemId: "",
			createdAt: now,
			updatedAt: now,
		};
		fusionRequests.push(request);
		return request;
	},

	async listFusionRequests(
		pageSize: number,
		pageToken: string,
		statusFilter: string,
	): Promise<ListFusionRequestsResult> {
		await delay(150);
		let filtered = [...fusionRequests];
		if (statusFilter) {
			filtered = filtered.filter((r) => r.status === statusFilter);
		}

		const startIndex = pageToken ? Number.parseInt(pageToken, 10) : 0;
		const endIndex = startIndex + pageSize;
		const pageItems = filtered.slice(startIndex, endIndex);
		const hasNext = endIndex < filtered.length;

		return {
			items: pageItems,
			pagination: {
				nextPageToken: hasNext ? String(endIndex) : "",
				totalCount: filtered.length,
			},
		};
	},

	async respondToFusionRequest(
		id: string,
		response: "APPROVED" | "REJECTED",
		foodItemId: string,
	): Promise<FusionRequest> {
		await delay(200);
		const index = fusionRequests.findIndex((r) => r.id === id);
		if (index === -1) {
			throw new Error("融通リクエストが見つかりません");
		}
		const existing = fusionRequests[index];
		if (!existing) {
			throw new Error("融通リクエストが見つかりません");
		}
		const newStatus: FusionRequestStatus = response === "APPROVED" ? "approved" : "rejected";
		const updated: FusionRequest = {
			...existing,
			status: newStatus,
			responderShokudoId: response === "APPROVED" ? "responder-default" : "",
			foodItemId: response === "APPROVED" ? foodItemId : "",
			updatedAt: nowIso(),
		};
		fusionRequests[index] = updated;
		return updated;
	},
};
