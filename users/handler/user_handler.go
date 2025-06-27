package handler

import (
	"backend/users/dto"
	"backend/users/service"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userSvc service.UserService
}

// NewUserHandler Constructor
func NewUserHandler(userSvc service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

func (h *UserHandler) HandleRegister(c *fiber.Ctx) error {
	var req dto.RegisterRequest // <-- [แก้ไข] ใช้ DTO จาก package ใหม่
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	// ส่วนที่เหลือเหมือนเดิม
	userResponse, err := h.userSvc.Register(req)
	if err != nil {
		if errors.Is(err, service.ErrEmailExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(userResponse)
}
func (h *UserHandler) HandleGetAllUsers(c *fiber.Ctx) error {
	// 1. สร้าง QueryParams พร้อมตั้งค่า Default
	params := dto.UserQueryParams{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),          // อาจจะใช้ limit น้อยกว่า Product
		SortBy: c.Query("sort_by", "created_at"), // Default เรียงตามวันที่สร้าง
		Order:  c.Query("order", "desc"),
	}

	// 2. Validate ค่าที่รับเข้ามา
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 || params.Limit > 100 { // กำหนด Limit สูงสุด
		params.Limit = 10
	}
	// ตรวจสอบค่าที่อนุญาตสำหรับ SortBy ของ User
	allowedSorts := map[string]bool{"name": true, "email": true, "created_at": true}
	if !allowedSorts[params.SortBy] {
		params.SortBy = "created_at"
	}
	if params.Order != "asc" && params.Order != "desc" {
		params.Order = "desc"
	}

	// 3. เรียกใช้ Service
	paginatedResult, err := h.userSvc.FindAllUsers(params)
	if err != nil {
		return err // ส่งให้ Error Middleware จัดการ
	}

	return c.Status(fiber.StatusOK).JSON(paginatedResult)
}
