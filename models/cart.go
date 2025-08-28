// models/cart.go
package models

type CartItem struct {
	ProductID     int `json:"product_id"`
	Quantity      int `json:"quantity"`
	Price         int `json:"price"` // integer currency (e.g., rupees)
	TotalDiscount int `json:"total_discount,omitempty"`
}

type Cart struct {
	Items []CartItem `json:"items"`
}
