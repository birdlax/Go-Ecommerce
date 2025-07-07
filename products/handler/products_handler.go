// products/handler/products_handler.go
package handler

import (
	"backend/domain"
	"backend/products/dto"
	"backend/products/service"
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
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
	productDataJSON := c.FormValue("data")
	if productDataJSON == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing product data field 'data'")
	}

	form, err := c.MultipartForm()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid form data")
	}
	formFiles := form.File["files"]

	var filesToAdd []dto.FileInput // <-- [แก้ไข] ใช้ FileInput จาก dto
	for _, fileHeader := range formFiles {
		file, err := fileHeader.Open()
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Cannot process uploaded file: "+fileHeader.Filename)
		}
		defer file.Close()
		filesToAdd = append(filesToAdd, dto.FileInput{
			Content:     file,
			Filename:    fileHeader.Filename,
			ContentType: fileHeader.Header.Get("Content-Type"),
		})
	}

	// [แก้ไข] ใช้ DTO จาก package dto
	req := dto.CreateProductRequest{
		Data:  productDataJSON,
		Files: filesToAdd,
	}

	// Validate DTO
	var productDataForValidation dto.CreateProductRequestData
	if err := json.Unmarshal([]byte(req.Data), &productDataForValidation); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid product data json")
	}
	if err := validator.New().Struct(productDataForValidation); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	productResponse, err := h.productSvc.CreateProductWithImages(c.Context(), req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(productResponse)
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

func (h *ProductHandler) HandleUpdateImages(c *fiber.Ctx) error {
	productID, err := c.ParamsInt("productId")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid product ID")
	}

	form, err := c.MultipartForm()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid form data")
	}

	var filesToAdd []dto.FileInput // <-- [แก้ไข] ใช้ FileInput จาก dto
	formFiles := form.File["add_files"]
	for _, fileHeader := range formFiles {
		file, err := fileHeader.Open()
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Cannot process uploaded file: "+fileHeader.Filename)
		}
		defer file.Close()
		filesToAdd = append(filesToAdd, dto.FileInput{
			Content:     file,
			Filename:    fileHeader.Filename,
			ContentType: fileHeader.Header.Get("Content-Type"),
		})
	}

	var imageIDsToDelete []uint
	deleteIDsStr := c.FormValue("delete_image_ids")
	if deleteIDsStr != "" {
		ids := strings.Split(deleteIDsStr, ",")
		for _, idStr := range ids {
			id, err := strconv.Atoi(strings.TrimSpace(idStr))
			if err == nil {
				imageIDsToDelete = append(imageIDsToDelete, uint(id))
			}
		}
	}

	// [แก้ไข] ใช้ DTO จาก package dto
	req := dto.UpdateImagesRequest{
		FilesToAdd:       filesToAdd,
		ImageIDsToDelete: imageIDsToDelete,
	}

	if err := h.productSvc.UpdateProductImages(c.Context(), uint(productID), req); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Product images updated successfully"})
}
