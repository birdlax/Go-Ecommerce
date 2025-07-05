// package products // แนะนำให้เปลี่ยนชื่อ package เป็น 'products' เพื่อให้ตรงกับ Domain
package products

import (
	"backend/middleware"
	"backend/products/handler"
	"backend/products/repository"
	"backend/products/service"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
)

func RegisterModule(app *fiber.App, db *gorm.DB, azureConnectionString string) {
	productDbRepo := repository.NewProductRepository(db)

	//สร้าง Repository สำหรับ Upload
	uploadRepo, err := repository.NewAzureUploadRepository(azureConnectionString, "uploads")
	if err != nil {
		log.Fatalf("FATAL: could not create upload repository for products: %v", err)
	}

	uow := repository.NewUnitOfWork(db)
	productSvc := service.NewProductService(productDbRepo, uploadRepo, uow)
	productHdl := handler.NewProductHandler(productSvc)

	// --- ลงทะเบียน Routes ---
	api := app.Group("/api/v1")
	productsAPI := api.Group("/products")

	productsAPI.Get("/", productHdl.HandleGetAllProducts)
	productsAPI.Get("/:id", productHdl.HandleGetProductByID)

	// == กลุ่มสำหรับ Product ที่เป็น Admin เท่านั้น ==
	// 1. สร้าง Group ใหม่สำหรับ Admin โดยเฉพาะ
	adminProductAPI := api.Group("/products/admin")
	// 2. ติดตั้ง Middleware ให้กับ Group ใหม่นี้เท่านั้น
	adminProductAPI.Use(middleware.Protected(), middleware.AdminRequired())

	// 3. ลงทะเบียน Route ของ Admin กับ Group ที่ป้องกันแล้ว
	adminProductAPI.Post("/", productHdl.HandleCreateProduct)
	adminProductAPI.Patch("/:id", productHdl.HandleUpdateProduct)
	adminProductAPI.Delete("/:id", productHdl.HandleDeleteProduct)

	// == กลุ่มสำหรับจัดการรูปภาพ (อาจจะให้ Admin เท่านั้นเช่นกัน) ==
	imagesAPI := api.Group("/products") // ใช้ group เดิม
	adminImageAPI := imagesAPI.Use(middleware.Protected(), middleware.AdminRequired())
	adminImageAPI.Patch("/:productId/images", productHdl.HandleUpdateImages)

	log.Println("✅ Product module registered successfully.")
}
