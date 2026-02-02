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
	UserID uint   `json:"user_id"` // เก็บ ID ของผู้ใช้
	Email  string `json:"email"`   // เก็บ Email ของผู้ใช้
	jwt.RegisteredClaims
}

// GenerateTokenPair สร้าง Access Token และ Refresh Token
func GenerateTokenPair(userID uint, email string) (string, string, error) {
	// สร้าง Access Token (อายุสั้น เช่น 15 นาที)
	accessTokenClaims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // หมดอายุใน 15 นาที
			IssuedAt:  jwt.NewNumericDate(time.Now()),                       // เวลาที่ออก token
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString(secretKey)
	if err != nil {
		return "", "", err
	}

	// สร้าง Refresh Token (อายุนาน เช่น 7 วัน)
	refreshTokenClaims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // หมดอายุใน 7 วัน
			IssuedAt:  jwt.NewNumericDate(time.Now()),                         // เวลาที่ออก token
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString(secretKey)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
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
