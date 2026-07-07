package repository

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

// newTestClient は Firestore エミュレータに接続したテスト用クライアントを返す。
//
// FIRESTORE_EMULATOR_HOST が未設定の場合はテストをスキップするため、
// エミュレータのないローカル環境でも `go test ./...` は成功する。CI では
// エミュレータを起動したうえで同環境変数を設定してテストを実行する。
//
// テストごとに一意のプロジェクトIDを用いることで、エミュレータ上のデータ名前空間を
// テスト間で分離し、並行実行時のデータ干渉を防ぐ。
func newTestClient(t *testing.T) *firestore.Client {
	t.Helper()

	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST is not set; skipping Firestore integration test")
	}

	projectID := "test-" + uuid.NewString()
	client, err := firestore.NewClient(context.Background(), projectID, option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("failed to create Firestore client: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := client.Close(); closeErr != nil {
			t.Logf("failed to close Firestore client: %v", closeErr)
		}
	})
	return client
}
