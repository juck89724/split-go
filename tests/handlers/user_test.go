package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"split-go/internal/handlers"
	"testing"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// 測試獲取用戶資料功能
func TestGetProfile(t *testing.T) {
	db := setupTestDB()
	handler := handlers.NewUserHandler(db)

	app := fiber.New()

	// 設置中間件來模擬已認證的用戶
	app.Use("/profile", func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		return c.Next()
	})

	app.Get("/profile", handler.GetProfile)

	// 創建測試用戶
	testUser := createTestUser(db, "profile@example.com", "profileuser")

	t.Run("成功獲取用戶資料", func(t *testing.T) {
		// 更新中間件以使用真實的用戶 ID
		app.Use("/profile", func(c *fiber.Ctx) error {
			c.Locals("user_id", testUser.ID)
			return c.Next()
		})

		req := httptest.NewRequest("GET", "/profile", nil)
		resp, err := app.Test(req)
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

		// 檢查返回的用戶數據
		data, ok := responseBody["data"].(map[string]interface{})
		if !ok {
			t.Error("應該返回用戶數據")
		} else {
			if data["email"] != testUser.Email {
				t.Errorf("期望 email %s，得到 %s", testUser.Email, data["email"])
			}
			if data["username"] != testUser.Username {
				t.Errorf("期望 username %s，得到 %s", testUser.Username, data["username"])
			}
		}
	})

	t.Run("未認證用戶", func(t *testing.T) {
		// 創建新的應用實例避免中間件衝突
		testApp := fiber.New()
		testApp.Use("/profile", func(c *fiber.Ctx) error {
			c.Locals("user_id", uint(0))
			return c.Next()
		})
		testApp.Get("/profile", handler.GetProfile)

		req := httptest.NewRequest("GET", "/profile", nil)
		resp, err := testApp.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusUnauthorized, resp.StatusCode)
		}
	})

	t.Run("用戶不存在", func(t *testing.T) {
		// 創建新的應用實例避免中間件衝突
		testApp := fiber.New()
		testApp.Use("/profile", func(c *fiber.Ctx) error {
			c.Locals("user_id", uint(999))
			return c.Next()
		})
		testApp.Get("/profile", handler.GetProfile)

		req := httptest.NewRequest("GET", "/profile", nil)
		resp, err := testApp.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusNotFound, resp.StatusCode)
		}
	})
}

// 測試更新用戶資料功能
func TestUpdateProfile(t *testing.T) {
	db := setupTestDB()
	handler := handlers.NewUserHandler(db)

	app := fiber.New()

	// 設置中間件來模擬已認證的用戶
	app.Use("/profile", func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		return c.Next()
	})

	app.Put("/profile", handler.UpdateProfile)

	// 創建測試用戶
	testUser := createTestUser(db, "update@example.com", "updateuser")

	tests := []struct {
		name         string
		userID       uint
		requestBody  map[string]interface{}
		expectedCode int
		expectError  bool
	}{
		{
			name:   "成功更新用戶資料",
			userID: testUser.ID,
			requestBody: map[string]interface{}{
				"name":      "更新後的名字",
				"avatar":    "https://example.com/avatar.jpg",
				"fcm_token": "new_fcm_token",
			},
			expectedCode: http.StatusOK,
			expectError:  false,
		},
		{
			name:   "部分更新用戶資料",
			userID: testUser.ID,
			requestBody: map[string]interface{}{
				"name": "只更新名字",
			},
			expectedCode: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "空的更新請求",
			userID:       testUser.ID,
			requestBody:  map[string]interface{}{},
			expectedCode: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "無效的用戶 ID",
			userID:       0,
			requestBody:  map[string]interface{}{},
			expectedCode: http.StatusUnauthorized,
			expectError:  true,
		},
		{
			name:   "用戶不存在",
			userID: 999,
			requestBody: map[string]interface{}{
				"name": "不存在的用戶",
			},
			expectedCode: http.StatusNotFound,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 為每個測試創建新的應用實例避免中間件衝突
			testApp := fiber.New()
			testApp.Use("/profile", func(c *fiber.Ctx) error {
				c.Locals("user_id", tt.userID)
				return c.Next()
			})
			testApp.Put("/profile", handler.UpdateProfile)

			jsonBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/profile", bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := testApp.Test(req)
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

			// 檢查成功更新時是否返回了更新後的資料
			if !tt.expectError && resp.StatusCode == http.StatusOK && tt.userID == testUser.ID {
				if data, exists := responseBody["data"]; exists {
					userData := data.(map[string]interface{})
					if name, ok := tt.requestBody["name"]; ok {
						if userData["name"] != name {
							t.Errorf("期望更新後的名字 %s，得到 %s", name, userData["name"])
						}
					}
				}
			}
		})
	}

	// 測試無效的 JSON 請求
	t.Run("無效的 JSON 請求", func(t *testing.T) {
		testApp := fiber.New()
		testApp.Use("/profile", func(c *fiber.Ctx) error {
			c.Locals("user_id", testUser.ID)
			return c.Next()
		})
		testApp.Put("/profile", handler.UpdateProfile)

		req := httptest.NewRequest("PUT", "/profile", bytes.NewReader([]byte("invalid json")))
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

// 測試更新 FCM Token 功能
func TestUpdateFCMToken(t *testing.T) {
	db := setupTestDB()
	handler := handlers.NewUserHandler(db)

	// 創建測試用戶
	testUser := createTestUser(db, "fcmtest@example.com", "fcmtestuser")

	app := fiber.New()

	// 設置認證中間件
	app.Use("/fcm-token", func(c *fiber.Ctx) error {
		c.Locals("user_id", testUser.ID)
		return c.Next()
	})

	app.Post("/fcm-token", handler.UpdateFCMToken)

	t.Run("成功更新 FCM Token", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"fcm_token": "test_fcm_token_12345",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/fcm-token", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusOK, resp.StatusCode)
		}

		var responseBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&responseBody)

		if responseBody["error"] != false {
			t.Error("期望 error 為 false")
		}

		expectedMessage := "FCM Token 更新成功"
		if message := responseBody["message"]; message != expectedMessage {
			t.Errorf("期望成功訊息 '%s'，得到 '%s'", expectedMessage, message)
		}
	})

	t.Run("缺少 FCM Token", func(t *testing.T) {
		requestBody := map[string]interface{}{}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/fcm-token", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	// 清理
	db.Delete(testUser)
}

// 測試用戶創建輔助函數
func TestCreateTestUser(t *testing.T) {
	db := setupTestDB()

	t.Run("成功創建測試用戶", func(t *testing.T) {
		user := createTestUser(db, "test@test.com", "testuser")

		if user.ID == 0 {
			t.Error("用戶 ID 不應該為 0")
		}

		if user.Email != "test@test.com" {
			t.Errorf("期望 email test@test.com，得到 %s", user.Email)
		}

		if user.Username != "testuser" {
			t.Errorf("期望 username testuser，得到 %s", user.Username)
		}

		// 檢查密碼是否已加密
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password123"))
		if err != nil {
			t.Error("密碼應該正確加密")
		}
	})

	t.Run("創建多個不同的測試用戶", func(t *testing.T) {
		user1 := createTestUser(db, "user1@test.com", "user1")
		user2 := createTestUser(db, "user2@test.com", "user2")

		if user1.ID == user2.ID {
			t.Error("不同用戶應該有不同的 ID")
		}

		if user1.Email == user2.Email {
			t.Error("不同用戶應該有不同的 email")
		}

		if user1.Username == user2.Username {
			t.Error("不同用戶應該有不同的 username")
		}
	})
}
