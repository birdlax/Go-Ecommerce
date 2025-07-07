package carts

import (
	"backend/carts/handler"
	"backend/carts/repository"
	"backend/carts/service"
	"backend/middleware"
	productRepo "backend/products/repository"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
)

func RegisterModule(api fiber.Router, db *gorm.DB, imageBaseURL string) {
	// สร้าง dependencies
	cartRepo := repository.NewCartRepository(db)
	prodRepo := productRepo.NewProductRepository(db) // Cart Service ต้องการ Product Repo
	cartSvc := service.NewCartService(cartRepo, prodRepo, imageBaseURL)
	cartHdl := handler.NewCartHandler(cartSvc)

	// สร้างกลุ่ม Route สำหรับ Cart และป้องกันด้วย Middleware
	cartAPI := api.Group("/cart", middleware.Protected())

	cartAPI.Get("/", cartHdl.HandleGetCart)
	cartAPI.Post("/items", cartHdl.HandleAddItemToCart)
	cartAPI.Patch("/items/:itemId", cartHdl.HandleUpdateCartItem)
	cartAPI.Delete("/items/:itemId", cartHdl.HandleRemoveCartItem)

	log.Println("✅ Cart module registered successfully.")
}
