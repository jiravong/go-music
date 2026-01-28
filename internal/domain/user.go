package domain // ประกาศ package domain สำหรับเก็บ entities และ interfaces

import (
	"context" // นำเข้า context สำหรับจัดการ timeout และ cancelation
	"time"    // นำเข้า time สำหรับจัดการเวลา
)

// User struct เก็บข้อมูลผู้ใช้งาน
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`         // ID ของผู้ใช้ (Primary Key)
	Email     string    `json:"email" gorm:"unique;not null"` // อีเมล ต้องไม่ซ้ำและห้ามว่าง
	Password  string    `json:"-" gorm:"not null"`            // รหัสผ่าน (ไม่ส่งกลับไปใน JSON)
	CreatedAt time.Time `json:"created_at"`                   // เวลาที่สร้าง
	UpdatedAt time.Time `json:"updated_at"`                   // เวลาที่แก้ไขล่าสุด
}

// UserRepository interface กำหนดเมธอดสำหรับจัดการข้อมูล User ในฐานข้อมูล
type UserRepository interface {
	Create(ctx context.Context, user *User) error                // สร้างผู้ใช้ใหม่
	GetByEmail(ctx context.Context, email string) (*User, error) // ค้นหาผู้ใช้ด้วยอีเมล
	GetByID(ctx context.Context, id uint) (*User, error)         // ค้นหาผู้ใช้ด้วย ID
}

// UserService interface กำหนดเมธอดสำหรับ business logic ของ User
type UserService interface {
	Register(ctx context.Context, user *User) error                    // ลงทะเบียนผู้ใช้
	Login(ctx context.Context, email, password string) (string, error) // เข้าสู่ระบบ (คืนค่า token)
}
