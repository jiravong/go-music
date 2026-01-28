package database // ประกาศ package database

import (
	"fmt" // นำเข้า package fmt สำหรับจัดรูปแบบข้อความ
	"log" // นำเข้า package log สำหรับแสดงข้อความ log
	"os"  // นำเข้า package os สำหรับอ่าน environment variables

	"go-music-api/internal/domain" // นำเข้า domain entities เพื่อใช้ในการ migrate

	"gorm.io/driver/postgres" // นำเข้า postgres driver สำหรับ gorm
	"gorm.io/gorm"            // นำเข้า gorm ORM
)

// NewPostgresDB สร้างการเชื่อมต่อกับฐานข้อมูล Postgres
func NewPostgresDB() (*gorm.DB, error) {
	// สร้าง DSN (Data Source Name) สำหรับเชื่อมต่อฐานข้อมูล
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Bangkok",
		os.Getenv("DB_HOST"),     // อ่าน host จาก env
		os.Getenv("DB_USER"),     // อ่าน user จาก env
		os.Getenv("DB_PASSWORD"), // อ่าน password จาก env
		os.Getenv("DB_NAME"),     // อ่าน db name จาก env
		os.Getenv("DB_PORT"),     // อ่าน port จาก env
	)

	// เปิดการเชื่อมต่อกับฐานข้อมูลโดยใช้ gorm
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// ถ้าเชื่อมต่อไม่สำเร็จ ส่งค่า nil และ error กลับไป
		return nil, err
	}

	// Auto Migrate
	// ทำการ migrate schema อัตโนมัติสำหรับ User และ Music
	err = db.AutoMigrate(&domain.User{}, &domain.Music{})
	if err != nil {
		// ถ้า migrate ไม่สำเร็จ ให้ log error และส่ง error กลับไป
		log.Printf("Failed to auto migrate: %v", err)
		return nil, err
	}

	// ส่งคืน connection ของฐานข้อมูล
	return db, nil
}
