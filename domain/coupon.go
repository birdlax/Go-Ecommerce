package domain

import (
	"gorm.io/gorm"
	"time"
)

type DiscountType string

const (
	DiscountTypeFixed      DiscountType = "fixed"
	DiscountTypePercentage DiscountType = "percentage"
)

type Coupon struct {
	gorm.Model
	Code          string       `gorm:"type:varchar(50);uniqueIndex;not null"`
	DiscountType  DiscountType `gorm:"type:varchar(20);not null"`
	DiscountValue float64      `gorm:"not null"`
	ExpiryDate    time.Time    `gorm:"not null"`
	UsageLimit    uint         `gorm:"not null;default:1"`
	UsageCount    uint         `gorm:"not null;default:0"`
	IsActive      bool         `gorm:"not null;default:true"`
}
