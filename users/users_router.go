package users

import (
	"backend/config"
	"backend/internal/datastore"
	"backend/middleware"
	"backend/users/handler"
	"backend/users/service"

	"github.com/gofiber/fiber/v2"
	"log"
)

func RegisterModule(api fiber.Router, uow datastore.UnitOfWork, cfg *config.Config) {
	// --- 1. ประกอบร่าง (Wiring) Dependencies ---
	// สร้างทุกอย่างจากชั้นในสุด (Repository) ออกมาข้างนอก (Handler)

	userSvc := service.NewUserService(uow)
	userHdl := handler.NewUserHandler(userSvc)

	// == กลุ่มสำหรับ Auth (Public) ==
	authAPI := api.Group("/auth")
	authAPI.Post("/register", userHdl.HandleRegister)
	authAPI.Post("/login", userHdl.HandleLogin)
	authAPI.Post("/refresh", userHdl.HandleRefreshToken)
	authAPI.Post("/logout", middleware.Protected(), userHdl.HandleLogout)

	// == กลุ่มสำหรับ Address (ต้อง Login) ==
	addressAPI := api.Group("/addresses", middleware.Protected())
	addressAPI.Post("/", userHdl.HandleAddAddress)
	addressAPI.Get("/", userHdl.HandleGetUserAddresses)
	addressAPI.Patch("/:addressId", userHdl.HandleUpdateAddress)
	addressAPI.Delete("/:addressId", userHdl.HandleDeleteAddress)

	// == กลุ่มสำหรับ User (มีหลายระดับสิทธิ์) ==
	usersAPI := api.Group("/users")

	// >> Route สำหรับ "ผู้ใช้ที่ล็อกอินแล้ว" ทุกคน (Public action for authenticated users) <<
	usersAPI.Get("/me", middleware.Protected(), userHdl.HandleGetMyProfile)

	// >> กลุ่มสำหรับ "Admin เท่านั้น" (ตามที่คุณแนะนำ) <<
	// 1. สร้าง Group สำหรับ Admin จาก /users
	// การใช้ .Group("") จะไม่สร้าง path เพิ่ม แต่จะให้ router instance ใหม่มาจัดการ
	adminAPI := usersAPI.Group("")
	// 2. ติดตั้ง Middleware ให้กับ Group นี้ครั้งเดียว
	adminAPI.Use(middleware.Protected(), middleware.AdminRequired())

	adminAPI.Get("/", userHdl.HandleGetAllUsers)
	adminAPI.Get("/:id", userHdl.HandleGetUserByID)
	adminAPI.Patch("/:id", userHdl.HandleUpdateUser)
	adminAPI.Delete("/:id", userHdl.HandleDeleteUser)

	log.Println("✅ User module registered successfully.")
}
