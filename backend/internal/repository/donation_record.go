package repository

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"

	"github.com/akaitigo/shokudo-nexus/backend/internal/domain"
)

const donationRecordsCollection = "donation_records"

// DonationRecordRepository はDonationRecordのFirestoreリポジトリ。
type DonationRecordRepository struct {
	client *firestore.Client
}

// NewDonationRecordRepository は新しいDonationRecordRepositoryを生成する。
func NewDonationRecordRepository(client *firestore.Client) *DonationRecordRepository {
	return &DonationRecordRepository{client: client}
}

// Create は寄付履歴を作成する。
func (r *DonationRecordRepository) Create(ctx context.Context, record *domain.DonationRecord) (*domain.DonationRecord, error) {
	record.ID = uuid.New().String()
	_, err := r.client.Collection(donationRecordsCollection).Doc(record.ID).Set(ctx, record)
	if err != nil {
		return nil, fmt.Errorf("failed to create donation record: %w", err)
	}
	return record, nil
}
