package auth

import (
	"context"
	"testing"
)

func TestWithUser_And_UserFromContext(t *testing.T) {
	user := &UserInfo{UID: "user-abc", Email: "abc@example.com"}
	ctx := WithUser(context.Background(), user)

	got := UserFromContext(ctx)
	if got == nil {
		t.Fatal("expected non-nil UserInfo")
	}
	if got.UID != "user-abc" {
		t.Errorf("expected UID 'user-abc', got %q", got.UID)
	}
	if got.Email != "abc@example.com" {
		t.Errorf("expected email 'abc@example.com', got %q", got.Email)
	}
}

func TestUserFromContext_EmptyContext(t *testing.T) {
	got := UserFromContext(context.Background())
	if got != nil {
		t.Errorf("expected nil for empty context, got %+v", got)
	}
}
