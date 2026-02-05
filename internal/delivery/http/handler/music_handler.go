package handler // ประกาศ package handler

import (
	"mime/multipart" // นำเข้า multipart สำหรับจัดการไฟล์
	"net/http"       // นำเข้า net/http
	"strconv"        // นำเข้า strconv

	"go-music-api/internal/domain" // นำเข้า domain entities

	"os"
	"strings"

	"github.com/gin-gonic/gin" // นำเข้า gin
)

func publicBaseURL() string {
	base := os.Getenv("PUBLIC_BASE_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	return strings.TrimRight(base, "/")
}

func toPublicURL(path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return publicBaseURL() + path
}

func hydrateMusicMediaURLs(m *domain.Music) {
	if m == nil {
		return
	}
	m.MP3URL = toPublicURL(m.MP3URL)
	m.MP4URL = toPublicURL(m.MP4URL)
	m.ImageURL = toPublicURL(m.ImageURL)
}

func hydrateMusicListMediaURLs(items []domain.Music) {
	for i := range items {
		hydrateMusicMediaURLs(&items[i])
	}
}

// MusicHandler struct สำหรับจัดการ HTTP request ที่เกี่ยวกับ Music
type MusicHandler struct {
	musicService domain.MusicService // ใช้ service ในการทำงาน
}

// NewMusicHandler สร้าง instance ของ MusicHandler
func NewMusicHandler(musicService domain.MusicService) *MusicHandler {
	return &MusicHandler{musicService: musicService}
}

type updateMusicRequest struct {
	Title  *string `json:"title"`
	Artist *string `json:"artist"`
	Lyrics *string `json:"lyrics"`
}

// Create จัดการ request สำหรับสร้างเพลงใหม่
func (h *MusicHandler) Create(c *gin.Context) {
	email, ok := c.Get("email")
	createdEmail, ok2 := email.(string)
	if !ok || !ok2 || createdEmail == "" {
		createdEmail = "system"
	}

	if form, err := c.MultipartForm(); err == nil && form != nil {
		if files := form.File["mp3_file"]; len(files) > 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "mp3_file must be a single file"})
			return
		}
		if files := form.File["mp4_file"]; len(files) > 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "mp4_file must be a single file"})
			return
		}
		if files := form.File["image"]; len(files) > 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "image must be a single file"})
			return
		}
	}

	title := c.PostForm("title")
	artist := c.PostForm("artist")
	lyrics := c.PostForm("lyrics")
	if title == "" || artist == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title and artist are required"})
		return
	}

	var mp3File *multipart.FileHeader
	if fh, err := c.FormFile("mp3_file"); err == nil {
		mp3File = fh
	} else if err != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var mp4File *multipart.FileHeader
	if fh, err := c.FormFile("mp4_file"); err == nil {
		mp4File = fh
	} else if err != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var imageFile *multipart.FileHeader
	if fh, err := c.FormFile("image"); err == nil {
		imageFile = fh
	} else if err != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	const maxUploadFileSize = 10 << 20
	if mp3File != nil && mp3File.Size > maxUploadFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "mp3_file is too large (max 10MB)"})
		return
	}
	if mp4File != nil && mp4File.Size > maxUploadFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "mp4_file is too large (max 10MB)"})
		return
	}
	if imageFile != nil && imageFile.Size > maxUploadFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image is too large (max 10MB)"})
		return
	}

	music := &domain.Music{
		Title:  title,
		Artist: artist,
		Lyrics: lyrics,
		BaseModel: domain.BaseModel{
			CreatedBy: createdEmail,
			UpdatedBy: createdEmail,
		},
	}

	if err := h.musicService.Create(c.Request.Context(), music, mp3File, mp4File, imageFile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hydrateMusicMediaURLs(music)
	c.JSON(http.StatusCreated, gin.H{"music": music})
}

// GetByID ดึงข้อมูลเพลงตาม ID
func (h *MusicHandler) GetByID(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	music, err := h.musicService.GetByID(c.Request.Context(), uint(id64))
	if err != nil {
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Music not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hydrateMusicMediaURLs(music)
	c.JSON(http.StatusOK, gin.H{"data": music})
}

// GetAll ดึงข้อมูลเพลงทั้งหมด
func (h *MusicHandler) GetAll(c *gin.Context) {
	musics, err := h.musicService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hydrateMusicListMediaURLs(musics)
	c.JSON(http.StatusOK, gin.H{"data": musics})
}

// Update แก้ไขข้อมูลเพลง
func (h *MusicHandler) Update(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	email, ok := c.Get("email")
	updatedEmail, ok2 := email.(string)
	if !ok || !ok2 || updatedEmail == "" {
		updatedEmail = "system"
	}

	existing, err := h.musicService.GetByID(c.Request.Context(), uint(id64))
	if err != nil {
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Music not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if form, err := c.MultipartForm(); err == nil && form != nil {
			if files := form.File["mp3_file"]; len(files) > 1 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "mp3_file must be a single file"})
				return
			}
			if files := form.File["mp4_file"]; len(files) > 1 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "mp4_file must be a single file"})
				return
			}
			if files := form.File["image"]; len(files) > 1 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "image must be a single file"})
				return
			}
		}

		title := c.PostForm("title")
		artist := c.PostForm("artist")
		lyrics := c.PostForm("lyrics")

		var mp3File *multipart.FileHeader
		if fh, err := c.FormFile("mp3_file"); err == nil {
			mp3File = fh
		} else if err != http.ErrMissingFile {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var mp4File *multipart.FileHeader
		if fh, err := c.FormFile("mp4_file"); err == nil {
			mp4File = fh
		} else if err != http.ErrMissingFile {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var imageFile *multipart.FileHeader
		if fh, err := c.FormFile("image"); err == nil {
			imageFile = fh
		} else if err != http.ErrMissingFile {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		const maxUploadFileSize = 10 << 20
		if mp3File != nil && mp3File.Size > maxUploadFileSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "mp3_file is too large (max 10MB)"})
			return
		}
		if mp4File != nil && mp4File.Size > maxUploadFileSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "mp4_file is too large (max 10MB)"})
			return
		}
		if imageFile != nil && imageFile.Size > maxUploadFileSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image is too large (max 10MB)"})
			return
		}

		merged := &domain.Music{
			BaseModel: domain.BaseModel{
				ID:        existing.ID,
				CreatedBy: existing.CreatedBy,
				UpdatedBy: updatedEmail,
			},
			Title:    existing.Title,
			Artist:   existing.Artist,
			Lyrics:   existing.Lyrics,
			MP3URL:   existing.MP3URL,
			MP4URL:   existing.MP4URL,
			ImageURL: existing.ImageURL,
		}
		if title != "" {
			merged.Title = title
		}
		if artist != "" {
			merged.Artist = artist
		}
		if lyrics != "" {
			merged.Lyrics = lyrics
		}

		if err := h.musicService.Update(c.Request.Context(), merged, mp3File, mp4File, imageFile); err != nil {
			if err == domain.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Music not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		updated, err := h.musicService.GetByID(c.Request.Context(), uint(id64))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		hydrateMusicMediaURLs(updated)
		c.JSON(http.StatusOK, gin.H{"data": updated})
		return
	}

	var req updateMusicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	merged := &domain.Music{
		BaseModel: domain.BaseModel{
			ID:        existing.ID,
			CreatedBy: existing.CreatedBy,
			UpdatedBy: updatedEmail,
		},
		Title:    existing.Title,
		Artist:   existing.Artist,
		Lyrics:   existing.Lyrics,
		MP3URL:   existing.MP3URL,
		MP4URL:   existing.MP4URL,
		ImageURL: existing.ImageURL,
	}
	if req.Title != nil {
		merged.Title = *req.Title
	}
	if req.Artist != nil {
		merged.Artist = *req.Artist
	}
	if req.Lyrics != nil {
		merged.Lyrics = *req.Lyrics
	}

	if err := h.musicService.Update(c.Request.Context(), merged, nil, nil, nil); err != nil {
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Music not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.musicService.GetByID(c.Request.Context(), uint(id64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	hydrateMusicMediaURLs(updated)
	c.JSON(http.StatusOK, gin.H{"data": updated})
}

// Delete ลบเพลง
func (h *MusicHandler) Delete(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.musicService.Delete(c.Request.Context(), uint(id64)); err != nil {
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Music not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Music deleted successfully"})
}
