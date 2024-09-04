package entity

type User struct {
	UserID      int    `json:"user_id" db:"user_id"`
	Username    string `json:"username" db:"username"`
	Password    string `json:"-" db:"password"`
	Email       string `json:"email" db:"email"`
	FirstName   string `json:"first_name" db:"first_name"`
	LastName    string `json:"last_name" db:"last_name"`
	PhoneNumber string `json:"phone_number" db:"phone_number"`
}
