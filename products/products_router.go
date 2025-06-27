// package products // แนะนำให้เปลี่ยนชื่อ package เป็น 'products' เพื่อให้ตรงกับ Domain
package products

import (
	"backend/products/handler"
	"backend/products/repository"
	"backend/products/service"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
)

func RegisterModule(app *fiber.App, db *gorm.DB, azureConnectionString string) {
	// --- ประกอบร่าง Dependencies ---

	// 1. สร้าง Repository สำหรับงาน Read (ใช้ DB connection หลัก)
	productDbRepo := repository.NewProductRepository(db)

	// 2. สร้าง Repository สำหรับ Upload
	uploadRepo, err := repository.NewAzureUploadRepository(azureConnectionString, "uploads")
	if err != nil {
		log.Fatalf("FATAL: could not create upload repository for products: %v", err)
	}

	// 3. สร้าง Unit of Work (ใช้ DB connection หลักเพื่อเริ่ม Transaction)
	uow := repository.NewUnitOfWork(db)

	// 4. สร้าง Service โดยฉีด Repository ปกติ และ Unit of Work เข้าไป
	productSvc := service.NewProductService(productDbRepo, uploadRepo, uow)

	// 5. สร้าง Handler (ส่วนนี้เหมือนเดิม)
	productHdl := handler.NewProductHandler(productSvc)

	// --- ลงทะเบียน Routes ---
	api := app.Group("/api/v1")
	productsAPI := api.Group("/products")

	productsAPI.Post("/", productHdl.HandleCreateProduct)
	productsAPI.Get("/", productHdl.HandleGetAllProducts)
	productsAPI.Get("/:id", productHdl.HandleGetProductByID)
	productsAPI.Patch("/:id", productHdl.HandleUpdateProduct)
	productsAPI.Delete("/:id", productHdl.HandleDeleteProduct)

	// Routes สำหรับจัดการรูปภาพ
	productsAPI.Patch("/:productId/images", productHdl.HandleUpdateImages)

	log.Println("✅ Product module registered successfully.")

	// --- ข้อแนะนำสำหรับอนาคต (ดูด้านล่าง) ---
	// Route สำหรับสร้างหมวดหมู่สินค้า
	// api.Post("/categories", categoryHdl.HandleCreateCategory)
}
