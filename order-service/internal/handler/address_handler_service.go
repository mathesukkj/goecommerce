package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mathesukkj/goecommerce/order-service/internal/dto"
	"github.com/mathesukkj/goecommerce/order-service/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupAddressHandler(t *testing.T) (*AddressHandler, *postgres.PostgresContainer) {
	t.Helper()

	db.MustExec("TRUNCATE TABLE address CASCADE")
	db.MustExec("ALTER SEQUENCE address_address_id_seq RESTART WITH 1")
	db.MustExec("TRUNCATE TABLE users CASCADE")
	db.MustExec("ALTER SEQUENCE users_user_id_seq RESTART WITH 1")

	addressHandler := NewAddressHandler(db)
	return addressHandler, pgContainer
}

func seedAddresss(t *testing.T) {
	t.Helper()

	params := map[string]interface{}{
		"addressname":  "address",
		"password":     "$2y$10$bsRLuOQN606nDdkFCF2D4eF74rON7JXEP.RxTAKbgTft2BgqtJgYu",
		"email":        "test@example.com",
		"first_name":   "address",
		"last_name":    "Address",
		"phone_number": "1234567890",
	}

	_, err := db.NamedExec(`
		INSERT INTO address (addressname, password, email, first_name, last_name, phone_number) 
		VALUES (:addressname, :password, :email, :first_name, :last_name, :phone_number)
	`, params)
	if err != nil {
		t.Fatalf("failed to seed address: %s", err)
	}
}

func TestListUserAddresses(t *testing.T) {
	addressHandler, _ := setupAddressHandler(t)
	seedAddresss(t)

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
			req, err := http.NewRequest(http.MethodGet, "/addresses", nil)
			assert.NoError(t, err)

			ctx := context.WithValue(req.Context(), keyUserId, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(addressHandler.ListUserAddresses)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var addresses []entity.Address
				err = json.NewDecoder(rr.Body).Decode(&addresses)
				assert.NoError(t, err)
				assert.Len(t, addresses, tt.expectedLength)
			}
		})
	}
}

func TestGetAddressByID(t *testing.T) {
	addressHandler, _ := setupAddressHandler(t)
	seedAddresss(t)
	seedUsers(t)

	tests := []struct {
		name            string
		addressID       string
		expectedStatus  int
		expectedAddress *entity.Address
	}{
		{
			name:           "success",
			addressID:      "1",
			expectedStatus: http.StatusOK,
			expectedAddress: &entity.Address{
				AddressID:     1,
				UserID:        1,
				StreetAddress: "123 Main St",
				City:          "Anytown",
				State:         "CA",
				PostalCode:    "12345",
				Country:       "USA",
			},
		},
		{
			name:            "invalid address id",
			addressID:       "invalid",
			expectedStatus:  http.StatusBadRequest,
			expectedAddress: nil,
		},
		{
			name:            "address not found",
			addressID:       "999",
			expectedStatus:  http.StatusNotFound,
			expectedAddress: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/addresses/"+tt.addressID, nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(addressHandler.GetAddressByID)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var address entity.Address
				err = json.NewDecoder(rr.Body).Decode(&address)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAddress, &address)
			}
		})
	}
}

func TestCreateAddress(t *testing.T) {
	addressHandler, _ := setupAddressHandler(t)
	seedAddresss(t)
	seedUsers(t)

	tests := []struct {
		name            string
		userID          int
		payload         dto.AddressPayload
		expectedStatus  int
		expectedAddress *entity.Address
	}{
		{
			name:   "success",
			userID: 1,
			payload: dto.AddressPayload{
				StreetAddress: "123 Main St",
				City:          "Anytown",
				State:         "CA",
				PostalCode:    "12345",
				Country:       "USA",
			},
			expectedStatus: http.StatusOK,
			expectedAddress: &entity.Address{
				AddressID:     1,
				UserID:        1,
				StreetAddress: "123 Main St",
				City:          "Anytown",
				State:         "CA",
				PostalCode:    "12345",
				Country:       "USA",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, _ := json.Marshal(tt.payload)
			req, err := http.NewRequest(http.MethodPost, "/addresses", bytes.NewBuffer(payload))
			assert.NoError(t, err)

			ctx := context.WithValue(req.Context(), keyUserId, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(addressHandler.CreateAddress)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var address entity.Address
				err = json.NewDecoder(rr.Body).Decode(&address)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAddress, &address)
			}
		})
	}
}

func TestUpdateAddress(t *testing.T) {
	addressHandler, _ := setupAddressHandler(t)
	seedAddresss(t)
	seedUsers(t)

	tests := []struct {
		name            string
		addressID       string
		payload         dto.AddressPayload
		expectedStatus  int
		expectedAddress *entity.Address
	}{
		{
			name:      "success",
			addressID: "1",
			payload: dto.AddressPayload{
				StreetAddress: "123 Main St",
				City:          "Anytown",
				State:         "CA",
				PostalCode:    "12345",
				Country:       "USA",
			},
			expectedStatus: http.StatusOK,
			expectedAddress: &entity.Address{
				AddressID:     1,
				UserID:        1,
				StreetAddress: "123 Main St",
				State:         "CA",
				PostalCode:    "12345",
				Country:       "USA",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, _ := json.Marshal(tt.payload)
			req, err := http.NewRequest(http.MethodPut, "/addresses/"+tt.addressID, bytes.NewBuffer(payload))
			assert.NoError(t, err)

			ctx := context.WithValue(req.Context(), keyUserId, 1)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(addressHandler.UpdateAddress)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var address entity.Address
				err = json.NewDecoder(rr.Body).Decode(&address)
				assert.NoError(t, err)

				assert.Equal(t, tt.expectedAddress, &address)
			}
		})
	}
}

func TestDeleteAddress(t *testing.T) {
	addressHandler, _ := setupAddressHandler(t)
	seedAddresss(t)
	seedUsers(t)

	tests := []struct {
		name           string
		addressID      string
		expectedStatus int
		serviceErr     error
	}{
		{
			name:           "success",
			addressID:      "1",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid address id",
			addressID:      "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "address not found",
			addressID:      "999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodDelete, "/addresses/"+tt.addressID, nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(addressHandler.DeleteAddress)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
