package handler

import (
	"backend/categories/dto"
	"backend/categories/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type CategoryHandler struct {
	categorySvc service.CategoryService
}

func NewCategoryHandler(categorySvc service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categorySvc: categorySvc}
}

func (h *CategoryHandler) HandleCreateCategory(c *fiber.Ctx) error {
	var req dto.CategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	if err := validator.New().Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	res, err := h.categorySvc.Create(req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(res)
}

func (h *CategoryHandler) HandleGetAllCategories(c *fiber.Ctx) error {
	res, err := h.categorySvc.GetAll()
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *CategoryHandler) HandleGetCategoryByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid category ID")
	}
	res, err := h.categorySvc.GetByID(uint(id))
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *CategoryHandler) HandleUpdateCategory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid category ID")
	}
	var req dto.CategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	if err := validator.New().Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	res, err := h.categorySvc.Update(uint(id), req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *CategoryHandler) HandleDeleteCategory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid category ID")
	}
	if err := h.categorySvc.Delete(uint(id)); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
