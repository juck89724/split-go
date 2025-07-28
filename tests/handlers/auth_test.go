package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"split-go/internal/config"
	"split-go/internal/handlers"
	"split-go/internal/models"
	"split-go/internal/services"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 設置測試資料庫
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("無法連接測試資料庫")
	}

	// 只遷移 User 模型，其他手動創建避免 PostgreSQL 語法問題
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		panic("無法執行用戶表遷移")
	}

	// 手動創建 UserSession 表，避免 PostgreSQL 特有語法問題
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			device_id TEXT NOT NULL,
			device_fingerprint TEXT NOT NULL,
			refresh_token_hash TEXT NOT NULL,
			access_token_version INTEGER DEFAULT 1,
			device_name TEXT,
			device_type TEXT,
			user_agent TEXT,
			ip_address TEXT,
			country TEXT,
			city TEXT,
			trust_level INTEGER DEFAULT 0,
			last_activity DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL,
			revoked_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		panic("無法創建 user_sessions 表")
	}

	// 手動創建 SecurityEvent 表，避免 PostgreSQL 特有語法問題
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS security_events (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			session_id TEXT,
			event_type TEXT NOT NULL,
			event_data TEXT,
			ip_address TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		panic("無法創建 security_events 表")
	}

	return db
}

// 設置測試配置
func setupTestConfig() *config.Config {
	return &config.Config{
		JWTSecret:            "test_jwt_secret",
		AccessTokenSecret:    "test_access_secret",
		RefreshTokenSecret:   "test_refresh_secret",
		DeviceTokenSecret:    "test_device_secret",
		AccessTokenDuration:  5 * time.Minute,
		RefreshTokenDuration: 1 * time.Hour,
		DeviceTokenDuration:  24 * time.Hour,
	}
}

// 創建測試用戶
func createTestUser(db *gorm.DB, email, username string) *models.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &models.User{
		Email:    email,
		Username: username,
		Password: string(hashedPassword),
		Name:     "測試用戶",
	}
	db.Create(user)
	return user
}

// 測試註冊功能
func TestRegister(t *testing.T) {
	db := setupTestDB()
	cfg := setupTestConfig()
	handler := handlers.NewAuthHandler(db, cfg)

	app := fiber.New()
	app.Post("/register", handler.Register)

	tests := []struct {
		name         string
		requestBody  map[string]interface{}
		expectedCode int
		expectError  bool
	}{
		{
			name: "成功註冊",
			requestBody: map[string]interface{}{
				"email":    "newuser@example.com",
				"username": "newuser",
				"password": "password123",
				"name":     "新用戶",
			},
			expectedCode: http.StatusCreated,
			expectError:  false,
		},
		{
			name: "Email 格式錯誤",
			requestBody: map[string]interface{}{
				"email":    "invalid-email",
				"username": "newuser",
				"password": "password123",
				"name":     "新用戶",
			},
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name: "密碼太短",
			requestBody: map[string]interface{}{
				"email":    "test2@example.com",
				"username": "testuser2",
				"password": "123",
				"name":     "測試用戶2",
			},
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name: "用戶名太短",
			requestBody: map[string]interface{}{
				"email":    "test3@example.com",
				"username": "ab",
				"password": "password123",
				"name":     "測試用戶3",
			},
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name: "缺少必要欄位",
			requestBody: map[string]interface{}{
				"email":    "test4@example.com",
				"password": "password123",
				"name":     "測試用戶4",
			},
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/register", bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("無法執行請求: %v", err)
			}

			if resp.StatusCode != tt.expectedCode {
				t.Errorf("期望狀態碼 %d，得到 %d", tt.expectedCode, resp.StatusCode)
			}

			var responseBody map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&responseBody)

			if tt.expectError && !responseBody["error"].(bool) {
				t.Error("期望錯誤但沒有收到錯誤")
			}

			if !tt.expectError && responseBody["error"].(bool) {
				t.Errorf("不期望錯誤但收到錯誤: %s", responseBody["message"])
			}
		})
	}

	// 測試重複 Email
	t.Run("Email 已存在", func(t *testing.T) {
		// 先創建一個用戶
		createTestUser(db, "existing@example.com", "existinguser")

		requestBody := map[string]interface{}{
			"email":    "existing@example.com",
			"username": "anotheruser",
			"password": "password123",
			"name":     "另一個用戶",
		}

		jsonBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/register", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusConflict, resp.StatusCode)
		}
	})

	// 測試重複用戶名
	t.Run("用戶名已存在", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"email":    "another@example.com",
			"username": "existinguser",
			"password": "password123",
			"name":     "另一個用戶",
		}

		jsonBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/register", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusConflict, resp.StatusCode)
		}
	})
}

// 測試登入功能
func TestLogin(t *testing.T) {
	db := setupTestDB()
	cfg := setupTestConfig()
	handler := handlers.NewAuthHandler(db, cfg)

	app := fiber.New()
	app.Post("/login", handler.Login)

	// 先創建測試用戶
	testUser := createTestUser(db, "login@example.com", "loginuser")

	tests := []struct {
		name         string
		requestBody  map[string]interface{}
		expectedCode int
		expectError  bool
	}{
		{
			name: "成功登入",
			requestBody: map[string]interface{}{
				"email":    "login@example.com",
				"password": "password123",
				"device_fingerprint": map[string]interface{}{
					"user_agent": "Mozilla/5.0",
					"platform":   "Web",
					"language":   "zh-TW",
					"timezone":   "Asia/Taipei",
					"screen": map[string]int{
						"width":  1920,
						"height": 1080,
					},
				},
				"device_name": "測試設備",
				"device_type": "desktop",
			},
			expectedCode: http.StatusOK,
			expectError:  false,
		},
		{
			name: "錯誤密碼",
			requestBody: map[string]interface{}{
				"email":    "login@example.com",
				"password": "wrongpassword",
				"device_fingerprint": map[string]interface{}{
					"user_agent": "Mozilla/5.0",
					"platform":   "Web",
					"language":   "zh-TW",
					"timezone":   "Asia/Taipei",
				},
				"device_name": "測試設備",
				"device_type": "desktop",
			},
			expectedCode: http.StatusUnauthorized,
			expectError:  true,
		},
		{
			name: "用戶不存在",
			requestBody: map[string]interface{}{
				"email":    "nonexistent@example.com",
				"password": "password123",
				"device_fingerprint": map[string]interface{}{
					"user_agent": "Mozilla/5.0",
					"platform":   "Web",
					"language":   "zh-TW",
					"timezone":   "Asia/Taipei",
				},
				"device_name": "測試設備",
				"device_type": "desktop",
			},
			expectedCode: http.StatusUnauthorized,
			expectError:  true,
		},
		{
			name: "無效的請求格式",
			requestBody: map[string]interface{}{
				"email": "login@example.com",
				// 缺少密碼
			},
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/login", bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("無法執行請求: %v", err)
			}

			if resp.StatusCode != tt.expectedCode {
				t.Errorf("期望狀態碼 %d，得到 %d", tt.expectedCode, resp.StatusCode)
			}

			var responseBody map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&responseBody)

			if tt.expectError && !responseBody["error"].(bool) {
				t.Error("期望錯誤但沒有收到錯誤")
			}

			if !tt.expectError && responseBody["error"].(bool) {
				t.Errorf("不期望錯誤但收到錯誤: %s", responseBody["message"])
			}

			// 檢查成功登入時是否返回了所需的 token
			if !tt.expectError && resp.StatusCode == http.StatusOK {
				data, ok := responseBody["data"].(map[string]interface{})
				if !ok {
					t.Error("成功登入應該返回 data 欄位")
				} else {
					if _, exists := data["access_token"]; !exists {
						t.Error("成功登入應該返回 access_token")
					}
					if _, exists := data["refresh_token"]; !exists {
						t.Error("成功登入應該返回 refresh_token")
					}
					if _, exists := data["device_token"]; !exists {
						t.Error("成功登入應該返回 device_token")
					}
				}
			}
		})
	}

	// 清理
	_ = testUser
}

// 測試刷新 Token 功能
func TestRefreshToken(t *testing.T) {
	db := setupTestDB()
	cfg := setupTestConfig()
	handler := handlers.NewAuthHandler(db, cfg)
	jwtService := services.NewJWTService(db, cfg)

	app := fiber.New()
	app.Post("/refresh", handler.RefreshToken)

	// 創建測試用戶並生成 Token
	testUser := createTestUser(db, "refresh@example.com", "refreshuser")

	// 創建設備指紋
	deviceFingerprint := services.DeviceFingerprint{
		UserAgent: "Mozilla/5.0",
		Platform:  "Web",
		Language:  "zh-TW",
		TimeZone:  "Asia/Taipei",
		IPAddress: "127.0.0.1",
	}

	// 生成有效的 tokens
	tokens, err := jwtService.HandleLogin(*testUser, deviceFingerprint, "測試設備", "desktop", "127.0.0.1")
	if err != nil {
		t.Fatalf("無法生成測試 tokens: %v", err)
	}

	tests := []struct {
		name         string
		refreshToken string
		expectedCode int
		expectError  bool
	}{
		{
			name:         "成功刷新 Token",
			refreshToken: tokens.RefreshToken,
			expectedCode: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "無效的 Refresh Token",
			refreshToken: "invalid_token",
			expectedCode: http.StatusUnauthorized,
			expectError:  true,
		},
		{
			name:         "空的 Refresh Token",
			refreshToken: "",
			expectedCode: http.StatusBadRequest,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody := map[string]interface{}{
				"refresh_token": tt.refreshToken,
			}

			jsonBytes, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/refresh", bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("無法執行請求: %v", err)
			}

			if resp.StatusCode != tt.expectedCode {
				t.Errorf("期望狀態碼 %d，得到 %d", tt.expectedCode, resp.StatusCode)
			}

			var responseBody map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&responseBody)

			if tt.expectError && !responseBody["error"].(bool) {
				t.Error("期望錯誤但沒有收到錯誤")
			}

			if !tt.expectError && responseBody["error"].(bool) {
				t.Errorf("不期望錯誤但收到錯誤: %s", responseBody["message"])
			}
		})
	}
}

// 測試登出功能
func TestLogout(t *testing.T) {
	db := setupTestDB()
	cfg := setupTestConfig()
	handler := handlers.NewAuthHandler(db, cfg)
	jwtService := services.NewJWTService(db, cfg)

	app := fiber.New()

	// 設置中間件來模擬已認證的用戶
	app.Use("/logout", func(c *fiber.Ctx) error {
		// 模擬從 JWT 中間件設置的用戶資訊
		c.Locals("user_id", uint(1))
		c.Locals("session_id", "test-session-id")
		return c.Next()
	})

	app.Post("/logout", handler.Logout)

	// 創建測試用戶和會話
	testUser := createTestUser(db, "logout@example.com", "logoutuser")

	deviceFingerprint := services.DeviceFingerprint{
		UserAgent: "Mozilla/5.0",
		Platform:  "Web",
		Language:  "zh-TW",
		TimeZone:  "Asia/Taipei",
		IPAddress: "127.0.0.1",
	}

	// 創建會話
	session, err := jwtService.CreateSession(*testUser, "test-device-id", deviceFingerprint, "測試設備", "desktop")
	if err != nil {
		t.Fatalf("無法創建測試會話: %v", err)
	}

	t.Run("成功登出", func(t *testing.T) {
		// 創建新的應用實例避免中間件衝突
		testApp := fiber.New()
		testApp.Use("/logout", func(c *fiber.Ctx) error {
			c.Locals("user_id", testUser.ID)
			c.Locals("session_id", session.ID)
			return c.Next()
		})
		testApp.Post("/logout", handler.Logout)

		req := httptest.NewRequest("POST", "/logout", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := testApp.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusOK, resp.StatusCode)
		}

		var responseBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&responseBody)

		if responseBody["error"].(bool) {
			t.Errorf("不期望錯誤但收到錯誤: %s", responseBody["message"])
		}
	})

	t.Run("無效會話登出", func(t *testing.T) {
		// 創建新的應用實例避免中間件衝突
		testApp := fiber.New()
		testApp.Use("/logout", func(c *fiber.Ctx) error {
			c.Locals("user_id", testUser.ID)
			c.Locals("session_id", "")
			return c.Next()
		})
		testApp.Post("/logout", handler.Logout)

		req := httptest.NewRequest("POST", "/logout", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := testApp.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusBadRequest, resp.StatusCode)
		}
	})
}
