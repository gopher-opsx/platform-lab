package redisrepo

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/gopher-opsx/platform-lab/services/cart-service/internal/domain"
)

type CartRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCartRepository(client *redis.Client, ttl time.Duration) *CartRepository {
	return &CartRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *CartRepository) Get(ctx context.Context, customerID string) (domain.Cart, error) {
	values, err := r.client.HGetAll(ctx, cartKey(customerID)).Result()
	if err != nil {
		return domain.Cart{}, fmt.Errorf("get cart from redis: %w", err)
	}

	items := make([]domain.CartItem, 0, len(values))
	for productID, rawQuantity := range values {
		quantity, err := strconv.Atoi(rawQuantity)
		if err != nil {
			return domain.Cart{}, fmt.Errorf("decode quantity for product %s: %w", productID, err)
		}

		items = append(items, domain.CartItem{
			ProductID: productID,
			Quantity:  quantity,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ProductID < items[j].ProductID
	})

	return domain.Cart{
		CustomerID: customerID,
		Items:      items,
	}, nil
}

func (r *CartRepository) SetItem(ctx context.Context, customerID string, item domain.CartItem) error {
	key := cartKey(customerID)

	pipe := r.client.TxPipeline()
	pipe.HSet(ctx, key, item.ProductID, item.Quantity)
	pipe.Expire(ctx, key, r.ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("set cart item in redis: %w", err)
	}

	return nil
}

func (r *CartRepository) RemoveItem(ctx context.Context, customerID, productID string) error {
	if err := r.client.HDel(ctx, cartKey(customerID), productID).Err(); err != nil {
		return fmt.Errorf("remove cart item from redis: %w", err)
	}
	return nil
}

func (r *CartRepository) Clear(ctx context.Context, customerID string) error {
	if err := r.client.Del(ctx, cartKey(customerID)).Err(); err != nil {
		return fmt.Errorf("clear cart in redis: %w", err)
	}
	return nil
}

func cartKey(customerID string) string {
	return "cart:" + customerID
}
