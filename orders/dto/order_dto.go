package dto

import (
	"backend/domain"
	"time"
)

// CreateOrderRequest คือ DTO สำหรับรับข้อมูลตอนสร้าง Order
type CreateOrderRequest struct {
	ShippingAddressID uint `json:"shipping_address_id" validate:"required"`
}

type OrderItemResponse struct {
	ProductID uint    `json:"product_id"`
	Name      string  `json:"name"`
	Sku       string  `json:"sku"`
	Price     float64 `json:"price"` // ราคา ณ ตอนที่สั่งซื้อ
	Quantity  uint    `json:"quantity"`
}

// OrderResponse คือ DTO สำหรับแสดงข้อมูล Order ฉบับเต็ม
type OrderResponse struct {
	ID                uint                `json:"id"`
	UserID            uint                `json:"user_id"`
	TotalPrice        float64             `json:"total_price"`
	Status            domain.OrderStatus  `json:"status"`
	ShippingAddressID uint                `json:"shipping_address_id"`
	CreatedAt         time.Time           `json:"created_at"`
	Items             []OrderItemResponse `json:"items"`
}
