package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddressPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload AddressPayload
		wantErr bool
	}{
		{
			name: "valid payload",
			payload: AddressPayload{
				StreetAddress: "123 Main St",
				City:          "Anytown",
				State:         "CA",
				PostalCode:    "12345",
				Country:       "USA",
			},
			wantErr: false,
		},
		{
			name: "missing street address",
			payload: AddressPayload{
				StreetAddress: "",
				City:          "Anytown",
				State:         "CA",
				PostalCode:    "12345",
				Country:       "USA",
			},
			wantErr: true,
		},
		{
			name: "missing city",
			payload: AddressPayload{
				StreetAddress: "123 Main St",
				City:          "",
				State:         "CA",
				PostalCode:    "12345",
				Country:       "USA",
			},
			wantErr: true,
		},
		{
			name: "missing state",
			payload: AddressPayload{
				StreetAddress: "123 Main St",
				City:          "Anytown",
				State:         "",
				PostalCode:    "12345",
				Country:       "USA",
			},
			wantErr: true,
		},
		{
			name: "missing postal code",
			payload: AddressPayload{
				StreetAddress: "123 Main St",
				City:          "Anytown",
				State:         "CA",
				PostalCode:    "",
				Country:       "USA",
			},
			wantErr: true,
		},
		{
			name: "missing country",
			payload: AddressPayload{
				StreetAddress: "123 Main St",
				City:          "Anytown",
				State:         "CA",
				PostalCode:    "12345",
				Country:       "",
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
