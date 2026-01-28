package domain // ประกาศ package domain

import (
	"context"        // นำเข้า context
	"mime/multipart" // นำเข้า multipart สำหรับจัดการไฟล์
)

// StorageService interface กำหนดเมธอดสำหรับจัดการไฟล์
type StorageService interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader) (string, error) // อัปโหลดไฟล์และคืนค่า URL
	DeleteFile(ctx context.Context, fileURL string) error                       // ลบไฟล์ตาม URL
}
