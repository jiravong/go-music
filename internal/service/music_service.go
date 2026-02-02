package service // ประกาศ package service

import (
	"context"        // นำเข้า context
	"mime/multipart" // นำเข้า multipart สำหรับจัดการไฟล์
	"time"           // นำเข้า time

	"go-music-api/internal/domain" // นำเข้า domain entities
)

// musicService struct สำหรับ implement interface MusicService
type musicService struct {
	musicRepo domain.MusicRepository // repository สำหรับจัดการข้อมูลเพลง
	storage   domain.StorageService  // service สำหรับจัดการไฟล์
	timeout   time.Duration          // ระยะเวลา timeout สำหรับ context
}

// NewMusicService สร้าง instance ของ MusicService
func NewMusicService(musicRepo domain.MusicRepository, storage domain.StorageService, timeout time.Duration) domain.MusicService {
	return &musicService{
		musicRepo: musicRepo,
		storage:   storage,
		timeout:   timeout,
	}
}

// Create สร้างเพลงใหม่พร้อมอัปโหลดไฟล์
func (s *musicService) Create(ctx context.Context, music *domain.Music, mp3File, mp4File *multipart.FileHeader) error {
	// สร้าง context ใหม่ที่มี timeout เพื่อป้องกันการทำงานนานเกินไป
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel() // ยกเลิก context เมื่อฟังก์ชันทำงานเสร็จ

	// ตรวจสอบว่ามีการอัปโหลดไฟล์ MP3 หรือไม่
	if mp3File != nil {
		// อัปโหลดไฟล์ MP3
		url, err := s.storage.UploadFile(ctx, mp3File)
		if err != nil {
			return err
		}
		// บันทึก URL ของไฟล์ MP3 ลงใน object music
		music.MP3URL = url
	}

	// ตรวจสอบว่ามีการอัปโหลดไฟล์ MP4 หรือไม่
	if mp4File != nil {
		// อัปโหลดไฟล์ MP4
		url, err := s.storage.UploadFile(ctx, mp4File)
		if err != nil {
			// ถ้าอัปโหลด MP4 ล้มเหลว อาจจะต้องลบ MP3 ที่อัปโหลดไปแล้ว (แต่ในที่นี้ละเว้นไว้เพื่อความเรียบง่าย)
			return err
		}
		// บันทึก URL ของไฟล์ MP4 ลงใน object music
		music.MP4URL = url
	}

	// บันทึกข้อมูลเพลงลงฐานข้อมูล
	return s.musicRepo.Create(ctx, music)
}

// GetByID ดึงข้อมูลเพลงตาม ID
func (s *musicService) GetByID(ctx context.Context, id uint) (*domain.Music, error) {
	// สร้าง context ที่มี timeout
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	// เรียก repository เพื่อดึงข้อมูล
	return s.musicRepo.GetByID(ctx, id)
}

// GetAll ดึงข้อมูลเพลงทั้งหมด
func (s *musicService) GetAll(ctx context.Context) ([]domain.Music, error) {
	// สร้าง context ที่มี timeout
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	// เรียก repository เพื่อดึงข้อมูลทั้งหมด
	return s.musicRepo.GetAll(ctx)
}

// Update อัปเดตข้อมูลเพลง
func (s *musicService) Update(ctx context.Context, music *domain.Music) error {
	// สร้าง context ที่มี timeout
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// ตรวจสอบว่ามีเพลงนี้อยู่ในระบบหรือไม่ และดึงข้อมูลเดิมมา
	existingMusic, err := s.musicRepo.GetByID(ctx, music.ID)
	if err != nil {
		return err
	}

	// อัปเดตเฉพาะข้อมูลที่แก้ไขได้
	existingMusic.Title = music.Title
	existingMusic.Artist = music.Artist
	existingMusic.Lyrics = music.Lyrics
	existingMusic.UpdatedBy = music.UpdatedBy
	// existingMusic.UpdatedAt จะถูกจัดการโดย GORM หรือเราจะ set เองก็ได้ แต่ GORM จัดการให้

	// บันทึกข้อมูลที่อัปเดตแล้วลงฐานข้อมูล
	return s.musicRepo.Update(ctx, existingMusic)
}

// Delete ลบเพลงตาม ID
func (s *musicService) Delete(ctx context.Context, id uint) error {
	// สร้าง context ที่มี timeout
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// ดึงข้อมูลเพลงเพื่อเอา URL ของไฟล์
	music, err := s.musicRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// ลบไฟล์ MP3 จาก storage ถ้ามี
	if music.MP3URL != "" {
		_ = s.storage.DeleteFile(ctx, music.MP3URL)
	}
	// ลบไฟล์ MP4 จาก storage ถ้ามี
	if music.MP4URL != "" {
		_ = s.storage.DeleteFile(ctx, music.MP4URL)
	}

	// ลบข้อมูลเพลงจากฐานข้อมูล
	return s.musicRepo.Delete(ctx, id)
}
