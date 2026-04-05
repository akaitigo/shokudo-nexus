package auth

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	firebaseAuth "firebase.google.com/go/v4/auth"
)

// FirebaseVerifier はFirebase Admin SDKを使ってIDトークンを検証する。
type FirebaseVerifier struct {
	client *firebaseAuth.Client
}

// NewFirebaseVerifier はFirebase Appからトークン検証クライアントを初期化する。
func NewFirebaseVerifier(ctx context.Context, app *firebase.App) (*FirebaseVerifier, error) {
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase Auth client: %w", err)
	}
	return &FirebaseVerifier{client: client}, nil
}

// VerifyIDToken はFirebase IDトークンを検証し、ユーザー情報を返す。
func (v *FirebaseVerifier) VerifyIDToken(ctx context.Context, idToken string) (*UserInfo, error) {
	token, err := v.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	email, _ := token.Claims["email"].(string)
	return &UserInfo{
		UID:   token.UID,
		Email: email,
	}, nil
}
