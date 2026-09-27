package httptransport

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gopher-opsx/platform-lab/services/cart-service/internal/repository"
	"github.com/gopher-opsx/platform-lab/services/cart-service/internal/service"
)

const customerIDHeader = "X-Customer-ID"

type CartHandler struct {
	cart *service.CartService
}

type itemRequest struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

type quantityRequest struct {
	Quantity int `json:"quantity"`
}

func NewCartHandler(cart *service.CartService) *CartHandler {
	return &CartHandler{cart: cart}
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	customerID, ok := customerIDFromRequest(w, r)
	if !ok {
		return
	}

	cart, err := h.cart.GetCart(r.Context(), customerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get cart")
		return
	}

	writeJSON(w, http.StatusOK, cart)
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	customerID, ok := customerIDFromRequest(w, r)
	if !ok {
		return
	}

	var request itemRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(request.ProductID) == "" {
		writeError(w, http.StatusBadRequest, "productId is required")
		return
	}

	if err := h.cart.AddItem(r.Context(), customerID, request.ProductID, request.Quantity); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	customerID, ok := customerIDFromRequest(w, r)
	if !ok {
		return
	}

	var request quantityRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.cart.UpdateItem(r.Context(), customerID, r.PathValue("productId"), request.Quantity); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	customerID, ok := customerIDFromRequest(w, r)
	if !ok {
		return
	}

	if err := h.cart.RemoveItem(r.Context(), customerID, r.PathValue("productId")); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to remove cart item")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	customerID, ok := customerIDFromRequest(w, r)
	if !ok {
		return
	}

	if err := h.cart.ClearCart(r.Context(), customerID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to clear cart")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func customerIDFromRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	customerID := strings.TrimSpace(r.Header.Get(customerIDHeader))
	if customerID == "" {
		writeError(w, http.StatusBadRequest, customerIDHeader+" header is required")
		return "", false
	}
	return customerID, true
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, repository.ErrInvalidQuantity) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeError(w, http.StatusInternalServerError, "cart operation failed")
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
