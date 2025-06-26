package main

import (
	"backend/config"
	"backend/middleware"
	"backend/products/domain"
	"backend/products/routes"
	"github.com/gofiber/fiber/v2"
	"log"
)

func main() {
	cfg := config.LoadConfig()
	db, err := config.ConnectDatabase(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to database: %v", err)
	}
	db.AutoMigrate(&domain.Category{}, &domain.Product{}, &domain.ProductImage{})

	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler, // <-- บอกให้ Fiber ใช้ Error Handler ของเรา
	})

	// Register the product module with the app and database connection
	routes.RegisterModule(app, db, cfg.AzureConnectionString)

	log.Println("Server started on :8080")
	app.Listen(":8080")

}
