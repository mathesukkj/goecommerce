package service

import (
	"database/sql"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/mathesukkj/goecommerce/order-service/internal/dto"
	"github.com/mathesukkj/goecommerce/order-service/internal/entity"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordTooLong = errors.New("password too long")
	ErrUserExists      = errors.New("user with this username or email already exists")
	ErrUserNotFound    = errors.New("user not found")
)

type UserService struct {
	db *sqlx.DB
}

func NewUserService(db *sqlx.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Signup(user dto.UserPayload) (*dto.LoginResponse, error) {
	userId, err := s.CreateUser(user)
	if err != nil {
		return nil, err
	}

	token, err := generateToken(userId)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{Token: token}, nil
}

func (s *UserService) CreateUser(user dto.UserPayload) (int, error) {
	query := `
		INSERT INTO users (username, password, email, first_name, last_name, phone_number)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING user_id
	`

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err == bcrypt.ErrPasswordTooLong {
		return 0, ErrPasswordTooLong
	} else if err != nil {
		return 0, err
	}

	var userId int
	err = s.db.QueryRowx(
		query,
		user.Username,
		string(hashedPassword),
		user.Email,
		user.FirstName,
		user.LastName,
		user.PhoneNumber,
	).Scan(
		&userId,
	)
	if err != nil {
		return 0, err
	}

	return userId, nil
}

func (s *UserService) Login(login dto.LoginPayload) (*dto.LoginResponse, error) {
	query := `SELECT user_id, password FROM users WHERE email = $1`

	var user entity.User
	err := s.db.QueryRowx(query, login.Email).Scan(&user.UserID, &user.Password)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)); err != nil {
		return nil, err
	}

	token, err := generateToken(user.UserID)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{Token: token}, nil
}

func generateToken(userId int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	secret := os.Getenv("JWT_SECRET")
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}
