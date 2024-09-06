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

func setupAddressHandler(t *testing.T) (*AddressHandler, *postgres.PostgresContainer) {
	t.Helper()

	db.MustExec("TRUNCATE TABLE addresses CASCADE")
	db.MustExec("ALTER SEQUENCE addresses_address_id_seq RESTART WITH 1")
	db.MustExec("TRUNCATE TABLE users CASCADE")
	db.MustExec("ALTER SEQUENCE users_user_id_seq RESTART WITH 1")

	addressHandler := NewAddressHandler(db)
	return addressHandler, pgContainer
}

func seedAddress(t *testing.T) {
	t.Helper()

	seedUsers(t)

	params := map[string]interface{}{
		"user_id":        1,
		"street_address": "123 Main St",
		"city":           "Anytown",
		"state":          "CA",
		"postal_code":    "12345",
		"country":        "USA",
	}

	_, err := db.NamedExec(`
		INSERT INTO addresses (user_id, street_address, city, state, postal_code, country) 
    VALUES (:user_id, :street_address, :city, :state, :postal_code, :country)
	`, params)
	if err != nil {
		t.Fatalf("failed to seed address: %s", err)
	}
}

func TestListUserAddresses(t *testing.T) {
	addressHandler, _ := setupAddressHandler(t)
	seedAddress(t)

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
			addressHandler.ListUserAddresses(rr, req)

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
	seedAddress(t)

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
			req.SetPathValue("address_id", tt.addressID)
			assert.NoError(t, err)
			req.SetPathValue("address_id", tt.addressID)

			rr := httptest.NewRecorder()
			addressHandler.GetAddressByID(rr, req)

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
	seedAddress(t)

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
			addressHandler.CreateAddress(rr, req)

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
	seedAddress(t)

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
			req, err := http.NewRequest(
				http.MethodPut,
				"/addresses/"+tt.addressID,
				bytes.NewBuffer(payload),
			)
			req.SetPathValue("address_id", tt.addressID)
			assert.NoError(t, err)

			ctx := context.WithValue(req.Context(), keyUserId, 1)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			addressHandler.UpdateAddress(rr, req)

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
	seedAddress(t)

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
			req.SetPathValue("address_id", tt.addressID)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			addressHandler.DeleteAddress(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
