package postgres // ประกาศ package postgres

import (
	"context" // นำเข้า context สำหรับจัดการ timeout และ cancelation
	"errors"  // นำเข้า errors สำหรับตรวจสอบ error type

	"go-music-api/internal/domain" // นำเข้า domain entities

	"gorm.io/gorm" // นำเข้า gorm ORM
)

// musicRepository struct สำหรับ implement interface MusicRepository
type musicRepository struct {
	db *gorm.DB // เก็บ connection ของ database
}

// NewMusicRepository สร้าง instance ของ MusicRepository
func NewMusicRepository(db *gorm.DB) domain.MusicRepository {
	return &musicRepository{db: db}
}

// Create บันทึกข้อมูลเพลงใหม่ลงฐานข้อมูล
func (r *musicRepository) Create(ctx context.Context, music *domain.Music) error {
	// ใช้ WithContext เพื่อให้ gorm เคารพ timeout หรือ cancelation ของ context
	return r.db.WithContext(ctx).Create(music).Error
}

// GetByID ดึงข้อมูลเพลงจาก ID
func (r *musicRepository) GetByID(ctx context.Context, id uint) (*domain.Music, error) {
	var music domain.Music
	// ค้นหาข้อมูล record แรกที่ตรงกับ ID
	if err := r.db.WithContext(ctx).First(&music, id).Error; err != nil {
		// ถ้าไม่พบข้อมูล ให้คืนค่า error เป็น ErrNotFound
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		// คืนค่า error อื่นๆ ถ้ามีปัญหา
		return nil, err
	}
	// คืนค่า pointer ของ music
	return &music, nil
}

// GetAll ดึงข้อมูลเพลงทั้งหมด
func (r *musicRepository) GetAll(ctx context.Context) ([]domain.Music, error) {
	var musics []domain.Music
	// ค้นหาข้อมูลทั้งหมดในตาราง musics
	if err := r.db.WithContext(ctx).Find(&musics).Error; err != nil {
		return nil, err
	}
	// คืนค่า slice ของ music
	return musics, nil
}

// Update อัปเดตข้อมูลเพลง
func (r *musicRepository) Update(ctx context.Context, music *domain.Music) error {
	// บันทึกการเปลี่ยนแปลงข้อมูลทั้งหมดของ object music ลงฐานข้อมูล
	return r.db.WithContext(ctx).Save(music).Error
}

// Delete ลบเพลงตาม ID
func (r *musicRepository) Delete(ctx context.Context, id uint) error {
	// ลบข้อมูลจากตาราง musics โดยระบุ ID
	return r.db.WithContext(ctx).Delete(&domain.Music{}, id).Error
}
