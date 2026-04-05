import { type FirebaseApp, initializeApp } from "firebase/app";
import {
	type Auth,
	signOut as firebaseSignOut,
	GoogleAuthProvider,
	getAuth,
	onAuthStateChanged,
	signInWithPopup,
	type User,
} from "firebase/auth";

let app: FirebaseApp | null = null;
let auth: Auth | null = null;

/**
 * Firebase Appの遅延初期化。
 * VITE_FIREBASE_API_KEY が未設定の場合は初期化をスキップする。
 */
function getFirebaseApp(): FirebaseApp | null {
	if (app !== null) return app;

	const apiKey = import.meta.env.VITE_FIREBASE_API_KEY;
	if (!apiKey) return null;

	app = initializeApp({
		apiKey,
		authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN ?? "",
		projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID ?? "",
		storageBucket: import.meta.env.VITE_FIREBASE_STORAGE_BUCKET ?? "",
		messagingSenderId: import.meta.env.VITE_FIREBASE_MESSAGING_SENDER_ID ?? "",
		appId: import.meta.env.VITE_FIREBASE_APP_ID ?? "",
	});
	return app;
}

/**
 * Firebase Auth インスタンスを取得する。
 * Firebase が未設定の場合は null を返す。
 */
export function getFirebaseAuth(): Auth | null {
	if (auth !== null) return auth;
	const firebaseApp = getFirebaseApp();
	if (firebaseApp === null) return null;
	auth = getAuth(firebaseApp);
	return auth;
}

/**
 * 現在のユーザーのIDトークンを取得する。
 * 未認証またはFirebase未設定の場合は null を返す。
 */
export async function getIdToken(): Promise<string | null> {
	const authInstance = getFirebaseAuth();
	if (authInstance === null) return null;
	const user = authInstance.currentUser;
	if (user === null) return null;
	return user.getIdToken();
}

/**
 * Googleアカウントでサインインする。
 */
export async function signInWithGoogle(): Promise<User> {
	const authInstance = getFirebaseAuth();
	if (authInstance === null) {
		throw new Error("Firebase is not configured. Set VITE_FIREBASE_API_KEY.");
	}
	const provider = new GoogleAuthProvider();
	const result = await signInWithPopup(authInstance, provider);
	return result.user;
}

/**
 * サインアウトする。
 */
export async function signOut(): Promise<void> {
	const authInstance = getFirebaseAuth();
	if (authInstance === null) return;
	await firebaseSignOut(authInstance);
}

/**
 * 認証状態の変化を監視する。
 * @returns unsubscribe関数
 */
export function onAuthStateChange(callback: (user: User | null) => void): () => void {
	const authInstance = getFirebaseAuth();
	if (authInstance === null) {
		// Firebase未設定時は即座にnullを通知
		callback(null);
		return () => {};
	}
	return onAuthStateChanged(authInstance, callback);
}
