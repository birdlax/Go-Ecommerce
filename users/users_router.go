// package products // แนะนำให้เปลี่ยนชื่อ package เป็น 'products' เพื่อให้ตรงกับ Domain
package users

import (
	"backend/users/handler"
	"backend/users/repository"
	"backend/users/service"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
)

func RegisterModule(api fiber.Router, db *gorm.DB) {
	// --- 1. ประกอบร่าง (Wiring) Dependencies ---
	// สร้างทุกอย่างจากชั้นในสุด (Repository) ออกมาข้างนอก (Handler)

	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo)
	userHdl := handler.NewUserHandler(userSvc)

	// --- 2. ลงทะเบียน Routes ---

	// สร้างกลุ่มสำหรับ User-related endpoints
	usersAPI := api.Group("/api/v1/users")
	usersAPI.Post("/register", userHdl.HandleRegister)
	// usersAPI.Get("/me", userHdl.HandleGetProfile) // ตัวอย่าง Endpoint สำหรับอนาคต (ต้องใช้ JWT)

	// สร้างกลุ่มสำหรับ Authentication endpoints
	// authAPI := api.Group("/auth")
	// authAPI.Post("/login", userHdl.HandleLogin) // Endpoint สำหรับทำ Login ในขั้นตอนถัดไป

	log.Println("✅ User module registered successfully.")
}
