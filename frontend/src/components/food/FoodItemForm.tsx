import { useFoodItems } from "@/hooks/useFoodItems";
import { FOOD_CATEGORIES, FOOD_UNITS } from "@/types/domain";
import type { CreateFoodItemInput } from "@/types/domain";
import { isExpiryDatePast, validateFoodItemInput } from "@/types/validation";
import type { FoodItemErrors } from "@/types/validation";
import { useCallback, useState } from "react";
import { useNavigate } from "react-router-dom";

/** 食品登録フォーム。 */
export function FoodItemForm(): React.ReactElement {
	const navigate = useNavigate();
	const { createItem, error: apiError } = useFoodItems();
	const [submitting, setSubmitting] = useState(false);
	const [errors, setErrors] = useState<FoodItemErrors>({});
	const [expiryWarning, setExpiryWarning] = useState(false);

	const [formData, setFormData] = useState<CreateFoodItemInput>({
		name: "",
		category: "",
		expiryDate: "",
		quantity: "",
		unit: "",
		donorId: "",
	});

	const handleChange = useCallback(
		(e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
			const { name, value } = e.target;
			setFormData((prev) => {
				if (name === "quantity") {
					const numVal = value === "" ? ("" as const) : Number.parseInt(value, 10);
					return { ...prev, [name]: numVal };
				}
				return { ...prev, [name]: value };
			});

			if (name === "expiryDate" && value) {
				setExpiryWarning(isExpiryDatePast(value));
			}

			// フィールドのエラーをクリア
			const fieldName = name as keyof FoodItemErrors;
			if (errors[fieldName]) {
				setErrors((prev) => ({ ...prev, [fieldName]: undefined }));
			}
		},
		[errors],
	);

	const handleSubmit = useCallback(
		async (e: React.FormEvent) => {
			e.preventDefault();
			const result = validateFoodItemInput(formData);
			if (!result.valid) {
				setErrors(result.errors);
				return;
			}

			setSubmitting(true);
			try {
				await createItem(formData);
				navigate("/food");
			} catch {
				// エラーはフックで処理済み
			} finally {
				setSubmitting(false);
			}
		},
		[formData, createItem, navigate],
	);

	return (
		<div>
			<div className="page-header">
				<h1 className="page-title">食品登録</h1>
				<p className="page-description">余剰食品の情報を登録します</p>
			</div>

			{apiError ? <div className="alert alert-error">{apiError}</div> : null}

			<form className="card" onSubmit={handleSubmit} noValidate>
				<div className="form-group">
					<label htmlFor="name" className="form-label">
						食品名<span className="required">*</span>
					</label>
					<input
						id="name"
						name="name"
						type="text"
						className={`form-input ${errors.name ? "error" : ""}`}
						value={formData.name}
						onChange={handleChange}
						maxLength={200}
						placeholder="例: にんじん"
						required
					/>
					{errors.name ? <div className="form-error">{errors.name}</div> : null}
				</div>

				<div className="form-row">
					<div className="form-group">
						<label htmlFor="category" className="form-label">
							カテゴリ<span className="required">*</span>
						</label>
						<select
							id="category"
							name="category"
							className={`form-select ${errors.category ? "error" : ""}`}
							value={formData.category}
							onChange={handleChange}
							required
						>
							<option value="">選択してください</option>
							{FOOD_CATEGORIES.map((cat) => (
								<option key={cat} value={cat}>
									{cat}
								</option>
							))}
						</select>
						{errors.category ? <div className="form-error">{errors.category}</div> : null}
					</div>

					<div className="form-group">
						<label htmlFor="expiryDate" className="form-label">
							消費期限<span className="required">*</span>
						</label>
						<input
							id="expiryDate"
							name="expiryDate"
							type="date"
							className={`form-input ${errors.expiryDate ? "error" : ""}`}
							value={formData.expiryDate}
							onChange={handleChange}
							required
						/>
						{errors.expiryDate ? <div className="form-error">{errors.expiryDate}</div> : null}
						{expiryWarning ? <div className="form-warning">消費期限が過去の日付です</div> : null}
					</div>
				</div>

				<div className="form-row">
					<div className="form-group">
						<label htmlFor="quantity" className="form-label">
							数量<span className="required">*</span>
						</label>
						<input
							id="quantity"
							name="quantity"
							type="number"
							className={`form-input ${errors.quantity ? "error" : ""}`}
							value={formData.quantity}
							onChange={handleChange}
							min={1}
							max={10000}
							placeholder="1-10000"
							required
						/>
						{errors.quantity ? <div className="form-error">{errors.quantity}</div> : null}
					</div>

					<div className="form-group">
						<label htmlFor="unit" className="form-label">
							単位<span className="required">*</span>
						</label>
						<select
							id="unit"
							name="unit"
							className={`form-select ${errors.unit ? "error" : ""}`}
							value={formData.unit}
							onChange={handleChange}
							required
						>
							<option value="">選択してください</option>
							{FOOD_UNITS.map((u) => (
								<option key={u} value={u}>
									{u}
								</option>
							))}
						</select>
						{errors.unit ? <div className="form-error">{errors.unit}</div> : null}
					</div>
				</div>

				<div className="form-actions">
					<button type="submit" className="btn btn-primary" disabled={submitting}>
						{submitting ? "登録中..." : "登録する"}
					</button>
					<button type="button" className="btn btn-outline" onClick={() => navigate("/food")} disabled={submitting}>
						キャンセル
					</button>
				</div>
			</form>
		</div>
	);
}
