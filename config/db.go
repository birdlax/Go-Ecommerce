// config/db.go
package config

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ไม่มี global variable แล้ว

// ฟังก์ชันจะรับ DSN (string) เข้ามาเป็น argument และคืนค่า *gorm.DB กับ error
// ทำให้ฟังก์ชันนี้บริสุทธิ์และทดสอบง่ายขึ้นมาก
func ConnectDatabase(dsn string) (*gorm.DB, error) {
	// ใช้ DSN ที่ได้รับมาเพื่อเชื่อมต่อ
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// คืนค่า error กลับไปให้ main.go จัดการ
		return nil, err
	}

	// ลบส่วน AutoMigrate ออกไป (จะย้ายไปทำที่ main.go)

	return db, nil
}
