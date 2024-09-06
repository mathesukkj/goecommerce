package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"

	"github.com/mathesukkj/goecommerce/order-service/internal/dto"
	"github.com/mathesukkj/goecommerce/order-service/internal/service"
)

type AddressHandler struct {
	service *service.AddressService
}

func NewAddressHandler(db *sqlx.DB) *AddressHandler {
	return &AddressHandler{service: service.NewAddressService(db)}
}

func (h *AddressHandler) ListUserAddresses(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(keyUserId).(int)
	if !ok || userID == 0 {
		http.Error(w, "user not logged in", http.StatusUnauthorized)
		return
	}

	addresses, err := h.service.ListUserAddresses(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(addresses)
}

func (h *AddressHandler) GetAddressByID(w http.ResponseWriter, r *http.Request) {
	addressID, err := strconv.Atoi(r.PathValue("address_id"))
	if err != nil {
		http.Error(w, "invalid address id", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(keyUserId).(int)
	if !ok || userID == 0 {
		http.Error(w, "user not logged in", http.StatusUnauthorized)
		return
	}

	address, err := h.service.GetAddressByID(addressID, userID)
	if err == service.ErrAddressNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(address)
}

func (h *AddressHandler) CreateAddress(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(keyUserId).(int)
	if !ok || userID == 0 {
		http.Error(w, "user not logged in", http.StatusUnauthorized)
		return
	}

	var payload dto.AddressPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	address, err := h.service.CreateAddress(payload, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(address)
}

func (h *AddressHandler) UpdateAddress(w http.ResponseWriter, r *http.Request) {
	addressID, err := strconv.Atoi(r.PathValue("address_id"))
	if err != nil {
		http.Error(w, "invalid address id", http.StatusBadRequest)
		return
	}

	var payload dto.AddressPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	address, err := h.service.UpdateAddress(addressID, payload)
	if err == service.ErrAddressNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(address)
}

func (h *AddressHandler) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	addressID, err := strconv.Atoi(r.PathValue("address_id"))
	if err != nil {
		http.Error(w, "invalid address id", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteAddress(addressID); err == service.ErrAddressNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
