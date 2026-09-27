package domain

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"priceCents"`
	Currency    string `json:"currency"`
	ImageURL    string `json:"imageUrl"`
	InStock     bool   `json:"inStock"`
}
