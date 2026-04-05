package auth

import "context"

// NoopVerifier は認証を無効化する開発・テスト用の検証器。
// DISABLE_AUTH=true で使用される。
// 本番環境では絶対に使用しないこと。
type NoopVerifier struct{}

// VerifyIDToken は常にダミーのユーザー情報を返す。
func (v *NoopVerifier) VerifyIDToken(_ context.Context, _ string) (*UserInfo, error) {
	return &UserInfo{
		UID:   "dev-user",
		Email: "dev@localhost",
	}, nil
}
