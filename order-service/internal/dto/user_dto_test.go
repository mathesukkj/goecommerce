package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoginPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload LoginPayload
		wantErr bool
	}{
		{
			name: "valid payload",
			payload: LoginPayload{
				Email:    "test@example.com",
				Password: "password",
			},
			wantErr: false,
		},
		{
			name: "missing email",
			payload: LoginPayload{
				Email:    "",
				Password: "password",
			},
			wantErr: true,
		},
		{
			name: "missing password",
			payload: LoginPayload{
				Email:    "test@example.com",
				Password: "",
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			payload: LoginPayload{
				Email:    "testexample.com",
				Password: "password",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestSignupPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload SignupPayload
		wantErr bool
	}{
		{
			name: "valid payload",
			payload: SignupPayload{
				Username:    "testuser",
				Password:    "password",
				Email:       "test@example.com",
				FirstName:   "John",
				LastName:    "Doe",
				PhoneNumber: "1234567890",
			},
			wantErr: false,
		},
		{
			name: "missing username",
			payload: SignupPayload{
				Password:    "password",
				Email:       "test@example.com",
				FirstName:   "John",
				LastName:    "Doe",
				PhoneNumber: "1234567890",
			},
			wantErr: true,
		},
		{
			name: "missing password",
			payload: SignupPayload{
				Username:    "testuser",
				Email:       "test@example.com",
				FirstName:   "John",
				LastName:    "Doe",
				PhoneNumber: "1234567890",
			},
			wantErr: true,
		},
		{
			name: "missing email",
			payload: SignupPayload{
				Username:    "testuser",
				Password:    "password",
				FirstName:   "John",
				LastName:    "Doe",
				PhoneNumber: "1234567890",
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			payload: SignupPayload{
				Username:    "testuser",
				Password:    "password",
				Email:       "testexample.com",
				FirstName:   "John",
				LastName:    "Doe",
				PhoneNumber: "1234567890",
			},
			wantErr: true,
		},
		{
			name: "missing first name",
			payload: SignupPayload{
				Username:    "testuser",
				Password:    "password",
				Email:       "test@example.com",
				LastName:    "Doe",
				PhoneNumber: "1234567890",
			},
			wantErr: true,
		},
		{
			name: "missing last name",
			payload: SignupPayload{
				Username:    "testuser",
				Password:    "password",
				Email:       "test@example.com",
				FirstName:   "John",
				PhoneNumber: "1234567890",
			},
			wantErr: true,
		},
		{
			name: "missing phone number",
			payload: SignupPayload{
				Username:  "testuser",
				Password:  "password",
				Email:     "test@example.com",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
