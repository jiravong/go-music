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
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Lyrics string `json:"lyrics"`
}

// Create จัดการ request สำหรับสร้างเพลงใหม่
func (h *MusicHandler) Create(c *gin.Context) {
	email, ok := c.Get("email")
	createdEmail, ok2 := email.(string)
	if !ok || !ok2 || createdEmail == "" {
		createdEmail = "system"
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

	const maxUploadFileSize = 10 << 20
	if mp3File != nil && mp3File.Size > maxUploadFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "mp3_file is too large (max 10MB)"})
		return
	}
	if mp4File != nil && mp4File.Size > maxUploadFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "mp4_file is too large (max 10MB)"})
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

	if err := h.musicService.Create(c.Request.Context(), music, mp3File, mp4File); err != nil {
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
	c.JSON(http.StatusOK, gin.H{"music": music})
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

	var req updateMusicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	music := &domain.Music{
		BaseModel: domain.BaseModel{
			ID:        uint(id64),
			UpdatedBy: updatedEmail,
		},
		Title:  req.Title,
		Artist: req.Artist,
		Lyrics: req.Lyrics,
	}

	if err := h.musicService.Update(c.Request.Context(), music); err != nil {
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
	c.JSON(http.StatusOK, gin.H{"music": updated})
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
