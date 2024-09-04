package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddToCartPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload AddToCartPayload
		wantErr bool
	}{
		{
			name: "valid payload",
			payload: AddToCartPayload{
				ProductID: 1,
				Quantity:  1,
			},
			wantErr: false,
		},
		{
			name: "missing product ID",
			payload: AddToCartPayload{
				Quantity: 1,
			},
			wantErr: true,
		},
		{
			name: "missing quantity",
			payload: AddToCartPayload{
				ProductID: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid quantity",
			payload: AddToCartPayload{
				ProductID: 1,
				Quantity:  0,
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
