// products/service/products_service.go
package service

import (
	"backend/products/domain"
	"backend/products/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	ReplaceProductImage(ctx context.Context, productID, imageID uint, file io.Reader, fileHeader *multipart.FileHeader) (*domain.ProductImage, error)
	UpdateProductImages(ctx context.Context, productID uint, req UpdateImagesRequest) error
}

type CreateProductRequest struct {
	Data  string // JSON string of product data
	Files []*multipart.FileHeader
}

type productService struct {
	productRepo repository.ProductRepository
	uploadRepo  repository.UploadRepository
	// และเก็บ Unit of Work สำหรับงาน Write ที่ต้องการ Transaction
	uow repository.UnitOfWork
}

func NewProductService(
	productRepo repository.ProductRepository,
	uploadRepo repository.UploadRepository,
	uow repository.UnitOfWork,
) ProductService {
	return &productService{
		productRepo: productRepo,
		uploadRepo:  uploadRepo,
		uow:         uow,
	}
}

type UpdateImagesRequest struct {
	FilesToAdd       []*multipart.FileHeader
	ImageIDsToDelete []uint
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
	imageBaseURL := "https://goecommerce.blob.core.windows.net/uploads/"
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
	imageBaseURL := "https://goecommerce.blob.core.windows.net/uploads/"
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

func (s *productService) ReplaceProductImage(ctx context.Context, productID, imageID uint, file io.Reader, fileHeader *multipart.FileHeader) (*domain.ProductImage, error) {
	// 1. ค้นหาข้อมูลรูปภาพเดิม เพื่อตรวจสอบและเอา path เก่ามาใช้
	oldImage, err := s.productRepo.FindImageByID(imageID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProductNotFound // หรืออาจจะสร้าง ErrImageNotFound แยก
		}
		return nil, err
	}

	// ตรวจสอบว่ารูปนี้เป็นของ Product นี้จริงหรือไม่
	if oldImage.ProductID != productID {
		return nil, errors.New("image does not belong to the specified product")
	}
	oldPath := oldImage.Path

	// 2. อัปโหลดไฟล์ใหม่ไปที่ Azure
	ext := filepath.Ext(fileHeader.Filename)
	newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	newPath := fmt.Sprintf("products/%d/%s", productID, newFileName)

	_, err = s.uploadRepo.UploadFile(ctx, newPath, file, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("failed to upload new file: %w", err)
	}

	// 3. อัปเดต Path ใหม่ลงใน Database
	if err := s.productRepo.UpdateImagePath(imageID, newPath); err != nil {
		// TODO: ถ้าขั้นตอนนี้ล้มเหลว ควรจะมี Logic ไปลบไฟล์ใหม่ที่เพิ่งอัปโหลดไปในข้อ 2 (Rollback)
		return nil, fmt.Errorf("failed to update image path in db: %w", err)
	}

	// 4. ลบไฟล์เก่าออกจาก Azure Storage
	if oldPath != "" {
		if err := s.uploadRepo.DeleteFile(ctx, oldPath); err != nil {
			// ถ้าลบไฟล์เก่าไม่สำเร็จ ควรทำอย่างไร? ส่วนใหญ่เราจะแค่ Log error ไว้
			// เพราะข้อมูลใน DB ถูกต้องแล้ว การมีไฟล์ขยะค้างอยู่ไม่กระทบระบบหลัก
			log.Printf("WARNING: Failed to delete old blob object '%s': %v", oldPath, err)
		}
	}

	// 5. ดึงข้อมูลรูปภาพล่าสุดมาคืนค่า
	updatedImage, err := s.productRepo.FindImageByID(imageID)
	if err != nil {
		return nil, ErrProductNotFound
	}

	// เพิ่ม Full URL ก่อนส่งกลับ
	imageBaseURL := "https://goecommerce.blob.core.windows.net/uploads/"
	updatedImage.URL = imageBaseURL + updatedImage.Path

	return updatedImage, nil
}

func (s *productService) UpdateProductImages(ctx context.Context, productID uint, req UpdateImagesRequest) error {
	// Logic การอัปโหลดไฟล์ไป Azure ยังคงทำนอก Transaction
	// เพราะเราไม่อยากให้ DB transaction ค้างนานระหว่างรออัปโหลด
	var newImagesData []domain.ProductImage
	if len(req.FilesToAdd) > 0 {
		for _, fileHeader := range req.FilesToAdd {
			file, err := fileHeader.Open()
			if err != nil {
				return err
			}
			defer file.Close()

			ext := filepath.Ext(fileHeader.Filename)
			newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
			blobPath := fmt.Sprintf("products/%d/%s", productID, newFileName)

			_, err = s.uploadRepo.UploadFile(ctx, blobPath, file, fileHeader.Header.Get("Content-Type"))
			if err != nil {
				return err
			}

			newImagesData = append(newImagesData, domain.ProductImage{
				ProductID: productID,
				Path:      blobPath,
				IsPrimary: false,
			})
		}
	}

	var pathsToDelete []string

	// --- เริ่มการทำงานกับ Database ผ่าน Unit of Work ---
	err := s.uow.Execute(func(repos *repository.Repositories) error {
		// Logic การลบและเพิ่มจะเกิดในนี้ โดยเรียกใช้ repo จาก UoW
		// repos.Product ตอนนี้คือ instance ที่ทำงานบน transaction

		if len(req.ImageIDsToDelete) > 0 {
			imagesToDelete, err := repos.Product.FindImagesByIDs(req.ImageIDsToDelete)
			if err != nil {
				return err
			}
			for _, img := range imagesToDelete {
				pathsToDelete = append(pathsToDelete, img.Path)
			}
			if err := repos.Product.DeleteImagesByIDs(req.ImageIDsToDelete); err != nil {
				return err
			}
		}

		if len(newImagesData) > 0 {
			if err := repos.Product.CreateImages(newImagesData); err != nil {
				return err
			}
		}

		return nil // คืนค่า nil เพื่อ Commit Transaction
	})

	if err != nil {
		// TODO: ควรมี Logic ลบไฟล์ที่เพิ่งอัปโหลดไป ถ้า DB Transaction ล้มเหลว
		return err
	}

	// ลบไฟล์เก่าออกจาก Azure หลังจาก DB Transaction สำเร็จ
	for _, path := range pathsToDelete {
		if err := s.uploadRepo.DeleteFile(ctx, path); err != nil {
			log.Printf("WARNING: Failed to delete old blob object '%s': %v", path, err)
		}
	}

	return nil
}
