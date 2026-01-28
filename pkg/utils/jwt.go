package utils // ประกาศ package utils

import (
	"errors" // นำเข้า errors
	"time"   // นำเข้า time

	"github.com/golang-jwt/jwt/v5" // นำเข้า jwt library
)

// secretKey สำหรับเซ็นลายเซ็น token (ควรเก็บใน .env ในการใช้งานจริง)
var secretKey = []byte("your-secret-key") // In real app, load from ENV

// Claims struct สำหรับเก็บข้อมูลใน Payload ของ JWT
type Claims struct {
	UserID uint `json:"user_id"` // เก็บ ID ของผู้ใช้
	jwt.RegisteredClaims
}

// GenerateToken สร้าง JWT token สำหรับผู้ใช้
func GenerateToken(userID uint) (string, error) {
	// สร้าง claims
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // หมดอายุใน 24 ชั่วโมง
			IssuedAt:  jwt.NewNumericDate(time.Now()),                     // เวลาที่ออก token
		},
	}

	// สร้าง token ด้วยวิธี signing HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// เซ็นลายเซ็นและคืนค่า token เป็น string
	return token.SignedString(secretKey)
}

// ValidateToken ตรวจสอบความถูกต้องของ JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	// Parse และตรวจสอบ token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// ตรวจสอบว่า signing method ถูกต้องหรือไม่ (ต้องเป็น HMAC)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		// คืนค่า secret key สำหรับการตรวจสอบลายเซ็น
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	// ตรวจสอบ claims และความถูกต้องของ token
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
