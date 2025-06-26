// /middleware/error_handler.go (สร้างไฟล์/โฟลเดอร์ใหม่)
package middleware

import (
	"backend/products/service"
	"errors"
	"github.com/gofiber/fiber/v2"
	"log"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	// Logger: บันทึก Error ที่เกิดขึ้นก่อนเสมอ
	log.Printf("An error occurred: %v", err)

	// Error Handler: ตัดสินใจจากประเภทของ Error
	var e *fiber.Error
	if errors.As(err, &e) {
		// ถ้าเป็น Error ของ Fiber เอง ก็ใช้ status code ของ Fiber
		return c.Status(e.Code).JSON(fiber.Map{"error": e.Message})
	}

	if errors.Is(err, service.ErrProductNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	// ... สามารถเพิ่มเงื่อนไข if errors.Is(...) สำหรับ custom error อื่นๆ ได้ที่นี่ ...

	// ถ้าเป็น Error ที่ไม่รู้จัก ให้ถือเป็น Internal Server Error
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": "Internal Server Error",
	})
}
