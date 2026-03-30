import { useFusionRequests } from "@/hooks/useFusionRequests";
import { FOOD_CATEGORIES, FOOD_UNITS } from "@/types/domain";
import type { CreateFusionRequestInput } from "@/types/domain";
import { validateFusionRequestInput } from "@/types/validation";
import type { FusionRequestErrors } from "@/types/validation";
import { useCallback, useState } from "react";
import { useNavigate } from "react-router-dom";

/** 融通リクエスト作成フォーム。 */
export function FusionRequestForm(): React.ReactElement {
	const navigate = useNavigate();
	const { createRequest, error: apiError } = useFusionRequests();
	const [submitting, setSubmitting] = useState(false);
	const [errors, setErrors] = useState<FusionRequestErrors>({});

	const [formData, setFormData] = useState<CreateFusionRequestInput>({
		requesterShokudoId: "",
		desiredCategory: "",
		desiredQuantity: "",
		unit: "",
		message: "",
	});

	const handleChange = useCallback(
		(e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
			const { name, value } = e.target;
			setFormData((prev) => {
				if (name === "desiredQuantity") {
					const numVal = value === "" ? ("" as const) : Number.parseInt(value, 10);
					return { ...prev, [name]: numVal };
				}
				return { ...prev, [name]: value };
			});

			const fieldName = name as keyof FusionRequestErrors;
			if (errors[fieldName]) {
				setErrors((prev) => ({ ...prev, [fieldName]: undefined }));
			}
		},
		[errors],
	);

	const handleSubmit = useCallback(
		async (e: React.FormEvent) => {
			e.preventDefault();
			const result = validateFusionRequestInput(formData);
			if (!result.valid) {
				setErrors(result.errors);
				return;
			}

			setSubmitting(true);
			try {
				await createRequest(formData);
				navigate("/fusion");
			} catch {
				// エラーはフックで処理済み
			} finally {
				setSubmitting(false);
			}
		},
		[formData, createRequest, navigate],
	);

	return (
		<div>
			<div className="page-header">
				<h1 className="page-title">融通リクエスト作成</h1>
				<p className="page-description">食材の融通をリクエストします</p>
			</div>

			{apiError ? <div className="alert alert-error">{apiError}</div> : null}

			<form className="card" onSubmit={handleSubmit} noValidate>
				<div className="form-group">
					<label htmlFor="requesterShokudoId" className="form-label">
						食堂ID<span className="required">*</span>
					</label>
					<input
						id="requesterShokudoId"
						name="requesterShokudoId"
						type="text"
						className={`form-input ${errors.requesterShokudoId ? "error" : ""}`}
						value={formData.requesterShokudoId}
						onChange={handleChange}
						placeholder="例: shokudo-001"
						required
					/>
					{errors.requesterShokudoId ? <div className="form-error">{errors.requesterShokudoId}</div> : null}
				</div>

				<div className="form-row">
					<div className="form-group">
						<label htmlFor="desiredCategory" className="form-label">
							希望カテゴリ<span className="required">*</span>
						</label>
						<select
							id="desiredCategory"
							name="desiredCategory"
							className={`form-select ${errors.desiredCategory ? "error" : ""}`}
							value={formData.desiredCategory}
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
						{errors.desiredCategory ? <div className="form-error">{errors.desiredCategory}</div> : null}
					</div>

					<div className="form-group">
						<label htmlFor="desiredQuantity" className="form-label">
							希望数量<span className="required">*</span>
						</label>
						<input
							id="desiredQuantity"
							name="desiredQuantity"
							type="number"
							className={`form-input ${errors.desiredQuantity ? "error" : ""}`}
							value={formData.desiredQuantity}
							onChange={handleChange}
							min={1}
							max={10000}
							placeholder="1-10000"
							required
						/>
						{errors.desiredQuantity ? <div className="form-error">{errors.desiredQuantity}</div> : null}
					</div>
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

				<div className="form-group">
					<label htmlFor="message" className="form-label">
						メッセージ
					</label>
					<input
						id="message"
						name="message"
						type="text"
						className="form-input"
						value={formData.message}
						onChange={handleChange}
						placeholder="例: 明日の子ども食堂で使いたいです"
					/>
				</div>

				<div className="form-actions">
					<button type="submit" className="btn btn-primary" disabled={submitting}>
						{submitting ? "作成中..." : "リクエストを作成"}
					</button>
					<button type="button" className="btn btn-outline" onClick={() => navigate("/fusion")} disabled={submitting}>
						キャンセル
					</button>
				</div>
			</form>
		</div>
	);
}
