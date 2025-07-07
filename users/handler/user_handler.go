package handler

import (
	"backend/middleware"
	"backend/users/dto"
	"backend/users/service"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"time"
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

func (h *UserHandler) HandleLogin(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate ข้อมูลหลังจาก Parse แล้ว
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	accessToken, refreshToken, err := h.userSvc.Login(req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	// --- ตั้งค่า Refresh Token เป็น HttpOnly Cookie ---
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(time.Hour * 24 * 7),
		HTTPOnly: true,
		Secure:   true, // ใน Production ควรเป็น true เสมอ
		SameSite: "Strict",
	})

	// ส่ง Access Token กลับไปใน JSON Body
	return c.Status(fiber.StatusOK).JSON(dto.LoginResponse{AccessToken: accessToken})
}

// --- เพิ่ม 2 Handler นี้เข้าไป ---
func (h *UserHandler) HandleRefreshToken(c *fiber.Ctx) error {
	// ดึง Refresh Token จาก Cookie ที่เบราว์เซอร์ส่งมาให้
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "refresh token not found"})
	}

	newAccessToken, err := h.userSvc.RefreshToken(refreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(dto.LoginResponse{AccessToken: newAccessToken})
}

func (h *UserHandler) HandleLogout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if err := h.userSvc.Logout(refreshToken); err != nil {
		return err // ให้ Error Middleware จัดการ
	}

	// ลบ Cookie ที่ฝั่ง Client โดยการส่ง Cookie ชื่อเดิมที่หมดอายุไปแล้ว
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // ตั้งเวลาในอดีต
		HTTPOnly: true,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "logged out successfully"})
}

func (h *UserHandler) HandleGetMyProfile(c *fiber.Ctx) error {
	// 1. ดึงข้อมูล Claims ที่ Middleware เก็บไว้ใน Locals context
	// เราต้องทำ Type Assertion เพื่อแปลง interface{} กลับมาเป็น *middleware.JwtClaims
	claims, ok := c.Locals("user").(*middleware.JwtClaims)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "Cannot parse user claims")
	}

	// 2. ตอนนี้เรามี UserID ที่เชื่อถือได้จาก Token แล้ว
	userID := claims.UserID

	// 3. เรียกใช้ Service เพื่อดึงข้อมูล User
	userResponse, err := h.userSvc.FindUserByID(userID)
	if err != nil {
		return err // ส่งให้ Error Middleware จัดการ
	}

	return c.Status(fiber.StatusOK).JSON(userResponse)
}

// HandleGetUserByID สำหรับ Admin ดึงข้อมูลผู้ใช้รายบุคคล
func (h *UserHandler) HandleGetUserByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID format")
	}

	userResponse, err := h.userSvc.FindUserByID(uint(id))
	if err != nil {
		return err // ส่งต่อให้ Centralized Error Handler จัดการ
	}

	return c.Status(fiber.StatusOK).JSON(userResponse)
}

// HandleUpdateUser สำหรับ Admin อัปเดตข้อมูลผู้ใช้ (เช่น Role, IsActive)
func (h *UserHandler) HandleUpdateUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID format")
	}

	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot parse JSON")
	}

	// ลบ field ที่ไม่ควรให้แก้ไขผ่าน API นี้ได้
	delete(updates, "id")
	delete(updates, "password")
	delete(updates, "email")

	updatedUser, err := h.userSvc.UpdateUser(uint(id), updates)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(updatedUser)
}

// HandleDeleteUser สำหรับ Admin ทำการ Soft Delete ผู้ใช้
func (h *UserHandler) HandleDeleteUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID format")
	}

	if err := h.userSvc.DeleteUser(uint(id)); err != nil {
		return err
	}

	// 204 No Content เป็นการตอบกลับที่เหมาะสมสำหรับการลบที่สำเร็จ
	return c.SendStatus(fiber.StatusNoContent)
}

// address related handlers
func (h *UserHandler) HandleAddAddress(c *fiber.Ctx) error {
	claims := c.Locals("user").(*middleware.JwtClaims) // ดึง user id จาก token
	var req dto.AddressRequest

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}
	address, err := h.userSvc.AddAddress(claims.UserID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(address)
}

func (h *UserHandler) HandleGetUserAddresses(c *fiber.Ctx) error {
	claims := c.Locals("user").(*middleware.JwtClaims)
	addresses, err := h.userSvc.GetUserAddresses(claims.UserID)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(addresses)
}

func (h *UserHandler) HandleUpdateAddress(c *fiber.Ctx) error {
	claims := c.Locals("user").(*middleware.JwtClaims)
	addressID, err := c.ParamsInt("addressId")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid address ID format")
	}

	var req dto.AddressRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	address, err := h.userSvc.UpdateAddress(claims.UserID, uint(addressID), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(address)
}
func (h *UserHandler) HandleDeleteAddress(c *fiber.Ctx) error {
	claims := c.Locals("user").(*middleware.JwtClaims)
	addressID, err := c.ParamsInt("addressId")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid address ID format")
	}

	if err := h.userSvc.DeleteAddress(claims.UserID, uint(addressID)); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent) // 204 No Content
}
