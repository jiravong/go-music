package handler // ประกาศ package handler

import (
	"net/http" // นำเข้า package net/http สำหรับจัดการ status codes
	"strconv"  // นำเข้า strconv สำหรับแปลง string เป็น int

	"go-music-api/internal/domain" // นำเข้า domain entities

	"github.com/gin-gonic/gin" // นำเข้า gin web framework
)

// MusicHandler struct สำหรับจัดการ HTTP request ที่เกี่ยวกับ Music
type MusicHandler struct {
	musicService domain.MusicService // ใช้ service ในการทำงาน
}

// NewMusicHandler สร้าง instance ของ MusicHandler
func NewMusicHandler(musicService domain.MusicService) *MusicHandler {
	return &MusicHandler{musicService: musicService}
}

// Create จัดการ request สำหรับสร้างเพลงใหม่ (POST /music)
func (h *MusicHandler) Create(c *gin.Context) {
	// Multipart form parsing
	// กำหนดขนาดสูงสุดของไฟล์ที่อนุญาตให้ parse คือ 32MB
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB max memory
		// ถ้า parse ไม่ผ่าน ให้ส่ง error 400
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse multipart form"})
		return
	}

	// รับค่าจาก form data
	title := c.PostForm("title")
	artist := c.PostForm("artist")
	lyrics := c.PostForm("lyrics")

	// ตรวจสอบข้อมูลที่จำเป็น
	if title == "" || artist == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title and Artist are required"})
		return
	}

	// รับไฟล์ MP3 จาก form data
	mp3File, err := c.FormFile("mp3_file")
	if err != nil && err != http.ErrMissingFile {
		// ถ้ามี error และไม่ใช่กรณีที่ไม่มีไฟล์ส่งมา ให้ส่ง error 400
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid MP3 file"})
		return
	}

	// รับไฟล์ MP4 จาก form data
	mp4File, err := c.FormFile("mp4_file")
	if err != nil && err != http.ErrMissingFile {
		// ถ้ามี error และไม่ใช่กรณีที่ไม่มีไฟล์ส่งมา ให้ส่ง error 400
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid MP4 file"})
		return
	}

	// ถ้าไม่ได้ส่งไฟล์มา ตัวแปร file จะเป็น nil ซึ่ง service จะจัดการต่อเอง

	// สร้าง object Music
	music := &domain.Music{
		Title:  title,
		Artist: artist,
		Lyrics: lyrics,
	}

	// เรียก service เพื่อสร้างเพลงและอัปโหลดไฟล์
	if err := h.musicService.Create(c.Request.Context(), music, mp3File, mp4File); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ส่ง response กลับเป็น JSON พร้อม status 201 Created
	c.JSON(http.StatusCreated, music)
}

// GetByID ดึงข้อมูลเพลงตาม ID (GET /music/:id)
func (h *MusicHandler) GetByID(c *gin.Context) {
	// แปลง id จาก string ใน URL param เป็น int
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// เรียก service เพื่อค้นหาเพลง
	music, err := h.musicService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Music not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ส่งข้อมูลเพลงกลับ
	c.JSON(http.StatusOK, music)
}

// GetAll ดึงข้อมูลเพลงทั้งหมด (GET /music)
func (h *MusicHandler) GetAll(c *gin.Context) {
	// เรียก service เพื่อดึงเพลงทั้งหมด
	musics, err := h.musicService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ส่งรายการเพลงกลับ
	c.JSON(http.StatusOK, musics)
}

// Update แก้ไขข้อมูลเพลง (PUT /music/:id)
func (h *MusicHandler) Update(c *gin.Context) {
	// แปลง id จาก URL param
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var music domain.Music
	// Bind JSON body เข้ากับตัวแปร music
	if err := c.ShouldBindJSON(&music); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	music.ID = uint(id)

	// เรียก service เพื่ออัปเดตข้อมูล
	if err := h.musicService.Update(c.Request.Context(), &music); err != nil {
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Music not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ส่งข้อมูลที่อัปเดตแล้วกลับ
	c.JSON(http.StatusOK, music)
}

// Delete ลบเพลง (DELETE /music/:id)
func (h *MusicHandler) Delete(c *gin.Context) {
	// แปลง id จาก URL param
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// เรียก service เพื่อลบเพลง
	if err := h.musicService.Delete(c.Request.Context(), uint(id)); err != nil {
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Music not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ส่งข้อความยืนยันการลบ
	c.JSON(http.StatusOK, gin.H{"message": "Music deleted successfully"})
}
