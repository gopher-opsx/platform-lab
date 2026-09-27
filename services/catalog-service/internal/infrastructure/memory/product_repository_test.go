package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/gopher-opsx/platform-lab/services/catalog-service/internal/repository"
)

func TestProductRepositoryList(t *testing.T) {
	repo := NewProductRepository()

	products, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}

	if len(products) != 3 {
		t.Fatalf("expected 3 products, got %d", len(products))
	}
}

func TestProductRepositoryGetByID(t *testing.T) {
	repo := NewProductRepository()

	product, err := repo.GetByID(context.Background(), "prod-001")
	if err != nil {
		t.Fatalf("GetByID() returned error: %v", err)
	}

	if product.ID != "prod-001" {
		t.Fatalf("expected prod-001, got %s", product.ID)
	}

	if product.Name != "CloudBook Pro 14" {
		t.Fatalf("unexpected product name: %s", product.Name)
	}
}

func TestProductRepositoryGetByIDNotFound(t *testing.T) {
	repo := NewProductRepository()

	_, err := repo.GetByID(context.Background(), "does-not-exist")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if !errors.Is(err, repository.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}
