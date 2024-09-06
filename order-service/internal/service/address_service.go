package service

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/mathesukkj/goecommerce/order-service/internal/dto"
	"github.com/mathesukkj/goecommerce/order-service/internal/entity"
)

type AddressService struct {
	db *sqlx.DB
}

var (
	ErrAddressNotFound = errors.New("address not found")
)

func NewAddressService(db *sqlx.DB) *AddressService {
	return &AddressService{db: db}
}

func (s *AddressService) ListUserAddresses(userID int) ([]entity.Address, error) {
	query := `SELECT address_id, user_id, street_address, city, state, postal_code, country FROM addresses WHERE user_id = $1`

	var addresses []entity.Address
	rows, err := s.db.Queryx(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var address entity.Address
		if err := rows.StructScan(&address); err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}

	return addresses, nil
}

func (s *AddressService) GetAddressByID(addressID, userID int) (*entity.Address, error) {
	query := `SELECT address_id, street_address, city, state, postal_code, country  FROM addresses WHERE address_id = $1 AND user_id = $2`

	var address entity.Address
	if err := s.db.QueryRowx(query, addressID, userID).StructScan(&address); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAddressNotFound
		}
		return nil, err
	}

	return &address, nil
}

func (s *AddressService) CreateAddress(
	address dto.AddressPayload,
	userId int,
) (*entity.Address, error) {
	query := `
		INSERT INTO addresses (user_id, street_address, city, state, postal_code, country)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING address_id, user_id, street_address, city, state, postal_code, country
	`

	var createdAddress entity.Address
	if err := s.db.QueryRowx(
		query,
		userId,
		address.StreetAddress,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
	).StructScan(&createdAddress); err != nil {
		return nil, err
	}

	return &createdAddress, nil
}

func (s *AddressService) UpdateAddress(
	addressID int,
	address dto.AddressPayload,
) (*entity.Address, error) {
	query := `
		UPDATE addresses
		SET street_address = $1, city = $2, state = $3, postal_code = $4, country = $5
		WHERE address_id = $6
		RETURNING address_id, user_id, street_address, city, state, postal_code, country
	`

	var updatedAddress entity.Address
	if err := s.db.QueryRowx(
		query,
		address.StreetAddress,
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		addressID,
	).StructScan(&updatedAddress); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAddressNotFound
		}
		return nil, err
	}

	return &updatedAddress, nil
}

func (s *AddressService) DeleteAddress(addressID int) error {
	query := `DELETE FROM addresses WHERE address_id = $1`

	result, err := s.db.Exec(query, addressID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrAddressNotFound
	}

	return nil
}
