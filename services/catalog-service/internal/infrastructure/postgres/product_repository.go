package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gopher-opsx/platform-lab/services/catalog-service/internal/domain"
	"github.com/gopher-opsx/platform-lab/services/catalog-service/internal/repository"
)

type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{
		pool: pool,
	}
}

func (r *ProductRepository) List(ctx context.Context) ([]domain.Product, error) {
	const query = `
		SELECT
			id,
			name,
			description,
			price_cents,
			currency,
			image_url,
			in_stock
		FROM products
		ORDER BY id
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	products := make([]domain.Product, 0)

	for rows.Next() {
		var product domain.Product

		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.PriceCents,
			&product.Currency,
			&product.ImageURL,
			&product.InStock,
		); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate products: %w", err)
	}

	return products, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (domain.Product, error) {
	const query = `
		SELECT
			id,
			name,
			description,
			price_cents,
			currency,
			image_url,
			in_stock
		FROM products
		WHERE id = $1
	`

	var product domain.Product

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.PriceCents,
		&product.Currency,
		&product.ImageURL,
		&product.InStock,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Product{}, repository.ErrProductNotFound
		}

		return domain.Product{}, fmt.Errorf("query product by id: %w", err)
	}

	return product, nil
}
