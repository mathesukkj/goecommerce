package service

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/mathesukkj/goecommerce/order-service/internal/dto"
	"github.com/mathesukkj/goecommerce/order-service/internal/entity"
)

func setupAddressService(t *testing.T) (*AddressService, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}

	addressService := NewAddressService(sqlx.NewDb(db, "postgres"))
	return addressService, mock
}

func seedAddresses(t *testing.T, service *AddressService) {
	t.Helper()

	params := map[string]interface{}{
		"user_id":        1,
		"street_address": "123 Main St",
		"city":           "New York",
		"state":          "NY",
		"postal_code":    "10001",
		"country":        "USA",
	}

	service.db.NamedExec(`
		INSERT INTO addresses (user_id, street_address, city, state, postal_code, country) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`, params)
}

func TestListUserAddresses(t *testing.T) {
	addressService, mock := setupAddressService(t)
	seedAddresses(t, addressService)
	query := "SELECT address_id, user_id, street_address, city, state, postal_code, country FROM addresses WHERE user_id = $1"

	tests := []struct {
		name    string
		userID  int
		want    int
		wantErr bool
	}{
		{
			name:    "existing user with addresses",
			userID:  1,
			want:    1,
			wantErr: false,
		},
		{
			name:    "user with no addresses",
			userID:  2,
			want:    0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := sqlmock.NewRows(
				[]string{
					"address_id",
					"user_id",
					"street_address",
					"city",
					"state",
					"postal_code",
					"country",
				},
			)

			if tt.want > 0 {
				rows.AddRow(1, tt.userID, "123 Main St", "New York", "NY", "10001", "USA")
			}

			mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(tt.userID).WillReturnRows(rows)

			got, err := addressService.ListUserAddresses(tt.userID)
			assert.Equal(
				t,
				tt.wantErr,
				err != nil,
				"ListUserAddresses() error = %v, wantErr %v",
				err,
				tt.wantErr,
			)
			if !tt.wantErr {
				assert.Len(
					t,
					got,
					tt.want,
					"ListUserAddresses() got = %v, want %v",
					len(got),
					tt.want,
				)
			}
		})
	}
}

func TestGetAddressByID(t *testing.T) {
	addressService, mock := setupAddressService(t)
	seedAddresses(t, addressService)
	query := `SELECT address_id, street_address, city, state, postal_code, country  FROM addresses WHERE address_id = $1 AND user_id = $2`

	tests := []struct {
		name      string
		addressID int
		userID    int
		want      *entity.Address
		wantErr   bool
	}{
		{
			name:      "existing address",
			addressID: 1,
			userID:    1,
			want: &entity.Address{
				AddressID:     1,
				UserID:        1,
				StreetAddress: "123 Main St",
				City:          "New York",
				State:         "NY",
				PostalCode:    "10001",
				Country:       "USA",
			},
			wantErr: false,
		},
		{
			name:      "non-existing address",
			addressID: 999,
			userID:    1,
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := sqlmock.NewRows(
				[]string{
					"address_id",
					"user_id",
					"street_address",
					"city",
					"state",
					"postal_code",
					"country",
				},
			)

			if tt.want != nil {
				rows.AddRow(tt.addressID, 1, "123 Main St", "New York", "NY", "10001", "USA")
			}

			mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(tt.addressID, tt.userID).WillReturnRows(rows)

			got, err := addressService.GetAddressByID(tt.addressID, tt.userID)
			if tt.wantErr {
				assert.Error(t, err, "GetAddressByID() should have returned an error")
				assert.Nil(t, got, "GetAddressByID() should have returned nil")
			} else {
				assert.NoError(t, err, "GetAddressByID() unexpected error")
				assert.Equal(t, tt.want, got, "GetAddressByID() returned unexpected result")
			}
		})
	}
}

func TestCreateAddress(t *testing.T) {
	addressService, mock := setupAddressService(t)
	query := `
		INSERT INTO addresses (user_id, street_address, city, state, postal_code, country)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING address_id, user_id, street_address, city, state, postal_code, country
	`

	tests := []struct {
		name    string
		address dto.AddressPayload
		userID  int
		want    *entity.Address
		wantErr bool
	}{
		{
			name: "valid address",
			address: dto.AddressPayload{
				StreetAddress: "123 Main St",
				City:          "New York",
				State:         "NY",
				PostalCode:    "10001",
				Country:       "USA",
			},
			userID: 1,
			want: &entity.Address{
				StreetAddress: "123 Main St",
				City:          "New York",
				State:         "NY",
				PostalCode:    "10001",
				Country:       "USA",
			},
			wantErr: false,
		},
		{
			name: "invalid address",
			address: dto.AddressPayload{
				StreetAddress: "",
				City:          "",
				State:         "",
				PostalCode:    "",
				Country:       "",
			},
			userID:  1,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := sqlmock.NewRows(
				[]string{"street_address", "city", "state", "postal_code", "country"},
			)

			if tt.want != nil {
				rows.AddRow(
					tt.address.StreetAddress,
					tt.address.City,
					tt.address.State,
					tt.address.PostalCode,
					tt.address.Country,
				)
			}

			mock.ExpectQuery(regexp.QuoteMeta(query)).
				WithArgs(tt.userID, tt.address.StreetAddress, tt.address.City, tt.address.State, tt.address.PostalCode, tt.address.Country).
				WillReturnRows(rows)

			got, err := addressService.CreateAddress(tt.address, tt.userID)
			if tt.wantErr {
				assert.Error(t, err, "CreateAddress() should have returned an error")
				assert.Nil(t, got, "CreateAddress() should have returned nil")
			} else {
				assert.NoError(t, err, "CreateAddress() unexpected error")
				assert.Equal(t, tt.want, got, "CreateAddress() returned unexpected result")
			}
		})
	}
}

func TestUpdateAddress(t *testing.T) {
	addressService, mock := setupAddressService(t)
	query := `
		UPDATE addresses
		SET street_address = $1, city = $2, state = $3, postal_code = $4, country = $5
		WHERE address_id = $6
		RETURNING address_id, user_id, street_address, city, state, postal_code, country
	`

	tests := []struct {
		name      string
		addressID int
		address   dto.AddressPayload
		want      *entity.Address
		wantErr   bool
	}{
		{
			name:      "valid update",
			addressID: 1,
			address: dto.AddressPayload{
				StreetAddress: "456 Elm St",
				City:          "Los Angeles",
				State:         "CA",
				PostalCode:    "90001",
				Country:       "USA",
			},
			want: &entity.Address{
				AddressID:     1,
				StreetAddress: "456 Elm St",
				City:          "Los Angeles",
				State:         "CA",
				PostalCode:    "90001",
				Country:       "USA",
			},
			wantErr: false,
		},
		{
			name:      "address not found",
			addressID: 999,
			address: dto.AddressPayload{
				StreetAddress: "789 Oak St",
				City:          "Chicago",
				State:         "IL",
				PostalCode:    "60601",
				Country:       "USA",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := sqlmock.NewRows(
				[]string{"address_id", "street_address", "city", "state", "postal_code", "country"},
			)
			if tt.want != nil {
				rows.AddRow(
					tt.want.AddressID,
					tt.want.StreetAddress,
					tt.want.City,
					tt.want.State,
					tt.want.PostalCode,
					tt.want.Country,
				)
			}

			mock.ExpectQuery(regexp.QuoteMeta(query)).
				WithArgs(tt.address.StreetAddress, tt.address.City, tt.address.State, tt.address.PostalCode, tt.address.Country, tt.addressID).
				WillReturnRows(rows)

			got, err := addressService.UpdateAddress(tt.addressID, tt.address)
			if tt.wantErr {
				assert.Error(t, err, "UpdateAddress() should have returned an error")
				assert.Nil(t, got, "UpdateAddress() should have returned nil")
			} else {
				assert.NoError(t, err, "UpdateAddress() unexpected error")
				assert.Equal(t, tt.want, got, "UpdateAddress() returned unexpected result")
			}
		})
	}
}

func TestDeleteAddress(t *testing.T) {
	addressService, mock := setupAddressService(t)
	query := "DELETE FROM addresses WHERE address_id = $1"

	tests := []struct {
		name      string
		addressID int
		wantErr   bool
	}{
		{
			name:      "existing address",
			addressID: 1,
			wantErr:   false,
		},
		{
			name:      "non-existing address",
			addressID: 999,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(tt.addressID).
					WillReturnResult(sqlmock.NewResult(0, 0))
			} else {
				mock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(tt.addressID).
					WillReturnResult(sqlmock.NewResult(0, 1))
			}

			err := addressService.DeleteAddress(tt.addressID)
			if tt.wantErr {
				assert.Error(t, err, "DeleteAddress() should have returned an error")
				assert.Equal(t, "address not found", err.Error(), "Unexpected error message")
			} else {
				assert.NoError(t, err, "DeleteAddress() unexpected error")
			}

			assert.NoError(t, mock.ExpectationsWereMet(), "Unfulfilled expectations")
		})
	}
}
