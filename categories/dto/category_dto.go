package dto

// CategoryRequest คือ DTO สำหรับรับข้อมูลตอนสร้างหรืออัปเดต Category
type CategoryRequest struct {
	Name        string `json:"name" validate:"required,min=3"`
	Description string `json:"description"`
}

// CategoryResponse คือ DTO สำหรับส่งข้อมูล Category กลับไป
type CategoryResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
