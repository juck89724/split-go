package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"split-go/internal/handlers"
	"split-go/internal/models"
	"testing"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// 設置分類測試資料庫
func setupCategoryTestDB() *gorm.DB {
	db := setupTestDB()

	// 創建分類表
	err := db.AutoMigrate(&models.Category{})
	if err != nil {
		panic("無法執行分類表遷移")
	}

	return db
}

// 創建測試分類
func createTestCategory(db *gorm.DB, name, icon, color string) *models.Category {
	category := &models.Category{
		Name:  name,
		Icon:  icon,
		Color: color,
	}

	db.Create(category)
	return category
}

// 測試獲取分類列表
func TestGetCategories(t *testing.T) {
	db := setupCategoryTestDB()
	handler := handlers.NewCategoryHandler(db)

	app := fiber.New()

	// 設置認證中間件
	app.Use("/categories", func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		return c.Next()
	})

	app.Get("/categories", handler.GetCategories)

	tests := []struct {
		name            string
		setupCategories []models.Category
		expectedStatus  int
		expectedCount   int
		expectedOrder   []string // 預期的排序順序
	}{
		{
			name: "成功獲取分類列表",
			setupCategories: []models.Category{
				{Name: "餐飲", Icon: "🍽️", Color: "#FF6B6B"},
				{Name: "交通", Icon: "🚗", Color: "#4ECDC4"},
				{Name: "娛樂", Icon: "🎬", Color: "#45B7D1"},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  3,
			expectedOrder:  []string{"交通", "娛樂", "餐飲"}, // 按名稱 ASC 排序
		},
		{
			name:            "空分類列表",
			setupCategories: []models.Category{},
			expectedStatus:  http.StatusOK,
			expectedCount:   0,
		},
		{
			name: "單一分類",
			setupCategories: []models.Category{
				{Name: "購物", Icon: "🛍️", Color: "#96CEB4"},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name: "多個分類測試排序",
			setupCategories: []models.Category{
				{Name: "住宿", Icon: "🏠", Color: "#DDA0DD"},
				{Name: "日用品", Icon: "🧴", Color: "#98D8C8"},
				{Name: "醫療", Icon: "💊", Color: "#F7DC6F"},
				{Name: "教育", Icon: "📚", Color: "#BB8FCE"},
				{Name: "其他", Icon: "📝", Color: "#AEB6BF"},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  5,
			expectedOrder:  []string{"住宿", "其他", "教育", "日用品", "醫療"}, // 按名稱 ASC 排序
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理資料庫
			db.Where("1 = 1").Delete(&models.Category{})

			// 創建測試分類
			var createdCategories []models.Category
			for _, cat := range tt.setupCategories {
				category := createTestCategory(db, cat.Name, cat.Icon, cat.Color)
				createdCategories = append(createdCategories, *category)
			}

			req := httptest.NewRequest("GET", "/categories", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("無法執行請求: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("期望狀態碼 %d，得到 %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedStatus == http.StatusOK {
				var responseBody map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&responseBody)

				// 檢查回應格式
				if responseBody["error"] != false {
					t.Error("期望 error 為 false")
				}

				data, ok := responseBody["data"].([]interface{})
				if !ok {
					t.Fatal("回應資料格式錯誤")
				}

				if len(data) != tt.expectedCount {
					t.Errorf("期望 %d 個分類，得到 %d 個", tt.expectedCount, len(data))
				}

				// 驗證每個分類的內容
				for i, item := range data {
					category, ok := item.(map[string]interface{})
					if !ok {
						t.Fatal("分類資料格式錯誤")
					}

					// 檢查必要欄位
					if category["id"] == nil {
						t.Error("分類 ID 不能為空")
					}
					if categoryName, ok := category["name"].(string); !ok || categoryName == "" {
						t.Error("分類名稱不能為空")
					}

					// 檢查排序順序
					if len(tt.expectedOrder) > 0 && i < len(tt.expectedOrder) {
						if categoryName := category["name"].(string); categoryName != tt.expectedOrder[i] {
							t.Errorf("位置 %d 期望分類名稱 '%s'，得到 '%s'", i, tt.expectedOrder[i], categoryName)
						}
					}

					// 檢查圖標和顏色（如果存在）
					if icon, exists := category["icon"]; exists && icon != "" {
						if iconStr, ok := icon.(string); !ok || iconStr == "" {
							t.Error("圖標格式錯誤")
						}
					}

					if color, exists := category["color"]; exists && color != "" {
						if colorStr, ok := color.(string); !ok || colorStr == "" {
							t.Error("顏色格式錯誤")
						}
					}
				}
			}

			// 清理創建的分類
			for _, cat := range createdCategories {
				db.Delete(&cat)
			}
		})
	}
}

// 測試分類內容驗證
func TestCategoryContentValidation(t *testing.T) {
	db := setupCategoryTestDB()
	handler := handlers.NewCategoryHandler(db)

	app := fiber.New()

	// 設置認證中間件
	app.Use("/categories", func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		return c.Next()
	})

	app.Get("/categories", handler.GetCategories)

	// 創建各種類型的測試分類
	testCategories := []struct {
		name        string
		icon        string
		color       string
		description string
	}{
		{
			name:        "餐飲",
			icon:        "🍽️",
			color:       "#FF6B6B",
			description: "有圖標和顏色的完整分類",
		},
		{
			name:        "交通",
			icon:        "",
			color:       "#4ECDC4",
			description: "沒有圖標的分類",
		},
		{
			name:        "娛樂",
			icon:        "🎬",
			color:       "",
			description: "沒有顏色的分類",
		},
		{
			name:        "其他",
			icon:        "",
			color:       "",
			description: "只有名稱的最小化分類",
		},
	}

	// 創建測試分類
	var createdCategories []models.Category
	for _, tc := range testCategories {
		category := createTestCategory(db, tc.name, tc.icon, tc.color)
		createdCategories = append(createdCategories, *category)
	}

	t.Run("驗證分類內容完整性", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/categories", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusOK, resp.StatusCode)
		}

		var responseBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&responseBody)

		data, ok := responseBody["data"].([]interface{})
		if !ok {
			t.Fatal("回應資料格式錯誤")
		}

		if len(data) != len(testCategories) {
			t.Errorf("期望 %d 個分類，得到 %d 個", len(testCategories), len(data))
		}

		// 建立名稱到測試資料的對應
		testDataMap := make(map[string]struct {
			icon        string
			color       string
			description string
		})
		for _, tc := range testCategories {
			testDataMap[tc.name] = struct {
				icon        string
				color       string
				description string
			}{tc.icon, tc.color, tc.description}
		}

		// 驗證每個分類
		for _, item := range data {
			category, ok := item.(map[string]interface{})
			if !ok {
				t.Fatal("分類資料格式錯誤")
			}

			categoryName := category["name"].(string)
			testData, exists := testDataMap[categoryName]
			if !exists {
				t.Errorf("未預期的分類名稱: %s", categoryName)
				continue
			}

			// 驗證圖標
			if testData.icon != "" {
				if icon, ok := category["icon"].(string); !ok || icon != testData.icon {
					t.Errorf("分類 %s 的圖標不正確，期望 '%s'，得到 '%s'", categoryName, testData.icon, icon)
				}
			} else {
				// 檢查空圖標是否正確處理（可能被省略）
				if icon, exists := category["icon"]; exists && icon != "" {
					t.Errorf("分類 %s 不應該有圖標，但得到 '%s'", categoryName, icon)
				}
			}

			// 驗證顏色
			if testData.color != "" {
				if color, ok := category["color"].(string); !ok || color != testData.color {
					t.Errorf("分類 %s 的顏色不正確，期望 '%s'，得到 '%s'", categoryName, testData.color, color)
				}
			} else {
				// 檢查空顏色是否正確處理（可能被省略）
				if color, exists := category["color"]; exists && color != "" {
					t.Errorf("分類 %s 不應該有顏色，但得到 '%s'", categoryName, color)
				}
			}
		}
	})

	// 清理創建的分類
	for _, cat := range createdCategories {
		db.Delete(&cat)
	}
}

// 測試分類資料庫錯誤處理
func TestCategoryDatabaseError(t *testing.T) {
	db := setupCategoryTestDB()
	handler := handlers.NewCategoryHandler(db)

	app := fiber.New()

	// 設置認證中間件
	app.Use("/categories", func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		return c.Next()
	})

	app.Get("/categories", handler.GetCategories)

	t.Run("資料庫連接正常情況", func(t *testing.T) {
		// 創建一個測試分類
		category := createTestCategory(db, "測試分類", "🔧", "#123456")

		req := httptest.NewRequest("GET", "/categories", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusOK, resp.StatusCode)
		}

		// 清理
		db.Delete(category)
	})

	// 注意：模擬資料庫錯誤在實際測試中比較困難，
	// 因為我們使用的是內存 SQLite 資料庫
	// 在實際專案中，可能需要使用 mock 或者其他技術來測試錯誤情況
}

// 性能測試 - 大量分類情況下的查詢性能
func TestCategoryPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳過性能測試")
	}

	db := setupCategoryTestDB()
	handler := handlers.NewCategoryHandler(db)

	app := fiber.New()

	// 設置認證中間件
	app.Use("/categories", func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		return c.Next()
	})

	app.Get("/categories", handler.GetCategories)

	t.Run("大量分類查詢性能", func(t *testing.T) {
		// 創建大量測試分類
		const categoryCount = 100
		var createdCategories []models.Category

		for i := 0; i < categoryCount; i++ {
			category := createTestCategory(db,
				fmt.Sprintf("分類-%03d", i),
				"📁",
				"#666666",
			)
			createdCategories = append(createdCategories, *category)
		}

		// 執行查詢
		req := httptest.NewRequest("GET", "/categories", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusOK, resp.StatusCode)
		}

		var responseBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&responseBody)

		data, ok := responseBody["data"].([]interface{})
		if !ok {
			t.Fatal("回應資料格式錯誤")
		}

		if len(data) != categoryCount {
			t.Errorf("期望 %d 個分類，得到 %d 個", categoryCount, len(data))
		}

		// 清理大量資料
		for _, cat := range createdCategories {
			db.Delete(&cat)
		}
	})
}
