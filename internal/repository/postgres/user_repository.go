package postgres // ประกาศ package postgres

import (
	"context" // นำเข้า context
	"errors"  // นำเข้า errors

	"go-music-api/internal/domain" // นำเข้า domain entities

	"gorm.io/gorm" // นำเข้า gorm ORM
)

// userRepository struct สำหรับ implement interface UserRepository
type userRepository struct {
	db *gorm.DB // เก็บ connection ของ database
}

// NewUserRepository สร้าง instance ของ UserRepository
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// Create สร้างผู้ใช้ใหม่
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	// บันทึกข้อมูล user ลงฐานข้อมูล
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		// ตรวจสอบว่า error คือ key ซ้ำหรือไม่ (เช่น email ซ้ำ)
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrConflict // คืนค่า ErrConflict ถ้าข้อมูลซ้ำ
		}
		return err
	}
	return nil
}

// GetByEmail ค้นหาผู้ใช้จากอีเมล
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	// ค้นหา record แรกที่ email ตรงกัน
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		// ถ้าไม่พบข้อมูล
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByID ค้นหาผู้ใช้จาก ID
func (r *userRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	// ค้นหา record แรกที่ id ตรงกัน
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		// ถ้าไม่พบข้อมูล
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdateProfile(ctx context.Context, id uint, updates map[string]any) error {
	res := r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
