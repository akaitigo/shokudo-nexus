// Package auth はFirebase Authenticationに基づくgRPC認証機能を提供する。
package auth

import "context"

// UserInfo はFirebase Auth トークンから抽出したユーザー情報。
type UserInfo struct {
	UID   string
	Email string
}

type contextKey struct{}

// WithUser はユーザー情報をコンテキストに格納する。
func WithUser(ctx context.Context, user *UserInfo) context.Context {
	return context.WithValue(ctx, contextKey{}, user)
}

// UserFromContext はコンテキストからユーザー情報を取得する。
// 認証されていない場合は nil を返す。
func UserFromContext(ctx context.Context) *UserInfo {
	user, _ := ctx.Value(contextKey{}).(*UserInfo)
	return user
}
