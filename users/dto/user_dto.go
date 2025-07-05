package dto

import (
	"backend/domain"
	"time"
)

// [แก้ไข] RegisterRequest จะรับ FirstName และ LastName แทน Name
type RegisterRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2"`
	LastName  string `json:"last_name" validate:"required,min=2"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
}

type UserResponse struct {
	ID        uint            `json:"id"`
	FirstName string          `json:"first_name"`
	LastName  string          `json:"last_name"`
	Email     string          `json:"email"`
	Role      domain.UserRole `json:"role"`
	IsActive  bool            `json:"is_active"`
	CreatedAt time.Time       `json:"created_at"`
}

// UpdateUserRequest คือ DTO สำหรับรับข้อมูลตอน `PATCH /users/{id}`
// เราใช้ Pointer (*) เพื่อให้สามารถแยกแยะได้ว่าฟิลด์ไหนที่ Client "ไม่ได้ส่งมา" (เป็น nil)
// กับฟิลด์ที่ "ส่งมาแต่เป็นค่าว่าง" (เช่น "") ซึ่งสำคัญมากสำหรับ PATCH
type UpdateUserRequest struct {
	FirstName *string          `json:"first_name" validate:"omitempty,min=2"`
	LastName  *string          `json:"last_name" validate:"omitempty,min=2"`
	IsActive  *bool            `json:"is_active"`
	Role      *domain.UserRole `json:"role" validate:"omitempty,oneof=customer admin staff"` // oneof คือการบังคับว่าค่าต้องเป็นหนึ่งในนี้เท่านั้น
}
type PaginatedUsersResponse struct {
	Data        []UserResponse `json:"data"`
	TotalItems  int64          `json:"total_items"`
	TotalPages  int            `json:"total_pages"`
	CurrentPage int            `json:"current_page"`
	Limit       int            `json:"limit"`
}

// เพิ่ม DTO สำหรับรับข้อมูลตอน Login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// DTO สำหรับส่ง Token กลับไป
type LoginResponse struct {
	AccessToken string `json:"access_token"`
}
