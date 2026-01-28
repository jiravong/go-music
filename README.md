# Go Music API

RESTful API for Music Management using Go, Clean Architecture, Gin, GORM, and PostgreSQL.

## Features

- **Music CRUD**: Manage music tracks (Title, Artist, Lyrics, MP3, MP4).
- **Authentication**: JWT based authentication (Register, Login).
- **File Upload**: Support for uploading MP3 and MP4 files (stored locally).
- **Clean Architecture**: Separation of concerns (Domain, Service, Repository, Delivery).

## Tech Stack

- **Language**: Go
- **Framework**: Gin
- **Database**: PostgreSQL
- **ORM**: GORM
- **Authentication**: JWT (golang-jwt)
- **File Storage**: Local filesystem

## Setup

1. **Prerequisites**:
   - Go 1.21+
   - PostgreSQL

2. **Clone the repository**:
   ```sh
   git clone <repo-url>
   cd go-music-api
   ```

3. **Environment Setup**:
   Copy the example environment file or create a `.env` file:
   ```sh
   DB_HOST=localhost
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_NAME=music_db
   DB_PORT=5432
   PORT=8080
   UPLOAD_DIR=./uploads
   BASE_URL=http://localhost:8080/uploads
   ```

4. **Install Dependencies**:
   ```sh
   go mod tidy
   ```

5. **Run the Application**:
   ```sh
   go run cmd/api/main.go
   ```

## API Endpoints

### Auth
- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login and get JWT

### Music (Requires Bearer Token)
- `POST /api/v1/music/` - Create a new music (Multipart form data: title, artist, lyrics, mp3_file, mp4_file)
- `GET /api/v1/music/` - Get all music
- `GET /api/v1/music/:id` - Get music by ID
- `PUT /api/v1/music/:id` - Update music details
- `DELETE /api/v1/music/:id` - Delete music

## Folder Structure

```
.
├── cmd
│   └── api
│       └── main.go           # Entry point
├── internal
│   ├── delivery
│   │   └── http              # HTTP Handlers and Middleware
│   ├── domain                # Business entities and Interfaces
│   ├── infrastructure        # External frameworks (DB, Storage)
│   ├── repository            # Data access implementation
│   └── service               # Business logic
└── pkg
    └── utils                 # Shared utilities
```
