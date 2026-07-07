// Package config はサービス層の調整可能なパラメータを環境変数から読み込む。
//
// マッチング・一覧・入力バリデーションの各しきい値をハードコードせず、運用環境
// ごとに環境変数で調整できるようにする。環境変数が未設定または不正な場合は、
// 従来のハードコード値と同じデフォルト値にフォールバックする。
package config

import (
	"os"
	"strconv"
)

// ServiceConfig はサービス層の調整可能なパラメータ。
type ServiceConfig struct {
	// DefaultPageSize は一覧APIでページサイズ未指定時に用いる既定値。
	DefaultPageSize int
	// MaxPageSize は一覧APIで許容する最大ページサイズ。
	MaxPageSize int
	// MinQuantity は数量の下限（食品登録・融通リクエスト共通）。
	MinQuantity int32
	// MaxQuantity は数量の上限（食品登録・融通リクエスト共通）。
	MaxQuantity int32
	// MaxNameLength は食品名の最大文字数。
	MaxNameLength int
	// MaxMessageLength は融通リクエストメッセージの最大文字数。
	MaxMessageLength int
}

// デフォルト値。環境変数が未設定・不正な場合のフォールバックに用いる。
const (
	defaultDefaultPageSize  = 20
	defaultMaxPageSize      = 100
	defaultMinQuantity      = 1
	defaultMaxQuantity      = 10000
	defaultMaxNameLength    = 200
	defaultMaxMessageLength = 5000
)

// 環境変数名。
const (
	envDefaultPageSize  = "SHOKUDO_DEFAULT_PAGE_SIZE"
	envMaxPageSize      = "SHOKUDO_MAX_PAGE_SIZE"
	envMinQuantity      = "SHOKUDO_MIN_QUANTITY"
	envMaxQuantity      = "SHOKUDO_MAX_QUANTITY"
	envMaxNameLength    = "SHOKUDO_MAX_NAME_LENGTH"
	envMaxMessageLength = "SHOKUDO_MAX_MESSAGE_LENGTH"
)

// Default はすべてデフォルト値を持つ ServiceConfig を返す。テストや
// 明示的にデフォルトを使いたい箇所で利用する。
func Default() ServiceConfig {
	return ServiceConfig{
		DefaultPageSize:  defaultDefaultPageSize,
		MaxPageSize:      defaultMaxPageSize,
		MinQuantity:      defaultMinQuantity,
		MaxQuantity:      defaultMaxQuantity,
		MaxNameLength:    defaultMaxNameLength,
		MaxMessageLength: defaultMaxMessageLength,
	}
}

// LoadServiceConfig は環境変数から ServiceConfig を読み込む。
// 各値が未設定・パース不能・非正の場合はデフォルト値にフォールバックする。
func LoadServiceConfig() ServiceConfig {
	return ServiceConfig{
		DefaultPageSize:  getEnvInt(envDefaultPageSize, defaultDefaultPageSize),
		MaxPageSize:      getEnvInt(envMaxPageSize, defaultMaxPageSize),
		MinQuantity:      getEnvInt32(envMinQuantity, defaultMinQuantity),
		MaxQuantity:      getEnvInt32(envMaxQuantity, defaultMaxQuantity),
		MaxNameLength:    getEnvInt(envMaxNameLength, defaultMaxNameLength),
		MaxMessageLength: getEnvInt(envMaxMessageLength, defaultMaxMessageLength),
	}
}

// getEnvInt は環境変数を正の整数として読み込む。
// 未設定・パース失敗・非正の値の場合は def を返す。
func getEnvInt(key string, def int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return def
	}
	return v
}

// getEnvInt32 は環境変数を正の int32 として読み込む。
// 未設定・パース失敗・非正・範囲外の場合は def を返す。
func getEnvInt32(key string, def int32) int32 {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}
	v, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || v <= 0 {
		return def
	}
	return int32(v)
}
