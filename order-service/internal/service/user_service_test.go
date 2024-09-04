package service

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/mathesukkj/goecommerce/order-service/internal/dto"
	"github.com/mathesukkj/goecommerce/order-service/internal/entity"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func setupUserService(t *testing.T) (*UserService, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}

	userService := NewUserService(sqlx.NewDb(db, "postgres"))
	return userService, mock
}

func seedUsers(t *testing.T, service *UserService) {
	t.Helper()

	params := map[string]interface{}{
		"username":     "user",
		"password":     "password",
		"email":        "user@example.com",
		"first_name":   "user",
		"last_name":    "User",
		"phone_number": "1234567890",
	}

	service.db.NamedExec(`
		INSERT INTO users (username, password, email, first_name, last_name, phone_number) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`, params)
}

func TestSignup(t *testing.T) {
	userService, mock := setupUserService(t)
	seedUsers(t, userService)
	query := `
		INSERT INTO users (username, password, email, first_name, last_name, phone_number)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING user_id
	`

	tests := []struct {
		name    string
		user    dto.SignupPayload
		err     error
		wantErr bool
	}{
		{
			name: "valid user payload",
			user: dto.SignupPayload{
				Username:    "testuser",
				Password:    "testpassword",
				Email:       "testuser@example.com",
				FirstName:   "Test",
				LastName:    "User",
				PhoneNumber: "1234567890",
			},
			err:     nil,
			wantErr: false,
		},
		{
			name: "non-unique username",
			user: dto.SignupPayload{
				Username:    "user",
				Password:    "password",
				Email:       "testuser@example.com",
				FirstName:   "Test",
				LastName:    "User",
				PhoneNumber: "1234567890",
			},
			err:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				mock.
					ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(tt.user.Username, sqlmock.AnyArg(), tt.user.Email, tt.user.FirstName, tt.user.LastName, tt.user.PhoneNumber).
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))
			}

			got, err := userService.Signup(tt.user)
			if tt.wantErr {
				assert.Error(t, err, "signup() should have returned an error")
				return
			}

			assert.NoError(t, err, "signup() unexpected error")
			assert.NotNil(t, got, "signup() returned unexpected user_id")
		})
	}
}

func TestCreateUser(t *testing.T) {
	userService, mock := setupUserService(t)
	seedUsers(t, userService)
	query := `
		INSERT INTO users (username, password, email, first_name, last_name, phone_number)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING user_id
	`

	tests := []struct {
		name     string
		user     dto.SignupPayload
		expected int
		err      error
	}{
		{
			name: "create user with valid payload",
			user: dto.SignupPayload{
				Username:    "testuser",
				Password:    "testpassword",
				Email:       "testuser@example.com",
				FirstName:   "Test",
				LastName:    "User",
				PhoneNumber: "1234567890",
			},
			expected: 1,
			err:      nil,
		},
		{
			name: "create user with password too long",
			user: dto.SignupPayload{
				Username:    "testuser",
				Password:    strings.Repeat("a", 73),
				Email:       "testuser@example.com",
				FirstName:   "Test",
				LastName:    "User",
				PhoneNumber: "1234567890",
			},
			err: ErrPasswordTooLong,
		},
		{
			name: "create user with non-unique username",
			user: dto.SignupPayload{
				Username:    "user",
				Password:    "password",
				Email:       "user1@example",
				FirstName:   "user",
				LastName:    "User",
				PhoneNumber: "1234567890",
			},
			err: errors.New("pq: duplicate key value violates unique constraint \"users_pkey\""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err != ErrPasswordTooLong {
				expectedQuery := mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(tt.user.Username, sqlmock.AnyArg(), tt.user.Email, tt.user.FirstName, tt.user.LastName, tt.user.PhoneNumber)
				if tt.expected != 0 {
					expectedQuery.WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(tt.expected))
				}
			}
			_, err := userService.CreateUser(tt.user)
			if tt.err == ErrPasswordTooLong {
				assert.Equal(t, tt.err, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestLogin(t *testing.T) {
	userService, mock := setupUserService(t)
	seedUsers(t, userService)
	query := `SELECT user_id, password FROM users WHERE email = $1`

	tests := []struct {
		name     string
		login    dto.LoginPayload
		mockUser *entity.User
		wantErr  bool
	}{
		{
			name: "successful login",
			login: dto.LoginPayload{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockUser: &entity.User{
				UserID:   1,
				Password: "$2y$10$zuljQprm6i1NQTGfQgB/xeC7wu44vtsb3./R8LuydUc6m1CdS8ziK",
			},
			wantErr: false,
		},
		{
			name: "user not found",
			login: dto.LoginPayload{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			mockUser: nil,
			wantErr:  true,
		},
		{
			name: "incorrect password",
			login: dto.LoginPayload{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			mockUser: &entity.User{
				UserID:   1,
				Password: "$2a$10$1234567890123456789012",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQuery := mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(tt.login.Email)

			if tt.mockUser == nil {
				mockQuery.WillReturnError(ErrUserNotFound)
			} else {
				mockQuery.WillReturnRows(sqlmock.NewRows([]string{"user_id", "password"}).
					AddRow(tt.mockUser.UserID, tt.mockUser.Password))
			}

			token, err := userService.Login(tt.login)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetUserById(t *testing.T) {
	userService, mock := setupUserService(t)
	seedUsers(t, userService)
	query := `SELECT user_id, username, email, first_name, last_name, phone_number FROM users WHERE user_id = $1`

	tests := []struct {
		name     string
		userID   int
		wantUser *entity.User
		wantErr  error
	}{
		{
			name:   "user found",
			userID: 1,
			wantUser: &entity.User{
				UserID:      1,
				Username:    "user",
				Email:       "user@example.com",
				FirstName:   "user",
				LastName:    "User",
				PhoneNumber: "1234567890",
			},
			wantErr: nil,
		},
		{
			name:     "user not found",
			userID:   999,
			wantUser: nil,
			wantErr:  ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQuery := mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(tt.userID)

			if tt.wantUser != nil {
				mockQuery.WillReturnRows(sqlmock.NewRows([]string{"user_id", "username", "email", "first_name", "last_name", "phone_number"}).
					AddRow(tt.wantUser.UserID, tt.wantUser.Username, tt.wantUser.Email, tt.wantUser.FirstName, tt.wantUser.LastName, tt.wantUser.PhoneNumber))
			} else {
				mockQuery.WillReturnError(sql.ErrNoRows)
			}

			gotUser, err := userService.GetUserByID(tt.userID)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, gotUser)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUser, gotUser)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name     string
		userId   int
		wantErr  bool
		jwtToken string
	}{
		{
			name:    "valid user ID",
			userId:  1,
			wantErr: false,
		},
		{
			name:    "zero user ID",
			userId:  0,
			wantErr: false,
		},
		{
			name:    "negative user ID",
			userId:  -1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("JWT_SECRET", "test_secret")
			defer os.Unsetenv("JWT_SECRET")

			got, err := generateToken(tt.userId)
			if tt.wantErr {
				assert.Error(t, err, "generateToken() should have returned an error")
				return
			}
			assert.NoError(t, err, "generateToken() unexpected error")

			token, err := jwt.Parse(got, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte("test_secret"), nil
			})
			assert.NoError(t, err, "failed to parse token")
			assert.True(t, token.Valid, "token is not valid")

			claims, ok := token.Claims.(jwt.MapClaims)
			assert.True(t, ok, "failed to parse claims")
			assert.Equal(t, float64(tt.userId), claims["user_id"], "unexpected user_id in token")
			assert.Contains(t, claims, "exp", "token missing expiration claim")

			exp, ok := claims["exp"].(float64)
			assert.True(t, ok, "failed to parse expiration claim")
			assert.Greater(t, exp, float64(time.Now().Unix()), "token has already expired")
			assert.Less(t, exp, float64(time.Now().Add(25*time.Hour).Unix()), "token expiration is too far in the future")
		})
	}
}
