// products/service/products_service.go
package service

import (
	"backend/products/domain"
	"backend/products/repository"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"mime/multipart"
	"path/filepath"
)

// ProductService Interface (ที่เราเคยออกแบบไว้)
type ProductService interface {
	CreateProductWithImages(ctx context.Context, req CreateProductRequest) (*domain.Product, error)
	CreateNewCategory(req domain.Category) error
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
