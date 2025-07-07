// domain/user.go
package domain

import (
	"gorm.io/gorm"
	"time"
)

type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleCustomer UserRole = "customer"
)

type User struct {
	gorm.Model
	FirstName              string     `json:"first_name" gorm:"type:varchar(100)"`
	LastName               string     `json:"last_name" gorm:"type:varchar(100)"`
	Email                  string     `json:"email" gorm:"type:varchar(100);uniqueIndex;not null"`
	Password               string     `json:"-" gorm:"not null"` // json:"-" คือการบอกว่าห้ามส่งฟิลด์นี้กลับไปใน JSON เด็ดขาด
	PhoneNumber            *string    `json:"phone_number" gorm:"type:varchar(20);uniqueIndex"`
	ProfilePicturePath     string     `json:"profile_picture_path"`
	Role                   UserRole   `json:"role" gorm:"type:varchar(20);default:'customer';not null"`
	IsActive               bool       `json:"is_active" gorm:"default:true;not null"`
	VerifiedAt             *time.Time `json:"verified_at"`
	LastLoginAt            *time.Time `json:"last_login_at"`
	PasswordResetToken     *string    `json:"-"`
	PasswordResetExpiresAt *time.Time `json:"-"`
	Addresses              []Address  `json:"addresses" gorm:"foreignKey:UserID"`
	RefreshToken           *string    `json:"-" gorm:"uniqueIndex"`
	RefreshTokenExpiresAt  *time.Time `json:"-"`
	Cart                   *Cart      `json:"cart" gorm:"foreignKey:UserID"`
}
