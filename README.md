
// README.md
# Monk Commerce — Coupons Management API (Go, In-Memory)

**Tech:** Go 1.22, Gorilla Mux, in-memory store (data resets on server restart).

## How to run
```bash
go mod tidy
go run .
```

## Endpoints
- `POST /coupons` — create coupon (body follows examples below)
- `GET /coupons` — list coupons
- `GET /coupons/{id}` — get by id
- `PUT /coupons/{id}` — update coupon
- `DELETE /coupons/{id}` — delete coupon
- `POST /applicable-coupons` — preview discounts for given cart
- `POST /apply-coupon/{id}` — apply coupon and return updated cart with per-item `total_discount`, plus totals

## Coupon JSON
Common fields:
```json
{
  "name": "Optional name",
  "type": "cart-wise | product-wise | bxgy",
  "details": { /* type-specific */ },
  "active": true,
  "expires_at": "2025-12-31T23:59:59Z" // optional RFC3339
}
```

### Cart-wise
```json
{
  "type": "cart-wise",
  "details": { "threshold": 100, "discount": 10 }
}
```

### Product-wise
```json
{
  "type": "product-wise",
  "details": { "product_id": 1, "discount": 20 }
}
```

### BxGy (Buy X Get Y)
```json
{
  "type": "bxgy",
  "details": {
    "buy_products": [ {"product_id": 1, "quantity": 3}, {"product_id": 2, "quantity": 3} ],
    "get_products": [ {"product_id": 3, "quantity": 1} ],
    "repition_limit": 2
  }
}
```

## Request/Response Examples

### `POST /applicable-coupons`
Request:
```json
{
  "cart": {
    "items": [
      {"product_id": 1, "quantity": 6, "price": 50},
      {"product_id": 2, "quantity": 3, "price": 30},
      {"product_id": 3, "quantity": 2, "price": 25}
    ]
  }
}
```
Response (example):
```json
{
  "applicable_coupons": [
    {"coupon_id": 1, "type": "cart-wise", "discount": 40},
    {"coupon_id": 3, "type": "bxgy", "discount": 50}
  ]
}
```

### `POST /apply-coupon/{id}`
Response (example for bxgy as in doc):
```json
{
  "updated_cart": {
    "items": [
      {"product_id": 1, "quantity": 6, "price": 50, "total_discount": 0},
      {"product_id": 2, "quantity": 3, "price": 30, "total_discount": 0},
      {"product_id": 3, "quantity": 4, "price": 25, "total_discount": 50}
    ]
  },
  "total_price": 490,
  "total_discount": 50,
  "final_price": 440
}
```

## Assumptions & Limitations
- **In-memory only**: data resets on server restart (per your requirement).
- **Prices are integers** (e.g., rupees). No floating point rounding issues.
- **BxGy interpretation**:
  - *Assumption*: Per repetition, required buys = sum of `quantity` across `buy_products`; free units = sum of `quantity` across `get_products`.
  - Eligible **buy** items can be any combination from the `buy_products` list.
  - Free items are granted **only for products already present in the cart** (we need their price).
  - We select the **cheapest** eligible get-items first to compute discount (merchant-friendly). This is documented and easy to flip to most-expensive if needed.
  - If get-products are not in the cart, no free line is added (lack of price).
- **Cart-wise distribution**: discount is distributed proportionally across items; rounding remainder goes to the line with the largest pre-discount value.
- **Expiration**: if `expires_at` is present and in the past, coupon is ignored/invalid.
- **Validation**: minimal shape validation; for production, add stronger schema checks.

## Possible Extensions (Unimplemented)
- Multiple product-wise products in one coupon.
- Stackable coupons and best-combination selection.
- Configurable policy for BxGy free item selection (cheapest vs most-expensive).
- Currency minor units and tax handling.
- Authentication, rate limits, logging middlewares.

## Quick Test Ideas
- Cart-wise: threshold boundary, zero discount, distribution rounding.
- Product-wise: product absent/present, multiple lines of same product.
- BxGy: limiting factor = buys / gets / repetition_limit; missing get item; price zero.
