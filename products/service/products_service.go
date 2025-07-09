package service

import (
	"backend/domain"
	"backend/internal/datastore"
	"backend/products/dto"
	"backend/products/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"math"
	"path/filepath"
)

var ErrProductNotFound = errors.New("product not found")

// ===================================================================
// ProductService Interface (อัปเดต Return Types ให้เป็น DTO)
// ===================================================================
type ProductService interface {
	CreateProductWithImages(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error)
	FindAllProducts(params dto.QueryParams) (*dto.PaginatedProductsDTO, error)
	FindProductByID(id uint) (*dto.ProductResponse, error)
	DeleteProduct(id uint) error
	UpdateProduct(id uint, updates map[string]interface{}) (*dto.ProductResponse, error)
	UpdateProductImages(ctx context.Context, productID uint, req dto.UpdateImagesRequest) error
}

// ===================================================================
// productService Implementation
// ===================================================================

type productService struct {
	uow          datastore.UnitOfWork
	imageBaseURL string
}

// NewProductService Constructor
func NewProductService(uow datastore.UnitOfWork, imageBaseURL string) ProductService {
	return &productService{
		uow:          uow,
		imageBaseURL: imageBaseURL,
	}
}

// --- เมธอดต่างๆ ที่แก้ไขให้เรียกใช้ Repository ผ่าน UoW ทั้งหมด ---

func (s *productService) CreateProductWithImages(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	var productData dto.CreateProductRequestData
	if err := json.Unmarshal([]byte(req.Data), &productData); err != nil {
		return nil, fmt.Errorf("invalid product data json: %w", err)
	}

	productToCreate := &domain.Product{
		Name:        productData.Name,
		Description: productData.Description,
		Price:       productData.Price,
		Quantity:    productData.Quantity,
		SKU:         productData.SKU,
		CategoryID:  productData.CategoryID,
	}

	uploadedFilePaths := make(map[string]string) // temp -> final

	// 1. อัปโหลดไฟล์ทั้งหมดไปที่ชั่วคราวก่อน
	for _, fileInput := range req.Files {
		tempPath := fmt.Sprintf("temp/%s%s", uuid.New().String(), filepath.Ext(fileInput.Filename))
		_, err := s.uow.UploadRepository().UploadFile(ctx, tempPath, fileInput.Content, fileInput.ContentType)
		if err != nil {
			return nil, fmt.Errorf("failed to upload file %s: %w", fileInput.Filename, err)
		}
		uploadedFilePaths[tempPath] = ""
	}

	// 2. ทำงานกับ Database ทั้งหมดใน Transaction เดียว
	err := s.uow.Execute(func(repos *datastore.Repositories) error {
		if err := repos.Product.CreateProduct(productToCreate); err != nil {
			return err
		}

		var newImagesData []domain.ProductImage
		i := 0
		for tempPath := range uploadedFilePaths {
			finalPath := fmt.Sprintf("products/%d/%s", productToCreate.ID, filepath.Base(tempPath))
			uploadedFilePaths[tempPath] = finalPath
			newImagesData = append(newImagesData, domain.ProductImage{
				ProductID: productToCreate.ID,
				Path:      finalPath,
				IsPrimary: i == 0,
			})
			i++
		}

		if len(newImagesData) > 0 {
			if err := repos.Product.CreateImages(newImagesData); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		for tempPath := range uploadedFilePaths {
			s.uow.UploadRepository().DeleteFile(context.Background(), tempPath)
		}
		return nil, fmt.Errorf("database transaction failed: %w", err)
	}

	// 3. ย้ายไฟล์ใน Azure
	for tempPath, finalPath := range uploadedFilePaths {
		if err := s.uow.UploadRepository().MoveFile(ctx, tempPath, finalPath); err != nil {
			log.Printf("WARNING: failed to move blob from %s to %s: %v", tempPath, finalPath, err)
		}
	}

	// 4. ดึงข้อมูลล่าสุดกลับมาในรูปแบบ DTO
	return s.FindProductByID(productToCreate.ID)
}

func (s *productService) FindProductByID(id uint) (*dto.ProductResponse, error) {
	product, err := s.uow.ProductRepository().FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return mapProductToProductResponse(product, s.imageBaseURL), nil
}

func (s *productService) FindAllProducts(params dto.QueryParams) (*dto.PaginatedProductsDTO, error) {
	var paginatedResponse *dto.PaginatedProductsDTO
	err := s.uow.Execute(func(repos *datastore.Repositories) error {
		totalItems, err := repos.Product.Count(params)
		if err != nil {
			return err
		}

		products, err := repos.Product.FindAll(params)
		if err != nil {
			return err
		}

		dtos := make([]dto.ProductListDTO, 0, len(products))
		for _, p := range products {
			dto := dto.ProductListDTO{
				ID:           p.ID,
				Name:         p.Name,
				Price:        p.Price,
				SKU:          p.SKU,
				CategoryName: p.Category.Name,
			}
			if len(p.Images) > 0 {
				dto.PrimaryImageURL = s.imageBaseURL + "/" + p.Images[0].Path
			}
			dtos = append(dtos, dto)
		}

		totalPages := int(math.Ceil(float64(totalItems) / float64(params.Limit)))
		paginatedResponse = &dto.PaginatedProductsDTO{
			Data:        dtos,
			TotalItems:  totalItems,
			TotalPages:  totalPages,
			CurrentPage: params.Page,
			Limit:       params.Limit,
		}
		return nil
	})
	return paginatedResponse, err
}

func (s *productService) UpdateProduct(id uint, updates map[string]interface{}) (*dto.ProductResponse, error) {
	err := s.uow.Execute(func(repos *datastore.Repositories) error {
		_, err := repos.Product.FindProductByID(id)
		if err != nil {
			return err
		}
		return repos.Product.Update(id, updates)
	})

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return s.FindProductByID(id)
}

func (s *productService) DeleteProduct(id uint) error {
	err := s.uow.Execute(func(repos *datastore.Repositories) error {
		return repos.Product.Delete(id)
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrProductNotFound
		}
		return err
	}
	return nil
}

func (s *productService) UpdateProductImages(ctx context.Context, productID uint, req dto.UpdateImagesRequest) error {
	var pathsToDeleteFromStorage []string

	err := s.uow.Execute(func(repos *datastore.Repositories) error {
		if len(req.ImageIDsToDelete) > 0 {
			imagesToDelete, err := repos.Product.FindImagesByIDs(req.ImageIDsToDelete)
			if err != nil {
				return err
			}
			for _, img := range imagesToDelete {
				pathsToDeleteFromStorage = append(pathsToDeleteFromStorage, img.Path)
			}
			if err := repos.Product.DeleteImagesByIDs(req.ImageIDsToDelete); err != nil {
				return err
			}
		}

		if len(req.FilesToAdd) > 0 {
			var newImagesData []domain.ProductImage
			for _, fileInput := range req.FilesToAdd {
				ext := filepath.Ext(fileInput.Filename)
				newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
				blobPath := fmt.Sprintf("products/%d/%s", productID, newFileName)

				_, err := s.uow.UploadRepository().UploadFile(ctx, blobPath, fileInput.Content, fileInput.ContentType)
				if err != nil {
					return err
				}

				newImagesData = append(newImagesData, domain.ProductImage{
					ProductID: productID,
					Path:      blobPath,
					IsPrimary: false,
				})
			}
			if err := repos.Product.CreateImages(newImagesData); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	for _, path := range pathsToDeleteFromStorage {
		if err := s.uow.UploadRepository().DeleteFile(ctx, path); err != nil {
			log.Printf("WARNING: Failed to delete old blob object '%s': %v", path, err)
		}
	}
	return nil
}

// ===================================================================
// Helper functions
// ===================================================================

func mapProductToProductResponse(product *domain.Product, imageBaseURL string) *dto.ProductResponse {
	imagesDto := make([]dto.ImageResponse, 0, len(product.Images))
	for _, img := range product.Images {
		imagesDto = append(imagesDto, dto.ImageResponse{
			ID:        img.ID,
			URL:       imageBaseURL + "/" + img.Path,
			IsPrimary: img.IsPrimary,
		})
	}

	return &dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Quantity:    product.Quantity,
		SKU:         product.SKU,
		Category: dto.CategoryResponse{
			ID:   product.Category.ID,
			Name: product.Category.Name,
		},
		Images:    imagesDto,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
	}
}
