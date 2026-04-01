import { createContext, useContext } from "react";
import type { ApiClient } from "@/lib/api-client";

const ApiClientContext = createContext<ApiClient | null>(null);

/** API クライアントプロバイダー。 */
export function ApiClientProvider({
	client,
	children,
}: {
	readonly client: ApiClient;
	readonly children: React.ReactNode;
}): React.ReactElement {
	return <ApiClientContext.Provider value={client}>{children}</ApiClientContext.Provider>;
}

/** API クライアントを取得するフック。ApiClientProvider 外で使用するとエラー。 */
export function useApiClient(): ApiClient {
	const client = useContext(ApiClientContext);
	if (client === null) {
		throw new Error("useApiClient must be used within an ApiClientProvider");
	}
	return client;
}
