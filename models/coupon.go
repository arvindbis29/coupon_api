// models/coupon.go
package models

import (
	"encoding/json"
	"time"
)

type CouponType string

const (
	CartWise    CouponType = "cart-wise"
	ProductWise CouponType = "product-wise"
	BxGy        CouponType = "bxgy"
)

type Coupon struct {
	ID        int             `json:"id"`
	Name      string          `json:"name,omitempty"`
	Type      CouponType      `json:"type"`
	Details   json.RawMessage `json:"details"` // json.RawMessage
	Active    bool            `json:"active"`
	ExpiresAt *time.Time      `json:"expires_at,omitempty"` // RFC3339, optional
}

// Per-type details (used when decoding Details)

type CartWiseDetails struct {
	Threshold int `json:"threshold"`
	Discount  int `json:"discount"` // percent
}

type ProductWiseDetails struct {
	ProductID int `json:"product_id"`
	Discount  int `json:"discount"` // percent
}

type BxGyProduct struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type BxGyDetails struct {
	BuyProducts     []BxGyProduct `json:"buy_products"`
	GetProducts     []BxGyProduct `json:"get_products"`
	RepetitionLimit int           `json:"repition_limit"`
}
