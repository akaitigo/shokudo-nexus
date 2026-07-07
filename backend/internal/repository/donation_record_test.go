package repository

import (
	"context"
	"testing"
	"time"

	"github.com/akaitigo/shokudo-nexus/backend/internal/domain"
)

func TestDonationRecordRepository_Create(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)
	repo := NewDonationRecordRepository(client)

	record := &domain.DonationRecord{
		FoodItemID:      "food-1",
		DonorID:         "donor-1",
		RecipientID:     "shokudo-1",
		FusionRequestID: "fusion-1",
		Category:        "野菜",
		Quantity:        5,
		Unit:            "kg",
		CreatedAt:       time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
	}

	created, err := repo.Create(ctx, record)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected generated ID, got empty string")
	}

	// ドキュメントが実際に永続化され、データ変換がラウンドトリップすることを確認する。
	doc, err := client.Collection(donationRecordsCollection).Doc(created.ID).Get(ctx)
	if err != nil {
		t.Fatalf("failed to read back donation record: %v", err)
	}
	var got domain.DonationRecord
	if dataErr := doc.DataTo(&got); dataErr != nil {
		t.Fatalf("failed to decode donation record: %v", dataErr)
	}
	if got.ID != created.ID {
		t.Errorf("ID mismatch: got %q, want %q", got.ID, created.ID)
	}
	if got.FoodItemID != "food-1" {
		t.Errorf("FoodItemID mismatch: got %q", got.FoodItemID)
	}
	if got.Category != "野菜" {
		t.Errorf("Category mismatch: got %q", got.Category)
	}
	if got.Quantity != 5 {
		t.Errorf("Quantity mismatch: got %d", got.Quantity)
	}
}
