import { HomePage } from "@/components/HomePage";
import { FoodItemForm } from "@/components/food/FoodItemForm";
import { FoodItemList } from "@/components/food/FoodItemList";
import { FusionRequestForm } from "@/components/fusion/FusionRequestForm";
import { FusionRequestList } from "@/components/fusion/FusionRequestList";
import { AppLayout } from "@/components/layout/AppLayout";
import { ApiClientProvider } from "@/lib/api-context";
import { mockApiClient } from "@/lib/mock-api-client";
import { BrowserRouter, Route, Routes } from "react-router-dom";

export function App(): React.ReactElement {
	return (
		<ApiClientProvider client={mockApiClient}>
			<BrowserRouter>
				<Routes>
					<Route element={<AppLayout />}>
						<Route path="/" element={<HomePage />} />
						<Route path="/food" element={<FoodItemList />} />
						<Route path="/food/new" element={<FoodItemForm />} />
						<Route path="/fusion" element={<FusionRequestList />} />
						<Route path="/fusion/new" element={<FusionRequestForm />} />
					</Route>
				</Routes>
			</BrowserRouter>
		</ApiClientProvider>
	);
}
