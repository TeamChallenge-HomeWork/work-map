package models

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr error
	}{
		{
			name:        "Valid data",
			email:       "address@gmail.com",
			password:    "Qwerty_123",
			expectedErr: nil,
		},
		{
			name:        "Invalid email",
			email:       "invalid-email",
			password:    "Qwerty_123",
			expectedErr: errors.New("invalid email"),
		},
		{
			name:        "Empty email",
			email:       "",
			password:    "Qwerty_123",
			expectedErr: errors.New("invalid request"),
		},
		{
			name:        "Empty password",
			email:       "address@gmail.com",
			password:    "",
			expectedErr: errors.New("invalid request"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				Email:    tt.email,
				Password: tt.password,
			}

			err := u.Validate()
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
