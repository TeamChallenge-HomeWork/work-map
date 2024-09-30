package handlers

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	pb "workmap/gateway/internal/gapi/proto_gen"
)

// MockAuthServiceClient is a mock for AuthServiceClient
type MockAuthServiceClient struct {
	mock.Mock
}

func (m *MockAuthServiceClient) Register(ctx context.Context, in *pb.RegisterRequest, opts ...grpc.CallOption) (*pb.RegisterReply, error) {
	args := m.Called(ctx, in)

	return args.Get(0).(*pb.RegisterReply), args.Error(1)
}

func (m *MockAuthServiceClient) Login(ctx context.Context, in *pb.LoginRequest, opts ...grpc.CallOption) (*pb.LoginReply, error) {
	args := m.Called(ctx, in)

	return args.Get(0).(*pb.LoginReply), args.Error(1)
}

func (m *MockAuthServiceClient) Logout(ctx context.Context, in *pb.LogoutRequest, opts ...grpc.CallOption) (*pb.LogoutReply, error) {
	return nil, nil
}

func (m *MockAuthServiceClient) RefreshToken(ctx context.Context, in *pb.RefreshTokenRequest, opts ...grpc.CallOption) (*pb.RefreshTokenReply, error) {
	return nil, nil
}

// MockRedis is a mock for Redis
type MockRedis struct {
	mock.Mock
}

func (m *MockRedis) GetAccessToken(accessToken string) error {
	args := m.Called(accessToken)

	return args.Error(0)
}

func (m *MockRedis) SaveAccessToken(accessToken string) error {
	args := m.Called(accessToken)

	return args.Error(0)
}

func (m *MockRedis) DeleteAccessToken(accessToken string) error {
	args := m.Called(accessToken)

	return args.Error(0)
}

type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) Validate() error {
	args := m.Called()

	fmt.Println(args.Error(0))

	return args.Error(0)
}
