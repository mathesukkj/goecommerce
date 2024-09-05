package entity

type Address struct {
	AddressID     int    `json:"address_id" db:"address_id"`
	UserID        int    `json:"-" db:"user_id"`
	StreetAddress string `json:"street_address" db:"street_address"`
	City          string `json:"city" db:"city"`
	State         string `json:"state" db:"state"`
	PostalCode    string `json:"postal_code" db:"postal_code"`
	Country       string `json:"country" db:"country"`
}
