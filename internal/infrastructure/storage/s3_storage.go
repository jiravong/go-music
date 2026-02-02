package storage // ประกาศ package storage

import (
	"context"        // นำเข้า context
	"fmt"            // นำเข้า fmt สำหรับจัดการข้อความ
	"mime/multipart" // นำเข้า multipart สำหรับจัดการไฟล์
	"path/filepath"  // นำเข้า filepath
	"time"           // นำเข้า time

	"github.com/aws/aws-sdk-go-v2/aws"        // นำเข้า aws sdk
	"github.com/aws/aws-sdk-go-v2/config"     // นำเข้า config
	"github.com/aws/aws-sdk-go-v2/service/s3" // นำเข้า s3 service
	"github.com/google/uuid"                  // นำเข้า uuid
)

// S3Storage struct สำหรับจัดการไฟล์บน AWS S3
type S3Storage struct {
	client     *s3.Client // AWS S3 Client
	bucketName string     // ชื่อ Bucket
	region     string     // Region ของ Bucket
}

// NewS3Storage สร้าง instance ใหม่ของ S3Storage
func NewS3Storage(bucketName, region string) (*S3Storage, error) {
	// โหลด Default Configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	// สร้าง S3 Client
	client := s3.NewFromConfig(cfg)

	return &S3Storage{
		client:     client,
		bucketName: bucketName,
		region:     region,
	}, nil
}

// UploadFile อัปโหลดไฟล์ขึ้น S3 และคืนค่า URL
func (s *S3Storage) UploadFile(ctx context.Context, file *multipart.FileHeader) (string, error) {
	// เปิดไฟล์
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// สร้างชื่อไฟล์ใหม่เพื่อไม่ให้ซ้ำกัน
	ext := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uuid.New().String(), ext)

	// อัปโหลดไฟล์ไปยัง S3
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(newFileName),
		Body:        src,
		ContentType: aws.String(file.Header.Get("Content-Type")),
		// ACL:         types.ObjectCannedACLPublicRead, // ถ้าต้องการให้เข้าถึงได้แบบ Public (ต้องตั้งค่า Bucket Policy ด้วย)
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %v", err)
	}

	// สร้าง URL ของไฟล์ (แบบ Virtual-hosted style)
	// รูปแบบ: https://bucket-name.s3.region.amazonaws.com/key
	fileURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucketName, s.region, newFileName)

	return fileURL, nil
}

// DeleteFile ลบไฟล์ออกจาก S3 (ยังไม่ implement logic การ parse key จาก URL ในตอนนี้)
func (s *S3Storage) DeleteFile(ctx context.Context, fileURL string) error {
	// ต้องเขียน logic เพื่อดึง key จาก fileURL ก่อน
	// ตัวอย่าง: https://my-bucket.s3.us-east-1.amazonaws.com/my-file.jpg -> key: my-file.jpg
	return nil
}
