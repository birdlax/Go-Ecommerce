package dto

// UserQueryParams คือ Struct สำหรับรับค่าจาก URL Query ของ User
type UserQueryParams struct {
	Page   int
	Limit  int
	SortBy string
	Order  string
}

// PaginatedUsersDTO คือ DTO สำหรับ Response ที่มีข้อมูลการแบ่งหน้าของ User
type PaginatedUsersDTO struct {
	Data        []UserResponse `json:"data"` // <-- ข้อมูลจะเป็น []UserResponse
	TotalItems  int64          `json:"total_items"`
	TotalPages  int            `json:"total_pages"`
	CurrentPage int            `json:"current_page"`
	Limit       int            `json:"limit"`
}
