package storage // ประกาศ package storage

import (
	"context"        // นำเข้า context
	"fmt"            // นำเข้า fmt สำหรับจัดรูปแบบข้อความ
	"io"             // นำเข้า io สำหรับการคัดลอกข้อมูลไฟล์
	"mime/multipart" // นำเข้า multipart สำหรับจัดการไฟล์อัปโหลด
	"os"             // นำเข้า os สำหรับจัดการไฟล์และโฟลเดอร์ในระบบ
	"path/filepath"  // นำเข้า filepath สำหรับจัดการ path ของไฟล์
	"time"           // นำเข้า time

	"github.com/google/uuid" // นำเข้า uuid สำหรับสร้างชื่อไฟล์ที่ไม่ซ้ำกัน
)

// LocalStorage struct เก็บค่า configuration สำหรับการเก็บไฟล์ในเครื่อง
type LocalStorage struct {
	UploadDir string // โฟลเดอร์ที่จะเก็บไฟล์
	BaseURL   string // URL พื้นฐานสำหรับเข้าถึงไฟล์
}

// NewLocalStorage สร้าง instance ของ LocalStorage และตรวจสอบ/สร้างโฟลเดอร์
func NewLocalStorage(uploadDir, baseURL string) (*LocalStorage, error) {
	// สร้างโฟลเดอร์ uploadDir ถ้ายังไม่มีอยู่ (0755 คือ permission)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, err
	}
	return &LocalStorage{
		UploadDir: uploadDir,
		BaseURL:   baseURL,
	}, nil
}

// UploadFile อัปโหลดไฟล์และคืนค่า URL ที่เข้าถึงไฟล์ได้
func (s *LocalStorage) UploadFile(ctx context.Context, file *multipart.FileHeader) (string, error) {
	// เปิดอ่านไฟล์ต้นทาง
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Generate unique filename
	// สร้างชื่อไฟล์ใหม่เพื่อไม่ให้ซ้ำกัน: วันเวลา + UUID + นามสกุลไฟล์เดิม
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s_%s%s", time.Now().Format("20060102150405"), uuid.New().String(), ext)
	filepath := filepath.Join(s.UploadDir, filename)

	// สร้างไฟล์ปลายทาง
	dst, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// คัดลอกข้อมูลจากต้นทางไปปลายทาง
	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	// คืนค่า URL สำหรับเข้าถึงไฟล์
	return fmt.Sprintf("%s/%s", s.BaseURL, filename), nil
}

// DeleteFile ลบไฟล์จากเครื่อง
func (s *LocalStorage) DeleteFile(ctx context.Context, fileURL string) error {
	// ดึงชื่อไฟล์จาก URL (แบบง่าย)
	// ในการใช้งานจริงอาจต้องมีการจัดการ URL parsing ที่ซับซ้อนกว่านี้
	filename := filepath.Base(fileURL)
	// ลบไฟล์ออกจากโฟลเดอร์
	return os.Remove(filepath.Join(s.UploadDir, filename))
}
