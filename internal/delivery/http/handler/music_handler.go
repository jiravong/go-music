package handler // ประกาศ package handler

import (
	"context"        // นำเข้า context
	"mime/multipart" // นำเข้า multipart สำหรับจัดการไฟล์

	"go-music-api/internal/domain" // นำเข้า domain entities

	"github.com/danielgtaylor/huma/v2" // นำเข้า huma
)

// MusicHandler struct สำหรับจัดการ HTTP request ที่เกี่ยวกับ Music
type MusicHandler struct {
	musicService domain.MusicService // ใช้ service ในการทำงาน
}

// NewMusicHandler สร้าง instance ของ MusicHandler
func NewMusicHandler(musicService domain.MusicService) *MusicHandler {
	return &MusicHandler{musicService: musicService}
}

// CreateMusicInput struct สำหรับรับข้อมูลสร้างเพลง
type CreateMusicInput struct {
	Body struct {
		Title   string                `form:"title" required:"true" doc:"Song title"`
		Artist  string                `form:"artist" required:"true" doc:"Artist name"`
		Lyrics  string                `form:"lyrics" doc:"Song lyrics"`
		MP3File *multipart.FileHeader `form:"mp3_file" doc:"MP3 Audio file"`
		MP4File *multipart.FileHeader `form:"mp4_file" doc:"MP4 Video file"`
	} `contentType:"multipart/form-data"`
}

// CreateMusicOutput struct สำหรับ response การสร้างเพลง
type CreateMusicOutput struct {
	Body struct {
		Music *domain.Music `json:"music" doc:"Created music object"`
	}
}

// Create จัดการ request สำหรับสร้างเพลงใหม่
func (h *MusicHandler) Create(ctx context.Context, input *CreateMusicInput) (*CreateMusicOutput, error) {
	// ดึง email จาก context
	email, ok := ctx.Value("email").(string)
	if !ok {
		email = "system"
	}

	// สร้าง object Music
	music := &domain.Music{
		Title:  input.Body.Title,
		Artist: input.Body.Artist,
		Lyrics: input.Body.Lyrics,
		BaseModel: domain.BaseModel{
			CreatedBy: email,
			UpdatedBy: email,
		},
	}

	// เรียก service เพื่อสร้างเพลงและอัปโหลดไฟล์
	if err := h.musicService.Create(ctx, music, input.Body.MP3File, input.Body.MP4File); err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}

	// ส่ง response กลับ
	return &CreateMusicOutput{
		Body: struct {
			Music *domain.Music `json:"music" doc:"Created music object"`
		}{
			Music: music,
		},
	}, nil
}

// GetMusicInput struct สำหรับรับ ID เพลง
type GetMusicInput struct {
	ID uint `path:"id" required:"true" doc:"Music ID"`
}

// GetMusicOutput struct สำหรับ response ข้อมูลเพลง
type GetMusicOutput struct {
	Body struct {
		Music *domain.Music `json:"music" doc:"Music object"`
	}
}

// GetByID ดึงข้อมูลเพลงตาม ID
func (h *MusicHandler) GetByID(ctx context.Context, input *GetMusicInput) (*GetMusicOutput, error) {
	// เรียก service เพื่อค้นหาเพลง
	music, err := h.musicService.GetByID(ctx, input.ID)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, huma.Error404NotFound("Music not found")
		}
		return nil, huma.Error500InternalServerError(err.Error())
	}

	// ส่งข้อมูลเพลงกลับ
	return &GetMusicOutput{
		Body: struct {
			Music *domain.Music `json:"music" doc:"Music object"`
		}{
			Music: music,
		},
	}, nil
}

// GetAllMusicOutput struct สำหรับ response รายการเพลง
type GetAllMusicOutput struct {
	Body struct {
		Musics []domain.Music `json:"musics" doc:"List of music objects"`
	}
}

// GetAll ดึงข้อมูลเพลงทั้งหมด
func (h *MusicHandler) GetAll(ctx context.Context, input *struct{}) (*GetAllMusicOutput, error) {
	// เรียก service เพื่อดึงเพลงทั้งหมด
	musics, err := h.musicService.GetAll(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}

	// ส่งรายการเพลงกลับ
	return &GetAllMusicOutput{
		Body: struct {
			Musics []domain.Music `json:"musics" doc:"List of music objects"`
		}{
			Musics: musics,
		},
	}, nil
}

// UpdateMusicInput struct สำหรับรับข้อมูลแก้ไขเพลง
type UpdateMusicInput struct {
	ID   uint `path:"id" required:"true" doc:"Music ID"`
	Body struct {
		Title  string `json:"title" doc:"Song title"`
		Artist string `json:"artist" doc:"Artist name"`
		Lyrics string `json:"lyrics" doc:"Song lyrics"`
	}
}

// UpdateMusicOutput struct สำหรับ response การแก้ไขเพลง
type UpdateMusicOutput struct {
	Body struct {
		Music *domain.Music `json:"music" doc:"Updated music object"`
	}
}

// Update แก้ไขข้อมูลเพลง
func (h *MusicHandler) Update(ctx context.Context, input *UpdateMusicInput) (*UpdateMusicOutput, error) {
	// ดึง email จาก context
	email, ok := ctx.Value("email").(string)
	if !ok {
		email = "system"
	}

	// Map ข้อมูลจาก input ไปยัง domain.Music
	music := &domain.Music{
		BaseModel: domain.BaseModel{
			ID:        input.ID,
			UpdatedBy: email,
		},
		Title:  input.Body.Title,
		Artist: input.Body.Artist,
		Lyrics: input.Body.Lyrics,
	}

	// เรียก service เพื่ออัปเดตข้อมูล
	if err := h.musicService.Update(ctx, music); err != nil {
		if err == domain.ErrNotFound {
			return nil, huma.Error404NotFound("Music not found")
		}
		return nil, huma.Error500InternalServerError(err.Error())
	}

	// ส่งข้อมูลที่อัปเดตแล้วกลับ
	return &UpdateMusicOutput{
		Body: struct {
			Music *domain.Music `json:"music" doc:"Updated music object"`
		}{
			Music: music,
		},
	}, nil
}

// DeleteMusicInput struct สำหรับรับ ID เพลงที่จะลบ
type DeleteMusicInput struct {
	ID uint `path:"id" required:"true" doc:"Music ID"`
}

// DeleteMusicOutput struct สำหรับ response การลบเพลง
type DeleteMusicOutput struct {
	Body struct {
		Message string `json:"message" doc:"Success message"`
	}
}

// Delete ลบเพลง
func (h *MusicHandler) Delete(ctx context.Context, input *DeleteMusicInput) (*DeleteMusicOutput, error) {
	// เรียก service เพื่อลบเพลง
	if err := h.musicService.Delete(ctx, input.ID); err != nil {
		if err == domain.ErrNotFound {
			return nil, huma.Error404NotFound("Music not found")
		}
		return nil, huma.Error500InternalServerError(err.Error())
	}

	// ส่งข้อความยืนยันการลบ
	return &DeleteMusicOutput{
		Body: struct {
			Message string `json:"message" doc:"Success message"`
		}{
			Message: "Music deleted successfully",
		},
	}, nil
}
