package service

import (
	"backend/categories/dto"
	"backend/categories/repository"
	"backend/domain"
	"backend/internal/datastore"
	"errors"
)

var ErrCategoryNotFound = errors.New("category not found")

type CategoryService interface {
	Create(req dto.CategoryRequest) (*dto.CategoryResponse, error)
	GetAll() ([]dto.CategoryResponse, error)
	GetByID(id uint) (*dto.CategoryResponse, error)
	Update(id uint, req dto.CategoryRequest) (*dto.CategoryResponse, error)
	Delete(id uint) error
}

type categoryService struct {
	uow datastore.UnitOfWork
}

func NewCategoryService(uow datastore.UnitOfWork) CategoryService {
	return &categoryService{uow: uow}
}

func (s *categoryService) Create(req dto.CategoryRequest) (*dto.CategoryResponse, error) {
	newCategory := &domain.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	err := s.uow.Execute(func(repos *datastore.Repositories) error {
		return repos.Category.Create(newCategory)
	})

	if err != nil {
		return nil, err
	}
	return mapCategoryToResponse(newCategory), nil
}

func (s *categoryService) GetAll() ([]dto.CategoryResponse, error) {
	categories, err := s.uow.CategoryRepository().FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]dto.CategoryResponse, 0, len(categories))
	for _, category := range categories {
		responses = append(responses, *mapCategoryToResponse(&category))
	}
	return responses, nil
}

func (s *categoryService) GetByID(id uint) (*dto.CategoryResponse, error) {
	category, err := s.uow.CategoryRepository().FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return mapCategoryToResponse(category), nil
}

func (s *categoryService) Update(id uint, req dto.CategoryRequest) (*dto.CategoryResponse, error) {
	var updatedCategory *domain.Category
	err := s.uow.Execute(func(repos *datastore.Repositories) error {
		category, err := repos.Category.FindByID(id)
		if err != nil {
			return err
		}
		category.Name = req.Name
		category.Description = req.Description
		updatedCategory = category
		return repos.Category.Update(category)
	})

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return mapCategoryToResponse(updatedCategory), nil
}

func (s *categoryService) Delete(id uint) error {
	err := s.uow.Execute(func(repos *datastore.Repositories) error {
		return repos.Category.Delete(id)
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	return nil
}

// Helper function
func mapCategoryToResponse(category *domain.Category) *dto.CategoryResponse {
	return &dto.CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
	}
}
