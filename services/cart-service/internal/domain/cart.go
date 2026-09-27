package domain

type CartItem struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

type Cart struct {
	CustomerID string     `json:"customerId"`
	Items      []CartItem `json:"items"`
}
