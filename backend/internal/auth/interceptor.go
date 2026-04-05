package auth

import (
	"context"
	"log/slog"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// skipMethods は認証をスキップするメソッド一覧。
// ヘルスチェックとリフレクションは未認証で利用可能にする。
var skipMethods = map[string]bool{
	"/grpc.health.v1.Health/Check": true,
	"/grpc.health.v1.Health/Watch": true,
}

// TokenVerifier はIDトークンを検証するインターフェース。
// テスト時にモック差し替えが可能。
type TokenVerifier interface {
	VerifyIDToken(ctx context.Context, idToken string) (*UserInfo, error)
}

// UnaryInterceptor はFirebase Auth トークン検証を行うgRPC UnaryServerInterceptor。
// リクエストのメタデータから "authorization" ヘッダーを取得し、
// Bearer トークンを検証してユーザー情報をコンテキストに注入する。
func UnaryInterceptor(verifier TokenVerifier) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// gRPC reflection はスキップ（開発用）
		if strings.HasPrefix(info.FullMethod, "/grpc.reflection.") {
			return handler(ctx, req)
		}

		// ヘルスチェック等はスキップ
		if skipMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// メタデータから authorization ヘッダーを取得
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			slog.Warn("missing metadata in request", "method", info.FullMethod)
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			slog.Warn("missing authorization header", "method", info.FullMethod)
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		// "Bearer <token>" 形式からトークンを抽出
		token, err := extractBearerToken(authHeaders[0])
		if err != nil {
			slog.Warn("invalid authorization header format", "method", info.FullMethod, "error", err)
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		// トークン検証
		user, err := verifier.VerifyIDToken(ctx, token)
		if err != nil {
			slog.Warn("token verification failed",
				"method", info.FullMethod,
				"error", err,
			)
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		slog.Debug("authenticated request",
			"method", info.FullMethod,
			"uid", user.UID,
		)

		// ユーザー情報をコンテキストに注入してハンドラに渡す
		return handler(WithUser(ctx, user), req)
	}
}

// extractBearerToken は "Bearer <token>" 形式の文字列からトークン部分を抽出する。
func extractBearerToken(authHeader string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", status.Error(codes.Unauthenticated, "authorization header must start with 'Bearer '")
	}
	token := strings.TrimPrefix(authHeader, prefix)
	if token == "" {
		return "", status.Error(codes.Unauthenticated, "empty bearer token")
	}
	return token, nil
}
