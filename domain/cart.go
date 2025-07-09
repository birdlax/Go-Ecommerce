package domain

import "gorm.io/gorm"

// Cart คือตะกร้าสินค้าหลักของผู้ใช้แต่ละคน
type Cart struct {
	gorm.Model
	UserID   uint `gorm:"uniqueIndex;not null"` // Foreign Key และกำหนดให้ 1 User มีได้แค่ 1 ตะกร้า
	User     User
	Items    []CartItem `gorm:"foreignKey:CartID"` // 1 ตะกร้า มีได้หลาย Item
	CouponID *uint      `gorm:"null"`              // <-- เพิ่ม: ID ของคูปองที่ใช้ (เป็น pointer เพราะอาจจะไม่มี)
	Coupon   *Coupon
}

// CartItem คือสินค้าแต่ละรายการที่อยู่ในตะกร้า
type CartItem struct {
	gorm.Model
	CartID    uint    `gorm:"not null"`
	ProductID uint    `gorm:"not null"`
	Product   Product // เพื่อให้ดึงข้อมูลสินค้ามาแสดงได้
	Quantity  uint    `gorm:"not null"`
}
