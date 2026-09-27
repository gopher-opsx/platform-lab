package memory

import (
	"context"

	"github.com/gopher-opsx/platform-lab/services/catalog-service/internal/domain"
	"github.com/gopher-opsx/platform-lab/services/catalog-service/internal/repository"
)

type ProductRepository struct {
	products []domain.Product
}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{
		products: []domain.Product{
			{
				ID:          "prod-001",
				Name:        "CloudBook Pro 14",
				Description: "14-inch developer laptop with 32 GB RAM and 1 TB SSD.",
				PriceCents:  169900,
				Currency:    "USD",
				ImageURL:    "/images/cloudbook-pro-14.jpg",
				InStock:     true,
			},
			{
				ID:          "prod-002",
				Name:        "CloudPhone X",
				Description: "5G smartphone with 256 GB storage and OLED display.",
				PriceCents:  89900,
				Currency:    "USD",
				ImageURL:    "/images/cloudphone-x.jpg",
				InStock:     true,
			},
			{
				ID:          "prod-003",
				Name:        "CloudPods",
				Description: "Wireless noise-cancelling earbuds with charging case.",
				PriceCents:  19900,
				Currency:    "USD",
				ImageURL:    "/images/cloudpods.jpg",
				InStock:     false,
			},
		},
	}
}

func (r *ProductRepository) List(ctx context.Context) ([]domain.Product, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	products := make([]domain.Product, len(r.products))
	copy(products, r.products)

	return products, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (domain.Product, error) {
	select {
	case <-ctx.Done():
		return domain.Product{}, ctx.Err()
	default:
	}

	for _, product := range r.products {
		if product.ID == id {
			return product, nil
		}
	}

	return domain.Product{}, repository.ErrProductNotFound
}
