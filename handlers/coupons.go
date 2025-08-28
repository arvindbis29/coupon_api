// handlers/coupons.go
package handlers

import (
	"coupon-api/models"
	"coupon-api/services"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func CreateCoupon(w http.ResponseWriter, r *http.Request) {
	var c models.Coupon
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	created := services.CreateCoupon(c)
	writeJSON(w, http.StatusOK, created)
}

func GetCoupons(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, services.GetAllCoupons())
}

func GetCouponByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	c, ok := services.GetCoupon(id)
	if !ok {
		http.Error(w, "coupon not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func UpdateCoupon(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var c models.Coupon
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	updated, ok := services.UpdateCoupon(id, c)
	if !ok {
		http.Error(w, "coupon not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func DeleteCoupon(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if !services.DeleteCoupon(id) {
		http.Error(w, "coupon not found", http.StatusNotFound)
		return
	}
	fmt.Println("Deletion success for the id:", id)
	w.WriteHeader(http.StatusNoContent)
}
