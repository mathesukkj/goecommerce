package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload OrderPayload
		wantErr bool
	}{
		{
			name: "valid payload",
			payload: OrderPayload{
				OrderDate:         "2022-01-01",
				TotalAmount:       100,
				PaymentMethodID:   1,
				ShippingAddressID: 1,
				OrderStatus:       "pending",
			},
			wantErr: false,
		},
		{
			name: "invalid order date",
			payload: OrderPayload{
				OrderDate:         "",
				TotalAmount:       100,
				PaymentMethodID:   1,
				ShippingAddressID: 1,
				OrderStatus:       "pending",
			},
			wantErr: true,
		},
		{
			name: "invalid total amount",
			payload: OrderPayload{
				OrderDate:         "2022-01-01",
				TotalAmount:       -100,
				PaymentMethodID:   1,
				ShippingAddressID: 1,
				OrderStatus:       "pending",
			},
			wantErr: true,
		},
		{
			name: "invalid payment method ID",
			payload: OrderPayload{
				OrderDate:         "2022-01-01",
				TotalAmount:       100,
				PaymentMethodID:   -1,
				ShippingAddressID: 1,
				OrderStatus:       "pending",
			},
			wantErr: true,
		},
		{
			name: "invalid shipping address ID",
			payload: OrderPayload{
				OrderDate:         "2022-01-01",
				TotalAmount:       100,
				PaymentMethodID:   1,
				ShippingAddressID: -1,
				OrderStatus:       "pending",
			},
			wantErr: true,
		},
		{
			name: "invalid order status",
			payload: OrderPayload{
				OrderDate:         "2022-01-01",
				TotalAmount:       100,
				PaymentMethodID:   1,
				ShippingAddressID: 1,
				OrderStatus:       "",
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
