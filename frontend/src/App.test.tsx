import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { App } from "./App";

describe("App", () => {
	it("renders the heading", () => {
		render(<App />);
		expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent("shokudo-nexus");
	});

	it("renders the description", () => {
		render(<App />);
		expect(screen.getByText("余剰食品と子ども食堂のマッチングプラットフォーム")).toBeDefined();
	});

	it("renders navigation links", () => {
		render(<App />);
		expect(screen.getByRole("link", { name: "食品登録" })).toBeInTheDocument();
		expect(screen.getByRole("link", { name: "在庫一覧" })).toBeInTheDocument();
		expect(screen.getByRole("link", { name: "融通リクエスト" })).toBeInTheDocument();
		expect(screen.getByRole("link", { name: "融通一覧" })).toBeInTheDocument();
	});
});
