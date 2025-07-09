package dto

import (
	"backend/domain" // [สำคัญ] แก้ไขชื่อ Module ให้ถูกต้อง
	"time"
)

// DTO สำหรับส่งข้อมูลสรุปทั้งหมดกลับไป
type DashboardStatsResponse struct {
	SalesSummary     SalesSummaryResponse      `json:"sales_summary"`
	CountSummary     CountSummaryResponse      `json:"count_summary"`
	RecentOrders     []RecentOrderResponse     `json:"recent_orders"`
	LowStockProducts []LowStockProductResponse `json:"low_stock_products"`
}

type SalesSummaryResponse struct {
	Today      float64 `json:"today"`
	Last7Days  float64 `json:"last_7_days"`
	Last30Days float64 `json:"last_30_days"`
}

type CountSummaryResponse struct {
	TotalUsers    int64 `json:"total_users"`
	TotalProducts int64 `json:"total_products"`
	TotalOrders   int64 `json:"total_orders"`
}

type RecentOrderResponse struct {
	OrderID      uint               `json:"order_id"`
	CustomerName string             `json:"customer_name"`
	TotalPrice   float64            `json:"total_price"`
	Status       domain.OrderStatus `json:"status"`
	CreatedAt    time.Time          `json:"created_at"`
}

type LowStockProductResponse struct {
	ProductID   uint   `json:"product_id"`
	ProductName string `json:"product_name"`
	SKU         string `json:"sku"`
	Quantity    int    `json:"quantity"`
}
