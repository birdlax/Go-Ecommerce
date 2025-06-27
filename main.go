package main

import (
	"backend/config"
	"backend/domain"
	"backend/middleware"
	"backend/products"
	"backend/users"
	"github.com/gofiber/fiber/v2"
	"log"
)

func main() {
	cfg := config.LoadConfig()
	db, err := config.ConnectDatabase(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to database: %v", err)
	}
	db.AutoMigrate(&domain.Category{}, &domain.Product{}, &domain.ProductImage{}, &domain.User{}, &domain.Address{})

	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	// Register the product module with the app and database connection
	products.RegisterModule(app, db, cfg.AzureConnectionString)
	// Register the user module with the app and database connection
	users.RegisterModule(app, db)

	log.Println("Server started on :8080")
	app.Listen(":8080")

}
