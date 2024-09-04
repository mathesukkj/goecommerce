package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/mathesukkj/goecommerce/order-service/internal/dto"
	"github.com/mathesukkj/goecommerce/order-service/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var db *sqlx.DB
var pgContainer *postgres.PostgresContainer

func TestMain(m *testing.M) {
	pgContainer, db = testutils.NewPostgresContainerDB()

	os.Exit(m.Run())
}

func setupUserHandler(t *testing.T) (*UserHandler, *postgres.PostgresContainer) {
	t.Helper()

	db.MustExec("TRUNCATE TABLE users CASCADE")

	userHandler := NewUserHandler(db)
	return userHandler, pgContainer
}

func seedUsers(t *testing.T) {
	t.Helper()

	params := map[string]interface{}{
		"username":     "user",
		"password":     "$2y$10$bsRLuOQN606nDdkFCF2D4eF74rON7JXEP.RxTAKbgTft2BgqtJgYu",
		"email":        "test@example.com",
		"first_name":   "user",
		"last_name":    "User",
		"phone_number": "1234567890",
	}

	_, err := db.NamedExec(`
		INSERT INTO users (username, password, email, first_name, last_name, phone_number) 
		VALUES (:username, :password, :email, :first_name, :last_name, :phone_number)
	`, params)
	if err != nil {
		t.Fatalf("failed to seed users: %s", err)
	}
}

func TestUserHandler_Signup(t *testing.T) {
	userHandler, _ := setupUserHandler(t)
	seedUsers(t)

	tests := []struct {
		name    string
		payload dto.SignupPayload
		want    int
	}{
		{
			name: "success",
			payload: dto.SignupPayload{
				Username:    "user2",
				Password:    "password",
				Email:       "test2@example.com",
				FirstName:   "user2",
				LastName:    "User2",
				PhoneNumber: "1234567890",
			},
			want: http.StatusOK,
		},
		{
			name: "invalid payload",
			payload: dto.SignupPayload{
				Username: "",
				Password: "password",
				Email:    "test@example.com",
			},
			want: http.StatusBadRequest,
		},
		{
			name: "user with email already exists",
			payload: dto.SignupPayload{
				Username:    "user2",
				Password:    "password",
				Email:       "test@example.com",
				FirstName:   "user",
				LastName:    "User",
				PhoneNumber: "1234567890",
			},
			want: http.StatusConflict,
		},
		{
			name: "user with username already exists",
			payload: dto.SignupPayload{
				Username:    "user",
				Password:    "password",
				Email:       "test2@example.com",
				FirstName:   "user",
				LastName:    "User",
				PhoneNumber: "1234567890",
			},
			want: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.payload)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			userHandler.Signup(rr, req)

			assert.Equal(t, tt.want, rr.Code)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	userHandler, _ := setupUserHandler(t)
	seedUsers(t)

	tests := []struct {
		name    string
		payload dto.LoginPayload
		want    int
	}{
		{
			name: "success",
			payload: dto.LoginPayload{
				Email:    "test@example.com",
				Password: "password",
			},
			want: http.StatusOK,
		},
		{
			name: "invalid password",
			payload: dto.LoginPayload{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			want: http.StatusUnauthorized,
		},
		{
			name: "invalid email",
			payload: dto.LoginPayload{
				Email:    "invalid@example.com",
				Password: "password",
			},
			want: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.payload)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			userHandler.Login(rr, req)

			assert.Equal(t, tt.want, rr.Code)
		})
	}
}
