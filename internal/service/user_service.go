package service // ประกาศ package service

import (
	"context" // นำเข้า context
	"time"    // นำเข้า time

	"go-music-api/internal/domain" // นำเข้า domain entities
	"go-music-api/pkg/utils"       // นำเข้า utils สำหรับช่วยจัดการ JWT

	"golang.org/x/crypto/bcrypt" // นำเข้า bcrypt สำหรับเข้ารหัสรหัสผ่าน
)

// userService struct สำหรับ implement interface UserService
type userService struct {
	userRepo domain.UserRepository // repository สำหรับจัดการข้อมูลผู้ใช้
	timeout  time.Duration         // ระยะเวลา timeout
}

// NewUserService สร้าง instance ของ UserService
func NewUserService(userRepo domain.UserRepository, timeout time.Duration) domain.UserService {
	return &userService{
		userRepo: userRepo,
		timeout:  timeout,
	}
}

// Register ลงทะเบียนผู้ใช้ใหม่
func (s *userService) Register(ctx context.Context, user *domain.User) error {
	// สร้าง context ที่มี timeout
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// เข้ารหัสรหัสผ่านด้วย bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	// แทนที่รหัสผ่านเดิมด้วยรหัสผ่านที่เข้ารหัสแล้ว
	user.Password = string(hashedPassword)

	// บันทึกข้อมูลผู้ใช้ลงฐานข้อมูล
	return s.userRepo.Create(ctx, user)
}

// Login ตรวจสอบข้อมูลการเข้าสู่ระบบและสร้าง Token
func (s *userService) Login(ctx context.Context, email, password string) (string, string, error) {
	// สร้าง context ที่มี timeout
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// ค้นหาผู้ใช้จากอีเมล
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// ถ้าหาไม่เจอ ให้คืนค่า error ว่า credentials ไม่ถูกต้อง
		return "", "", domain.ErrInvalidCreds
	}

	// ตรวจสอบรหัสผ่านว่าตรงกับที่เข้ารหัสไว้หรือไม่
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// ถ้ารหัสผ่านไม่ตรง ให้คืนค่า error ว่า credentials ไม่ถูกต้อง
		return "", "", domain.ErrInvalidCreds
	}

	// สร้าง JWT token pair (Access Token และ Refresh Token)
	accessToken, refreshToken, err := utils.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return "", "", err
	}

	// คืนค่า token ทั้งคู่
	return accessToken, refreshToken, nil
}

// RefreshToken ตรวจสอบ Refresh Token และสร้าง Access Token ใหม่
func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// ตรวจสอบความถูกต้องของ Refresh Token
	claims, err := utils.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	// สร้าง Token Pair ใหม่ (จริงๆ เราต้องการแค่ Access Token ใหม่ แต่ใช้ฟังก์ชันเดิมเพื่อความสะดวก)
	// หมายเหตุ: ในระบบจริงอาจจะมีการตรวจสอบเพิ่มเติม เช่น Blacklist หรือเช็คว่า user ยัง active อยู่ไหม
	accessToken, _, err := utils.GenerateTokenPair(claims.UserID, claims.Email)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}
