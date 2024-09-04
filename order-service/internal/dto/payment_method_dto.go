package dto

import "github.com/go-playground/validator/v10"

type PaymentMethodPayload struct {
	PaymentType    string `json:"payment_type" validate:"required"`
	CardNumber     string `json:"card_number" validate:"required"`
	ExpirationDate string `json:"expiration_date" validate:"required"`
	CardHolderName string `json:"card_holder_name" validate:"required"`
}

func (p *PaymentMethodPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}
