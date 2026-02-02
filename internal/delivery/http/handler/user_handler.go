package handler // ประกาศ package handler

import (
	"net/http" // นำเข้า package net/http

	"go-music-api/internal/domain" // นำเข้า domain entities

	"github.com/gin-gonic/gin" // นำเข้า gin
)

// UserHandler struct สำหรับจัดการ HTTP request ที่เกี่ยวกับ User
type UserHandler struct {
	userService domain.UserService // ใช้ service ในการทำงาน
}

// NewUserHandler สร้าง instance ของ UserHandler
func NewUserHandler(userService domain.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Register ลงทะเบียนผู้ใช้ใหม่ (POST /auth/register)
func (h *UserHandler) Register(c *gin.Context) {
	// ใช้ struct แยกสำหรับรับข้อมูล เพื่อแก้ปัญหา json:"-" ใน domain.User ที่ทำให้รับ password ไม่ได้
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	// Bind JSON body เข้ากับตัวแปร req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Map ข้อมูลจาก req ไปยัง domain.User
	user := &domain.User{
		Email:    req.Email,
		Password: req.Password,
	}

	// เรียก service เพื่อลงทะเบียน
	if err := h.userService.Register(c.Request.Context(), user); err != nil {
		// ถ้ามี email ซ้ำ
		if err == domain.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ส่ง response แจ้งว่าลงทะเบียนสำเร็จ
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// Login เข้าสู่ระบบ (POST /auth/login)
func (h *UserHandler) Login(c *gin.Context) {
	// struct สำหรับรับค่า request body
	var req struct {
		Email    string `json:"email" binding:"required,email"` // อีเมล จำเป็นต้องมีและรูปแบบถูกต้อง
		Password string `json:"password" binding:"required"`    // รหัสผ่าน จำเป็นต้องมี
	}

	// Bind JSON body เข้ากับตัวแปร req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// เรียก service เพื่อเข้าสู่ระบบและรับ token
	accessToken, refreshToken, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		// ถ้าข้อมูลไม่ถูกต้อง
		if err == domain.ErrInvalidCreds {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ส่ง token กลับไป
	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// RefreshToken ขอ Access Token ใหม่ (POST /auth/refresh-token)
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, err := h.userService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}
