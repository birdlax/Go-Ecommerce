package main

import (
	// "backend/carts"
	"backend/config"
	"backend/domain"
	"backend/internal/datastore"
	"backend/middleware"
	// "backend/orders"
	"backend/products"
	// "backend/users"
	"github.com/gofiber/fiber/v2"

	"log"
)

func main() {
	// 1. โหลด Config ทั้งหมด
	cfg := config.LoadConfig()

	// 2. เชื่อมต่อ Database
	db, err := config.ConnectDatabase(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to database: %v", err)
	}
	db.AutoMigrate(
		&domain.Category{}, &domain.Product{}, &domain.ProductImage{},
		&domain.User{}, &domain.Address{},
		&domain.Cart{}, &domain.CartItem{},
		&domain.Order{}, &domain.OrderItem{},
	)

	// 3. สร้าง Dependencies ส่วนกลาง
	uploadRepo, err := datastore.NewAzureUploadRepository(cfg.AzureConnectionString, "uploads")
	if err != nil {
		log.Fatalf("FATAL: could not create upload repository: %v", err)
	}
	uow := datastore.NewUnitOfWork(db, uploadRepo)

	// 4. สร้าง Fiber App
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	api := app.Group("/api/v1")

	// --- [แก้ไข] 5. ลงทะเบียน Module โดยส่ง uow และ cfg เข้าไป ---
	// เราส่ง cfg ทั้งก้อนเข้าไป เพื่อให้แต่ละ Module เลือกใช้ค่า Config ที่ตัวเองต้องการได้
	products.RegisterModule(api, uow, cfg)
	// users.RegisterModule(api, uow, cfg) // แก้ไข User Module ให้รับ cfg ด้วย
	// carts.RegisterModule(api, uow, cfg)
	// orders.RegisterModule(api, uow) // Order อาจจะไม่ต้องการ config เพิ่มเติม

	log.Println("Server started on :8080")
	app.Listen(":8080")
}
