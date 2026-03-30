import { Link } from "react-router-dom";

/** ホーム画面。主要な機能へのナビゲーションを提供する。 */
export function HomePage(): React.ReactElement {
	return (
		<div>
			<div className="page-header">
				<h1 className="page-title">shokudo-nexus</h1>
				<p className="page-description">余剰食品と子ども食堂のマッチングプラットフォーム</p>
			</div>

			<div className="form-row" style={{ gap: "1rem" }}>
				<Link to="/food" className="card" style={{ textDecoration: "none", color: "inherit" }}>
					<h2 style={{ fontSize: "1.125rem", marginBottom: "0.5rem" }}>在庫管理</h2>
					<p style={{ color: "var(--color-text-muted)", fontSize: "0.875rem" }}>余剰食品の登録・確認・管理を行います</p>
				</Link>

				<Link to="/fusion" className="card" style={{ textDecoration: "none", color: "inherit" }}>
					<h2 style={{ fontSize: "1.125rem", marginBottom: "0.5rem" }}>食材融通</h2>
					<p style={{ color: "var(--color-text-muted)", fontSize: "0.875rem" }}>
						拠点間の食材融通リクエストを管理します
					</p>
				</Link>
			</div>
		</div>
	);
}
