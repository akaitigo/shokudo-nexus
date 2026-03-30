package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	pb "github.com/akaitigo/shokudo-nexus/backend/gen/shokudo/v1"
	"github.com/akaitigo/shokudo-nexus/backend/internal/repository"
	"github.com/akaitigo/shokudo-nexus/backend/internal/service"
)

func main() {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "9090"
	}

	ctx := context.Background()
	fsClient, err := newFirestoreClient(ctx)
	if err != nil {
		log.Fatalf("failed to create Firestore client: %v", err)
	}
	defer func() {
		if closeErr := fsClient.Close(); closeErr != nil {
			log.Printf("failed to close Firestore client: %v", closeErr)
		}
	}()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	// Repositories
	foodItemRepo := repository.NewFoodItemRepository(fsClient)
	fusionRequestRepo := repository.NewFusionRequestRepository(fsClient)
	donationRecordRepo := repository.NewDonationRecordRepository(fsClient)

	// Services
	foodInventorySvc := service.NewFoodInventoryService(foodItemRepo)
	fusionSvc := service.NewFusionService(fusionRequestRepo, foodItemRepo, donationRecordRepo)

	pb.RegisterFoodInventoryServiceServer(s, foodInventorySvc)
	pb.RegisterFusionServiceServer(s, fusionSvc)

	// Health check
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(s, healthServer)

	// Reflection (dev only)
	reflection.Register(s)

	log.Printf("gRPC server listening on :%s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
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
