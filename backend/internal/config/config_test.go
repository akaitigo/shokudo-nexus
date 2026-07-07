package config

import "testing"

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.DefaultPageSize != 20 {
		t.Errorf("DefaultPageSize: got %d, want 20", cfg.DefaultPageSize)
	}
	if cfg.MaxPageSize != 100 {
		t.Errorf("MaxPageSize: got %d, want 100", cfg.MaxPageSize)
	}
	if cfg.MinQuantity != 1 {
		t.Errorf("MinQuantity: got %d, want 1", cfg.MinQuantity)
	}
	if cfg.MaxQuantity != 10000 {
		t.Errorf("MaxQuantity: got %d, want 10000", cfg.MaxQuantity)
	}
	if cfg.MaxNameLength != 200 {
		t.Errorf("MaxNameLength: got %d, want 200", cfg.MaxNameLength)
	}
	if cfg.MaxMessageLength != 5000 {
		t.Errorf("MaxMessageLength: got %d, want 5000", cfg.MaxMessageLength)
	}
}

// clearConfigEnv は全設定の環境変数を空文字にし（=未設定扱い）、
// 各テストがデフォルトから開始できるようにする。t.Setenv でテスト終了時に復元される。
func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		envDefaultPageSize, envMaxPageSize, envMinQuantity,
		envMaxQuantity, envMaxNameLength, envMaxMessageLength,
	} {
		t.Setenv(key, "")
	}
}

func TestLoadServiceConfig_UsesDefaultsWhenUnset(t *testing.T) {
	clearConfigEnv(t)
	if got := LoadServiceConfig(); got != Default() {
		t.Errorf("expected defaults, got %+v", got)
	}
}

func TestLoadServiceConfig_ReadsEnv(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv(envDefaultPageSize, "10")
	t.Setenv(envMaxPageSize, "50")
	t.Setenv(envMinQuantity, "2")
	t.Setenv(envMaxQuantity, "500")
	t.Setenv(envMaxNameLength, "80")
	t.Setenv(envMaxMessageLength, "1000")

	cfg := LoadServiceConfig()
	want := ServiceConfig{
		DefaultPageSize:  10,
		MaxPageSize:      50,
		MinQuantity:      2,
		MaxQuantity:      500,
		MaxNameLength:    80,
		MaxMessageLength: 1000,
	}
	if cfg != want {
		t.Errorf("got %+v, want %+v", cfg, want)
	}
}

func TestLoadServiceConfig_InvalidValuesFallBackToDefault(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"non-numeric", "abc"},
		{"zero", "0"},
		{"negative", "-5"},
		{"float", "3.14"},
		{"empty", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearConfigEnv(t)
			t.Setenv(envMaxPageSize, tt.value)
			if got := LoadServiceConfig().MaxPageSize; got != Default().MaxPageSize {
				t.Errorf("MaxPageSize with %q: got %d, want default %d", tt.value, got, Default().MaxPageSize)
			}
		})
	}
}

func TestLoadServiceConfig_Int32FieldsParse(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv(envMinQuantity, "3")
	t.Setenv(envMaxQuantity, "9999")

	cfg := LoadServiceConfig()
	if cfg.MinQuantity != 3 {
		t.Errorf("MinQuantity: got %d, want 3", cfg.MinQuantity)
	}
	if cfg.MaxQuantity != 9999 {
		t.Errorf("MaxQuantity: got %d, want 9999", cfg.MaxQuantity)
	}
}

func TestLoadServiceConfig_Int32OverflowFallsBack(t *testing.T) {
	clearConfigEnv(t)
	// int32 の範囲を超える値はデフォルトにフォールバックする。
	t.Setenv(envMaxQuantity, "3000000000")
	if got := LoadServiceConfig().MaxQuantity; got != Default().MaxQuantity {
		t.Errorf("expected fallback to default %d for overflow, got %d", Default().MaxQuantity, got)
	}
}
