package dto

import (
	"backend/domain" // [สำคัญ] แก้ไขชื่อ Module ให้ถูกต้อง
	"time"
)

// CouponRequest คือ DTO สำหรับรับข้อมูลตอนสร้างหรืออัปเดต
type CouponRequest struct {
	Code          string              `json:"code" validate:"required,alphanum,uppercase,min=4"`
	DiscountType  domain.DiscountType `json:"discount_type" validate:"required,oneof=fixed percentage"`
	DiscountValue float64             `json:"discount_value" validate:"required,gt=0"`
	ExpiryDate    time.Time           `json:"expiry_date" validate:"required,gt"`
	UsageLimit    uint                `json:"usage_limit" validate:"required,gte=1"`
	IsActive      bool                `json:"is_active"`
}

// CouponResponse คือ DTO สำหรับส่งข้อมูลกลับไป
type CouponResponse struct {
	ID            uint                `json:"id"`
	Code          string              `json:"code"`
	DiscountType  domain.DiscountType `json:"discount_type"`
	DiscountValue float64             `json:"discount_value"`
	ExpiryDate    time.Time           `json:"expiry_date"`
	UsageLimit    uint                `json:"usage_limit"`
	UsageCount    uint                `json:"usage_count"`
	IsActive      bool                `json:"is_active"`
	CreatedAt     time.Time           `json:"created_at"`
}
