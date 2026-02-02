package domain // ประกาศ package domain สำหรับเก็บ entities และ interfaces

import (
	"context" // นำเข้า context สำหรับจัดการ timeout และ cancelation
)

// User struct เก็บข้อมูลผู้ใช้งาน
type User struct {
	BaseModel
	Email    string `json:"email" gorm:"unique;not null"` // อีเมล ต้องไม่ซ้ำและห้ามว่าง
	Password string `json:"-" gorm:"not null"`            // รหัสผ่าน (ไม่ส่งกลับไปใน JSON)
}

// UserRepository interface กำหนดเมธอดสำหรับจัดการข้อมูล User ในฐานข้อมูล
type UserRepository interface {
	Create(ctx context.Context, user *User) error                // สร้างผู้ใช้ใหม่
	GetByEmail(ctx context.Context, email string) (*User, error) // ค้นหาผู้ใช้ด้วยอีเมล
	GetByID(ctx context.Context, id uint) (*User, error)         // ค้นหาผู้ใช้ด้วย ID
}

// UserService interface กำหนดเมธอดสำหรับ business logic ของ User
type UserService interface {
	Register(ctx context.Context, user *User) error                            // ลงทะเบียนผู้ใช้
	Login(ctx context.Context, email, password string) (string, string, error) // เข้าสู่ระบบ (คืนค่า access token และ refresh token)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)     // ขอ Access Token ใหม่ด้วย Refresh Token
}
