package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"net/http/httptest"
	"testing"
	pb "workmap/gateway/internal/gapi/proto_gen"
	"workmap/gateway/internal/models"
	store "workmap/gateway/internal/redis"
)

func TestUserRegister(t *testing.T) {
	logger := zap.NewNop()

	var (
		email    = gofakeit.Email()
		password = gofakeit.Password(true, true, true, true, false, 12)
		at       = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InVzZXIzOUBlbWFpbC5jb20iLCJuYmYiOjE3MjEyNTI5NjksImV4cCI6MTcyMTI1Mjk3OSwiaWF0IjoxNzIxMjUyOTY5fQ.kgAoGtXbJgHGDWtE2QTeZACjhZ4EOoz10gq6HW_zbCSg3g7QSagOToYHgWaEecBJpg7yQ-DaCjY6BCyiEClA7Q"
		rt       = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InVzZXIzOUBlbWFpbC5jb20iLCJuYmYiOjE3MjEyNTI5NjksImV4cCI6MTcyMTI1Mjk3OSwiaWF0IjoxNzIxMjUyOTY5fQ.kgAoGtXbJgHGDWtE2QTeZACjhZ4EOoz10gq6HW_zbCSg3g7QSagOToYHgWaEecBJpg7yQ-DaCjY6BCyiEClA7Q"
	)

	tests := []struct {
		name             string
		input            models.User
		mockAuthResponse *pb.RegisterReply
		mockAuthError    error
		mockRedisError   error
		expectedStatus   int
		expectedMessage  string
	}{
		{
			name: "valid request",
			input: models.User{
				Email:    email,
				Password: password,
			},
			mockAuthResponse: &pb.RegisterReply{
				RefreshToken: rt,
				AccessToken:  at,
			},
			mockAuthError:   nil,
			expectedStatus:  http.StatusCreated,
			expectedMessage: "",
		},
		{
			name: "user already exist",
			input: models.User{
				Email:    email,
				Password: password,
			},
			mockAuthResponse: nil,
			mockAuthError:    status.New(codes.AlreadyExists, "User email taken").Err(),
			mockRedisError:   nil,
			expectedStatus:   http.StatusConflict,
			expectedMessage:  "User email taken\n",
		},
		{
			name:             "empty request body",
			input:            models.User{},
			mockAuthResponse: nil,
			mockAuthError:    nil,
			mockRedisError:   nil,
			expectedStatus:   http.StatusBadRequest,
			expectedMessage:  "Invalid request\n",
		},
		{
			name:             "wrong request body",
			input:            models.User{},
			mockAuthResponse: nil,
			mockAuthError:    nil,
			mockRedisError:   nil,
			expectedStatus:   http.StatusBadRequest,
			expectedMessage:  "Invalid request\n",
		},
		{
			name: "missing email",
			input: models.User{
				Password: password,
			},
			mockAuthResponse: nil,
			mockAuthError:    nil,
			mockRedisError:   nil,
			expectedStatus:   http.StatusBadRequest,
			expectedMessage:  "Invalid request\n",
		},
		{
			name: "missing password",
			input: models.User{
				Email: email,
			},
			mockAuthResponse: nil,
			mockAuthError:    nil,
			mockRedisError:   nil,
			expectedStatus:   http.StatusBadRequest,
			expectedMessage:  "Invalid request\n",
		},
		{
			name: "auth service error with code",
			input: models.User{
				Email:    email,
				Password: password,
			},
			mockAuthResponse: nil,
			mockAuthError:    status.New(codes.Unavailable, "auth service error").Err(),
			mockRedisError:   nil,
			expectedStatus:   http.StatusBadRequest,
			expectedMessage:  "auth service error\n",
		},
		{
			name: "auth service unexpected error",
			input: models.User{
				Email:    email,
				Password: password,
			},
			mockAuthResponse: nil,
			mockAuthError:    errors.New("unexpected error"),
			mockRedisError:   nil,
			expectedStatus:   http.StatusInternalServerError,
			expectedMessage:  "Internal server error\n",
		},
		{
			name: "wrong refresh token",
			input: models.User{
				Email:    email,
				Password: password,
			},
			mockAuthResponse: &pb.RegisterReply{
				RefreshToken: "wrongToken",
				AccessToken:  at,
			},
			mockAuthError:   nil,
			mockRedisError:  nil,
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "Internal server error\n",
		},
		{
			name: "save access token error",
			input: models.User{
				Email:    email,
				Password: password,
			},
			mockAuthResponse: &pb.RegisterReply{
				RefreshToken: rt,
				AccessToken:  "wrongToken",
			},
			mockAuthError:   nil,
			mockRedisError:  errors.New("tokenStore error"),
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "Internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthService := new(MockAuthServiceClient)
			mockValidator := new(MockValidator)
			mockRedis := new(MockRedis)

			var mockRedisStore store.TokenStore = mockRedis

			handler := &Handler{
				logger:     logger,
				auth:       mockAuthService,
				tokenStore: mockRedisStore,
			}

			mockAuthService.On("Register", mock.Anything, mock.Anything).Return(tt.mockAuthResponse, tt.mockAuthError)
			mockRedis.On("SaveAccessToken", mock.Anything).Return(tt.mockRedisError)
			mockValidator.On("Validate").Return()

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

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedMessage, rr.Body.String())
		})
	}
}
