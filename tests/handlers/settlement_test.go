package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"split-go/internal/handlers"
	"split-go/internal/models"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// 設置結算測試資料庫
func setupSettlementTestDB() *gorm.DB {
	db := setupTestDB()

	// 創建相關表
	err := db.AutoMigrate(
		&models.Group{},
		&models.GroupMember{},
		&models.Settlement{},
		&models.Transaction{},
		&models.TransactionSplit{},
	)
	if err != nil {
		panic("無法執行結算表遷移")
	}

	return db
}

// 創建測試結算記錄
func createTestSettlement(db *gorm.DB, groupID, fromUserID, toUserID uint, amount float64) *models.Settlement {
	settlement := &models.Settlement{
		GroupID:    groupID,
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Amount:     amount,
		Currency:   "TWD",
		Status:     "pending",
		Notes:      "測試結算",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	db.Create(settlement)
	return settlement
}

// 創建測試交易記錄（用於平衡計算）
func createTestTransaction(db *gorm.DB, groupID, paidBy, createdBy uint, amount float64) *models.Transaction {
	transaction := &models.Transaction{
		GroupID:     groupID,
		Description: "測試交易",
		Amount:      amount,
		Currency:    "TWD",
		PaidBy:      paidBy,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	db.Create(transaction)
	return transaction
}

// 創建測試分帳記錄
func createTestTransactionSplit(db *gorm.DB, transactionID, userID uint, amount float64) {
	split := &models.TransactionSplit{
		TransactionID: transactionID,
		UserID:        userID,
		Amount:        amount,
		Percentage:    0, // 假設是固定金額分帳
		SplitType:     models.SplitFixed,
	}

	db.Create(split)
}

// 測試獲取結算記錄列表
func TestGetSettlements(t *testing.T) {
	db := setupSettlementTestDB()
	handler := handlers.NewSettlementHandler(db)

	app := fiber.New()

	// 設置認證中間件
	app.Use("/settlements", func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		return c.Next()
	})

	app.Get("/settlements", handler.GetSettlements)

	// 創建測試資料
	user1 := createTestUser(db, "user1@example.com", "user1")
	user2 := createTestUser(db, "user2@example.com", "user2")
	group := createTestGroup(db, "測試群組", "測試描述", user1.ID)

	// 添加用戶到群組
	addGroupMember(db, group.ID, user2.ID, "member")

	// 創建結算記錄
	settlement := createTestSettlement(db, group.ID, user1.ID, user2.ID, 100.0)

	tests := []struct {
		name           string
		userID         uint
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "成功獲取結算記錄",
			userID:         user1.ID,
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "收款者查看結算記錄",
			userID:         user2.ID,
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "無相關結算記錄的用戶",
			userID:         createTestUser(db, "norelation@example.com", "norelation").ID,
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理之前的結算記錄，避免干擾
			db.Where("1 = 1").Delete(&models.Settlement{})

			// 重新創建結算記錄（只為相關用戶）
			if tt.expectedCount > 0 {
				createTestSettlement(db, group.ID, user1.ID, user2.ID, 100.0)
			}

			// 創建新的 app 實例避免中間件衝突
			testApp := fiber.New()
			testApp.Use("/settlements", func(c *fiber.Ctx) error {
				c.Locals("user_id", tt.userID)
				return c.Next()
			})
			testApp.Get("/settlements", handler.GetSettlements)

			req := httptest.NewRequest("GET", "/settlements", nil)
			resp, err := testApp.Test(req)
			if err != nil {
				t.Fatalf("無法執行請求: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("期望狀態碼 %d，得到 %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedStatus == http.StatusOK {
				var responseBody map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&responseBody)

				data, ok := responseBody["data"].([]interface{})
				if !ok {
					t.Fatal("回應資料格式錯誤")
				}

				if len(data) != tt.expectedCount {
					t.Errorf("期望 %d 筆記錄，得到 %d 筆", tt.expectedCount, len(data))
				}
			}
		})
	}

	// 清理
	db.Delete(settlement)
	db.Delete(group)
	db.Delete(user1)
	db.Delete(user2)
}

// 測試創建結算記錄
func TestCreateSettlement(t *testing.T) {
	db := setupSettlementTestDB()
	handler := handlers.NewSettlementHandler(db)

	app := fiber.New()

	// 設置認證中間件
	app.Use("/settlements", func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		return c.Next()
	})

	app.Post("/settlements", handler.CreateSettlement)

	// 創建測試資料
	user1 := createTestUser(db, "user1@example.com", "user1")
	user2 := createTestUser(db, "user2@example.com", "user2")
	user3 := createTestUser(db, "user3@example.com", "user3")
	group := createTestGroup(db, "測試群組", "測試描述", user1.ID)

	// 添加用戶到群組
	addGroupMember(db, group.ID, user2.ID, "member")

	tests := []struct {
		name           string
		userID         uint
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "成功創建結算記錄",
			userID: user1.ID,
			requestBody: map[string]interface{}{
				"group_id":   group.ID,
				"to_user_id": user2.ID,
				"amount":     150.0,
				"currency":   "TWD",
				"notes":      "餐費分攤",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "收款者不是群組成員",
			userID: user1.ID,
			requestBody: map[string]interface{}{
				"group_id":   group.ID,
				"to_user_id": user3.ID,
				"amount":     100.0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "用戶不是群組成員",
		},
		{
			name:   "向自己結算",
			userID: user1.ID,
			requestBody: map[string]interface{}{
				"group_id":   group.ID,
				"to_user_id": user1.ID,
				"amount":     100.0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "不能向自己結算",
		},
		{
			name:   "缺少必要欄位",
			userID: user1.ID,
			requestBody: map[string]interface{}{
				"group_id": group.ID,
				"amount":   100.0,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "金額為零",
			userID: user1.ID,
			requestBody: map[string]interface{}{
				"group_id":   group.ID,
				"to_user_id": user2.ID,
				"amount":     0,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理之前的結算記錄
			db.Where("1 = 1").Delete(&models.Settlement{})

			// 創建新的 app 實例避免中間件衝突
			testApp := fiber.New()
			testApp.Use("/settlements", func(c *fiber.Ctx) error {
				c.Locals("user_id", tt.userID)
				return c.Next()
			})
			testApp.Post("/settlements", handler.CreateSettlement)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/settlements", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := testApp.Test(req)
			if err != nil {
				t.Fatalf("無法執行請求: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("期望狀態碼 %d，得到 %d", tt.expectedStatus, resp.StatusCode)
			}

			var responseBody map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&responseBody)

			if tt.expectedError != "" {
				message, ok := responseBody["message"].(string)
				if !ok || message != tt.expectedError {
					t.Errorf("期望錯誤訊息 '%s'，得到 '%s'", tt.expectedError, message)
				}
			}

			if tt.expectedStatus == http.StatusCreated {
				// 驗證結算記錄是否正確創建
				var count int64
				db.Model(&models.Settlement{}).Where("group_id = ? AND from_user_id = ? AND to_user_id = ?",
					tt.requestBody["group_id"], tt.userID, tt.requestBody["to_user_id"]).Count(&count)

				if count == 0 {
					t.Error("結算記錄未正確創建")
				}
			}
		})
	}

	// 清理
	db.Where("group_id = ?", group.ID).Delete(&models.Settlement{})
	db.Delete(group)
	db.Delete(user1)
	db.Delete(user2)
	db.Delete(user3)
}

// 測試標記結算為已付款
func TestMarkAsPaid(t *testing.T) {
	db := setupSettlementTestDB()
	handler := handlers.NewSettlementHandler(db)

	app := fiber.New()

	// 設置認證中間件
	app.Use("/settlements/:id/paid", func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		return c.Next()
	})

	app.Put("/settlements/:id/paid", handler.MarkAsPaid)

	// 創建測試資料
	user1 := createTestUser(db, "user1@example.com", "user1")
	user2 := createTestUser(db, "user2@example.com", "user2")
	group := createTestGroup(db, "測試群組", "測試描述", user1.ID)
	addGroupMember(db, group.ID, user2.ID, "member")

	settlement := createTestSettlement(db, group.ID, user1.ID, user2.ID, 100.0)

	tests := []struct {
		name           string
		userID         uint
		settlementID   uint
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "收款者成功標記已付款",
			userID:         user2.ID,
			settlementID:   settlement.ID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "付款者無權限標記",
			userID:         user1.ID,
			settlementID:   settlement.ID,
			expectedStatus: http.StatusForbidden,
			expectedError:  "只有收款者可以標記為已付款",
		},
		{
			name:           "無效的結算 ID",
			userID:         user2.ID,
			settlementID:   999,
			expectedStatus: http.StatusNotFound,
			expectedError:  "結算記錄不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置結算狀態
			db.Model(settlement).Updates(map[string]interface{}{
				"status":     "pending",
				"settled_at": nil,
			})

			// 創建新的 app 實例避免中間件衝突
			testApp := fiber.New()
			testApp.Use("/settlements/:id/paid", func(c *fiber.Ctx) error {
				c.Locals("user_id", tt.userID)
				return c.Next()
			})
			testApp.Put("/settlements/:id/paid", handler.MarkAsPaid)

			req := httptest.NewRequest("PUT", fmt.Sprintf("/settlements/%d/paid", tt.settlementID), nil)
			resp, err := testApp.Test(req)
			if err != nil {
				t.Fatalf("無法執行請求: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("期望狀態碼 %d，得到 %d", tt.expectedStatus, resp.StatusCode)
			}

			var responseBody map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&responseBody)

			if tt.expectedError != "" {
				message, ok := responseBody["message"].(string)
				if !ok || message != tt.expectedError {
					t.Errorf("期望錯誤訊息 '%s'，得到 '%s'", tt.expectedError, message)
				}
			}

			if tt.expectedStatus == http.StatusOK {
				// 驗證狀態是否已更新
				var updatedSettlement models.Settlement
				db.First(&updatedSettlement, settlement.ID)
				if updatedSettlement.Status != "paid" {
					t.Error("結算狀態未正確更新為 paid")
				}
				if updatedSettlement.SettledAt == nil {
					t.Error("結算時間未設置")
				}
			}
		})
	}

	// 清理
	db.Delete(settlement)
	db.Delete(group)
	db.Delete(user1)
	db.Delete(user2)
}

// 測試取消結算記錄
func TestCancelSettlement(t *testing.T) {
	db := setupSettlementTestDB()
	handler := handlers.NewSettlementHandler(db)

	app := fiber.New()

	// 設置認證中間件
	app.Use("/settlements/:id", func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		return c.Next()
	})

	app.Delete("/settlements/:id", handler.CancelSettlement)

	// 創建測試資料
	user1 := createTestUser(db, "user1@example.com", "user1")
	user2 := createTestUser(db, "user2@example.com", "user2")
	user3 := createTestUser(db, "user3@example.com", "user3")
	group := createTestGroup(db, "測試群組", "測試描述", user1.ID)
	addGroupMember(db, group.ID, user2.ID, "member")

	settlement := createTestSettlement(db, group.ID, user1.ID, user2.ID, 100.0)

	tests := []struct {
		name             string
		userID           uint
		settlementID     uint
		settlementStatus string
		expectedStatus   int
		expectedError    string
	}{
		{
			name:             "付款者成功取消結算",
			userID:           user1.ID,
			settlementID:     settlement.ID,
			settlementStatus: "pending",
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "收款者成功取消結算",
			userID:           user2.ID,
			settlementID:     settlement.ID,
			settlementStatus: "pending",
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "無關用戶無權限取消",
			userID:           user3.ID,
			settlementID:     settlement.ID,
			settlementStatus: "pending",
			expectedStatus:   http.StatusForbidden,
			expectedError:    "無權限取消此結算記錄",
		},
		{
			name:             "無法取消已付款的結算",
			userID:           user1.ID,
			settlementID:     settlement.ID,
			settlementStatus: "paid",
			expectedStatus:   http.StatusBadRequest,
			expectedError:    "只能取消待付款的結算記錄",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 確保結算記錄存在且設置正確狀態
			db.Unscoped().Delete(&models.Settlement{}, settlement.ID) // 完全刪除

			// 重新創建結算記錄
			freshSettlement := createTestSettlement(db, group.ID, user1.ID, user2.ID, 100.0)
			db.Model(freshSettlement).Update("status", tt.settlementStatus)

			// 創建新的 app 實例避免中間件衝突
			testApp := fiber.New()
			testApp.Use("/settlements/:id", func(c *fiber.Ctx) error {
				c.Locals("user_id", tt.userID)
				return c.Next()
			})
			testApp.Delete("/settlements/:id", handler.CancelSettlement)

			req := httptest.NewRequest("DELETE", fmt.Sprintf("/settlements/%d", freshSettlement.ID), nil)
			resp, err := testApp.Test(req)
			if err != nil {
				t.Fatalf("無法執行請求: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("期望狀態碼 %d，得到 %d", tt.expectedStatus, resp.StatusCode)
			}

			var responseBody map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&responseBody)

			if tt.expectedError != "" {
				message, ok := responseBody["message"].(string)
				if !ok || message != tt.expectedError {
					t.Errorf("期望錯誤訊息 '%s'，得到 '%s'", tt.expectedError, message)
				}
			}

			if tt.expectedStatus == http.StatusOK {
				// 驗證記錄是否被軟刪除
				var count int64
				db.Model(&models.Settlement{}).Where("id = ?", freshSettlement.ID).Count(&count)
				if count != 0 {
					t.Error("結算記錄未正確刪除")
				}
			}
		})
	}

	// 清理
	db.Unscoped().Delete(settlement)
	db.Delete(group)
	db.Delete(user1)
	db.Delete(user2)
	db.Delete(user3)
}

// 測試獲取結算建議
func TestGetSettlementSuggestions(t *testing.T) {
	db := setupSettlementTestDB()
	handler := handlers.NewSettlementHandler(db)

	app := fiber.New()

	// 設置認證中間件
	app.Use("/groups/:id/settlement-suggestions", func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(1))
		return c.Next()
	})

	app.Get("/groups/:id/settlement-suggestions", handler.GetSettlementSuggestions)

	// 創建測試資料
	user1 := createTestUser(db, "user1@example.com", "user1")
	user2 := createTestUser(db, "user2@example.com", "user2")
	user3 := createTestUser(db, "user3@example.com", "user3")
	group := createTestGroup(db, "測試群組", "測試描述", user1.ID)

	// 添加用戶到群組
	addGroupMember(db, group.ID, user2.ID, "member")
	addGroupMember(db, group.ID, user3.ID, "member")

	// 創建交易記錄模擬場景：
	// user1 付了 300，user2 和 user3 各分 150
	transaction := createTestTransaction(db, group.ID, user1.ID, user1.ID, 300.0)
	createTestTransactionSplit(db, transaction.ID, user2.ID, 150.0)
	createTestTransactionSplit(db, transaction.ID, user3.ID, 150.0)

	tests := []struct {
		name           string
		userID         uint
		groupID        uint
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "成功獲取結算建議",
			userID:         user1.ID,
			groupID:        group.ID,
			expectedStatus: http.StatusOK,
			expectedCount:  2, // user2 和 user3 各需付 150 給 user1
		},
		{
			name:           "非群組成員無權限",
			userID:         99,
			groupID:        group.ID,
			expectedStatus: http.StatusNotFound, // 用戶不存在，會先返回 404
		},
		{
			name:           "無效的群組 ID",
			userID:         user1.ID,
			groupID:        999,
			expectedStatus: http.StatusForbidden, // RequireGroupMember 會返回 403
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 創建新的 app 實例避免中間件衝突
			testApp := fiber.New()
			testApp.Use("/groups/:id/settlement-suggestions", func(c *fiber.Ctx) error {
				c.Locals("user_id", tt.userID)
				return c.Next()
			})
			testApp.Get("/groups/:id/settlement-suggestions", handler.GetSettlementSuggestions)

			req := httptest.NewRequest("GET", fmt.Sprintf("/groups/%d/settlement-suggestions", tt.groupID), nil)
			resp, err := testApp.Test(req)
			if err != nil {
				t.Fatalf("無法執行請求: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("期望狀態碼 %d，得到 %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedStatus == http.StatusOK {
				var responseBody map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&responseBody)

				data, ok := responseBody["data"].([]interface{})
				if !ok {
					t.Fatal("回應資料格式錯誤")
				}

				if len(data) != tt.expectedCount {
					t.Errorf("期望 %d 個結算建議，得到 %d 個", tt.expectedCount, len(data))
				}

				// 驗證結算建議內容
				if len(data) > 0 {
					suggestion := data[0].(map[string]interface{})
					if suggestion["amount"] == nil || suggestion["from_user"] == nil || suggestion["to_user"] == nil {
						t.Error("結算建議內容不完整")
					}
				}
			}
		})
	}

	// 清理
	db.Delete(&models.TransactionSplit{}, "transaction_id = ?", transaction.ID)
	db.Delete(transaction)
	db.Delete(group)
	db.Delete(user1)
	db.Delete(user2)
	db.Delete(user3)
}
