package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
	pb "workmap/gateway/internal/gapi/proto_gen"
	"workmap/gateway/internal/store"
)

func TestGetTTL(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		exp           float64
		expectedError error
	}{
		{
			name:          "valid token",
			input:         "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InVzZXIzOUBlbWFpbC5jb20iLCJuYmYiOjE3MjEyNTI5NjksImV4cCI6MTcyMTI1Mjk3OSwiaWF0IjoxNzIxMjUyOTY5fQ.kgAoGtXbJgHGDWtE2QTeZACjhZ4EOoz10gq6HW_zbCSg3g7QSagOToYHgWaEecBJpg7yQ-DaCjY6BCyiEClA7Q",
			exp:           1721252979,
			expectedError: nil,
		},
		{
			name:          "invalid token",
			input:         "invalid.token",
			exp:           0,
			expectedError: errors.New("cannot split the token string"),
		},
		{
			name:          "wrong token",
			input:         "not.a.token",
			exp:           0,
			expectedError: errors.New("illegal base64 data at input byte 0"),
		},
		{
			name:          "token without \"exp\" field",
			input:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			exp:           0,
			expectedError: errors.New("exp not found in the token"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ttl, err := getTTL(tt.input)

			if tt.expectedError != nil {
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("unexpected error: got %v, want %v", err, tt.expectedError)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				expString := strconv.FormatFloat(tt.exp, 'f', -1, 64)
				i, err := strconv.ParseInt(expString, 10, 64)
				if err != nil {
					t.Fatal(err)
				}

				expectedTTL := time.Until(time.Unix(i, 0))
				if expectedTTL.Round(time.Second) != ttl.Round(time.Second) {
					t.Errorf("unexpected response: got %v, want %v", ttl, expectedTTL)
				}
			}
		})
	}
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

func TestUserRegister(t *testing.T) {
	logger := zap.NewNop()

	var (
		email    = gofakeit.Email()
		password = gofakeit.Password(true, true, true, true, false, 12)
		at       = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InVzZXIzOUBlbWFpbC5jb20iLCJuYmYiOjE3MjEyNTI5NjksImV4cCI6MTcyMTI1Mjk3OSwiaWF0IjoxNzIxMjUyOTY5fQ.kgAoGtXbJgHGDWtE2QTeZACjhZ4EOoz10gq6HW_zbCSg3g7QSagOToYHgWaEecBJpg7yQ-DaCjY6BCyiEClA7Q"
		rt       = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InVzZXIzOUBlbWFpbC5jb20iLCJuYmYiOjE3MjEyNTI5NjksImV4cCI6MTcyMTI1Mjk3OSwiaWF0IjoxNzIxMjUyOTY5fQ.kgAoGtXbJgHGDWtE2QTeZACjhZ4EOoz10gq6HW_zbCSg3g7QSagOToYHgWaEecBJpg7yQ-DaCjY6BCyiEClA7Q"
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
			expectedStatus: http.StatusCreated,
		},
		{
			name: "user already exist",
			input: user{
				Email:    email,
				Password: password,
			},
			mockResponse:   nil,
			mockError:      status.New(codes.AlreadyExists, "auth service error").Err(),
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "empty request body",
			input:          user{},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "wrong request body",
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
		{
			name: "no connection to redis",
			input: user{
				Email:    email,
				Password: password,
			},
			mockResponse: &pb.RegisterReply{
				RefreshToken: rt,
				AccessToken:  at,
			},
			mockError:      nil,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "wrong access token",
			input: user{
				Email:    email,
				Password: password,
			},
			mockResponse: &pb.RegisterReply{
				RefreshToken: rt,
				AccessToken:  "wrongToken",
			},
			mockError:      nil,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthService := new(MockAuthServiceClient)

			server, _ := miniredis.Run()
			rc := redis.NewClient(&redis.Options{
				Addr: server.Addr(),
			})

			r := &store.Redis{Client: *rc}

			handler := &Handler{
				logger: logger,
				auth:   mockAuthService,
				redis:  *r,
			}

			if tt.name == "no connection to redis" {
				server.Close()
			}

			mockAuthService.On("Register", mock.Anything, mock.Anything).Return(tt.mockResponse, tt.mockError)

			body, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("failed to marshal input: %v", err)
			}

			if tt.name == "wrong request body" {
				body = make([]byte, 0)
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

			if tt.expectedStatus == http.StatusCreated {
				resp := rr.Header().Get("Authorization")
				exp := fmt.Sprintf("Bearer %s", tt.mockResponse.AccessToken)
				if resp != exp {
					t.Errorf("unexpected access token: got %v want %v", resp, exp)
				}

				cookies := rr.Result().Cookies()
				if len(cookies) == 0 {
					t.Errorf("expected a cookie but got none")
				} else if cookies[0].Value != tt.mockResponse.RefreshToken {
					t.Errorf("unexpected cookie: got %v want %v", cookies[0].Value, tt.mockResponse.RefreshToken)
				}
			} else if tt.expectedStatus == http.StatusConflict {
				expected := "User already exist"
				if strings.TrimSpace(rr.Body.String()) != expected {
					t.Errorf("handler returned unexpected body: got %v want %v",
						strings.TrimSpace(rr.Body.String()), expected)
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

func TestUserLogin(t *testing.T) {
	logger := zap.NewNop()

	var (
		email    = gofakeit.Email()
		password = gofakeit.Password(true, true, true, true, false, 12)
		at       = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InVzZXIzOUBlbWFpbC5jb20iLCJuYmYiOjE3MjEyNTI5NjksImV4cCI6MTcyMTI1Mjk3OSwiaWF0IjoxNzIxMjUyOTY5fQ.kgAoGtXbJgHGDWtE2QTeZACjhZ4EOoz10gq6HW_zbCSg3g7QSagOToYHgWaEecBJpg7yQ-DaCjY6BCyiEClA7Q"
		rt       = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InVzZXIzOUBlbWFpbC5jb20iLCJuYmYiOjE3MjEyNTI5NjksImV4cCI6MTcyMTI1Mjk3OSwiaWF0IjoxNzIxMjUyOTY5fQ.kgAoGtXbJgHGDWtE2QTeZACjhZ4EOoz10gq6HW_zbCSg3g7QSagOToYHgWaEecBJpg7yQ-DaCjY6BCyiEClA7Q"
	)

	tests := []struct {
		name           string
		input          user
		mockResponse   *pb.LoginReply
		mockError      error
		expectedStatus int
	}{
		{
			name: "valid request",
			input: user{
				Email:    email,
				Password: password,
			},
			mockResponse: &pb.LoginReply{
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
			name:           "wrong request body",
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
		{
			name: "no connection to redis",
			input: user{
				Email:    email,
				Password: password,
			},
			mockResponse: &pb.LoginReply{
				RefreshToken: rt,
				AccessToken:  at,
			},
			mockError:      nil,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "wrong access token",
			input: user{
				Email:    email,
				Password: password,
			},
			mockResponse: &pb.LoginReply{
				RefreshToken: rt,
				AccessToken:  "wrongToken",
			},
			mockError:      nil,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthService := new(MockAuthServiceClient)

			server, _ := miniredis.Run()
			rc := redis.NewClient(&redis.Options{
				Addr: server.Addr(),
			})

			r := &store.Redis{Client: *rc}

			handler := &Handler{
				logger: logger,
				auth:   mockAuthService,
				redis:  *r,
			}

			if tt.name == "no connection to redis" {
				server.Close()
			}

			mockAuthService.On("Login", mock.Anything, mock.Anything).Return(tt.mockResponse, tt.mockError)

			body, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("failed to marshal input: %v", err)
			}

			if tt.name == "wrong request body" {
				body = make([]byte, 0)
			}

			// TODO method and url is not necessary, but why?
			req, err := http.NewRequest("POST", "/user/login", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.UserLogin(rr, req)

			if status1 := rr.Code; status1 != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status1, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				resp := rr.Header().Get("Authorization")
				exp := fmt.Sprintf("Bearer %s", tt.mockResponse.AccessToken)
				if resp != exp {
					t.Errorf("unexpected access token: got %v want %v", resp, exp)
				}

				cookies := rr.Result().Cookies()
				if len(cookies) == 0 {
					t.Errorf("expected a cookie but got none")
				} else if cookies[0].Value != tt.mockResponse.RefreshToken {
					t.Errorf("unexpected cookie: got %v want %v", cookies[0].Value, tt.mockResponse.RefreshToken)
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
