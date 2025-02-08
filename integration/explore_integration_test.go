// +build integration

package integration_test

import (
	"context"
	"log"
	"testing"
	"time"

	"google.golang.org/grpc"

	pb "github.com/KEdore/explore/proto"
)

// waitForServer is a helper that waits until the gRPC service is available.
func waitForServer(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
		cancel()
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return context.DeadlineExceeded
}

// TestCountLikedYou verifies that when no user has liked the given recipient, the count is zero.
func TestCountLikedYou(t *testing.T) {
	addr := "localhost:9090"
	if err := waitForServer(addr, 10*time.Second); err != nil {
		t.Fatalf("gRPC server not ready at %s: %v", addr, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewExploreServiceClient(conn)

	req := &pb.CountLikedYouRequest{
		RecipientUserId: "test-recipient",
	}
	resp, err := client.CountLikedYou(ctx, req)
	if err != nil {
		t.Fatalf("CountLikedYou RPC failed: %v", err)
	}

	if resp.Count != 0 {
		t.Errorf("Expected count 0, got %d", resp.Count)
	} else {
		log.Printf("CountLikedYou for recipient %q returned: %d", req.RecipientUserId, resp.Count)
	}
}
