// =============================================================================
// Playwright 設定テンプレート
//
// 使い方:
//   1. baseURL をプロジェクトに合わせて変更
//   2. webServer.command をプロジェクトの dev server コマンドに変更
//   3. 必要に応じてブラウザ・ビューポート設定を調整
// =============================================================================

import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  // テストディレクトリ
  testDir: "./test/e2e",

  // テスト実行の並列数
  fullyParallel: true,

  // CI では retry しない、ローカルでは1回リトライ
  retries: process.env.CI ? 0 : 1,

  // CI では並列ワーカー数を制限
  workers: process.env.CI ? 1 : undefined,

  // レポーター
  reporter: process.env.CI ? "github" : "html",

  // 共通設定
  use: {
    // dev server のURL
    baseURL: "http://localhost:3000",

    // テスト失敗時にスクリーンショットを取得
    screenshot: "only-on-failure",

    // テスト失敗時にトレースを取得
    trace: "on-first-retry",
  },

  // テスト対象ブラウザ
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
    {
      name: "firefox",
      use: { ...devices["Desktop Firefox"] },
    },
    // モバイルビューポート
    {
      name: "mobile-chrome",
      use: { ...devices["Pixel 5"] },
    },
  ],

  // dev server の自動起動
  webServer: {
    command: "npm run dev",
    url: "http://localhost:3000",
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});
