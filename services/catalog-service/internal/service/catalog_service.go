package service

import (
	"context"

	"github.com/gopher-opsx/platform-lab/services/catalog-service/internal/domain"
	"github.com/gopher-opsx/platform-lab/services/catalog-service/internal/repository"
)

type CatalogService struct {
	products repository.ProductRepository
}

func NewCatalogService(products repository.ProductRepository) *CatalogService {
	return &CatalogService{
		products: products,
	}
}

func (s *CatalogService) ListProducts(ctx context.Context) ([]domain.Product, error) {
	return s.products.List(ctx)
}

func (s *CatalogService) GetProduct(ctx context.Context, id string) (domain.Product, error) {
	return s.products.GetByID(ctx, id)
}
