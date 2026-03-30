import { Link, Outlet, useLocation } from "react-router-dom";

/** アプリケーション共通レイアウト。ヘッダー + ナビゲーション + メインコンテンツ。 */
export function AppLayout(): React.ReactElement {
	const location = useLocation();

	function navClass(path: string): string {
		return location.pathname === path ? "active" : "";
	}

	return (
		<div className="app-layout">
			<header className="app-header">
				<div className="app-header-inner">
					<Link to="/" className="app-title">
						shokudo-nexus
					</Link>
					<nav className="app-nav">
						<Link to="/food/new" className={navClass("/food/new")}>
							食品登録
						</Link>
						<Link to="/food" className={navClass("/food")}>
							在庫一覧
						</Link>
						<Link to="/fusion/new" className={navClass("/fusion/new")}>
							融通リクエスト
						</Link>
						<Link to="/fusion" className={navClass("/fusion")}>
							融通一覧
						</Link>
					</nav>
				</div>
			</header>
			<main className="app-main">
				<Outlet />
			</main>
		</div>
	);
}
