package service

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/mathesukkj/goecommerce/order-service/internal/dto"
	"github.com/mathesukkj/goecommerce/order-service/internal/entity"
	"github.com/stretchr/testify/assert"
)

func setupPaymentMethodService(t *testing.T) (*PaymentMethodService, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}

	paymentMethodService := NewPaymentMethodService(sqlx.NewDb(db, "postgres"))
	return paymentMethodService, mock
}

func seedPaymentMethods(t *testing.T, service *PaymentMethodService) {
	t.Helper()

	params := map[string]interface{}{
		"user_id":              1,
		"street_paymentMethod": "123 Main St",
		"city":                 "New York",
		"state":                "NY",
		"postal_code":          "10001",
		"country":              "USA",
	}

	service.db.NamedExec(`
		INSERT INTO payment_methods (user_id, payment_type, card_number, expiration_date, card_holder_name) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`, params)
}

func TestListUserPaymentMethods(t *testing.T) {
	paymentMethodService, mock := setupPaymentMethodService(t)
	seedPaymentMethods(t, paymentMethodService)
	query := "SELECT * FROM payment_methods WHERE user_id = $1"

	tests := []struct {
		name    string
		userID  int
		want    int
		wantErr bool
	}{
		{
			name:    "existing user with payment methods",
			userID:  1,
			want:    1,
			wantErr: false,
		},
		{
			name:    "user with no payment methods",
			userID:  2,
			want:    0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"payment_method_id", "user_id", "payment_type", "card_number", "expiration_date", "card_holder_name"})

			if tt.want > 0 {
				rows.AddRow(1, tt.userID, "Credit Card", "1234567890123456", "2025-12-31", "John Doe")
			}

			mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(tt.userID).WillReturnRows(rows)

			got, err := paymentMethodService.ListUserPaymentMethods(tt.userID)
			assert.Equal(t, tt.wantErr, err != nil, "ListUserPaymentMethods() error = %v, wantErr %v", err, tt.wantErr)
			if !tt.wantErr {
				assert.Len(t, got, tt.want, "ListUserPaymentMethods() got = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func TestGetPaymentMethodByID(t *testing.T) {
	paymentMethodService, mock := setupPaymentMethodService(t)
	seedPaymentMethods(t, paymentMethodService)
	query := "SELECT * FROM payment_methods WHERE payment_method_id = $1"

	tests := []struct {
		name            string
		paymentMethodID int
		want            *entity.PaymentMethod
		wantErr         bool
	}{
		{
			name:            "existing payment method",
			paymentMethodID: 1,
			want: &entity.PaymentMethod{
				PaymentMethodID: 1,
				UserID:          1,
				PaymentType:     "Credit Card",
				CardNumber:      "1234567890123456",
				ExpirationDate:  "2025-12-31",
				CardHolderName:  "John Doe",
			},
			wantErr: false,
		},
		{
			name:            "non-existing payment method",
			paymentMethodID: 999,
			want:            nil,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"payment_method_id", "user_id", "payment_type", "card_number", "expiration_date", "card_holder_name"})

			if tt.want != nil {
				rows.AddRow(tt.paymentMethodID, 1, "Credit Card", "1234567890123456", "2025-12-31", "John Doe")
			}

			mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(tt.paymentMethodID).WillReturnRows(rows)

			got, err := paymentMethodService.GetPaymentMethodByID(tt.paymentMethodID)
			assert.Equal(t, tt.wantErr, err != nil, "GetPaymentMethodByID() error = %v, wantErr %v", err, tt.wantErr)
			if !tt.wantErr {
				assert.Equal(t, tt.want, got, "GetPaymentMethodByID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreatePaymentMethod(t *testing.T) {
	paymentMethodService, mock := setupPaymentMethodService(t)

	query := `
		INSERT INTO payment_methods (user_id, payment_type, card_number, expiration_date, card_holder_name)
		VALUES (:user_id, :payment_type, :card_number, :expiration_date, :card_holder_name)
		RETURNING *
	`

	tests := []struct {
		name    string
		payload dto.PaymentMethodPayload
		userID  int
		want    *entity.PaymentMethod
		wantErr bool
	}{
		{
			name: "valid payment method",
			payload: dto.PaymentMethodPayload{
				PaymentType:    "Credit Card",
				CardNumber:     "1234567890123456",
				ExpirationDate: "2025-12-31",
				CardHolderName: "John Doe",
			},
			userID: 1,
			want: &entity.PaymentMethod{
				PaymentMethodID: 1,
				UserID:          1,
				PaymentType:     "Credit Card",
				CardNumber:      "1234567890123456",
				ExpirationDate:  "2025-12-31",
				CardHolderName:  "John Doe",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(tt.userID, tt.payload.PaymentType, tt.payload.CardNumber, tt.payload.ExpirationDate, tt.payload.CardHolderName).WillReturnRows(sqlmock.NewRows([]string{"payment_method_id", "user_id", "payment_type", "card_number", "expiration_date", "card_holder_name"}).AddRow(tt.want.PaymentMethodID, tt.want.UserID, tt.want.PaymentType, tt.want.CardNumber, tt.want.ExpirationDate, tt.want.CardHolderName))

			got, err := paymentMethodService.CreatePaymentMethod(tt.payload, tt.userID)
			assert.Equal(t, tt.wantErr, err != nil, "CreatePaymentMethod() error = %v, wantErr %v", err, tt.wantErr)
			if !tt.wantErr {
				assert.Equal(t, tt.want, got, "CreatePaymentMethod() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdatePaymentMethod(t *testing.T) {
	paymentMethodService, mock := setupPaymentMethodService(t)
	query := `
		UPDATE payment_methods
		SET payment_type = :payment_type, card_number = :card_number, expiration_date = :expiration_date, card_holder_name = :card_holder_name
		WHERE payment_method_id = :payment_method_id
		RETURNING *
	`

	tests := []struct {
		name            string
		paymentMethodID int
		payload         dto.PaymentMethodPayload
		want            *entity.PaymentMethod
		wantErr         bool
	}{
		{
			name:            "valid update",
			paymentMethodID: 1,
			payload: dto.PaymentMethodPayload{
				PaymentType:    "Credit Card",
				CardNumber:     "1234567890123456",
				ExpirationDate: "2025-12-31",
				CardHolderName: "John Doe",
			},
			want: &entity.PaymentMethod{
				PaymentMethodID: 1,
				UserID:          1,
				PaymentType:     "Credit Card",
				CardNumber:      "1234567890123456",
				ExpirationDate:  "2025-12-31",
				CardHolderName:  "John Doe",
			},
			wantErr: false,
		},
		{
			name:            "non-existing payment method",
			paymentMethodID: 999,
			payload: dto.PaymentMethodPayload{
				PaymentType:    "Credit Card",
				CardNumber:     "1234567890123456",
				ExpirationDate: "2025-12-31",
				CardHolderName: "John Doe",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{
				"payment_type",
				"user_id",
				"card_number",
				"expiration_date",
				"card_holder_name",
				"payment_method_id",
			})
			if tt.want != nil {
				rows.AddRow(tt.want.PaymentType, tt.want.UserID, tt.want.CardNumber, tt.want.ExpirationDate, tt.want.CardHolderName, tt.want.PaymentMethodID)
			}

			mock.ExpectQuery(regexp.QuoteMeta(query)).
				WithArgs(tt.payload.PaymentType, tt.payload.CardNumber, tt.payload.ExpirationDate, tt.payload.CardHolderName, tt.paymentMethodID).
				WillReturnRows(rows)

			got, err := paymentMethodService.UpdatePaymentMethod(tt.paymentMethodID, tt.payload)
			if tt.wantErr {
				assert.Error(t, err, "UpdatePaymentMethod() should have returned an error")
				assert.Nil(t, got, "UpdatePaymentMethod() should have returned nil")
			} else {
				assert.NoError(t, err, "UpdatePaymentMethod() unexpected error")
				assert.Equal(t, tt.want, got, "UpdatePaymentMethod() returned unexpected result")
			}
		})
	}
}

func TestDeletePaymentMethod(t *testing.T) {
	paymentMethodService, mock := setupPaymentMethodService(t)
	query := "DELETE FROM payment_methods WHERE payment_method_id = $1"

	tests := []struct {
		name            string
		paymentMethodID int
		wantErr         bool
	}{
		{
			name:            "existing payment method",
			paymentMethodID: 1,
			wantErr:         false,
		},
		{
			name:            "non-existing payment method",
			paymentMethodID: 999,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(tt.paymentMethodID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			} else {
				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(tt.paymentMethodID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			}

			err := paymentMethodService.DeletePaymentMethod(tt.paymentMethodID)
			if tt.wantErr {
				assert.Error(t, err, "DeletePaymentMethod() should have returned an error")
				assert.Equal(t, "payment method not found", err.Error(), "Unexpected error message")
			} else {
				assert.NoError(t, err, "DeletePaymentMethod() unexpected error")
			}
		})
	}
}
