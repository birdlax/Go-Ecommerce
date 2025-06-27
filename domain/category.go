// /products/domain/category.go (อาจจะสร้างโฟลเดอร์ domain เพิ่มเพื่อความชัดเจน)
package domain

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Name        string `json:"name" gorm:"type:varchar(100);unique;not null"`
	Description string `json:"description"`
}
