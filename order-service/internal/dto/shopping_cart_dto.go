package dto

import "github.com/go-playground/validator/v10"

type AddToCartPayload struct {
	ProductID int `json:"product_id" validate:"required"`
	Quantity  int `json:"quantity" validate:"required,min=1"`
}

func (p *AddToCartPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}
