package domain // ประกาศ package domain

import "errors" // นำเข้า errors standard library

// กำหนดตัวแปร error มาตรฐานที่ใช้ในโปรเจค
var (
	ErrNotFound     = errors.New("record not found")      // ไม่พบข้อมูล
	ErrConflict     = errors.New("record already exists") // ข้อมูลซ้ำ
	ErrInternal     = errors.New("internal server error") // ข้อผิดพลาดภายในเซิร์ฟเวอร์
	ErrInvalidCreds = errors.New("invalid credentials")   // รหัสผ่านหรือข้อมูลยืนยันตัวตนไม่ถูกต้อง
	ErrUnauthorized = errors.New("unauthorized")          // ไม่มีสิทธิ์เข้าถึง
)
