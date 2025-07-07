package dto

// AddressRequest คือ DTO สำหรับรับข้อมูลที่อยู่จาก Client
type AddressRequest struct {
	AddressLine1 string `json:"address_line_1" validate:"required"`
	AddressLine2 string `json:"address_line_2"`
	City         string `json:"city" validate:"required"`
	State        string `json:"state" validate:"required"`       // จังหวัด
	PostalCode   string `json:"postal_code" validate:"required"` // รหัสไปรษณีย์
	Country      string `json:"country" validate:"required"`
	IsDefault    bool   `json:"is_default"` // รับค่าว่าต้องการตั้งเป็นที่อยู่หลักหรือไม่
}
