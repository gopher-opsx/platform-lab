package httptransport

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gopher-opsx/platform-lab/services/order-service/internal/domain"
	"github.com/gopher-opsx/platform-lab/services/order-service/internal/repository"
	"github.com/gopher-opsx/platform-lab/services/order-service/internal/service"
)

const customerIDHeader = "X-Customer-ID"

type createOrderRequest struct {
	Currency string             `json:"currency"`
	Items    []domain.OrderItem `json:"items"`
}

type OrderHandler struct {
	orders *service.OrderService
}

func NewOrderHandler(orders *service.OrderService) *OrderHandler {
	return &OrderHandler{orders: orders}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	customerID, ok := customerIDFromRequest(w, r)
	if !ok {
		return
	}

	var request createOrderRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	order, err := h.orders.CreateOrder(r.Context(), service.CreateOrderInput{
		CustomerID: customerID,
		Currency:   request.Currency,
		Items:      request.Items,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	order, err := h.orders.GetOrder(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get order")
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	customerID, ok := customerIDFromRequest(w, r)
	if !ok {
		return
	}

	orders, err := h.orders.ListOrders(r.Context(), customerID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, orders)
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
	switch {
	case errors.Is(err, service.ErrCustomerRequired),
		errors.Is(err, service.ErrItemsRequired),
		errors.Is(err, service.ErrProductRequired),
		errors.Is(err, service.ErrInvalidQuantity),
		errors.Is(err, service.ErrInvalidPrice),
		errors.Is(err, service.ErrCurrencyRequired):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "order operation failed")
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
