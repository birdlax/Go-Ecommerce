// products/handler/products_handler.go
package handler

import (
	"backend/products/domain"
	"backend/products/service"
	"log"
	"strconv"
	"strings"

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

func (h *ProductHandler) HandleGetAllProducts(c *fiber.Ctx) error {
	// 1. สร้าง QueryParams พร้อมตั้งค่า Default
	params := domain.QueryParams{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 20),
		SortBy: c.Query("sort_by", "created_at"), // ค่าเริ่มต้นเรียงตาม "ล่าสุด"
		Order:  c.Query("order", "desc"),
	}

	// 2. Validate ค่าที่รับเข้ามา
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}
	// แปลงค่า "latest" เป็น "created_at"
	if params.SortBy == "latest" {
		params.SortBy = "created_at"
	}
	// ตรวจสอบค่าที่อนุญาตสำหรับ SortBy เพื่อป้องกัน SQL Injection
	allowedSorts := map[string]bool{"price": true, "name": true, "created_at": true}
	if !allowedSorts[params.SortBy] {
		params.SortBy = "created_at" // ถ้าไม่ถูกต้องให้กลับไปใช้ค่า Default
	}

	// 3. เรียกใช้ Service
	paginatedResult, err := h.productSvc.FindAllProducts(params)
	if err != nil {
		log.Printf("Error finding all products: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not retrieve products",
		})
	}

	return c.Status(fiber.StatusOK).JSON(paginatedResult)
}

func (h *ProductHandler) HandleGetProductByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid product ID format") // คืนเป็น Fiber Error
	}

	product, err := h.productSvc.FindProductByID(uint(id))
	if err != nil {
		return err // <-- คืน error ออกไปตรงๆ เลย! ไม่ต้องเขียน if/else แล้ว
	}

	return c.Status(fiber.StatusOK).JSON(product)
}

func (h *ProductHandler) HandleDeleteProduct(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		// ใช้ Fiber Error Handler กลางที่เราสร้างไว้
		return fiber.NewError(fiber.StatusBadRequest, "Invalid product ID format")
	}

	err = h.productSvc.DeleteProduct(uint(id))
	if err != nil {
		// คืน error ให้ Middleware จัดการต่อ
		// Middleware จะเช็คว่าเป็น ErrProductNotFound หรือไม่ แล้วส่ง 404 กลับไป
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Delete successfully",
	})
}

func (h *ProductHandler) HandleUpdateProduct(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid product ID format")
	}

	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot parse JSON")
	}

	// ลบฟิลด์ที่ไม่ควรให้ user อัปเดตได้เองออกจาก map
	delete(updates, "id")

	// เรียก service ให้อัปเดต
	updatedProduct, err := h.productSvc.UpdateProduct(uint(id), updates)
	if err != nil {
		// คืน error ที่ได้รับจาก service ไปตรงๆ
		// Middleware Error Handler ของเราจะดักจับและแปลงเป็น HTTP Response ที่เหมาะสม
		// เช่น ถ้า err คือ service.ErrProductNotFound มันจะแปลงเป็น 404 Not Found
		return err
	}

	return c.Status(fiber.StatusOK).JSON(updatedProduct)
}

func (h *ProductHandler) HandleReplaceProductImage(c *fiber.Ctx) error {
	productID, err := c.ParamsInt("productId")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid product ID format")
	}
	imageID, err := c.ParamsInt("imageId")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid image ID format")
	}

	fileHeader, err := c.FormFile("file") // รับไฟล์ใหม่จาก field ชื่อ 'file'
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Missing image file")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Cannot open file")
	}
	defer file.Close()

	updatedImage, err := h.productSvc.ReplaceProductImage(c.Context(), uint(productID), uint(imageID), file, fileHeader)
	if err != nil {
		// Middleware จัดการ Error ต่อให้
		return err
	}

	return c.Status(fiber.StatusOK).JSON(updatedImage)
}

func (h *ProductHandler) HandleUpdateImages(c *fiber.Ctx) error {
	productID, err := c.ParamsInt("productId")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid product ID")
	}

	// รับไฟล์ที่จะเพิ่ม
	form, err := c.MultipartForm()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid form data")
	}
	filesToAdd := form.File["add_files"]

	// รับ ID ของไฟล์ที่จะลบ
	deleteIDsStr := c.FormValue("delete_image_ids")
	var imageIDsToDelete []uint
	if deleteIDsStr != "" {
		ids := strings.Split(deleteIDsStr, ",")
		for _, idStr := range ids {
			id, err := strconv.Atoi(strings.TrimSpace(idStr))
			if err == nil {
				imageIDsToDelete = append(imageIDsToDelete, uint(id))
			}
		}
	}

	req := service.UpdateImagesRequest{
		FilesToAdd:       filesToAdd,
		ImageIDsToDelete: imageIDsToDelete,
	}

	if err := h.productSvc.UpdateProductImages(c.Context(), uint(productID), req); err != nil {
		return err // ส่งต่อให้ Error Middleware จัดการ
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Product images updated successfully"})
}
