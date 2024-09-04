package entity

type Address struct {
	AddressID     int    `json:"address_id"`
	UserID        int    `json:"user_id"`
	StreetAddress string `json:"street_address"`
	City          string `json:"city"`
	State         string `json:"state"`
	PostalCode    string `json:"postal_code"`
	Country       string `json:"country"`
}
