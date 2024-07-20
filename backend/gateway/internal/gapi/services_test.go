package gapi

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	pb "workmap/gateway/internal/gapi/proto_gen"
)

// MockAuthServiceServer is a mock implementation of AuthServiceServer.
type MockAuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *MockAuthServiceServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterReply, error) {
	return &pb.RegisterReply{RefreshToken: "mockToken", AccessToken: "mockAccess"}, nil
}

func startMockGRPCServer(t *testing.T) (net.Listener, *grpc.Server) {
	lis, err := net.Listen("tcp", ":0") // ":0" means to use a random available port
	assert.NoError(t, err)

	server := grpc.NewServer()
	pb.RegisterAuthServiceServer(server, &MockAuthServiceServer{})

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	return lis, server
}

func TestNewAuthService(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		lis, server := startMockGRPCServer(t)
		defer server.Stop()

		addr := lis.Addr().String()
		t.Log(addr)
		cfg := &AuthConfig{
			Host: "localhost",
			Port: addr[strings.LastIndex(addr, ":")+1:],
		}

		client, err := NewAuthService(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, client)

		// Perform a simple call to ensure the connection works
		resp, err := client.Register(context.Background(), &pb.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
		})
		assert.NoError(t, err)
		assert.Equal(t, "mockToken", resp.RefreshToken)
		assert.Equal(t, "mockAccess", resp.AccessToken)
	})
}
