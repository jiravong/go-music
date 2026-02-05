package domain // ประกาศ package domain

import (
	"context"        // นำเข้า context
	"mime/multipart" // นำเข้า multipart สำหรับจัดการไฟล์อัปโหลด
)

// Music struct เก็บข้อมูลเพลง
type Music struct {
	BaseModel
	Title    string `json:"title" gorm:"not null"`  // ชื่อเพลง
	Artist   string `json:"artist" gorm:"not null"` // ชื่อศิลปิน
	Lyrics   string `json:"lyrics"`                 // เนื้อเพลง
	MP3URL   string `json:"mp3_url"`                // URL ไฟล์ MP3
	MP4URL   string `json:"mp4_url"`                // URL ไฟล์ MP4
	ImageURL string `json:"image_url"`              // URL รูปหน้าปก
}

// MusicRepository interface กำหนดเมธอดสำหรับจัดการข้อมูล Music ในฐานข้อมูล
type MusicRepository interface {
	Create(ctx context.Context, music *Music) error       // สร้างเพลงใหม่
	GetByID(ctx context.Context, id uint) (*Music, error) // ดึงข้อมูลเพลงตาม ID
	GetAll(ctx context.Context) ([]Music, error)          // ดึงข้อมูลเพลงทั้งหมด
	Update(ctx context.Context, music *Music) error       // อัปเดตข้อมูลเพลง
	Delete(ctx context.Context, id uint) error            // ลบเพลง
}

// MusicService interface กำหนดเมธอดสำหรับ business logic ของ Music
type MusicService interface {
	Create(ctx context.Context, music *Music, mp3File, mp4File, imageFile *multipart.FileHeader) error // สร้างเพลงพร้อมอัปโหลดไฟล์
	GetByID(ctx context.Context, id uint) (*Music, error)                                              // ดึงข้อมูลเพลงตาม ID
	GetAll(ctx context.Context) ([]Music, error)                                                       // ดึงข้อมูลเพลงทั้งหมด
	Update(ctx context.Context, music *Music, mp3File, mp4File, imageFile *multipart.FileHeader) error // อัปเดตข้อมูลเพลง
	Delete(ctx context.Context, id uint) error                                                         // ลบเพลง
}
