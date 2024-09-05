package handler

import (
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/mathesukkj/goecommerce/order-service/internal/dto"
	"github.com/mathesukkj/goecommerce/order-service/internal/service"
)

type key string

const (
	keyUserId key = "user_id"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(db *sqlx.DB) *UserHandler {
	return &UserHandler{service: service.NewUserService(db)}
}

func (h *UserHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var body dto.SignupPayload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := body.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.service.Signup(body)
	switch err {
	case service.ErrUserAlreadyExists:
		http.Error(w, err.Error(), http.StatusConflict)
		return
	case service.ErrPasswordTooLong:
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	case nil:
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := dto.LoginResponse{
		Token: token,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body dto.LoginPayload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := body.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.service.Login(body)
	switch err {
	case service.ErrUserNotFound:
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	case service.ErrInvalidPassword:
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	case nil:
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := dto.LoginResponse{
		Token: token,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) GetLoggedInUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(keyUserId).(int)
	if !ok {
		http.Error(w, "user not logged in", http.StatusUnauthorized)
		return
	}

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(keyUserId).(int)
	if !ok {
		http.Error(w, "user not logged in", http.StatusUnauthorized)
		return
	}

	var body dto.UpdateUserPayload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := body.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedUser, err := h.service.UpdateUser(userID, body)
	if err == service.ErrUserNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updatedUser)
}
