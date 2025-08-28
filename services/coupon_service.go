// services/coupon_service.go
package services

import (
	"coupon-api/models"
	"coupon-api/store"
)

func CreateCoupon(c models.Coupon) models.Coupon {
	store.Mu.Lock()
	defer store.Mu.Unlock()
	c.ID = store.IDCounter
	store.IDCounter++
	if !c.Active { // default active
		c.Active = true
	}
	store.CouponStore[c.ID] = c
	return c
}

func GetAllCoupons() []models.Coupon {
	store.Mu.RLock()
	defer store.Mu.RUnlock()
	list := make([]models.Coupon, 0, len(store.CouponStore))
	for _, c := range store.CouponStore {
		list = append(list, c)
	}
	return list
}

func GetCoupon(id int) (models.Coupon, bool) {
	store.Mu.RLock()
	defer store.Mu.RUnlock()
	c, ok := store.CouponStore[id]
	return c, ok
}

func UpdateCoupon(id int, updated models.Coupon) (models.Coupon, bool) {
	store.Mu.Lock()
	defer store.Mu.Unlock()
	if _, ok := store.CouponStore[id]; !ok {
		return models.Coupon{}, false
	}
	updated.ID = id
	store.CouponStore[id] = updated
	return updated, true
}

func DeleteCoupon(id int) bool {
	store.Mu.Lock()
	defer store.Mu.Unlock()
	if _, ok := store.CouponStore[id]; !ok {
		return false
	}
	delete(store.CouponStore, id)
	return true
}
