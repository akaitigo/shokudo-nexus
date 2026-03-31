import { createContext, useContext } from "react";
import type { ApiClient } from "@/lib/api-client";
import { mockApiClient } from "@/lib/mock-api-client";

const ApiClientContext = createContext<ApiClient>(mockApiClient);

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

/** API クライアントを取得するフック。 */
export function useApiClient(): ApiClient {
	return useContext(ApiClientContext);
}
