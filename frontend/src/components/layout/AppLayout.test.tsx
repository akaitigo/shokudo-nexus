import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { AppLayout } from "@/components/layout/AppLayout";
import { AuthProvider } from "@/lib/auth-context";

/**
 * テスト環境ではFirebaseが未設定（VITE_FIREBASE_API_KEY なし）のため、
 * isConfigured=false → 開発モードとして動作する。
 */
function renderWithRouter(initialPath = "/"): void {
	render(
		<AuthProvider>
			<MemoryRouter initialEntries={[initialPath]}>
				<Routes>
					<Route element={<AppLayout />}>
						<Route index element={<div>Home Content</div>} />
						<Route path="food" element={<div>Food Content</div>} />
					</Route>
				</Routes>
			</MemoryRouter>
		</AuthProvider>,
	);
}

describe("AppLayout", () => {
	it("shows DEV MODE badge when Firebase is not configured", () => {
		renderWithRouter();
		expect(screen.getByText("DEV MODE")).toBeInTheDocument();
	});

	it("renders navigation links in dev mode", () => {
		renderWithRouter();
		expect(screen.getByRole("link", { name: "食品登録" })).toBeInTheDocument();
		expect(screen.getByRole("link", { name: "在庫一覧" })).toBeInTheDocument();
		expect(screen.getByRole("link", { name: "融通リクエスト" })).toBeInTheDocument();
		expect(screen.getByRole("link", { name: "融通一覧" })).toBeInTheDocument();
	});

	it("renders child route content in dev mode", () => {
		renderWithRouter();
		expect(screen.getByText("Home Content")).toBeInTheDocument();
	});

	it("does not show sign-in button in dev mode", () => {
		renderWithRouter();
		expect(screen.queryByRole("button", { name: /サインイン/ })).not.toBeInTheDocument();
	});

	it("applies active class to current nav link", () => {
		renderWithRouter("/food");
		const foodLink = screen.getByRole("link", { name: "在庫一覧" });
		expect(foodLink.className).toContain("active");
	});
});
