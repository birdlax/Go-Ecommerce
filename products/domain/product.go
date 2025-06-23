// /products/domain/product.go
package domain

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Name        string         `json:"name" gorm:"type:varchar(255);not null"`
	Description string         `json:"description"`
	Price       float64        `json:"price" gorm:"not null"`
	Quantity    int            `json:"quantity"`
	SKU         string         `json:"sku" gorm:"type:varchar(100);unique"`
	Images      []ProductImage `gorm:"foreignKey:ProductID" json:"images"`
	CategoryID  uint           `json:"category_id"`
	Category    Category       `gorm:"foreignKey:CategoryID" json:"category"`
}

// type CreateProductRequest struct {
// 	Name        string
// 	Description string
// 	Price       float64
// 	Quantity    int
// 	SKU         string
// 	CategoryID  uint
// }
