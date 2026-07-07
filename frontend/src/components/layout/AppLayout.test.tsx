import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { AppLayout } from "@/components/layout/AppLayout";

// useAuth をモックし、AppLayout の認証状態ごとの描画・インタラクションを検証する。
// vi.hoisted で先に生成することで vi.mock のファクトリから安全に参照できる。
const { mockUseAuth } = vi.hoisted(() => ({ mockUseAuth: vi.fn() }));

vi.mock("@/lib/auth-context", () => ({
	useAuth: () => mockUseAuth(),
}));

type MockUser = { readonly displayName: string | null; readonly email: string | null };

interface MockAuthState {
	readonly user: MockUser | null;
	readonly loading: boolean;
	readonly isConfigured: boolean;
	readonly signIn: () => Promise<void>;
	readonly handleSignOut: () => Promise<void>;
}

function setAuth(overrides: Partial<MockAuthState> = {}): {
	readonly signIn: ReturnType<typeof vi.fn>;
	readonly handleSignOut: ReturnType<typeof vi.fn>;
} {
	const signIn = overrides.signIn ?? vi.fn().mockResolvedValue(undefined);
	const handleSignOut = overrides.handleSignOut ?? vi.fn().mockResolvedValue(undefined);
	const state: MockAuthState = {
		user: overrides.user ?? null,
		loading: overrides.loading ?? false,
		isConfigured: overrides.isConfigured ?? false,
		signIn,
		handleSignOut,
	};
	mockUseAuth.mockReturnValue(state);
	return { signIn: signIn as ReturnType<typeof vi.fn>, handleSignOut: handleSignOut as ReturnType<typeof vi.fn> };
}

function renderLayout(initialPath = "/"): void {
	render(
		<MemoryRouter initialEntries={[initialPath]}>
			<Routes>
				<Route element={<AppLayout />}>
					<Route index element={<div>Home Content</div>} />
					<Route path="food" element={<div>Food Content</div>} />
				</Route>
			</Routes>
		</MemoryRouter>,
	);
}

beforeEach(() => {
	mockUseAuth.mockReset();
});

describe("AppLayout - 開発モード (Firebase未設定)", () => {
	it("DEV MODE バッジを表示する", () => {
		setAuth({ isConfigured: false });
		renderLayout();
		expect(screen.getByText("DEV MODE")).toBeInTheDocument();
	});

	it("ナビゲーションリンクを表示する", () => {
		setAuth({ isConfigured: false });
		renderLayout();
		expect(screen.getByRole("link", { name: "食品登録" })).toBeInTheDocument();
		expect(screen.getByRole("link", { name: "在庫一覧" })).toBeInTheDocument();
		expect(screen.getByRole("link", { name: "融通リクエスト" })).toBeInTheDocument();
		expect(screen.getByRole("link", { name: "融通一覧" })).toBeInTheDocument();
	});

	it("子ルートのコンテンツを表示する", () => {
		setAuth({ isConfigured: false });
		renderLayout();
		expect(screen.getByText("Home Content")).toBeInTheDocument();
	});

	it("サインインボタンを表示しない", () => {
		setAuth({ isConfigured: false });
		renderLayout();
		expect(screen.queryByRole("button", { name: /サインイン/ })).not.toBeInTheDocument();
	});

	it("現在のナビリンクに active クラスを付与する", () => {
		setAuth({ isConfigured: false });
		renderLayout("/food");
		const foodLink = screen.getByRole("link", { name: "在庫一覧" });
		expect(foodLink.className).toContain("active");
	});
});

describe("AppLayout - 認証状態の読み込み中", () => {
	it("読み込み中メッセージを表示しナビを隠す", () => {
		setAuth({ isConfigured: true, loading: true, user: null });
		renderLayout();
		expect(screen.getByText("認証状態を確認中...")).toBeInTheDocument();
		expect(screen.queryByRole("link", { name: "在庫一覧" })).not.toBeInTheDocument();
		expect(screen.queryByText("Home Content")).not.toBeInTheDocument();
	});
});

describe("AppLayout - 未認証 (Firebase設定済み)", () => {
	it("サインインを促す空状態とボタンを表示する", () => {
		setAuth({ isConfigured: true, loading: false, user: null });
		renderLayout();
		expect(screen.getByText("この機能を利用するにはサインインが必要です")).toBeInTheDocument();
		expect(screen.getByRole("button", { name: "Googleアカウントでサインイン" })).toBeInTheDocument();
		// ナビゲーションとコンテンツは隠れる。
		expect(screen.queryByRole("link", { name: "在庫一覧" })).not.toBeInTheDocument();
		expect(screen.queryByText("Home Content")).not.toBeInTheDocument();
	});

	it("ヘッダーにサインインボタンを表示する", () => {
		setAuth({ isConfigured: true, loading: false, user: null });
		renderLayout();
		expect(screen.getByRole("button", { name: "サインイン" })).toBeInTheDocument();
	});

	it("サインインボタンをクリックすると signIn が呼ばれる", async () => {
		const user = userEvent.setup();
		const { signIn } = setAuth({ isConfigured: true, loading: false, user: null });
		renderLayout();

		await user.click(screen.getByRole("button", { name: "Googleアカウントでサインイン" }));

		await waitFor(() => {
			expect(signIn).toHaveBeenCalledTimes(1);
		});
	});

	it("サインイン失敗時にエラーメッセージを表示する", async () => {
		const user = userEvent.setup();
		setAuth({
			isConfigured: true,
			loading: false,
			user: null,
			signIn: vi.fn().mockRejectedValue(new Error("popup closed")),
		});
		renderLayout();

		await user.click(screen.getByRole("button", { name: "Googleアカウントでサインイン" }));

		await waitFor(() => {
			expect(screen.getByText("サインインに失敗しました。もう一度お試しください。")).toBeInTheDocument();
		});
	});
});

describe("AppLayout - 認証済み (Firebase設定済み)", () => {
	it("ユーザー名とサインアウトボタン、ナビ、コンテンツを表示する", () => {
		setAuth({
			isConfigured: true,
			loading: false,
			user: { displayName: "山田太郎", email: "yamada@example.com" },
		});
		renderLayout();

		expect(screen.getByText("山田太郎")).toBeInTheDocument();
		expect(screen.getByRole("button", { name: "サインアウト" })).toBeInTheDocument();
		expect(screen.getByRole("link", { name: "在庫一覧" })).toBeInTheDocument();
		expect(screen.getByText("Home Content")).toBeInTheDocument();
	});

	it("displayName が無い場合は email を表示する", () => {
		setAuth({
			isConfigured: true,
			loading: false,
			user: { displayName: null, email: "noname@example.com" },
		});
		renderLayout();
		expect(screen.getByText("noname@example.com")).toBeInTheDocument();
	});

	it("サインアウトボタンをクリックすると handleSignOut が呼ばれる", async () => {
		const user = userEvent.setup();
		const { handleSignOut } = setAuth({
			isConfigured: true,
			loading: false,
			user: { displayName: "山田太郎", email: "yamada@example.com" },
		});
		renderLayout();

		await user.click(screen.getByRole("button", { name: "サインアウト" }));

		await waitFor(() => {
			expect(handleSignOut).toHaveBeenCalledTimes(1);
		});
	});
});
