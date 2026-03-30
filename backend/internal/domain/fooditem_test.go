package domain

import (
	"testing"
	"time"
)

func TestFoodItem_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		expiry   time.Time
		now      time.Time
		expected bool
	}{
		{
			name:     "not expired - expiry in future",
			expiry:   time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC),
			now:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "expired - expiry in past",
			expiry:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			now:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "not expired - exact same time",
			expiry:   time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
			now:      time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "expired - one second after",
			expiry:   time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
			now:      time.Date(2026, 4, 1, 12, 0, 1, 0, time.UTC),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &FoodItem{ExpiryDate: tt.expiry}
			got := item.IsExpired(tt.now)
			if got != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFoodItemStatus_Valid(t *testing.T) {
	statuses := []FoodItemStatus{
		FoodItemStatusAvailable,
		FoodItemStatusReserved,
		FoodItemStatusConsumed,
		FoodItemStatusExpired,
		FoodItemStatusDeleted,
	}

	for _, s := range statuses {
		if !ValidFoodItemStatuses[s] {
			t.Errorf("expected %q to be a valid status", s)
		}
	}

	if ValidFoodItemStatuses["invalid"] {
		t.Error("expected 'invalid' to not be a valid status")
	}
}
