package entity

type PaymentMethod struct {
	PaymentMethodID int    `json:"payment_method_id"`
	UserID          int    `json:"user_id"`
	PaymentType     string `json:"payment_type"`
	CardNumber      string `json:"card_number"`
	ExpirationDate  string `json:"expiration_date"`
	CardHolderName  string `json:"card_holder_name"`
}
