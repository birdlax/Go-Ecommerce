package routes

import (
	"backend/products/handler"
	"backend/products/repository"
	"backend/products/service"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
)

func RegisterModule(app *fiber.App, db *gorm.DB, azureConnectionString string) {
	// สร้าง Repository ทั้งหมดที่ Product Domain ต้องการ
	productDbRepo := repository.NewProductRepository(db)
	uploadRepo, err := repository.NewAzureUploadRepository(azureConnectionString, "uploads")
	if err != nil {
		log.Fatalf("FATAL: could not create upload repository for products: %v", err)
	}

	productSvc := service.NewProductService(productDbRepo, uploadRepo)
	productHdl := handler.NewProductHandler(productSvc)

	api := app.Group("/api/v1")
	// Route สำหรับสร้างหมวดหมู่สินค้า
	api.Post("/categories", productHdl.HandleCreateCategory)
	// Route สำหรับสร้างสินค้า
	productsAPI := api.Group("/products")
	productsAPI.Post("/", productHdl.HandleCreateProduct)
	productsAPI.Get("/", productHdl.HandleGetAllProducts)
	productsAPI.Get("/:id", productHdl.HandleGetProductByID)
	productsAPI.Delete("/:id", productHdl.HandleDeleteProduct)
	productsAPI.Patch("/:id", productHdl.HandleUpdateProduct)

	log.Println("✅ Product module registered successfully.")
}
