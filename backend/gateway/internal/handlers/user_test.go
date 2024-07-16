package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	pb "workmap/gateway/internal/gapi/proto_gen"
	"workmap/gateway/internal/store"
)

func (m *MockAuthServiceClient) Register(ctx context.Context, in *pb.RegisterRequest, opts ...grpc.CallOption) (*pb.RegisterReply, error) {
	args := m.Called(ctx, in)

	return args.Get(0).(*pb.RegisterReply), args.Error(1)
}

// Empty implementations for other methods of pb.AuthServiceClient
func (m *MockAuthServiceClient) Login(ctx context.Context, in *pb.LoginRequest, opts ...grpc.CallOption) (*pb.LoginReply, error) {
	return nil, nil
}

func (m *MockAuthServiceClient) Logout(ctx context.Context, in *pb.LogoutRequest, opts ...grpc.CallOption) (*pb.LogoutReply, error) {
	return nil, nil
}

func (m *MockAuthServiceClient) RefreshToken(ctx context.Context, in *pb.RefreshTokenRequest, opts ...grpc.CallOption) (*pb.RefreshTokenReply, error) {
	return nil, nil
}

func TestUserRegister(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	var (
		email    = gofakeit.Email()
		password = gofakeit.Password(true, true, true, true, false, 12)
		at       = gofakeit.UUID()
		rt       = gofakeit.UUID()
	)

	tests := []struct {
		name           string
		input          user
		mockResponse   *pb.RegisterReply
		mockError      error
		expectedStatus int
	}{
		{
			name: "valid request",
			input: user{
				Email:    email,
				Password: password,
			},
			mockResponse: &pb.RegisterReply{
				RefreshToken: rt,
				AccessToken:  at,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty request body",
			input:          user{},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing email",
			input: user{
				Password: password,
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			input: user{
				Email: email,
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "auth service error with code",
			input: user{
				Email:    email,
				Password: password,
			},
			mockResponse:   nil,
			mockError:      status.New(codes.Unavailable, "auth service error").Err(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "auth service unexpected error",
			input: user{
				Email:    email,
				Password: password,
			},
			mockResponse:   nil,
			mockError:      errors.New("unexpected error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthService := new(MockAuthServiceClient)

			testRedis, err := store.NewRedis(&store.RedisConfig{
				Host:     "100.104.232.63",
				Port:     "6366",
				Password: "password",
			})
			if err != nil {
				t.Fatal("failed to create test redis client")
			}

			handler := &Handler{
				logger: logger,
				auth:   mockAuthService,
				redis:  testRedis,
			}

			mockAuthService.On("Register", mock.Anything, mock.Anything).Return(tt.mockResponse, tt.mockError)

			body, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("failed to marshal input: %v", err)
			}

			// TODO method and url is not necessary, but why?
			req, err := http.NewRequest("POST", "/user/register", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.UserRegister(rr, req)

			if status1 := rr.Code; status1 != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status1, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp accessTokenResponse
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if resp.AccessToken != tt.mockResponse.AccessToken {
					t.Errorf("unexpected access token: got %v want %v", resp.AccessToken, tt.mockResponse.AccessToken)
				}
				cookie := rr.Result().Cookies()
				if len(cookie) == 0 || cookie[0].Value != tt.mockResponse.RefreshToken {
					t.Errorf("unexpected cookie: got %v want %v", cookie[0].Value, tt.mockResponse.RefreshToken)
				}
			} else if tt.expectedStatus == http.StatusBadRequest {
				expected := "Invalid request"
				if strings.TrimSpace(rr.Body.String()) != expected {
					t.Errorf("handler returned unexpected body: got %v want %v",
						strings.TrimSpace(rr.Body.String()), expected)
				}
			} else {
				expected := "Internal server error"
				if strings.TrimSpace(rr.Body.String()) != expected {
					t.Errorf("handler returned unexpected body: got %v want %v",
						strings.TrimSpace(rr.Body.String()), expected)
				}
			}
		})
	}
}
