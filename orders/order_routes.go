package orders

import (
	"backend/middleware"
	"backend/orders/handler"

	"backend/config"
	"backend/internal/datastore"
	"backend/orders/service"
	"github.com/gofiber/fiber/v2"

	"log"
)

func RegisterModule(api fiber.Router, uow datastore.UnitOfWork, cfg *config.Config) {

	orderSvc := service.NewOrderService(uow)
	orderHdl := handler.NewOrderHandler(orderSvc)

	// สร้างกลุ่ม Route และป้องกันด้วย Middleware
	orderAPI := api.Group("/orders", middleware.Protected())

	orderAPI.Post("/", orderHdl.HandleCreateOrder)
	orderAPI.Get("/", orderHdl.HandleGetMyOrders)
	orderAPI.Get("/:id", orderHdl.HandleGetOrderByID)
	// --- เพิ่ม Route สำหรับ Webhook (เป็น Public แต่ในระบบจริงต้องมี Signature Verification) ---
	orderAPI.Post("/payments/webhook", orderHdl.HandlePaymentWebhook)
	// --- เพิ่ม Route สำหรับ Ship Order ---

	adminOrderAPI := orderAPI.Group("", middleware.AdminRequired())
	adminOrderAPI.Post("/:id/confirm-payment", orderHdl.HandleConfirmPayment)
	adminOrderAPI.Post("/:id/ship", orderHdl.HandleShipOrder)

	log.Println("✅ Order module registered successfully.")
}
