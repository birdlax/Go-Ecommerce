package carts

import (
	"backend/carts/handler"
	"backend/carts/service"
	"backend/config"
	"backend/internal/datastore"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
	"log"
)

func RegisterModule(api fiber.Router, uow datastore.UnitOfWork, cfg *config.Config) {
	// สร้าง dependencies
	cartSvc := service.NewCartService(uow, cfg.ImageBaseURL)
	cartHdl := handler.NewCartHandler(cartSvc)

	// สร้างกลุ่ม Route สำหรับ Cart และป้องกันด้วย Middleware
	cartAPI := api.Group("/cart", middleware.Protected())

	cartAPI.Get("/", cartHdl.HandleGetCart)
	cartAPI.Post("/items", cartHdl.HandleAddItemToCart)
	cartAPI.Patch("/items/:itemId", cartHdl.HandleUpdateCartItem)
	cartAPI.Delete("/items/:itemId", cartHdl.HandleRemoveCartItem)

	log.Println("✅ Cart module registered successfully.")
}
