package entity

type PaymentMethod struct {
	PaymentMethodID int    `json:"payment_method_id" db:"payment_method_id"`
	UserID          int    `json:"-" db:"user_id"`
	PaymentType     string `json:"payment_type" db:"payment_type"`
	CardNumber      string `json:"card_number" db:"card_number"`
	ExpirationDate  string `json:"expiration_date" db:"expiration_date"`
	CardHolderName  string `json:"card_holder_name" db:"card_holder_name"`
}
