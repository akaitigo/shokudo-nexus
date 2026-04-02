package domain

import "testing"

func TestIsValidCategory(t *testing.T) {
	tests := []struct {
		category string
		expected bool
	}{
		{"野菜", true},
		{"肉", true},
		{"魚", true},
		{"乳製品", true},
		{"穀物", true},
		{"その他", true},
		{"invalid", false},
		{"", false},
		{"果物", false},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			got := IsValidCategory(tt.category)
			if got != tt.expected {
				t.Errorf("IsValidCategory(%q) = %v, want %v", tt.category, got, tt.expected)
			}
		})
	}
}

func TestIsValidUnit(t *testing.T) {
	tests := []struct {
		unit     string
		expected bool
	}{
		{"kg", true},
		{"個", true},
		{"パック", true},
		{"本", true},
		{"袋", true},
		{"箱", true},
		{"リットル", false},
		{"", false},
		{"g", false},
	}

	for _, tt := range tests {
		t.Run(tt.unit, func(t *testing.T) {
			got := IsValidUnit(tt.unit)
			if got != tt.expected {
				t.Errorf("IsValidUnit(%q) = %v, want %v", tt.unit, got, tt.expected)
			}
		})
	}
}
