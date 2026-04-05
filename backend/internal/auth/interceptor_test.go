package auth

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// mockVerifier はテスト用のTokenVerifierモック。
type mockVerifier struct {
	verifyFunc func(ctx context.Context, idToken string) (*UserInfo, error)
}

func (m *mockVerifier) VerifyIDToken(ctx context.Context, idToken string) (*UserInfo, error) {
	return m.verifyFunc(ctx, idToken)
}

func newMockVerifier(user *UserInfo, err error) *mockVerifier {
	return &mockVerifier{
		verifyFunc: func(_ context.Context, _ string) (*UserInfo, error) {
			return user, err
		},
	}
}

// noopHandler はテスト用のgRPCハンドラ。コンテキストからユーザー情報を返す。
func noopHandler(ctx context.Context, _ any) (any, error) {
	user := UserFromContext(ctx)
	return user, nil
}

func TestUnaryInterceptor_ValidToken(t *testing.T) {
	verifier := newMockVerifier(&UserInfo{UID: "user-123", Email: "test@example.com"}, nil)
	interceptor := UnaryInterceptor(verifier)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer valid-token"))
	info := &grpc.UnaryServerInfo{FullMethod: "/shokudo.v1.FoodInventoryService/CreateFoodItem"}

	resp, err := interceptor(ctx, nil, info, noopHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user, ok := resp.(*UserInfo)
	if !ok || user == nil {
		t.Fatal("expected UserInfo in response")
	}
	if user.UID != "user-123" {
		t.Errorf("expected UID 'user-123', got %q", user.UID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", user.Email)
	}
}

func TestUnaryInterceptor_MissingMetadata(t *testing.T) {
	verifier := newMockVerifier(nil, nil)
	interceptor := UnaryInterceptor(verifier)

	// メタデータなしのコンテキスト
	ctx := context.Background()
	info := &grpc.UnaryServerInfo{FullMethod: "/shokudo.v1.FoodInventoryService/CreateFoodItem"}

	_, err := interceptor(ctx, nil, info, noopHandler)
	if err == nil {
		t.Fatal("expected error for missing metadata")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestUnaryInterceptor_MissingAuthorizationHeader(t *testing.T) {
	verifier := newMockVerifier(nil, nil)
	interceptor := UnaryInterceptor(verifier)

	// authorization ヘッダーなしのメタデータ
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("content-type", "application/grpc"))
	info := &grpc.UnaryServerInfo{FullMethod: "/shokudo.v1.FoodInventoryService/CreateFoodItem"}

	_, err := interceptor(ctx, nil, info, noopHandler)
	if err == nil {
		t.Fatal("expected error for missing authorization header")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestUnaryInterceptor_InvalidBearerFormat(t *testing.T) {
	verifier := newMockVerifier(nil, nil)
	interceptor := UnaryInterceptor(verifier)

	tests := []struct {
		name       string
		authHeader string
	}{
		{"no bearer prefix", "just-a-token"},
		{"empty bearer", "Bearer "},
		{"basic auth", "Basic dXNlcjpwYXNz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", tt.authHeader))
			info := &grpc.UnaryServerInfo{FullMethod: "/shokudo.v1.FoodInventoryService/CreateFoodItem"}

			_, err := interceptor(ctx, nil, info, noopHandler)
			if err == nil {
				t.Fatal("expected error for invalid bearer format")
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("expected gRPC status error, got: %v", err)
			}
			if st.Code() != codes.Unauthenticated {
				t.Errorf("expected Unauthenticated, got %v", st.Code())
			}
		})
	}
}

func TestUnaryInterceptor_TokenVerificationFailed(t *testing.T) {
	verifier := newMockVerifier(nil, errors.New("token expired"))
	interceptor := UnaryInterceptor(verifier)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer expired-token"))
	info := &grpc.UnaryServerInfo{FullMethod: "/shokudo.v1.FoodInventoryService/CreateFoodItem"}

	_, err := interceptor(ctx, nil, info, noopHandler)
	if err == nil {
		t.Fatal("expected error for failed token verification")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestUnaryInterceptor_HealthCheckSkipped(t *testing.T) {
	verifier := newMockVerifier(nil, errors.New("should not be called"))
	interceptor := UnaryInterceptor(verifier)

	// 認証なしのコンテキスト
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(nil))
	info := &grpc.UnaryServerInfo{FullMethod: "/grpc.health.v1.Health/Check"}

	_, err := interceptor(ctx, nil, info, noopHandler)
	if err != nil {
		t.Fatalf("expected no error for health check, got: %v", err)
	}
}

func TestUnaryInterceptor_ReflectionSkipped(t *testing.T) {
	verifier := newMockVerifier(nil, errors.New("should not be called"))
	interceptor := UnaryInterceptor(verifier)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(nil))
	info := &grpc.UnaryServerInfo{FullMethod: "/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo"}

	_, err := interceptor(ctx, nil, info, noopHandler)
	if err != nil {
		t.Fatalf("expected no error for reflection, got: %v", err)
	}
}

func TestExtractBearerToken_Valid(t *testing.T) {
	token, err := extractBearerToken("Bearer my-valid-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "my-valid-token" {
		t.Errorf("expected 'my-valid-token', got %q", token)
	}
}

func TestExtractBearerToken_NoBearerPrefix(t *testing.T) {
	_, err := extractBearerToken("not-a-bearer-token")
	if err == nil {
		t.Fatal("expected error for missing Bearer prefix")
	}
}

func TestExtractBearerToken_EmptyToken(t *testing.T) {
	_, err := extractBearerToken("Bearer ")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}
