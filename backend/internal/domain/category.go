package domain

// ValidCategories は許可された食品カテゴリ一覧。
var ValidCategories = map[string]bool{
	"野菜":  true,
	"肉":   true,
	"魚":   true,
	"乳製品": true,
	"穀物":  true,
	"その他": true,
}

// IsValidCategory はカテゴリが定義済みかどうかを判定する。
func IsValidCategory(category string) bool {
	return ValidCategories[category]
}

// ValidUnits は許可された単位一覧。
var ValidUnits = map[string]bool{
	"kg":  true,
	"個":   true,
	"パック": true,
	"本":   true,
	"袋":   true,
	"箱":   true,
}

// IsValidUnit は単位が定義済みかどうかを判定する。
func IsValidUnit(unit string) bool {
	return ValidUnits[unit]
}
