/// <reference types="vite/client" />

interface ImportMetaEnv {
	/** gRPC バックエンドの URL。未設定時はモック API を使用。 */
	readonly VITE_API_URL?: string;
}

interface ImportMeta {
	readonly env: ImportMetaEnv;
}
