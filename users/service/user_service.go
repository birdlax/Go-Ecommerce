package service

import (
	"backend/domain"
	"backend/internal/datastore"
	"backend/users/dto"
	"backend/users/repository"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"math"
	"os"
	"time"
)

// Custom Error
var ErrEmailExists = errors.New("email already exists")
var ErrUserNotFound = errors.New("user not found")

// UserService Interface
type UserService interface {
	FindAllUsers(params dto.UserQueryParams) (*dto.PaginatedUsersDTO, error)
	FindUserByID(id uint) (*dto.UserResponse, error)
	UpdateUser(id uint, req map[string]interface{}) (*dto.UserResponse, error)
	DeleteUser(id uint) error

	// เพิ่ม Method สำหรับ Register,Login, Refresh Token และ Logout
	Register(req dto.RegisterRequest) (*dto.UserResponse, error)
	Login(req dto.LoginRequest) (accessToken string, refreshToken string, err error)
	RefreshToken(tokenString string) (newAccessToken string, err error)
	Logout(hashedToken string) error

	// Address related methods
	AddAddress(userID uint, req dto.AddressRequest) (*domain.Address, error)
	GetUserAddresses(userID uint) ([]domain.Address, error)
	UpdateAddress(userID, addressID uint, req dto.AddressRequest) (*domain.Address, error)
	DeleteAddress(userID, addressID uint) error
}

type userService struct {
	uow datastore.UnitOfWork
}

func NewUserService(uow datastore.UnitOfWork) UserService {
	return &userService{uow: uow}
}

func (s *userService) Register(req dto.RegisterRequest) (*dto.UserResponse, error) {
	// 1. ตรวจสอบ Email ซ้ำ (เหมือนเดิม)
	_, err := s.uow.UserRepository().FindByEmail(req.Email)
	if !errors.Is(err, repository.ErrNotFound) {
		// ถ้า err ไม่ใช่ RecordNotFound แสดงว่ามี user แล้ว หรือเกิด error อื่น
		return nil, ErrEmailExists
	}

	// 2. Hash รหัสผ่าน (เหมือนเดิม)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 3. สร้าง User object ใหม่ด้วยข้อมูลจาก Model ล่าสุด
	newUser := &domain.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      domain.RoleCustomer,
		IsActive:  true,
	}

	err = s.uow.Execute(func(repos *datastore.Repositories) error {
		return repos.User.Create(newUser)
	})
	if err != nil {
		return nil, err
	}

	return mapUserToUserResponse(newUser), nil
}

func (s *userService) FindUserByID(id uint) (*dto.UserResponse, error) {
	user, err := s.uow.UserRepository().FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound // สร้าง ErrUserNotFound คล้ายๆ ErrProductNotFound
		}
		return nil, err
	}
	return mapUserToUserResponse(user), nil
}

func (s *userService) FindAllUsers(params dto.UserQueryParams) (*dto.PaginatedUsersDTO, error) {
	// 1. ดึงจำนวนผู้ใช้รวมทั้งหมด
	totalItems, err := s.uow.UserRepository().Count()
	if err != nil {
		return nil, err
	}

	// 2. ดึงข้อมูลผู้ใช้ในหน้าที่ต้องการ
	users, err := s.uow.UserRepository().FindAll(params)
	if err != nil {
		return nil, err
	}

	// 3. แปลง Domain Model เป็น UserResponse DTO
	userResponses := make([]dto.UserResponse, 0, len(users))
	for _, u := range users {
		userResponses = append(userResponses, *mapUserToUserResponse(&u))
	}

	// 4. คำนวณค่า Pagination
	totalPages := int(math.Ceil(float64(totalItems) / float64(params.Limit)))

	// 5. สร้าง Response DTO สุดท้ายด้วย PaginatedUsersDTO
	paginatedResponse := &dto.PaginatedUsersDTO{
		Data:        userResponses,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		CurrentPage: params.Page,
		Limit:       params.Limit,
	}

	return paginatedResponse, nil
}

func (s *userService) UpdateUser(id uint, updates map[string]interface{}) (*dto.UserResponse, error) {
	user, err := s.uow.UserRepository().FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	// อัปเดตฟิลด์ที่ต้องการ
	if firstName, ok := updates["first_name"].(string); ok {
		user.FirstName = firstName
	}
	if lastName, ok := updates["last_name"].(string); ok {
		user.LastName = lastName
	}
	if email, ok := updates["email"].(string); ok {
		// ตรวจสอบว่า Email ซ้ำหรือไม่
		existingUser, err := s.uow.UserRepository().FindByEmail(email)
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != user.ID {
			return nil, ErrEmailExists // ถ้า Email ซ้ำกับ User อื่น
		}
		user.Email = email
	}
	if password, ok := updates["password"].(string); ok {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashedPassword)
	}
	// อัปเดต Role ถ้ามีการส่งมา
	if role, ok := updates["role"].(string); ok {
		// ควรมีการ validate ว่า role ที่ส่งมาถูกต้องหรือไม่
		user.Role = domain.UserRole(role)
	}

	if err := s.uow.UserRepository().Update(user); err != nil {
		return nil, err
	}
	return mapUserToUserResponse(user), nil
}

func (s *userService) DeleteUser(id uint) error {
	err := s.uow.UserRepository().Delete(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

// Helper function สำหรับแปลง Domain เป็น DTO
func mapUserToUserResponse(user *domain.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      user.Role,
	}
}

func (s *userService) Login(req dto.LoginRequest) (string, string, error) {
	// 1. ค้นหาผู้ใช้และเปรียบเทียบรหัสผ่าน (เหมือนเดิม)
	user, err := s.uow.UserRepository().FindByEmail(req.Email)
	if err != nil {
		return "", "", errors.New("invalid credentials 1")
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", "", errors.New("invalid credentials or password")
	}

	// 2. สร้าง Access Token (อายุสั้น)
	accessToken, err := createAccessToken(user)
	if err != nil {
		return "", "", err
	}

	// 3. สร้าง Refresh Token (อายุยาว)
	// 3.1 สร้าง Token แบบสุ่ม
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", err
	}
	refreshToken := hex.EncodeToString(randomBytes)

	// 3.2 Hash Token ก่อนเก็บลง DB
	hash := sha256.New()
	hash.Write([]byte(refreshToken))
	hashedRefreshToken := hex.EncodeToString(hash.Sum(nil))

	// 3.3 กำหนดวันหมดอายุ
	refreshTokenExpiresAt := time.Now().Add(time.Hour * 24 * 7) // 7 วัน

	// 3.4 บันทึกลง DB
	user.RefreshToken = &hashedRefreshToken
	user.RefreshTokenExpiresAt = &refreshTokenExpiresAt
	if err := s.uow.UserRepository().Update(user); err != nil {
		return "", "", err
	}

	// 4. คืนค่า Access Token (ใน body) และ Refresh Token (สำหรับตั้งเป็น cookie)
	return accessToken, refreshToken, nil
}

func (s *userService) RefreshToken(tokenString string) (string, error) {
	// 1. Hash token ที่ได้รับมาเพื่อนำไปค้นหาใน DB
	hash := sha256.New()
	hash.Write([]byte(tokenString))
	hashedToken := hex.EncodeToString(hash.Sum(nil))

	// 2. ค้นหา User จาก Hashed Refresh Token
	user, err := s.uow.UserRepository().FindByRefreshToken(hashedToken)
	if err != nil {
		return "", errors.New("invalid or expired refresh token")
	}

	// 3. ตรวจสอบว่า Token หมดอายุหรือยัง
	if user.RefreshTokenExpiresAt == nil || time.Now().After(*user.RefreshTokenExpiresAt) {
		return "", errors.New("invalid or expired refresh token")
	}

	// 4. ถ้าทุกอย่างถูกต้อง ให้สร้าง Access Token ตัวใหม่
	return createAccessToken(user)
}

func (s *userService) Logout(tokenString string) error {
	if tokenString == "" {
		return nil // ถ้าไม่มี token ก็ไม่ต้องทำอะไร
	}
	hash := sha256.New()
	hash.Write([]byte(tokenString))
	hashedToken := hex.EncodeToString(hash.Sum(nil))

	user, err := s.uow.UserRepository().FindByRefreshToken(hashedToken)
	if err != nil {
		return nil // หาไม่เจอก็ไม่เป็นไร
	}

	// ล้างค่า Refresh Token ใน DB
	user.RefreshToken = nil
	user.RefreshTokenExpiresAt = nil
	return s.uow.UserRepository().Update(user)
}

func createAccessToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 1).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// address
func (s *userService) AddAddress(userID uint, req dto.AddressRequest) (*domain.Address, error) {
	newAddress := &domain.Address{
		UserID:       userID,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		State:        req.State,
		PostalCode:   req.PostalCode,
		Country:      req.Country,
		IsDefault:    req.IsDefault,
	}

	// [Logic สำคัญ] ถ้าตั้งที่อยู่นี้เป็น Default ต้องไปเคลียร์ Default เก่าก่อน
	if newAddress.IsDefault {
		if err := s.uow.AddressRepository().ClearDefault(userID); err != nil {
			return nil, err
		}
	}

	if err := s.uow.AddressRepository().Create(newAddress); err != nil {
		return nil, err
	}
	return newAddress, nil
}

func (s *userService) GetUserAddresses(userID uint) ([]domain.Address, error) {
	return s.uow.AddressRepository().FindByUserID(userID)
}

func (s *userService) UpdateAddress(userID, addressID uint, req dto.AddressRequest) (*domain.Address, error) {
	address, err := s.uow.AddressRepository().FindByID(addressID)
	if err != nil {
		return nil, errors.New("address not found")
	}

	// ตรวจสอบความเป็นเจ้าของ
	if address.UserID != userID {
		return nil, errors.New("you do not own this address")
	}

	// อัปเดตข้อมูล
	address.AddressLine1 = req.AddressLine1
	// ... อัปเดตฟิลด์อื่นๆ ...
	address.IsDefault = req.IsDefault

	if address.IsDefault {
		if err := s.uow.AddressRepository().ClearDefault(userID); err != nil {
			return nil, err
		}
	}

	if err := s.uow.AddressRepository().Update(address); err != nil {
		return nil, err
	}
	return address, nil
}

func (s *userService) DeleteAddress(userID, addressID uint) error {
	address, err := s.uow.AddressRepository().FindByID(addressID)
	if err != nil {
		return errors.New("address not found")
	}

	if address.UserID != userID {
		return errors.New("you do not own this address")
	}

	return s.uow.AddressRepository().Delete(addressID)
}
