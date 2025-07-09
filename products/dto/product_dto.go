package dto

import (
	"io"
	"time"
)

// ===================================================================
// DTOs for Requests (ข้อมูลที่ Client ส่งเข้ามา)
// ===================================================================

// FileInput คือ struct กลางสำหรับข้อมูลไฟล์ที่ Handler เตรียมให้ Service
// ทำให้ Service ไม่ต้องรู้จัก `multipart.FileHeader` เลย
type FileInput struct {
	Content     io.Reader
	Filename    string
	ContentType string
}
type QueryParams struct {
	Page   int
	Limit  int
	SortBy string
	Order  string
	// --- ฟิลด์ใหม่ ---
	Search     string
	CategoryID uint
	MinPrice   float64
	MaxPrice   float64
}

// CreateProductRequestData คือ Struct สำหรับข้อมูลสินค้าที่เป็น JSON
type CreateProductRequestData struct {
	Name        string  `json:"name" validate:"required,min=3"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Quantity    int     `json:"quantity" validate:"required,gte=0"`
	SKU         string  `json:"sku" validate:"required"`
	CategoryID  uint    `json:"category_id" validate:"required"`
}

// CreateProductRequest คือ DTO ที่ Handler สร้างเพื่อส่งให้ Service ตอนสร้างสินค้า
type CreateProductRequest struct {
	Data  string      // JSON string of product data
	Files []FileInput // Slice ของไฟล์ที่แกะข้อมูลแล้ว
}

// UpdateImagesRequest คือ DTO สำหรับจัดการชุดรูปภาพ
type UpdateImagesRequest struct {
	FilesToAdd       []FileInput
	ImageIDsToDelete []uint
}

// ===================================================================
// DTOs for Responses (ข้อมูลที่ Server ส่งกลับไป)
// ===================================================================

// ImageResponse เป็น DTO สำหรับข้อมูลรูปภาพ
type ImageResponse struct {
	ID        uint   `json:"id"`
	URL       string `json:"url"`
	IsPrimary bool   `json:"is_primary"`
}

// CategoryResponse เป็น DTO สำหรับข้อมูลหมวดหมู่
type CategoryResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// ProductResponse คือ DTO สำหรับแสดงผลข้อมูลสินค้า 1 ชิ้นแบบ "จัดเต็ม"
type ProductResponse struct {
	ID          uint             `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Price       float64          `json:"price"`
	Quantity    int              `json:"quantity"`
	SKU         string           `json:"sku"`
	Category    CategoryResponse `json:"category"`
	Images      []ImageResponse  `json:"images"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// ProductListDTO คือ DTO สำหรับแสดงผลในหน้ารายการสินค้า (ข้อมูลย่อ)
type ProductListDTO struct {
	ID              uint    `json:"id"`
	Name            string  `json:"name"`
	Price           float64 `json:"price"`
	SKU             string  `json:"sku"`
	CategoryName    string  `json:"category_name"`
	PrimaryImageURL string  `json:"primary_image_url"`
}

// PaginatedProductsDTO คือ DTO สำหรับ Response ที่มีข้อมูลการแบ่งหน้า
type PaginatedProductsDTO struct {
	Data        []ProductListDTO `json:"data"`
	TotalItems  int64            `json:"total_items"`
	TotalPages  int              `json:"total_pages"`
	CurrentPage int              `json:"current_page"`
	Limit       int              `json:"limit"`
}
