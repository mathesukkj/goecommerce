package dto

import "github.com/go-playground/validator/v10"

type AddressPayload struct {
	StreetAddress string `json:"street_address" validate:"required"`
	City          string `json:"city" validate:"required"`
	State         string `json:"state" validate:"required"`
	PostalCode    string `json:"postal_code" validate:"required"`
	Country       string `json:"country" validate:"required"`
}

func (a *AddressPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}
