package middleware // ประกาศ package middleware

import (
	"context"
	"net/http" // นำเข้า package net/http
	"strings"  // นำเข้า strings สำหรับจัดการข้อความ

	"go-music-api/pkg/utils" // นำเข้า utils สำหรับตรวจสอบ JWT

	"github.com/danielgtaylor/huma/v2" // นำเข้า huma
	"github.com/gin-gonic/gin"         // นำเข้า gin
)

// AuthMiddleware ตรวจสอบ JWT token ใน request header (สำหรับ Gin)
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// อ่าน header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// ถ้าไม่มี header ให้ส่ง error 401
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort() // หยุดการทำงานของ handler ถัดไป
			return
		}

		// แยกส่วน Bearer และ Token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// ถ้ารูปแบบไม่ถูกต้อง (ต้องเป็น "Bearer <token>") ให้ส่ง error 401
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// ตรวจสอบความถูกต้องของ Token
		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			// ถ้า token ไม่ถูกต้อง ให้ส่ง error 401
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// บันทึก user_id ลงใน context เพื่อให้ handler ถัดไปใช้งานได้
		c.Set("user_id", claims.UserID)
		// ไปทำงานต่อที่ handler ถัดไป
		c.Next()
	}
}

// HumaAuthMiddleware ตรวจสอบ JWT token (สำหรับ Huma)
func HumaAuthMiddleware(api huma.API) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		// อ่าน header Authorization
		authHeader := ctx.Header("Authorization")
		if authHeader == "" {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Authorization header is required")
			return
		}

		// แยกส่วน Bearer และ Token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		// ตรวจสอบความถูกต้องของ Token
		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "Invalid token")
			return
		}

		// บันทึก user_id ลงใน context
		// Huma ใช้ context.Context ที่อยู่ใน ctx
		// เราต้องสร้าง context ใหม่ที่มี user_id แล้ว set กลับไปที่ ctx
		newCtx := context.WithValue(ctx.Context(), "user_id", claims.UserID)
		ctx = huma.WithContext(ctx, newCtx)

		next(ctx)
	}
}

// CORSMiddleware จัดการ Cross-Origin Resource Sharing (CORS)
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// กำหนด headers เพื่ออนุญาตการเข้าถึงข้ามโดเมน
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // อนุญาตทุก origin (ควรระบุเจาะจงใน production)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		// จัดการ Preflight request (OPTIONS)
		if c.Request.Method == "OPTIONS" {
			// ตอบกลับ status 204 No Content ทันที
			c.AbortWithStatus(204)
			return
		}

		// ไปทำงานต่อที่ handler ถัดไป
		c.Next()
	}
}
