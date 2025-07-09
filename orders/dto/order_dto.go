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
	ShippingAddress   *AddressResponse    `json:"shipping_address,omitempty"`
	PaymentMethod     *string             `json:"payment_method,omitempty"`
	Items             []OrderItemResponse `json:"items"`
}

type PaymentWebhookRequest struct {
	OrderID       uint   `json:"order_id" validate:"required"`
	Status        string `json:"status" validate:"required,oneof=success failed"`
	PaymentMethod string `json:"payment_method" validate:"required,oneof=credit_card bank_transfer"`
}

type ShipOrderRequest struct {
	OrderID        uint   `json:"order_id" validate:"required"`
	TrackingNumber string `json:"tracking_number" validate:"required"`
}

type AddressResponse struct {
	ID           uint   `json:"id"`
	AddressLine1 string `json:"address_line_1"`
	City         string `json:"city"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
	State        string `json:"state"`
	AddressLine2 string `json:"address_line_2,omitempty"` // ใช้ omitempty
	IsDefault    bool   `json:"is_default"`
}
