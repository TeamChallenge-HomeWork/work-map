package gapi

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "workmap/gateway/internal/gapi/proto_gen"
)

type AuthConfig struct {
	Host string
	Port string
}

func NewAuthService(cfg *AuthConfig) (pb.AuthServiceClient, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewAuthServiceClient(conn)

	return client, nil
}
