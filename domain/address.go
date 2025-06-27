// domain/address.go
package domain

import "gorm.io/gorm"

type Address struct {
	gorm.Model
	UserID       uint   `gorm:"not null"`
	AddressLine1 string `json:"address_line_1" gorm:"not null"`
	AddressLine2 string `json:"address_line_2"`
	City         string `json:"city" gorm:"not null"`
	State        string `json:"state" gorm:"not null"`
	PostalCode   string `json:"postal_code" gorm:"not null"`
	Country      string `json:"country" gorm:"not null"`
	IsDefault    bool   `json:"is_default" gorm:"default:false"`
}
