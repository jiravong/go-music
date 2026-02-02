package domain

import "time"

// BaseModel struct สำหรับเก็บฟิลด์พื้นฐานที่ใช้ร่วมกัน (Core Model)
type BaseModel struct {
	ID        uint      `json:"id" gorm:"primaryKey"`               // ID (Primary Key)
	CreatedBy string    `json:"created_by" gorm:"default:'system'"` // ผู้สร้าง
	UpdatedBy string    `json:"updated_by" gorm:"default:'system'"` // ผู้แก้ไขล่าสุด
	CreatedAt time.Time `json:"created_at"`                         // เวลาที่สร้าง
	UpdatedAt time.Time `json:"updated_at"`                         // เวลาที่แก้ไขล่าสุด
}
