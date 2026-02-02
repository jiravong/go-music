package handler // ประกาศ package handler

import (
	"context" // นำเข้า context

	"go-music-api/internal/domain" // นำเข้า domain entities

	"github.com/danielgtaylor/huma/v2" // นำเข้า huma
)

// UserHandler struct สำหรับจัดการ HTTP request ที่เกี่ยวกับ User
type UserHandler struct {
	userService domain.UserService // ใช้ service ในการทำงาน
}

// NewUserHandler สร้าง instance ของ UserHandler
func NewUserHandler(userService domain.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// RegisterInput struct สำหรับรับข้อมูลลงทะเบียน
type RegisterInput struct {
	Body struct {
		Email    string `json:"email" required:"true" format:"email" doc:"User email address"`
		Password string `json:"password" required:"true" minLength:"6" doc:"User password (min 6 chars)"`
	}
}

// RegisterOutput struct สำหรับ response การลงทะเบียน
type RegisterOutput struct {
	Body struct {
		Message string `json:"message" doc:"Success message"`
	}
}

// Register ลงทะเบียนผู้ใช้ใหม่
func (h *UserHandler) Register(ctx context.Context, input *RegisterInput) (*RegisterOutput, error) {
	// Map ข้อมูลจาก input ไปยัง domain.User
	user := &domain.User{
		Email:    input.Body.Email,
		Password: input.Body.Password,
	}

	// เรียก service เพื่อลงทะเบียน
	if err := h.userService.Register(ctx, user); err != nil {
		// ถ้ามี email ซ้ำ
		if err == domain.ErrConflict {
			return nil, huma.Error409Conflict("Email already exists")
		}
		return nil, huma.Error500InternalServerError(err.Error())
	}

	// ส่ง response แจ้งว่าลงทะเบียนสำเร็จ
	return &RegisterOutput{
		Body: struct {
			Message string `json:"message" doc:"Success message"`
		}{
			Message: "User registered successfully",
		},
	}, nil
}

// LoginInput struct สำหรับรับข้อมูล Login
type LoginInput struct {
	Body struct {
		Email    string `json:"email" required:"true" format:"email" doc:"User email"`
		Password string `json:"password" required:"true" doc:"User password"`
	}
}

// LoginOutput struct สำหรับ response Login
type LoginOutput struct {
	Body struct {
		AccessToken  string `json:"access_token" doc:"JWT Access Token"`
		RefreshToken string `json:"refresh_token" doc:"JWT Refresh Token"`
	}
}

// Login เข้าสู่ระบบ
func (h *UserHandler) Login(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
	// เรียก service เพื่อเข้าสู่ระบบและรับ token
	accessToken, refreshToken, err := h.userService.Login(ctx, input.Body.Email, input.Body.Password)
	if err != nil {
		// ถ้าข้อมูลไม่ถูกต้อง
		if err == domain.ErrInvalidCreds {
			return nil, huma.Error401Unauthorized("Invalid email or password")
		}
		return nil, huma.Error500InternalServerError(err.Error())
	}

	// ส่ง token กลับไป
	return &LoginOutput{
		Body: struct {
			AccessToken  string `json:"access_token" doc:"JWT Access Token"`
			RefreshToken string `json:"refresh_token" doc:"JWT Refresh Token"`
		}{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}

// RefreshTokenInput struct สำหรับรับ Refresh Token
type RefreshTokenInput struct {
	Body struct {
		RefreshToken string `json:"refresh_token" required:"true" doc:"Refresh Token to get new Access Token"`
	}
}

// RefreshTokenOutput struct สำหรับ response Refresh Token
type RefreshTokenOutput struct {
	Body struct {
		AccessToken string `json:"access_token" doc:"New Access Token"`
	}
}

// RefreshToken ขอ Access Token ใหม่
func (h *UserHandler) RefreshToken(ctx context.Context, input *RefreshTokenInput) (*RefreshTokenOutput, error) {
	accessToken, err := h.userService.RefreshToken(ctx, input.Body.RefreshToken)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid refresh token")
	}

	return &RefreshTokenOutput{
		Body: struct {
			AccessToken string `json:"access_token" doc:"New Access Token"`
		}{
			AccessToken: accessToken,
		},
	}, nil
}
