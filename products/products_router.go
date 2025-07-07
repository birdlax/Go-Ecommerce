package products

import (
	// [สำคัญ] แก้ไข "my-ecommerce-app" เป็นชื่อ Module ใน go.mod ของคุณ
	"backend/config"
	"backend/internal/datastore"
	// "backend/middleware"
	"backend/products/handler"
	"backend/products/service"

	"github.com/gofiber/fiber/v2"
	"log"
)

// [แก้ไข] Signature ของฟังก์ชันจะเปลี่ยนไป รับ uow และ cfg
func RegisterModule(api fiber.Router, uow datastore.UnitOfWork, cfg *config.Config) {

	// --- การประกอบร่างจะเกิดขึ้นที่นี่ โดยใช้ Dependencies ที่ได้รับมา ---
	// Service จะถูกสร้างโดยรับแค่ UoW และ Config ที่จำเป็น (ImageBaseURL)
	productSvc := service.NewProductService(uow, cfg.ImageBaseURL)
	productHdl := handler.NewProductHandler(productSvc)

	// --- ลงทะเบียน Routes (ส่วนนี้ของคุณถูกต้องและดีมากแล้ว) ---
	productsAPI := api.Group("/products")

	// >> Public Routes <<
	productsAPI.Get("/", productHdl.HandleGetAllProducts)
	productsAPI.Get("/:id", productHdl.HandleGetProductByID)

	// >> Admin-Only Routes <<
	adminAPI := productsAPI.Group("")
	// adminAPI.Use(middleware.Protected(), middleware.AdminRequired())

	adminAPI.Post("/", productHdl.HandleCreateProduct)
	adminAPI.Patch("/:id", productHdl.HandleUpdateProduct)
	adminAPI.Delete("/:id", productHdl.HandleDeleteProduct)
	adminAPI.Patch("/:productId/images", productHdl.HandleUpdateImages)

	log.Println("✅ Product module registered successfully.")
}
