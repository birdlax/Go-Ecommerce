package orders

import (
	uow "backend/" // Import UoW
	"backend/middleware"
	"backend/orders/handler"
	"backend/orders/repository"
	"backend/orders/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
)

func RegisterModule(api fiber.Router, db *gorm.DB) {
	// สร้าง UoW และ Service
	unitOfWork := uow.NewUnitOfWork(db)
	orderSvc := service.NewOrderService(unitOfWork)
	orderHdl := handler.NewOrderHandler(orderSvc)

	// สร้างกลุ่ม Route และป้องกันด้วย Middleware
	orderAPI := api.Group("/orders", middleware.Protected())

	orderAPI.Post("/", orderHdl.HandleCreateOrder)
	orderAPI.Get("/", orderHdl.HandleGetMyOrders)
	orderAPI.Get("/:id", orderHdl.HandleGetOrderByID)

	log.Println("✅ Order module registered successfully.")
}
