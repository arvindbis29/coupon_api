package main

import (
	"coupon-api/handlers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	// Coupon CRUD
	r.HandleFunc("/coupons", handlers.CreateCoupon).Methods("POST")
	r.HandleFunc("/coupons", handlers.GetCoupons).Methods("GET")
	r.HandleFunc("/coupons/{id}", handlers.GetCouponByID).Methods("GET")
	r.HandleFunc("/coupons/{id}", handlers.UpdateCoupon).Methods("PUT")
	r.HandleFunc("/coupons/{id}", handlers.DeleteCoupon).Methods("DELETE")

	// Application endpoints
	r.HandleFunc("/applicable-coupons", handlers.ApplicableCoupons).Methods("POST")
	r.HandleFunc("/apply-coupon/{id}", handlers.ApplyCoupon).Methods("POST")

	log.Println("Server running on :8080 (in-memory, data resets on restart)")
	log.Fatal(http.ListenAndServe(":8080", r))
}
