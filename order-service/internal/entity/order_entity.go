package entity

type Order struct {
	OrderID           int    `json:"order_id" db:"order_id"`
	UserID            int    `json:"-" db:"user_id"`
	OrderDate         string `json:"order_date" db:"order_date"`
	TotalAmount       int    `json:"total_amount" db:"total_amount"`
	PaymentMethodID   int    `json:"payment_method_id" db:"payment_method_id"`
	ShippingAddressID int    `json:"shipping_address_id" db:"shipping_address_id"`
	OrderStatus       string `json:"order_status" db:"order_status"`
}
