package token

import (
	"encoding/base64"
	"errors"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestExtractTTL(t *testing.T) {
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

	// TODO refactor this
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ttl, err := ExtractTTL(tt.input)

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

func TestExtractEmail(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedEmail string
		expectedError error
	}{
		{
			name:          "valid token",
			input:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImVtYWlsQGdtYWlsLmNvbSIsInN1YiI6IjEyMzQ1Njc4OTAiLCJuYW1lIjoiSm9obiBEb2UiLCJpYXQiOjE1MTYyMzkwMjJ9.MWI1UQVlomIW5wy_fR9YlofQAdX4yt_fvyx2lj5GlzE",
			expectedEmail: "email@gmail.com",
			expectedError: nil,
		},
		{
			name:          "invalid token",
			input:         "invalid.token",
			expectedEmail: "",
			expectedError: errors.New("cannot split the token string"),
		},
		{
			name:          "wrong token",
			input:         "not.a.token",
			expectedEmail: "",
			expectedError: base64.CorruptInputError(0),
		},
		{
			name:          "token without \"email\" field",
			input:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expectedEmail: "",
			expectedError: errors.New("email not found in the token"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := ExtractEmail(tt.input)

			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedEmail, email)
		})
	}
}
