package dto

import "github.com/go-playground/validator/v10"

type SignupPayload struct {
	Username    string         `json:"username" validate:"required"`
	Password    string         `json:"password" validate:"required"`
	Email       string         `json:"email" validate:"required,email"`
	FirstName   string         `json:"first_name" validate:"required"`
	LastName    string         `json:"last_name" validate:"required"`
	PhoneNumber string         `json:"phone_number" validate:"required"`
	Address     AddressPayload `json:"address" validate:"required"`
}

func (s *SignupPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

type LoginPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (l *LoginPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(l)
}

// used for login and signup response
type LoginResponse struct {
	UserID int    `json:"user_id"`
	Token  string `json:"token"`
}
