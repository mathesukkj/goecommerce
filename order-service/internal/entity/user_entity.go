package entity

type User struct {
	UserID      int    `json:"user_id"`
	Username    string `json:"username"`
	Password    string `json:"-"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
}
