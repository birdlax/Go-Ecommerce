package domain

import (
	"gorm.io/gorm"
)

// OrderStatus คือประเภทข้อมูลสำหรับสถานะของ Order
type OrderStatus string

// ค่าคงที่สำหรับ OrderStatus
const (
	StatusPending    OrderStatus = "pending"    // รอการชำระเงิน
	StatusProcessing OrderStatus = "processing" // จ่ายเงินแล้ว, กำลังเตรียมของ
	StatusShipped    OrderStatus = "shipped"    // จัดส่งแล้ว
	StatusCompleted  OrderStatus = "completed"  // ได้รับของแล้ว
	StatusCancelled  OrderStatus = "cancelled"  // ยกเลิก
)

// Order คือข้อมูลหลักของคำสั่งซื้อ
type Order struct {
	gorm.Model
	UserID            uint `gorm:"not null"`
	User              User
	OrderItems        []OrderItem `gorm:"foreignKey:OrderID"`
	TotalPrice        float64     `gorm:"not null"`
	Discount          float64     `gorm:"not null;default:0"` // <-- เพิ่ม: ยอดส่วนลด
	FinalPrice        float64     `gorm:"not null"`           // <-- เพิ่ม: ยอดที่ต้องจ่ายจริง
	AppliedCouponCode *string     `gorm:"type:varchar(50)"`   // <-- เพิ่ม: โค้ดคูปองที่ใช้
	ShippingAddressID uint        `gorm:"not null"`           // ID ของที่อยู่ที่จะจัดส่ง
	ShippingAddress   Address     `gorm:"foreignKey:ShippingAddressID"`
	PaymentMethod     *string     `gorm:"type:varchar(50)"`
	Status            OrderStatus `gorm:"type:varchar(20);not null;default:'pending'"`
	TrackingNumber    *string     `gorm:"type:varchar(50);default:null"` // หมายเลขติดตามพัสดุ
}

// OrderItem คือสินค้าแต่ละรายการในคำสั่งซื้อ
type OrderItem struct {
	gorm.Model
	OrderID   uint `gorm:"not null"`
	ProductID uint `gorm:"not null"`
	Product   Product
	Quantity  uint    `gorm:"not null"`
	Price     float64 `gorm:"not null"` // ราคาของสินค้า ณ เวลาที่สั่งซื้อ
}
