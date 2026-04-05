package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	pb "github.com/akaitigo/shokudo-nexus/backend/gen/shokudo/v1"
	"github.com/akaitigo/shokudo-nexus/backend/internal/auth"
	"github.com/akaitigo/shokudo-nexus/backend/internal/repository"
	"github.com/akaitigo/shokudo-nexus/backend/internal/service"
)

func main() {
	// 構造化ログの初期化
	initLogger()

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "9090"
	}

	ctx := context.Background()
	fsClient, err := newFirestoreClient(ctx)
	if err != nil {
		slog.Error("failed to create Firestore client", "error", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := fsClient.Close(); closeErr != nil {
			slog.Error("failed to close Firestore client", "error", closeErr)
		}
	}()

	// Firebase Auth Interceptor の初期化
	verifier, err := newTokenVerifier(ctx)
	if err != nil {
		slog.Error("failed to initialize Firebase Auth verifier", "error", err)
		os.Exit(1)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(auth.UnaryInterceptor(verifier)),
	)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		slog.Error("failed to listen", "port", port, "error", err)
		os.Exit(1)
	}

	// Repositories
	foodItemRepo := repository.NewFoodItemRepository(fsClient)
	fusionRequestRepo := repository.NewFusionRequestRepository(fsClient)
	_ = repository.NewDonationRecordRepository(fsClient) // Phase 2: CompleteFusionRequest で使用予定

	// Services
	foodInventorySvc := service.NewFoodInventoryService(foodItemRepo)
	fusionSvc := service.NewFusionService(fusionRequestRepo, foodItemRepo)

	pb.RegisterFoodInventoryServiceServer(s, foodInventorySvc)
	pb.RegisterFusionServiceServer(s, fusionSvc)

	// Health check
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(s, healthServer)

	// Reflection (dev only)
	reflection.Register(s)

	slog.Info("gRPC server listening", "port", port)
	if err := s.Serve(lis); err != nil {
		slog.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}

// initLogger は構造化ログ（slog）を初期化する。
// LOG_FORMAT=json で JSON 出力、デフォルトはテキスト出力。
func initLogger() {
	logLevel := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: logLevel}

	var handler slog.Handler
	if os.Getenv("LOG_FORMAT") == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

func newFirestoreClient(ctx context.Context) (*firestore.Client, error) {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = "shokudo-nexus-dev"
	}

	// Firestore エミュレータ使用時
	if os.Getenv("FIRESTORE_EMULATOR_HOST") != "" {
		return firestore.NewClient(ctx, projectID, option.WithoutAuthentication())
	}

	return firestore.NewClient(ctx, projectID)
}

// newTokenVerifier はFirebase Auth トークン検証器を作成する。
// FIREBASE_AUTH_EMULATOR_HOST が設定されている場合、または
// DISABLE_AUTH=true の場合はエミュレータ/開発用のnoop検証器を使用する。
func newTokenVerifier(ctx context.Context) (auth.TokenVerifier, error) {
	if os.Getenv("DISABLE_AUTH") == "true" {
		slog.Warn("authentication is disabled (DISABLE_AUTH=true)")
		return &auth.NoopVerifier{}, nil
	}

	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	verifier, err := auth.NewFirebaseVerifier(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firebase verifier: %w", err)
	}

	return verifier, nil
}
