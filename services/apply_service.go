
// services/apply_service.go
package services

import (
	"coupon-api/models"
	"encoding/json"
	"sort"
	"time"
)

// ApplicableCoupon holds preview discount per coupon
 type ApplicableCoupon struct {
	CouponID int                `json:"coupon_id"`
	Type     models.CouponType  `json:"type"`
	Discount int                `json:"discount"`
}

func CalculateApplicableCoupons(cart models.Cart, now time.Time) []ApplicableCoupon {
	coupons := GetAllCoupons()
	var result []ApplicableCoupon

	for _, c := range coupons {
		if !c.Active { continue }
		if c.ExpiresAt != nil && c.ExpiresAt.Before(now) { continue }

		switch c.Type {
		case models.CartWise:
			var d models.CartWiseDetails
			if json.Unmarshal(c.Details, &d) == nil {
				total := totalCart(cart)
				if total > d.Threshold && d.Discount > 0 {
					disc := total * d.Discount / 100
					result = append(result, ApplicableCoupon{c.ID, c.Type, disc})
				}
			}
		case models.ProductWise:
			var d models.ProductWiseDetails
			if json.Unmarshal(c.Details, &d) == nil {
				disc := 0
				for _, it := range cart.Items {
					if it.ProductID == d.ProductID && d.Discount > 0 {
						disc += it.Price * it.Quantity * d.Discount / 100
					}
				}
				if disc > 0 { result = append(result, ApplicableCoupon{c.ID, c.Type, disc}) }
			}
		case models.BxGy:
			if disc := bxgyPreviewDiscount(cart, c); disc > 0 {
				result = append(result, ApplicableCoupon{c.ID, c.Type, disc})
			}
		}
	}
	return result
}

func totalCart(cart models.Cart) int {
	sum := 0
	for _, it := range cart.Items { sum += it.Price * it.Quantity }
	return sum
}

// Assumption documented in README:
// - For BxGy, required buys per repetition = sum of quantities in BuyProducts (across the listed product IDs).
// - Free items per repetition = sum of quantities in GetProducts.
// - We only grant free items for products that are present in the cart (we need their price). No new items are added if not present.
// - To be merchant-friendly we select the CHEAPEST eligible get-items first when computing free value.

func bxgyPreviewDiscount(cart models.Cart, c models.Coupon) int {
	var d models.BxGyDetails
	if json.Unmarshal(c.Details, &d) != nil { return 0 }
	if len(d.BuyProducts) == 0 || len(d.GetProducts) == 0 { return 0 }

	buyRequired := 0
	for _, bp := range d.BuyProducts { buyRequired += max(0, bp.Quantity) }
	getPerRep := 0
	for _, gp := range d.GetProducts { getPerRep += max(0, gp.Quantity) }
	if buyRequired == 0 || getPerRep == 0 { return 0 }

	// counts in cart
	buySet := make(map[int]struct{})
	for _, bp := range d.BuyProducts { buySet[bp.ProductID] = struct{}{} }
	getSet := make(map[int]struct{})
	for _, gp := range d.GetProducts { getSet[gp.ProductID] = struct{}{} }

	totalBuy := 0
	getItems := make([]models.CartItem, 0)
	for _, it := range cart.Items {
		if _, ok := buySet[it.ProductID]; ok { totalBuy += it.Quantity }
		if _, ok := getSet[it.ProductID]; ok && it.Quantity > 0 && it.Price > 0 {
			getItems = append(getItems, it)
		}
	}
	if totalBuy < buyRequired { return 0 }

	maxRepsByBuy := totalBuy / buyRequired

	// total get availability in cart
	totalGetAvail := 0
	for _, gi := range getItems { totalGetAvail += gi.Quantity }
	if totalGetAvail == 0 { return 0 }
	maxRepsByGet := totalGetAvail / getPerRep

	reps := min(maxRepsByBuy, maxRepsByGet)
	if d.RepetitionLimit > 0 { reps = min(reps, d.RepetitionLimit) }
	if reps <= 0 { return 0 }

	// choose cheapest get items first for discount preview
	sort.Slice(getItems, func(i, j int) bool { return getItems[i].Price < getItems[j].Price })
	freeUnits := reps * getPerRep
	discount := 0
	for _, gi := range getItems {
		if freeUnits == 0 { break }
		use := min(gi.Quantity, freeUnits)
		discount += use * gi.Price
		freeUnits -= use
	}
	return discount
}

func max(a, b int) int { if a > b { return a }; return b }
func min(a, b int) int { if a < b { return a }; return b }

// ApplyCoupon mutates a copy of cart and returns updated cart + totals

type ApplyResult struct {
	UpdatedCart  models.Cart `json:"updated_cart"`
	TotalPrice   int         `json:"total_price"`
	TotalDiscount int        `json:"total_discount"`
	FinalPrice   int         `json:"final_price"`
}

func ApplyCouponToCart(cart models.Cart, c models.Coupon, now time.Time) (ApplyResult, bool, string) {
	if !c.Active { return ApplyResult{}, false, "coupon inactive" }
	if c.ExpiresAt != nil && c.ExpiresAt.Before(now) { return ApplyResult{}, false, "coupon expired" }

	switch c.Type {
	case models.CartWise:
		var d models.CartWiseDetails
		if json.Unmarshal(c.Details, &d) != nil { return ApplyResult{}, false, "invalid details" }
		return applyCartWise(cart, d)
	case models.ProductWise:
		var d models.ProductWiseDetails
		if json.Unmarshal(c.Details, &d) != nil { return ApplyResult{}, false, "invalid details" }
		return applyProductWise(cart, d)
	case models.BxGy:
		var d models.BxGyDetails
		if json.Unmarshal(c.Details, &d) != nil { return ApplyResult{}, false, "invalid details" }
		return applyBxGy(cart, d)
	default:
		return ApplyResult{}, false, "unsupported type"
	}
}

func applyCartWise(cart models.Cart, d models.CartWiseDetails) (ApplyResult, bool, string) {
	total := totalCart(cart)
	if total <= d.Threshold || d.Discount <= 0 {
		return ApplyResult{}, false, "conditions not met"
	}
	disc := total * d.Discount / 100

	// distribute proportionally (rounded), remainder to most expensive item
	updated := cart
	// find most expensive line
	maxIdx := 0
	maxVal := 0
	for i, it := range updated.Items {
		line := it.Price * it.Quantity
		part := 0
		if total > 0 { part = line * disc / total }
		updated.Items[i].TotalDiscount = part
		if line > maxVal { maxVal, maxIdx = line, i }
	}
	// handle rounding remainder
	sumParts := 0
	for _, it := range updated.Items { sumParts += it.TotalDiscount }
	if sumParts < disc {
		updated.Items[maxIdx].TotalDiscount += (disc - sumParts)
	}

	return ApplyResult{UpdatedCart: updated, TotalPrice: total, TotalDiscount: disc, FinalPrice: total - disc}, true, ""
}

func applyProductWise(cart models.Cart, d models.ProductWiseDetails) (ApplyResult, bool, string) {
	updated := cart
	total := totalCart(cart)
	disc := 0
	for i, it := range updated.Items {
		if it.ProductID == d.ProductID && d.Discount > 0 {
			lineDisc := it.Price * it.Quantity * d.Discount / 100
			disc += lineDisc
			updated.Items[i].TotalDiscount = lineDisc
		}
	}
	if disc == 0 {
		return ApplyResult{}, false, "conditions not met"
	}
	return ApplyResult{UpdatedCart: updated, TotalPrice: total, TotalDiscount: disc, FinalPrice: total - disc}, true, ""
}

func applyBxGy(cart models.Cart, d models.BxGyDetails) (ApplyResult, bool, string) {
	if len(d.BuyProducts) == 0 || len(d.GetProducts) == 0 { return ApplyResult{}, false, "invalid details" }
	buyRequired := 0
	for _, bp := range d.BuyProducts { buyRequired += max(0, bp.Quantity) }
	getPerRep := 0
	for _, gp := range d.GetProducts { getPerRep += max(0, gp.Quantity) }
	if buyRequired == 0 || getPerRep == 0 { return ApplyResult{}, false, "invalid details" }

	buySet := make(map[int]struct{})
	getSet := make(map[int]struct{})
	for _, bp := range d.BuyProducts { buySet[bp.ProductID] = struct{}{} }
	for _, gp := range d.GetProducts { getSet[gp.ProductID] = struct{}{} }

	updated := cart
	total := totalCart(cart)

	// compute reps possible
	totalBuy := 0
	for _, it := range updated.Items { if _, ok := buySet[it.ProductID]; ok { totalBuy += it.Quantity } }
	if totalBuy < buyRequired { return ApplyResult{}, false, "conditions not met" }
	maxRepsByBuy := totalBuy / buyRequired

	totalGetAvail := 0
	for _, it := range updated.Items { if _, ok := getSet[it.ProductID]; ok { totalGetAvail += it.Quantity } }
	if totalGetAvail == 0 { return ApplyResult{}, false, "conditions not met" }
	maxRepsByGet := totalGetAvail / getPerRep

	reps := min(maxRepsByBuy, maxRepsByGet)
	if d.RepetitionLimit > 0 { reps = min(reps, d.RepetitionLimit) }
	if reps <= 0 { return ApplyResult{}, false, "conditions not met" }

	freeUnits := reps * getPerRep

	// Choose cheapest get items first and mark their discount; also increase quantity by free units (as per sample response)
	// Build a slice of pointers to items in updated cart that are in getSet
	indices := make([]int, 0)
	for i, it := range updated.Items { if _, ok := getSet[it.ProductID]; ok { indices = append(indices, i) } }
	sort.Slice(indices, func(i, j int) bool {
		return updated.Items[indices[i]].Price < updated.Items[indices[j]].Price
	})

	disc := 0
	for _, idx := range indices {
		if freeUnits == 0 { break }
		it := &updated.Items[idx]
		use := min(it.Quantity, freeUnits) // we only free up to existing quantity per item line
		if use > 0 {
			disc += use * it.Price
			it.TotalDiscount += use * it.Price
			it.Quantity += use // add free pieces to quantity
			freeUnits -= use
		}
	}
	if disc == 0 { return ApplyResult{}, false, "conditions not met" }

	return ApplyResult{UpdatedCart: updated, TotalPrice: total, TotalDiscount: disc, FinalPrice: total - disc}, true, ""
}
