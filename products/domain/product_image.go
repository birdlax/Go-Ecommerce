// /products/domain/product_image.go
package domain

import "gorm.io/gorm"

type ProductImage struct {
	gorm.Model
	ProductID uint   `json:"-"` // ไม่ต้องส่ง ProductID กลับไปใน JSON ก็ได้ เพราะมันซ้อนอยู่ใน Product อยู่แล้ว
	Path      string `json:"path" gorm:"type:varchar(255);not null"`
	IsPrimary bool   `json:"is_primary" gorm:"default:false"`
	// เราสามารถสร้าง URL แบบเต็มได้ตอนส่งข้อมูลกลับ โดยไม่เก็บลง DB
	URL string `json:"url" gorm:"-"` // gorm:"-" หมายถึงไม่ต้องสร้างคอลัมน์นี้ใน DB
}
