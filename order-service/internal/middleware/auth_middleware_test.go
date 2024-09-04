package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddlewareToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "secret")
	defer os.Unsetenv("JWT_SECRET")

	tests := []struct {
		name        string
		setupAuth   func(r *http.Request)
		checkResult func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "valid token",
			setupAuth: func(r *http.Request) {
				token := generateTestToken(1)
				r.Header.Set("Authorization", "Bearer "+token)
			},
			checkResult: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "no authorization header",
			setupAuth: func(r *http.Request) {
			},
			checkResult: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
				assert.Contains(t, recorder.Body.String(), "user not logged in")
			},
		},
		{
			name: "invalid token format",
			setupAuth: func(r *http.Request) {
				r.Header.Set("Authorization", "InvalidFormat")
			},
			checkResult: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
				assert.Contains(t, recorder.Body.String(), "invalid token")
			},
		},
		{
			name: "invalid token",
			setupAuth: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer invalidtoken")
			},
			checkResult: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
				assert.Contains(t, recorder.Body.String(), "invalid token")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			tt.setupAuth(req)

			recorder := httptest.NewRecorder()
			handler := JwtUserId(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID := r.Context().Value(jwtUserIdKey)
				assert.NotNil(t, userID)
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(recorder, req)
			tt.checkResult(t, recorder)
		})
	}
}

func generateTestToken(userId int) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return ""
	}
	return tokenStr
}
