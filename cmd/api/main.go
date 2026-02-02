package main // ประกาศ package main เป็นจุดเริ่มต้นของโปรแกรม

import (
	"log" // นำเข้า package log สำหรับการแสดงข้อความ log
	"net/http"
	"os"   // นำเข้า package os สำหรับการจัดการกับระบบปฏิบัติการ เช่น อ่าน environment variables
	"time" // นำเข้า package time สำหรับการจัดการเรื่องเวลา

	// นำเข้า packages ภายในโปรเจค
	"go-music-api/internal/delivery/http/handler"    // นำเข้า handler สำหรับจัดการ HTTP request
	"go-music-api/internal/delivery/http/middleware" // นำเข้า middleware สำหรับจัดการ request ก่อนถึง handler
	"go-music-api/internal/domain"

	// นำเข้า domain entities
	"go-music-api/internal/infrastructure/database" // นำเข้า database สำหรับจัดการการเชื่อมต่อฐานข้อมูล
	"go-music-api/internal/infrastructure/storage"  // นำเข้า storage สำหรับจัดการไฟล์
	"go-music-api/internal/repository/postgres"     // นำเข้า repository สำหรับจัดการข้อมูลกับ Postgres
	"go-music-api/internal/service"                 // นำเข้า service สำหรับ business logic

	"github.com/danielgtaylor/huma/v2"                  // นำเข้า huma
	"github.com/danielgtaylor/huma/v2/adapters/humagin" // นำเข้า humagin adapter
	"github.com/gin-gonic/gin"                          // นำเข้า gin web framework
	"github.com/joho/godotenv"                          // นำเข้า godotenv สำหรับโหลดไฟล์ .env
)

func main() { // ฟังก์ชัน main เป็นจุดเริ่มต้นการทำงานของโปรแกรม
	// Load .env
	// โหลดตัวแปรสภาพแวดล้อมจากไฟล์ .env
	if err := godotenv.Load(); err != nil {
		// ถ้าไม่เจอไฟล์ .env ให้แสดง log เตือนและใช้ system environment variables แทน
		log.Println("No .env file found, using system environment variables")
	}

	// Init Database
	// เริ่มต้นการเชื่อมต่อฐานข้อมูล Postgres
	db, err := database.NewPostgresDB()
	if err != nil {
		// ถ้าเชื่อมต่อไม่ได้ ให้จบการทำงานและแสดง error
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Init Storage
	// ตรวจสอบประเภท Storage ที่ต้องการใช้ (local หรือ s3)
	storageType := os.Getenv("STORAGE_TYPE")
	var storageService domain.StorageService
	// var err error // ลบออกเพราะประกาศ err ในแต่ละ block หรือใช้ short decl อย่างระวัง

	// ตัวแปรสำหรับเก็บ path upload (ใช้เฉพาะ Local Storage)
	var uploadDir string

	if storageType == "s3" {
		// ถ้าเลือกใช้ S3
		bucketName := os.Getenv("AWS_BUCKET_NAME")
		region := os.Getenv("AWS_REGION")
		if bucketName == "" || region == "" {
			log.Fatal("AWS_BUCKET_NAME and AWS_REGION are required for s3 storage")
		}
		// เริ่มต้น S3 Storage
		var err error
		storageService, err = storage.NewS3Storage(bucketName, region)
		if err != nil {
			log.Fatalf("Failed to initialize S3 storage: %v", err)
		}
		log.Println("Using S3 Storage")
	} else {
		// Default ใช้ Local Storage
		// อ่านค่า UPLOAD_DIR จาก environment variable
		uploadDir = os.Getenv("UPLOAD_DIR")
		if uploadDir == "" {
			// ถ้าไม่ได้ตั้งค่าไว้ ให้ใช้ค่า default เป็น "./uploads"
			uploadDir = "./uploads"
		}
		// อ่านค่า BASE_URL จาก environment variable
		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			// ถ้าไม่ได้ตั้งค่าไว้ ให้ใช้ค่า default
			baseURL = "http://localhost:8080/uploads"
		}

		// เริ่มต้น service สำหรับจัดการไฟล์ (Local Storage)
		var err error
		storageService, err = storage.NewLocalStorage(uploadDir, baseURL)
		if err != nil {
			// ถ้าเริ่มต้นไม่ได้ ให้จบการทำงานและแสดง error
			log.Fatalf("Failed to initialize storage: %v", err)
		}
		log.Println("Using Local Storage")
	}

	// Init Repositories
	// สร้าง repository สำหรับจัดการข้อมูล Music โดยใช้ db connection ที่สร้างไว้
	musicRepo := postgres.NewMusicRepository(db)
	// สร้าง repository สำหรับจัดการข้อมูล User
	userRepo := postgres.NewUserRepository(db)

	// Init Services
	// กำหนด timeout สำหรับ context เป็น 5 วินาที
	timeout := 5 * time.Second
	// สร้าง service สำหรับ Music โดยส่ง repository, storage service และ timeout เข้าไป
	musicService := service.NewMusicService(musicRepo, storageService, timeout)
	// สร้าง service สำหรับ User
	userService := service.NewUserService(userRepo, timeout)

	// Init Handlers
	// สร้าง handler สำหรับ Music โดยส่ง service เข้าไป
	musicHandler := handler.NewMusicHandler(musicService)
	// สร้าง handler สำหรับ User
	userHandler := handler.NewUserHandler(userService)

	// Init Router
	// สร้าง router ของ Gin (Default จะมี Logger และ Recovery middleware มาให้)
	r := gin.Default()

	// Middleware
	// เรียกใช้ CORS Middleware เพื่ออนุญาตการเข้าถึงข้ามโดเมน
	r.Use(middleware.CORSMiddleware())

	// Static files for uploads
	// กำหนด path /uploads ให้เข้าถึงไฟล์ในโฟลเดอร์ uploadDir ได้โดยตรง (เฉพาะ Local Storage)
	if storageType != "s3" && uploadDir != "" {
		r.Static("/uploads", uploadDir)
	}

	// Init Huma Configuration
	config := huma.DefaultConfig("Go Music API", "1.0.0")
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
	}
	humaApi := humagin.New(r, config)

	// Public Routes (Auth)
	huma.Register(humaApi, huma.Operation{
		OperationID: "register-user",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/register",
		Summary:     "Register a new user",
		Description: "Register a new user with email and password",
		Tags:        []string{"Auth"},
	}, userHandler.Register)

	huma.Register(humaApi, huma.Operation{
		OperationID: "login-user",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/login",
		Summary:     "Login user",
		Description: "Login with email and password to get access and refresh tokens",
		Tags:        []string{"Auth"},
	}, userHandler.Login)

	huma.Register(humaApi, huma.Operation{
		OperationID: "refresh-token",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/refresh-token",
		Summary:     "Refresh Access Token",
		Description: "Get a new access token using a refresh token",
		Tags:        []string{"Auth"},
	}, userHandler.RefreshToken)

	// Protected Routes (Music)
	// Create middleware
	authMiddleware := middleware.HumaAuthMiddleware(humaApi)

	huma.Register(humaApi, huma.Operation{
		OperationID: "create-music",
		Method:      http.MethodPost,
		Path:        "/api/v1/music",
		Summary:     "Create music",
		Description: "Create a new music entry with MP3 and MP4 files",
		Tags:        []string{"Music"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
		Middlewares: huma.Middlewares{authMiddleware},
	}, musicHandler.Create)

	huma.Register(humaApi, huma.Operation{
		OperationID: "get-all-music",
		Method:      http.MethodGet,
		Path:        "/api/v1/music",
		Summary:     "Get all music",
		Description: "Retrieve a list of all music entries",
		Tags:        []string{"Music"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
		Middlewares: huma.Middlewares{authMiddleware},
	}, musicHandler.GetAll)

	huma.Register(humaApi, huma.Operation{
		OperationID: "get-music-by-id",
		Method:      http.MethodGet,
		Path:        "/api/v1/music/{id}",
		Summary:     "Get music by ID",
		Description: "Retrieve a specific music entry by its ID",
		Tags:        []string{"Music"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
		Middlewares: huma.Middlewares{authMiddleware},
	}, musicHandler.GetByID)

	huma.Register(humaApi, huma.Operation{
		OperationID: "update-music",
		Method:      http.MethodPut,
		Path:        "/api/v1/music/{id}",
		Summary:     "Update music",
		Description: "Update an existing music entry",
		Tags:        []string{"Music"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
		Middlewares: huma.Middlewares{authMiddleware},
	}, musicHandler.Update)

	huma.Register(humaApi, huma.Operation{
		OperationID: "delete-music",
		Method:      http.MethodDelete,
		Path:        "/api/v1/music/{id}",
		Summary:     "Delete music",
		Description: "Delete a music entry by its ID",
		Tags:        []string{"Music"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
		Middlewares: huma.Middlewares{authMiddleware},
	}, musicHandler.Delete)

	// อ่านค่า PORT จาก environment variable
	port := os.Getenv("PORT")
	if port == "" {
		// ถ้าไม่ได้ตั้งค่าไว้ ให้ใช้ port 8080
		port = "8080"
	}

	// แสดง log ว่า server กำลังเริ่มทำงานที่ port ไหน
	log.Printf("Server starting on port %s", port)
	// สั่งให้ router เริ่มทำงานและรอรับ request
	if err := r.Run(":" + port); err != nil {
		// ถ้า start server ไม่ได้ ให้จบการทำงานและแสดง error
		log.Fatalf("Failed to start server: %v", err)
	}
}
