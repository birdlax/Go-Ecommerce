package middleware

import (
	"backend/domain"
	"errors"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JwtClaims คือ struct ที่เราจะใช้เก็บข้อมูลใน Payload ของ JWT
// เราจะฝัง RegisteredClaims เพื่อให้มีฟิลด์มาตรฐานเช่น 'exp' (วันหมดอายุ) มาด้วย
type JwtClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Protected คือ Middleware ที่จะตรวจสอบ JWT
func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. ดึงค่า Authorization Header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		// 2. ตรวจสอบรูปแบบ "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}
		tokenString := parts[1]

		// 3. ดึง JWT Secret Key จาก Environment Variable
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "JWT secret not configured",
			})
		}

		// 4. Parse และ Validate Token
		claims := &JwtClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// ตรวจสอบ Signing Method Algorithm ให้แน่ใจว่าเป็น HS256
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		// 5. จัดการกับ Error ที่เกิดจากการ Parse
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token has expired"})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}

		if !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}

		// 6. ถ้า Token ถูกต้อง ให้เก็บข้อมูล Claims ไว้ใน Context
		// เพื่อให้ Handler ที่ทำงานต่อไปสามารถดึงข้อมูล User ไปใช้ได้
		c.Locals("user", claims)

		// 7. ไปยัง Middleware หรือ Handler ตัวถัดไป
		return c.Next()
	}
}

func AdminRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. ดึงข้อมูล Claims ที่ Middleware 'Protected' ได้เก็บไว้ให้
		claims, ok := c.Locals("user").(*JwtClaims)
		if !ok {
			// นี่เป็นกรณีที่ผิดปกติมาก ไม่ควรจะเกิดขึ้นถ้าเราวาง Middleware ถูกต้อง
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Cannot parse user claims",
			})
		}

		// 2. ตรวจสอบ Role จาก Claims
		// เราเปรียบเทียบกับค่าคงที่ที่เราสร้างไว้เพื่อความปลอดภัยและป้องกันการพิมพ์ผิด
		if claims.Role != string(domain.RoleAdmin) {
			// ถ้า Role ไม่ใช่ 'admin' ให้ส่ง 403 Forbidden กลับไป
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You do not have permission to access this resource",
			})
		}

		// 3. ถ้าเป็น Admin ให้ผ่านไปยัง Handler ตัวถัดไป
		return c.Next()
	}
}
