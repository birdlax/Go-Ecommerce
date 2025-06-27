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

type ProductListDTO struct {
	ID              uint    `json:"id"`
	Name            string  `json:"name"`
	Price           float64 `json:"price"`
	SKU             string  `json:"sku"`
	CategoryName    string  `json:"category_name"`
	PrimaryImageURL string  `json:"primary_image_url"`
}

type QueryParams struct {
	Page   int
	Limit  int
	SortBy string
	Order  string
}

// PaginatedProductsDTO คือ DTO สำหรับ Response ที่มีข้อมูลการแบ่งหน้า
type PaginatedProductsDTO struct {
	Data        []ProductListDTO `json:"data"` // ข้อมูลสินค้าในหน้านั้นๆ
	TotalItems  int64            `json:"total_items"`
	TotalPages  int              `json:"total_pages"`
	CurrentPage int              `json:"current_page"`
	Limit       int              `json:"limit"`
}
