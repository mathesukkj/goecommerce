package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/mathesukkj/goecommerce/order-service/internal/dto"
	"github.com/mathesukkj/goecommerce/order-service/internal/service"
)

type PaymentMethodHandler struct {
	service *service.PaymentMethodService
}

func NewPaymentMethodHandler(db *sqlx.DB) *PaymentMethodHandler {
	return &PaymentMethodHandler{service: service.NewPaymentMethodService(db)}
}

func (h *PaymentMethodHandler) ListUserPaymentMethods(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(keyUserId).(int)
	if !ok || userID == 0 {
		http.Error(w, "user not logged in", http.StatusUnauthorized)
		return
	}

	paymentMethods, err := h.service.ListUserPaymentMethods(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(paymentMethods)
}

func (h *PaymentMethodHandler) GetPaymentMethodByID(w http.ResponseWriter, r *http.Request) {
	paymentMethodID, err := strconv.Atoi(r.PathValue("payment_method_id"))
	if err != nil {
		http.Error(w, "invalid paymentMethod id", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(keyUserId).(int)
	if !ok || userID == 0 {
		http.Error(w, "user not logged in", http.StatusUnauthorized)
		return
	}

	paymentMethod, err := h.service.GetPaymentMethodByID(paymentMethodID, userID)
	if err == service.ErrPaymentMethodNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(paymentMethod)
}

func (h *PaymentMethodHandler) CreatePaymentMethod(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(keyUserId).(int)
	if !ok || userID == 0 {
		http.Error(w, "user not logged in", http.StatusUnauthorized)
		return
	}

	var payload dto.PaymentMethodPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	paymentMethod, err := h.service.CreatePaymentMethod(payload, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(paymentMethod)
}

func (h *PaymentMethodHandler) UpdatePaymentMethod(w http.ResponseWriter, r *http.Request) {
	paymentMethodID, err := strconv.Atoi(r.PathValue("payment_method_id"))
	if err != nil {
		http.Error(w, "invalid paymentMethod id", http.StatusBadRequest)
		return
	}

	var payload dto.PaymentMethodPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	paymentMethod, err := h.service.UpdatePaymentMethod(paymentMethodID, payload)
	if err == service.ErrPaymentMethodNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(paymentMethod)
}

func (h *PaymentMethodHandler) DeletePaymentMethod(w http.ResponseWriter, r *http.Request) {
	paymentMethodID, err := strconv.Atoi(r.PathValue("payment_method_id"))
	if err != nil {
		http.Error(w, "invalid paymentMethod id", http.StatusBadRequest)
		return
	}

	if err := h.service.DeletePaymentMethod(paymentMethodID); err == service.ErrPaymentMethodNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
