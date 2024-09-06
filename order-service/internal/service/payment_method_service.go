package service

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/mathesukkj/goecommerce/order-service/internal/dto"
	"github.com/mathesukkj/goecommerce/order-service/internal/entity"
)

type PaymentMethodService struct {
	db *sqlx.DB
}

var (
	ErrPaymentMethodNotFound = errors.New("payment method not found")
)

func NewPaymentMethodService(db *sqlx.DB) *PaymentMethodService {
	return &PaymentMethodService{db: db}
}

func (s *PaymentMethodService) ListUserPaymentMethods(userID int) ([]entity.PaymentMethod, error) {
	query := `SELECT payment_method_id, user_id, payment_type, card_number, expiration_date, card_holder_name FROM payment_methods WHERE user_id = $1`

	var paymentMethods []entity.PaymentMethod
	rows, err := s.db.Queryx(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var paymentMethod entity.PaymentMethod
		if err := rows.StructScan(&paymentMethod); err != nil {
			return nil, err
		}
		paymentMethods = append(paymentMethods, paymentMethod)
	}

	return paymentMethods, nil
}

func (s *PaymentMethodService) GetPaymentMethodByID(paymentMethodID, userID int) (*entity.PaymentMethod, error) {
	query := `SELECT payment_method_id, user_id, payment_type, card_number, expiration_date, card_holder_name FROM payment_methods WHERE payment_method_id = $1 AND user_id = $2`

	var paymentMethod entity.PaymentMethod
	if err := s.db.QueryRowx(query, paymentMethodID, userID).StructScan(&paymentMethod); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPaymentMethodNotFound
		}
		return nil, err
	}

	return &paymentMethod, nil
}

func (s *PaymentMethodService) CreatePaymentMethod(payload dto.PaymentMethodPayload, userId int) (*entity.PaymentMethod, error) {
	query := `
		INSERT INTO payment_methods (user_id,  payment_type, card_number, expiration_date, card_holder_name)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING payment_method_id, user_id, payment_type, card_number, expiration_date, card_holder_name
	`

	var createdPaymentMethod entity.PaymentMethod
	if err := s.db.QueryRowx(
		query,
		userId,
		payload.PaymentType,
		payload.CardNumber,
		payload.ExpirationDate,
		payload.CardHolderName,
	).StructScan(&createdPaymentMethod); err != nil {
		return nil, err
	}

	return &createdPaymentMethod, nil
}

func (s *PaymentMethodService) UpdatePaymentMethod(paymentMethodID int, paymentMethod dto.PaymentMethodPayload) (*entity.PaymentMethod, error) {
	query := `
		UPDATE payment_methods
		SET payment_type = $1, card_number = $2, expiration_date = $3, card_holder_name = $4
		WHERE payment_method_id = $5
		RETURNING payment_method_id, user_id, payment_type, card_number, expiration_date, card_holder_name
	`

	var updatedPaymentMethod entity.PaymentMethod
	if err := s.db.QueryRowx(
		query,
		paymentMethod.PaymentType,
		paymentMethod.CardNumber,
		paymentMethod.ExpirationDate,
		paymentMethod.CardHolderName,
		paymentMethodID,
	).StructScan(&updatedPaymentMethod); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPaymentMethodNotFound
		}
		return nil, err
	}

	return &updatedPaymentMethod, nil
}

func (s *PaymentMethodService) DeletePaymentMethod(paymentMethodID int) error {
	query := `DELETE FROM payment_methods WHERE payment_method_id = $1`

	result, err := s.db.Exec(query, paymentMethodID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrPaymentMethodNotFound
	}

	return nil
}
