package service

import (
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
		user     dto.UserPayload
		expected int
		err      error
	}{
		{
			name: "create user with valid payload",
			user: dto.UserPayload{
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
			user: dto.UserPayload{
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
			user: dto.UserPayload{
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
