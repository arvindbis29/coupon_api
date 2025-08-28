
// handlers/apply.go
package handlers

import (
	"coupon-api/models"
	"coupon-api/services"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func ApplicableCoupons(w http.ResponseWriter, r *http.Request) {
	var body struct { Cart models.Cart `json:"cart"` }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	apps := services.CalculateApplicableCoupons(body.Cart, time.Now())
	writeJSON(w, http.StatusOK, map[string]any{"applicable_coupons": apps})
}

func ApplyCoupon(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil { http.Error(w, "invalid id", http.StatusBadRequest); return }
	var body struct { Cart models.Cart `json:"cart"` }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	coupon, ok := services.GetCoupon(id)
	if !ok { http.Error(w, "coupon not found", http.StatusNotFound); return }
	res, applied, reason := services.ApplyCouponToCart(body.Cart, coupon, time.Now())
	if !applied {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": reason})
		return
	}
	writeJSON(w, http.StatusOK, res)
}
