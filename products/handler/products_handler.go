// products/handler/products_handler.go
package handler

import (
	"backend/products/domain"
	"backend/products/service"
	"log"

	"github.com/gofiber/fiber/v2"
)

type ProductHandler struct {
	productSvc service.ProductService
}

func NewProductHandler(productSvc service.ProductService) *ProductHandler {
	return &ProductHandler{
		productSvc: productSvc,
	}
}

func (h *ProductHandler) HandleCreateProduct(c *fiber.Ctx) error {
	// 1. รับค่า 'data' ที่เป็น JSON String
	productDataJSON := c.FormValue("data")
	if productDataJSON == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing product data field 'data'"})
	}

	// 2. รับไฟล์ทั้งหมดจาก field 'files'
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid form data"})
	}
	files := form.File["files"] // รับ slice ของไฟล์

	// 3. สร้าง Request object เพื่อส่งให้ Service
	req := service.CreateProductRequest{
		Data:  productDataJSON,
		Files: files,
	}

	// 4. เรียกใช้ Service
	createdProduct, err := h.productSvc.CreateProductWithImages(c.Context(), req)
	if err != nil {
		log.Printf("Error creating product with images: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create product"})
	}

	return c.Status(fiber.StatusCreated).JSON(createdProduct)
}
func (h *ProductHandler) HandleCreateCategory(c *fiber.Ctx) error {
	var req domain.Category
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}
	err := h.productSvc.CreateNewCategory(req)
	if err != nil {
		log.Printf("Error creating category: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not create category",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Category created successfully",
		"category": req,
	})
}
