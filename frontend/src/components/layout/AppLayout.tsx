import { useCallback, useState } from "react";
import { Link, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "@/lib/auth-context";

/** アプリケーション共通レイアウト。ヘッダー + ナビゲーション + メインコンテンツ。 */
export function AppLayout(): React.ReactElement {
	const location = useLocation();
	const { user, loading, isConfigured, signIn, handleSignOut } = useAuth();
	const [signingIn, setSigningIn] = useState(false);
	const [signInError, setSignInError] = useState<string | null>(null);

	function navClass(path: string): string {
		return location.pathname === path ? "active" : "";
	}

	const onSignIn = useCallback(async () => {
		setSigningIn(true);
		setSignInError(null);
		try {
			await signIn();
		} catch {
			setSignInError("サインインに失敗しました。もう一度お試しください。");
		} finally {
			setSigningIn(false);
		}
	}, [signIn]);

	const onSignOut = useCallback(async () => {
		try {
			await handleSignOut();
		} catch {
			// サインアウト失敗は無視
		}
	}, [handleSignOut]);

	/** 認証済みまたはFirebase未設定（開発モード）であればtrue */
	const canAccess = !isConfigured || user !== null;

	return (
		<div className="app-layout">
			<header className="app-header">
				<div className="app-header-inner">
					<Link to="/" className="app-title">
						shokudo-nexus
					</Link>
					{canAccess ? (
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
					) : null}
					<div className="app-auth">
						<AuthStatus
							user={user}
							loading={loading}
							isConfigured={isConfigured}
							signingIn={signingIn}
							onSignIn={onSignIn}
							onSignOut={onSignOut}
						/>
					</div>
				</div>
			</header>
			{signInError ? (
				<div className="alert alert-error" style={{ margin: "1rem" }}>
					{signInError}
				</div>
			) : null}
			<main className="app-main">
				{canAccess ? (
					<Outlet />
				) : loading ? (
					<div className="loading">認証状態を確認中...</div>
				) : (
					<div className="empty-state">
						<p>この機能を利用するにはサインインが必要です</p>
						<button type="button" className="btn btn-primary" onClick={onSignIn} disabled={signingIn}>
							{signingIn ? "サインイン中..." : "Googleアカウントでサインイン"}
						</button>
					</div>
				)}
			</main>
		</div>
	);
}

/** ヘッダー内の認証ステータス表示。 */
function AuthStatus({
	user,
	loading,
	isConfigured,
	signingIn,
	onSignIn,
	onSignOut,
}: {
	readonly user: { readonly displayName: string | null; readonly email: string | null } | null;
	readonly loading: boolean;
	readonly isConfigured: boolean;
	readonly signingIn: boolean;
	readonly onSignIn: () => void;
	readonly onSignOut: () => void;
}): React.ReactElement {
	if (!isConfigured) {
		return <span className="auth-dev-mode">DEV MODE</span>;
	}

	if (loading) {
		return <span className="auth-loading">...</span>;
	}

	if (user !== null) {
		const displayLabel = user.displayName ?? user.email ?? "ユーザー";
		return (
			<span className="auth-user-info">
				<span className="auth-user-name">{displayLabel}</span>
				<button type="button" className="btn btn-outline btn-sm" onClick={onSignOut}>
					サインアウト
				</button>
			</span>
		);
	}

	return (
		<button type="button" className="btn btn-primary btn-sm" onClick={onSignIn} disabled={signingIn}>
			{signingIn ? "サインイン中..." : "サインイン"}
		</button>
	);
}
