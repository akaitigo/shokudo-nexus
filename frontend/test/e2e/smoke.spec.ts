// =============================================================================
// 最小 E2E テスト（Playwright）
//
// プロジェクトの「最低限動いている」ことを確認するスモークテスト。
// ここを起点に、プロジェクト固有のテストを追加していく。
// =============================================================================

import { expect, test } from "@playwright/test";

test.describe("スモークテスト", () => {
	test("トップページが表示される", async ({ page }) => {
		await page.goto("/");

		// ページが正常にロードされることを確認
		await expect(page).toHaveTitle(/.+/);

		// 主要なランドマーク要素が存在することを確認
		const main = page.locator("main");
		await expect(main).toBeVisible();
	});

	test("404ページが適切に表示される", async ({ page }) => {
		const response = await page.goto("/this-page-does-not-exist");

		// 404ステータスまたはカスタム404ページを確認
		// フレームワークによってはSPA routingで200を返す場合がある
		if (response) {
			expect([200, 404]).toContain(response.status());
		}
	});

	test("ナビゲーションが動作する", async ({ page }) => {
		await page.goto("/");

		// TODO: プロジェクト固有のナビゲーション要素に置き換える
		// 例: メニューリンクをクリックして遷移を確認
		// const nav = page.getByRole("navigation");
		// await expect(nav).toBeVisible();
		// await nav.getByRole("link", { name: "About" }).click();
		// await expect(page).toHaveURL(/.*about/);
	});
});

test.describe("アクセシビリティ", () => {
	test("画像にalt属性がある", async ({ page }) => {
		await page.goto("/");

		const images = page.locator("img");
		const count = await images.count();

		for (let i = 0; i < count; i++) {
			const alt = await images.nth(i).getAttribute("alt");
			// alt属性が存在すること（空文字はデコレーティブ画像として許容）
			expect(alt).not.toBeNull();
		}
	});

	test("ページにh1が1つ存在する", async ({ page }) => {
		await page.goto("/");

		const h1 = page.locator("h1");
		await expect(h1).toHaveCount(1);
	});
});

test.describe("レスポンシブ", () => {
	// mobile-chrome プロジェクトで自動的にモバイルビューポートでテストされる
	test("モバイルでもメインコンテンツが表示される", async ({ page }) => {
		await page.goto("/");

		const main = page.locator("main");
		await expect(main).toBeVisible();
	});
});
