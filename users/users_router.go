// package products // แนะนำให้เปลี่ยนชื่อ package เป็น 'products' เพื่อให้ตรงกับ Domain
package users

import (
	"backend/middleware"
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

	apiv1 := api.Group("/api/v1")

	authAPI := apiv1.Group("/auth")
	authAPI.Post("/register", userHdl.HandleRegister)
	authAPI.Post("/login", userHdl.HandleLogin)
	authAPI.Post("/refresh", userHdl.HandleRefreshToken)
	authAPI.Post("/logout", middleware.Protected(), userHdl.HandleLogout)

	usersAPI := apiv1.Group("/users")
	usersAPI.Get("/me", middleware.Protected(), userHdl.HandleGetMyProfile)
	// --- 3. ลงทะเบียน Middleware ---
	adminOnlyAPI := usersAPI.Use(middleware.Protected(), middleware.AdminRequired())
	adminOnlyAPI.Get("/", userHdl.HandleGetAllUsers)
	// adminOnlyAPI.Get("/:id", userHdl.HandleGetUserByID)
	// adminOnlyAPI.Patch("/:id", userHdl.HandleUpdateUser)
	// adminOnlyAPI.Delete("/:id", userHdl.HandleDeleteUser)

	log.Println("✅ User module registered successfully.")
}
