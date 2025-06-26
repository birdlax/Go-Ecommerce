// products/service/products_service.go
package service

import (
	"backend/products/domain"
	"backend/products/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"mime/multipart"
	"path/filepath"

	"github.com/google/uuid"
)

var ErrProductNotFound = errors.New("product not found")

type ProductService interface {
	CreateProductWithImages(ctx context.Context, req CreateProductRequest) (*domain.Product, error)
	CreateNewCategory(req domain.Category) error
	FindAllProducts(params domain.QueryParams) (*domain.PaginatedProductsDTO, error)
	FindProductByID(id uint) (*domain.Product, error)
	DeleteProduct(id uint) error
	UpdateProduct(id uint, updates map[string]interface{}) (*domain.Product, error)
}

type CreateProductRequest struct {
	Data  string // JSON string of product data
	Files []*multipart.FileHeader
}

type productService struct {
	productRepo repository.ProductRepository
	uploadRepo  repository.UploadRepository
}

func NewProductService(productRepo repository.ProductRepository, uploadRepo repository.UploadRepository) ProductService {
	return &productService{
		productRepo: productRepo,
		uploadRepo:  uploadRepo,
	}
}

func (s *productService) CreateProductWithImages(ctx context.Context, req CreateProductRequest) (*domain.Product, error) {
	// 1. Unmarshal ข้อมูลสินค้าจาก JSON String
	var productData domain.Product
	if err := json.Unmarshal([]byte(req.Data), &productData); err != nil {
		return nil, fmt.Errorf("invalid product data json: %w", err)
	}

	// 2. สร้าง Product record ใน DB ก่อน 1 ครั้ง เพื่อเอา ProductID
	// ตอนนี้ productData จะมี ID, CreatedAt, etc. จาก DB แล้ว
	if err := s.productRepo.CreateProduct(&productData); err != nil {
		return nil, fmt.Errorf("failed to create product record: %w", err)
	}
	log.Printf("Successfully created product record with ID: %d", productData.ID)

	// 3. วนลูปอัปโหลดไฟล์ไป Azure และเตรียมข้อมูล ProductImage
	var imagesToCreate []domain.ProductImage
	for i, fileHeader := range req.Files {
		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("cannot open file %s: %w", fileHeader.Filename, err)
		}
		defer file.Close()

		// สร้าง Path โดยใช้ ProductID ที่เพิ่งได้มา
		ext := filepath.Ext(fileHeader.Filename)
		newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
		blobPath := fmt.Sprintf("products/%d/%s", productData.ID, newFileName)

		// เรียก UploadRepository เพื่ออัปโหลดไฟล์ไป Azure
		_, err = s.uploadRepo.UploadFile(ctx, blobPath, file, fileHeader.Header.Get("Content-Type"))
		if err != nil {
			// หมายเหตุ: ในระบบจริงควรมี Logic ลบ Product ที่สร้างไปแล้วถ้าอัปโหลดไฟล์ล้มเหลว (Rollback)
			return nil, fmt.Errorf("failed to upload file %s: %w", fileHeader.Filename, err)
		}

		// เตรียมข้อมูล Image เพื่อรอการบันทึกลง DB
		imagesToCreate = append(imagesToCreate, domain.ProductImage{
			ProductID: productData.ID, // ใช้ ID ของ Product ที่สร้างเสร็จแล้ว
			Path:      blobPath,
			IsPrimary: i == 0,
		})
	}

	// 4. บันทึกข้อมูล Image ทั้งหมดลง DB ในครั้งเดียว (Bulk Insert)
	if err := s.productRepo.CreateImages(imagesToCreate); err != nil {
		// หมายเหตุ: ในระบบจริงควรมี Logic ลบไฟล์ที่อัปโหลดไปแล้วทั้งหมดถ้าบันทึก DB ล้มเหลว
		return nil, fmt.Errorf("failed to save image records to db: %w", err)
	}
	log.Printf("Successfully saved %d image records to database.", len(imagesToCreate))

	// 5. ประกอบร่างข้อมูล Product ที่สมบูรณ์เพื่อส่งกลับ
	// GORM จะไม่โหลด relation ให้อัตโนมัติ เราต้อง query มาใหม่ หรือประกอบเอง
	// เพื่อความง่าย เราจะประกอบร่างเอง
	productData.Images = imagesToCreate

	// ถ้าทุกอย่างสำเร็จ จะไม่มี error และคืนค่า product ที่สมบูรณ์กลับไป
	return &productData, nil
}

func (s *productService) CreateNewCategory(req domain.Category) error {
	newCategory := &domain.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	// เรียกใช้ Repository เพื่อบันทึกข้อมูล
	err := s.productRepo.CreateCategory(newCategory)
	if err != nil {
		return err
	}

	return nil
}

// FindProductByID คือ Logic การดึงข้อมูลชิ้นเดียว และ "แปล" Error
func (s *productService) FindProductByID(id uint) (*domain.Product, error) {
	product, err := s.productRepo.FindProductByID(id)
	if err != nil {
		// "แปล" error จาก Repository เป็น error ของ Service
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	// เพิ่ม URL เต็มให้กับทุกรูปภาพก่อนส่งกลับไป
	imageBaseURL := "https://<your-account-name>.blob.core.windows.net/<your-container-name>/"
	for i := range product.Images {
		product.Images[i].URL = imageBaseURL + product.Images[i].Path
	}

	return product, nil
}

func (s *productService) FindAllProducts(params domain.QueryParams) (*domain.PaginatedProductsDTO, error) {
	// 1. ดึงจำนวนสินค้ารวมทั้งหมดจาก Repository
	totalItems, err := s.productRepo.Count()
	if err != nil {
		return nil, err
	}

	// 2. ดึงข้อมูลสินค้าในหน้าที่ต้องการ
	products, err := s.productRepo.FindAll(params)
	if err != nil {
		return nil, err
	}

	// 3. แปลง Domain Model เป็น DTO (เหมือนเดิม)
	imageBaseURL := "https://<your-account-name>.blob.core.windows.net/<your-container-name>/"
	dtos := make([]domain.ProductListDTO, 0, len(products))
	for _, p := range products {
		// สร้าง DTO พร้อมคัดลอกข้อมูลทั้งหมดจาก p มาใส่ทันที
		dto := domain.ProductListDTO{
			ID:           p.ID,
			Name:         p.Name,
			Price:        p.Price,
			SKU:          p.SKU,
			CategoryName: p.Category.Name, // สามารถดึงได้โดยตรงเพราะ Repository ทำ Preload("Category") ไว้แล้ว
		}

		// ค้นหารูปปก (ยังคงเหมือนเดิม)
		if len(p.Images) > 0 {
			// สมมติเอารูปแรกเป็นรูปปก
			dto.PrimaryImageURL = imageBaseURL + p.Images[0].Path
		}

		dtos = append(dtos, dto)
	}

	// 4. คำนวณค่า Pagination
	totalPages := int(math.Ceil(float64(totalItems) / float64(params.Limit)))

	// 5. สร้าง Response DTO สุดท้าย
	paginatedResponse := &domain.PaginatedProductsDTO{
		Data:        dtos,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		CurrentPage: params.Page,
		Limit:       params.Limit,
	}

	return paginatedResponse, nil
}

func (s *productService) DeleteProduct(id uint) error {
	err := s.productRepo.Delete(id)
	if err != nil {
		// แปล error จาก repository เป็น error ของ service
		if errors.Is(err, repository.ErrNotFound) {
			return ErrProductNotFound
		}
		return err
	}
	return nil
}

func (s *productService) UpdateProduct(id uint, updates map[string]interface{}) (*domain.Product, error) {
	// 1. ตรวจสอบว่ามีสินค้านี้อยู่จริงหรือไม่ก่อนทำการอัปเดต
	// โดยการเรียก FindProductByID ซึ่งจะคืนค่า ErrProductNotFound ถ้าไม่มี
	_, err := s.productRepo.FindProductByID(id)
	if err != nil {
		// "แปล" error จาก repository เป็น error ของ service
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	// 2. ถ้ามีอยู่จริง ก็สั่งให้อัปเดตข้อมูล
	if err := s.productRepo.Update(id, updates); err != nil {
		return nil, err
	}

	// 3. ดึงข้อมูลตัวเต็มที่อัปเดตแล้วกลับไปแสดง
	return s.productRepo.FindProductByID(id)
}
