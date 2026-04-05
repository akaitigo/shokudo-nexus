import type { User } from "firebase/auth";
import { createContext, useCallback, useContext, useEffect, useState } from "react";
import { getFirebaseAuth, onAuthStateChange, signInWithGoogle, signOut } from "@/lib/firebase";

/** 認証状態。 */
interface AuthState {
	/** 現在のユーザー。未認証の場合はnull。 */
	readonly user: User | null;
	/** 認証状態の読み込み中かどうか。 */
	readonly loading: boolean;
	/** Firebase Auth が設定されているかどうか。 */
	readonly isConfigured: boolean;
	/** Googleアカウントでサインインする。 */
	readonly signIn: () => Promise<void>;
	/** サインアウトする。 */
	readonly handleSignOut: () => Promise<void>;
}

const AuthContext = createContext<AuthState | null>(null);

/** 認証プロバイダー。 */
export function AuthProvider({ children }: { readonly children: React.ReactNode }): React.ReactElement {
	const [user, setUser] = useState<User | null>(null);
	const [loading, setLoading] = useState(true);
	const isConfigured = getFirebaseAuth() !== null;

	useEffect(() => {
		const unsubscribe = onAuthStateChange((authUser) => {
			setUser(authUser);
			setLoading(false);
		});
		return unsubscribe;
	}, []);

	const signIn = useCallback(async () => {
		await signInWithGoogle();
	}, []);

	const handleSignOut = useCallback(async () => {
		await signOut();
	}, []);

	return (
		<AuthContext.Provider value={{ user, loading, isConfigured, signIn, handleSignOut }}>
			{children}
		</AuthContext.Provider>
	);
}

/** 認証状態を取得するフック。AuthProvider外で使用するとエラー。 */
export function useAuth(): AuthState {
	const state = useContext(AuthContext);
	if (state === null) {
		throw new Error("useAuth must be used within an AuthProvider");
	}
	return state;
}
