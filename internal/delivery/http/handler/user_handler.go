package handler // ประกาศ package handler

import (
	"net/http" // นำเข้า net/http

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

type registerRequest struct {
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=6"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	ImageProfile string `json:"image_profile"`
}

// Register ลงทะเบียนผู้ใช้ใหม่
func (h *UserHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &domain.User{
		Email:        req.Email,
		Password:     req.Password,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		ImageProfile: req.ImageProfile,
	}

	if err := h.userService.Register(c.Request.Context(), user); err != nil {
		if err == domain.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login เข้าสู่ระบบ
func (h *UserHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err == domain.ErrInvalidCreds {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken ขอ Access Token ใหม่
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req refreshTokenRequest
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

func (h *UserHandler) GetMe(c *gin.Context) {
	userIDAny, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDAny.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

type updateMeRequest struct {
	FirstName    *string `json:"first_name"`
	LastName     *string `json:"last_name"`
	ImageProfile *string `json:"image_profile"`
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	userIDAny, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDAny.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req updateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]any{}
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.ImageProfile != nil {
		updates["image_profile"] = *req.ImageProfile
	}

	if emailAny, ok := c.Get("email"); ok {
		if email, ok := emailAny.(string); ok && email != "" {
			updates["updated_by"] = email
		}
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, updates)
	if err != nil {
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}
