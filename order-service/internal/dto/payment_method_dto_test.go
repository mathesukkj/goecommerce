package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaymentMethodPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload PaymentMethodPayload
		wantErr bool
	}{
		{
			name: "valid payload",
			payload: PaymentMethodPayload{
				PaymentType:    "credit_card",
				CardNumber:     "1234567890123456",
				ExpirationDate: "12/25",
				CardHolderName: "John Doe",
			},
			wantErr: false,
		},
		{
			name: "missing payment type",
			payload: PaymentMethodPayload{
				CardNumber:     "1234567890123456",
				ExpirationDate: "12/25",
				CardHolderName: "John Doe",
			},
			wantErr: true,
		},
		{
			name: "missing card number",
			payload: PaymentMethodPayload{
				PaymentType:    "credit_card",
				ExpirationDate: "12/25",
				CardHolderName: "John Doe",
			},
			wantErr: true,
		},
		{
			name: "missing expiration date",
			payload: PaymentMethodPayload{
				PaymentType:    "credit_card",
				CardNumber:     "1234567890123456",
				CardHolderName: "John Doe",
			},
			wantErr: true,
		},
		{
			name: "missing card holder name",
			payload: PaymentMethodPayload{
				PaymentType:    "credit_card",
				CardNumber:     "1234567890123456",
				ExpirationDate: "12/25",
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
