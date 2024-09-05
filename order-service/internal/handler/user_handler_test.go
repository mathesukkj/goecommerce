package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/mathesukkj/goecommerce/order-service/internal/dto"
	"github.com/mathesukkj/goecommerce/order-service/internal/entity"
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
	db.MustExec("ALTER SEQUENCE users_user_id_seq RESTART WITH 1")

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

func TestSignup(t *testing.T) {
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

			req := httptest.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			userHandler.Signup(rr, req)

			assert.Equal(t, tt.want, rr.Code)
		})
	}
}

func TestLogin(t *testing.T) {
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

			req := httptest.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			userHandler.Login(rr, req)

			assert.Equal(t, tt.want, rr.Code)
		})
	}
}

func TestGetLoggedInUser(t *testing.T) {
	userHandler, _ := setupUserHandler(t)
	seedUsers(t)

	tests := []struct {
		name       string
		userID     int
		wantStatus int
		wantUser   *entity.User
	}{
		{
			name:       "success",
			userID:     1,
			wantStatus: http.StatusOK,
			wantUser: &entity.User{
				UserID:      1,
				Username:    "user",
				Email:       "test@example.com",
				FirstName:   "user",
				LastName:    "User",
				PhoneNumber: "1234567890",
			},
		},
		{
			name:       "user not logged in",
			userID:     0,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
			ctx := req.Context()
			ctx = context.WithValue(ctx, keyUserId, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			userHandler.GetLoggedInUser(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantUser != nil {
				var gotUser entity.User
				err := json.NewDecoder(rr.Body).Decode(&gotUser)
				assert.NoError(t, err)
				assert.Equal(t, *tt.wantUser, gotUser)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	userHandler, _ := setupUserHandler(t)
	seedUsers(t)

	tests := []struct {
		name       string
		userID     int
		payload    dto.UpdateUserPayload
		wantStatus int
		wantUser   *entity.User
	}{
		{
			name:   "success",
			userID: 1,
			payload: dto.UpdateUserPayload{
				Username:    "updateduser",
				Email:       "updated@example.com",
				FirstName:   "Updated",
				LastName:    "User",
				PhoneNumber: "9876543210",
			},
			wantStatus: http.StatusOK,
			wantUser: &entity.User{
				UserID:      1,
				Username:    "updateduser",
				Email:       "updated@example.com",
				FirstName:   "Updated",
				LastName:    "User",
				PhoneNumber: "9876543210",
			},
		},
		{
			name:   "user not logged in",
			userID: 0,
			payload: dto.UpdateUserPayload{
				Username: "updateduser",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "invalid payload",
			userID: 1,
			payload: dto.UpdateUserPayload{
				Username: "",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "user not found",
			userID: 999,
			payload: dto.UpdateUserPayload{
				Username:    "updateduser",
				Email:       "updated@example.com",
				FirstName:   "Updated",
				LastName:    "User",
				PhoneNumber: "9876543210",
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")
			if tt.userID != 0 {
				ctx := req.Context()
				ctx = context.WithValue(ctx, keyUserId, tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			userHandler.UpdateUser(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantUser != nil {
				var gotUser entity.User
				err := json.NewDecoder(rr.Body).Decode(&gotUser)
				assert.NoError(t, err)
				assert.Equal(t, *tt.wantUser, gotUser)
			}
		})
	}

}
