// store/memory.go
package store

import (
	"coupon-api/models"
	"sync"
)

var (
	CouponStore = make(map[int]models.Coupon)
	IDCounter   = 1
	Mu          sync.RWMutex
)