package dto

import "github.com/go-playground/validator/v10"

type OrderPayload struct {
	OrderDate         string `json:"order_date" validate:"required"`
	TotalAmount       int    `json:"total_amount" validate:"required,min=1"`
	PaymentMethodID   int    `json:"payment_method_id" validate:"required,min=1"`
	ShippingAddressID int    `json:"shipping_address_id" validate:"required,min=1"`
	OrderStatus       string `json:"order_status" validate:"required"`
}

func (a *OrderPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}
