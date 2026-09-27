package httptransport

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gopher-opsx/platform-lab/services/catalog-service/internal/repository"
	"github.com/gopher-opsx/platform-lab/services/catalog-service/internal/service"
)

type CatalogHandler struct {
	catalog *service.CatalogService
}

func NewCatalogHandler(catalog *service.CatalogService) *CatalogHandler {
	return &CatalogHandler{
		catalog: catalog,
	}
}

func (h *CatalogHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.catalog.ListProducts(r.Context())
	if err != nil {
		http.Error(w, "failed to list products", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, products)
}

func (h *CatalogHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	product, err := h.catalog.GetProduct(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to get product", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, product)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
