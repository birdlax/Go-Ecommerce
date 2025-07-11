package dto

// AddItemRequest คือ DTO สำหรับรับข้อมูลตอนเพิ่มสินค้าลงตะกร้า
type AddItemRequest struct {
	ProductID uint `json:"product_id" validate:"required"`
	Quantity  uint `json:"quantity" validate:"required,min=1"`
}

// CartItemResponse คือ DTO สำหรับสินค้าแต่ละรายการในตะกร้า
type CartItemResponse struct {
	ID        uint    `json:"id"`
	ProductID uint    `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  uint    `json:"quantity"`
	ImageURL  string  `json:"image_url"`
}

// CartResponse คือ DTO สำหรับตะกร้าสินค้าทั้งหมด
type CartResponse struct {
	ID            uint               `json:"id"`
	UserID        uint               `json:"user_id"`
	Items         []CartItemResponse `json:"items"`
	Subtotal      float64            `json:"subtotal"`                 // <-- เพิ่ม: ราคารวมก่อนหักส่วนลด
	Discount      float64            `json:"discount"`                 // <-- เพิ่ม: ยอดเงินส่วนลด
	GrandTotal    float64            `json:"grand_total"`              // <-- เพิ่ม: ราคาสุทธิ
	AppliedCoupon *string            `json:"applied_coupon,omitempty"` // <-- เพิ่ม: โค้ดคูปองที่ใช้
}
type ApplyCouponRequest struct {
	CouponCode string `json:"coupon_code" validate:"required"`
}
type UpdateItemRequest struct {
	// ใช้ gte=0 เพื่อให้สามารถส่งค่า 0 มาเพื่อลบสินค้าได้
	Quantity uint `json:"quantity" validate:"gte=0"`
}
