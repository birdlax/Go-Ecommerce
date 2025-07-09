package main

import (
	// "backend/carts"
	"backend/carts"
	"backend/categories"
	"backend/config"
	"backend/coupons"
	"backend/dashboard"
	"backend/domain"
	"backend/internal/datastore"
	"backend/middleware"
	"backend/orders"
	"backend/products"
	"backend/users"
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
		&domain.Coupon{},
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

	products.RegisterModule(api, uow, cfg)
	categories.RegisterModule(api, uow, cfg)
	users.RegisterModule(api, uow, cfg)
	carts.RegisterModule(api, uow, cfg)
	orders.RegisterModule(api, uow, cfg)
	dashboard.RegisterModule(api, uow, cfg)
	coupons.RegisterModule(api, uow, cfg)

	log.Println("Server started on :8080")
	app.Listen(":8080")
}
