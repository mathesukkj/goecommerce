package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/mathesukkj/goecommerce/order-service/internal/dto"
	"github.com/mathesukkj/goecommerce/order-service/internal/entity"
)

func setupPaymentMethodHandler(t *testing.T) (*PaymentMethodHandler, *postgres.PostgresContainer) {
	t.Helper()

	db.MustExec("TRUNCATE TABLE payment_methods CASCADE")
	db.MustExec("ALTER SEQUENCE payment_methods_payment_method_id_seq RESTART WITH 1")
	db.MustExec("TRUNCATE TABLE users CASCADE")
	db.MustExec("ALTER SEQUENCE users_user_id_seq RESTART WITH 1")

	paymentMethodHandler := NewPaymentMethodHandler(db)
	return paymentMethodHandler, pgContainer
}

func seedPaymentMethods(t *testing.T) {
	t.Helper()

	seedUsers(t)

	params := map[string]interface{}{
		"user_id":          1,
		"payment_type":     "credit_card",
		"card_number":      "1234567890123456",
		"expiration_date":  "2025-12-31",
		"card_holder_name": "John Doe",
	}

	_, err := db.NamedExec(`
		INSERT INTO payment_methods (user_id, payment_type, card_number, expiration_date, card_holder_name) 
		VALUES (:user_id, :payment_type, :card_number, :expiration_date, :card_holder_name)
	`, params)
	if err != nil {
		t.Fatalf("failed to seed paymentMethod: %s", err)
	}
}

func TestListUserPaymentMethods(t *testing.T) {
	paymentMethodHandler, _ := setupPaymentMethodHandler(t)
	seedPaymentMethods(t)

	tests := []struct {
		name           string
		userID         int
		expectedStatus int
		expectedLength int
	}{
		{
			name:           "success",
			userID:         1,
			expectedStatus: http.StatusOK,
			expectedLength: 1,
		},
		{
			name:           "user not logged in",
			userID:         0,
			expectedStatus: http.StatusUnauthorized,
			expectedLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/payment-methods", nil)
			assert.NoError(t, err)

			ctx := context.WithValue(req.Context(), keyUserId, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			paymentMethodHandler.ListUserPaymentMethods(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var paymentMethods []entity.PaymentMethod
				err = json.NewDecoder(rr.Body).Decode(&paymentMethods)
				assert.NoError(t, err)
				assert.Len(t, paymentMethods, tt.expectedLength)
			}
		})
	}
}

func TestGetPaymentMethodByID(t *testing.T) {
	paymentMethodHandler, _ := setupPaymentMethodHandler(t)
	seedPaymentMethods(t)

	tests := []struct {
		name                  string
		paymentMethodID       string
		expectedStatus        int
		expectedPaymentMethod *entity.PaymentMethod
	}{
		{
			name:            "success",
			paymentMethodID: "1",
			expectedStatus:  http.StatusOK,
			expectedPaymentMethod: &entity.PaymentMethod{
				PaymentMethodID: 1,
				UserID:          1,
				PaymentType:     "credit_card",
				CardNumber:      "1234-5678-9012-3456",
				ExpirationDate:  "2025-12-31",
				CardHolderName:  "John Doe",
			},
		},
		{
			name:                  "invalid paymentMethod id",
			paymentMethodID:       "invalid",
			expectedStatus:        http.StatusBadRequest,
			expectedPaymentMethod: nil,
		},
		{
			name:                  "paymentMethod not found",
			paymentMethodID:       "999",
			expectedStatus:        http.StatusNotFound,
			expectedPaymentMethod: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/payment-methods/"+tt.paymentMethodID, nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			paymentMethodHandler.GetPaymentMethodByID(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var paymentMethod entity.PaymentMethod
				err = json.NewDecoder(rr.Body).Decode(&paymentMethod)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPaymentMethod, &paymentMethod)
			}
		})
	}
}

func TestCreatePaymentMethod(t *testing.T) {
	paymentMethodHandler, _ := setupPaymentMethodHandler(t)
	seedPaymentMethods(t)

	tests := []struct {
		name                  string
		userID                int
		payload               dto.PaymentMethodPayload
		expectedStatus        int
		expectedPaymentMethod *entity.PaymentMethod
	}{
		{
			name:   "success",
			userID: 1,
			payload: dto.PaymentMethodPayload{
				PaymentType:    "credit_card",
				CardNumber:     "1234-5678-9012-3456",
				ExpirationDate: "2025-12-31",
				CardHolderName: "John Doe",
			},
			expectedStatus: http.StatusOK,
			expectedPaymentMethod: &entity.PaymentMethod{
				PaymentMethodID: 1,
				UserID:          1,
				PaymentType:     "credit_card",
				CardNumber:      "1234-5678-9012-3456",
				ExpirationDate:  "2025-12-31",
				CardHolderName:  "John Doe",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, _ := json.Marshal(tt.payload)
			req, err := http.NewRequest(
				http.MethodPost,
				"/payment-methods",
				bytes.NewBuffer(payload),
			)
			assert.NoError(t, err)

			ctx := context.WithValue(req.Context(), keyUserId, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			paymentMethodHandler.CreatePaymentMethod(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var paymentMethod entity.PaymentMethod
				err = json.NewDecoder(rr.Body).Decode(&paymentMethod)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPaymentMethod, &paymentMethod)
			}
		})
	}
}

func TestUpdatePaymentMethod(t *testing.T) {
	paymentMethodHandler, _ := setupPaymentMethodHandler(t)
	seedPaymentMethods(t)

	tests := []struct {
		name                  string
		paymentMethodID       string
		payload               dto.PaymentMethodPayload
		expectedStatus        int
		expectedPaymentMethod *entity.PaymentMethod
	}{
		{
			name:            "success",
			paymentMethodID: "1",
			payload: dto.PaymentMethodPayload{
				PaymentType:    "credit_card",
				CardNumber:     "1235-5678-9012-3456",
				ExpirationDate: "2025-12-31",
				CardHolderName: "John Doe",
			},
			expectedStatus: http.StatusOK,
			expectedPaymentMethod: &entity.PaymentMethod{
				PaymentMethodID: 1,
				UserID:          1,
				PaymentType:     "credit_card",
				CardNumber:      "1235-5678-9012-3456",
				ExpirationDate:  "2025-12-31",
				CardHolderName:  "John Doe",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, _ := json.Marshal(tt.payload)
			req, err := http.NewRequest(
				http.MethodPut,
				"/payment-methods/"+tt.paymentMethodID,
				bytes.NewBuffer(payload),
			)
			assert.NoError(t, err)

			ctx := context.WithValue(req.Context(), keyUserId, 1)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			paymentMethodHandler.UpdatePaymentMethod(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var paymentMethod entity.PaymentMethod
				err = json.NewDecoder(rr.Body).Decode(&paymentMethod)
				assert.NoError(t, err)

				assert.Equal(t, tt.expectedPaymentMethod, &paymentMethod)
			}
		})
	}
}

func TestDeletePaymentMethod(t *testing.T) {
	paymentMethodHandler, _ := setupPaymentMethodHandler(t)
	seedPaymentMethods(t)

	tests := []struct {
		name            string
		paymentMethodID string
		expectedStatus  int
		serviceErr      error
	}{
		{
			name:            "success",
			paymentMethodID: "1",
			expectedStatus:  http.StatusNoContent,
		},
		{
			name:            "invalid paymentMethod id",
			paymentMethodID: "invalid",
			expectedStatus:  http.StatusBadRequest,
		},
		{
			name:            "paymentMethod not found",
			paymentMethodID: "999",
			expectedStatus:  http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(
				http.MethodDelete,
				"/payment-methods/"+tt.paymentMethodID,
				nil,
			)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			paymentMethodHandler.DeletePaymentMethod(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
