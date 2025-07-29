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

// 設置群組測試資料庫
func setupGroupTestDB() *gorm.DB {
	db := setupTestDB()

	// 創建群組相關表
	err := db.AutoMigrate(&models.Group{}, &models.GroupMember{})
	if err != nil {
		panic("無法執行群組表遷移")
	}

	return db
}

// 創建測試群組
func createTestGroup(db *gorm.DB, name string, description string, creatorID uint) *models.Group {
	group := &models.Group{
		Name:        name,
		Description: description,
		CreatedBy:   creatorID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	db.Create(group)

	// 添加創建者為管理員成員
	member := &models.GroupMember{
		GroupID:  group.ID,
		UserID:   creatorID,
		Role:     "admin",
		JoinedAt: time.Now(),
	}
	db.Create(member)

	return group
}

// 添加群組成員
func addGroupMember(db *gorm.DB, groupID, userID uint, role string) {
	member := &models.GroupMember{
		GroupID:  groupID,
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now(),
	}
	db.Create(member)
}

// 測試獲取用戶群組列表
func TestGetUserGroups(t *testing.T) {
	db := setupGroupTestDB()
	handler := handlers.NewGroupHandler(db)

	// 創建測試用戶
	user1 := createTestUser(db, "user1@example.com", "user1")
	user2 := createTestUser(db, "user2@example.com", "user2")

	// 創建測試群組
	_ = createTestGroup(db, "測試群組1", "這是測試群組1", user1.ID) // user1 作為創建者
	group2 := createTestGroup(db, "測試群組2", "這是測試群組2", user2.ID)

	// 將 user1 加入 group2
	addGroupMember(db, group2.ID, user1.ID, "member")

	t.Run("成功獲取用戶群組列表", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups", func(c *fiber.Ctx) error {
			c.Locals("user_id", user1.ID)
			return c.Next()
		})
		app.Get("/groups", handler.GetUserGroups)

		req := httptest.NewRequest("GET", "/groups", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusOK, resp.StatusCode)
		}

		var responseBody map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&responseBody)
		if err != nil {
			t.Fatalf("無法解析回應: %v", err)
		}

		if responseBody["error"].(bool) {
			t.Errorf("期望成功回應，得到錯誤: %v", responseBody["message"])
		}

		// 檢查群組數量（user1 應該有 2 個群組）
		data := responseBody["data"].([]interface{})
		if len(data) != 2 {
			t.Errorf("期望 2 個群組，得到 %d 個", len(data))
		}
	})

	t.Run("未認證用戶無法獲取群組列表", func(t *testing.T) {
		app := fiber.New()
		app.Get("/groups", handler.GetUserGroups)

		req := httptest.NewRequest("GET", "/groups", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusUnauthorized, resp.StatusCode)
		}
	})
}

// 測試創建群組
func TestCreateGroup(t *testing.T) {
	db := setupGroupTestDB()
	handler := handlers.NewGroupHandler(db)

	user := createTestUser(db, "creator@example.com", "creator")

	t.Run("成功創建群組", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups", func(c *fiber.Ctx) error {
			c.Locals("user_id", user.ID)
			return c.Next()
		})
		app.Post("/groups", handler.CreateGroup)

		reqBody := map[string]interface{}{
			"name":        "新群組",
			"description": "這是一個新的測試群組",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/groups", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusCreated, resp.StatusCode)
		}

		var responseBody map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&responseBody)
		if err != nil {
			t.Fatalf("無法解析回應: %v", err)
		}

		if responseBody["error"].(bool) {
			t.Errorf("期望成功回應，得到錯誤: %v", responseBody["message"])
		}

		// 檢查群組是否真的被創建
		var group models.Group
		err = db.Where("name = ?", "新群組").First(&group).Error
		if err != nil {
			t.Errorf("群組未被創建到資料庫")
		}

		// 檢查創建者是否被加為管理員
		var member models.GroupMember
		err = db.Where("group_id = ? AND user_id = ? AND role = ?", group.ID, user.ID, "admin").First(&member).Error
		if err != nil {
			t.Errorf("創建者未被設為管理員")
		}
	})

	t.Run("群組名稱為空時創建失敗", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups", func(c *fiber.Ctx) error {
			c.Locals("user_id", user.ID)
			return c.Next()
		})
		app.Post("/groups", handler.CreateGroup)

		reqBody := map[string]interface{}{
			"name":        "",
			"description": "描述",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/groups", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("無效JSON格式創建失敗", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups", func(c *fiber.Ctx) error {
			c.Locals("user_id", user.ID)
			return c.Next()
		})
		app.Post("/groups", handler.CreateGroup)

		req := httptest.NewRequest("POST", "/groups", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusBadRequest, resp.StatusCode)
		}
	})
}

// 測試獲取群組詳細資訊
func TestGetGroup(t *testing.T) {
	db := setupGroupTestDB()
	handler := handlers.NewGroupHandler(db)

	user1 := createTestUser(db, "user1@example.com", "user1")
	user2 := createTestUser(db, "user2@example.com", "user2")
	user3 := createTestUser(db, "user3@example.com", "user3")

	group := createTestGroup(db, "測試群組", "群組描述", user1.ID)
	addGroupMember(db, group.ID, user2.ID, "member")

	t.Run("群組成員成功獲取群組詳情", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups/:id", func(c *fiber.Ctx) error {
			c.Locals("user_id", user2.ID)
			return c.Next()
		})
		app.Get("/groups/:id", handler.GetGroup)

		req := httptest.NewRequest("GET", fmt.Sprintf("/groups/%d", group.ID), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusOK, resp.StatusCode)
		}

		var responseBody map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&responseBody)
		if err != nil {
			t.Fatalf("無法解析回應: %v", err)
		}

		if responseBody["error"].(bool) {
			t.Errorf("期望成功回應，得到錯誤: %v", responseBody["message"])
		}

		data := responseBody["data"].(map[string]interface{})
		if data["name"] != "測試群組" {
			t.Errorf("期望群組名稱 '測試群組'，得到 '%v'", data["name"])
		}

		// 檢查成員列表
		members := data["members"].([]interface{})
		if len(members) != 2 {
			t.Errorf("期望 2 個成員，得到 %d 個", len(members))
		}

		// 檢查用戶角色
		if data["my_role"] != "member" {
			t.Errorf("期望角色 'member'，得到 '%v'", data["my_role"])
		}
	})

	t.Run("非群組成員無法獲取群組詳情", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups/:id", func(c *fiber.Ctx) error {
			c.Locals("user_id", user3.ID)
			return c.Next()
		})
		app.Get("/groups/:id", handler.GetGroup)

		req := httptest.NewRequest("GET", fmt.Sprintf("/groups/%d", group.ID), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusForbidden, resp.StatusCode)
		}
	})

	t.Run("群組不存在", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups/:id", func(c *fiber.Ctx) error {
			c.Locals("user_id", user1.ID)
			return c.Next()
		})
		app.Get("/groups/:id", handler.GetGroup)

		req := httptest.NewRequest("GET", "/groups/99999", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		// 因為 RequireGroupMember 會先檢查成員資格，所以不存在的群組會返回 403 而不是 404
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusForbidden, resp.StatusCode)
		}
	})

	t.Run("無效的群組ID", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups/:id", func(c *fiber.Ctx) error {
			c.Locals("user_id", user1.ID)
			return c.Next()
		})
		app.Get("/groups/:id", handler.GetGroup)

		req := httptest.NewRequest("GET", "/groups/invalid", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusBadRequest, resp.StatusCode)
		}
	})
}

// 測試更新群組
func TestUpdateGroup(t *testing.T) {
	db := setupGroupTestDB()
	handler := handlers.NewGroupHandler(db)

	user1 := createTestUser(db, "admin@example.com", "admin")
	user2 := createTestUser(db, "member@example.com", "member")

	group := createTestGroup(db, "原始群組名", "原始描述", user1.ID)
	addGroupMember(db, group.ID, user2.ID, "member")

	t.Run("管理員成功更新群組", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups/:id", func(c *fiber.Ctx) error {
			c.Locals("user_id", user1.ID)
			return c.Next()
		})
		app.Put("/groups/:id", handler.UpdateGroup)

		reqBody := map[string]interface{}{
			"name":        "更新後的群組名",
			"description": "更新後的描述",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", fmt.Sprintf("/groups/%d", group.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusOK, resp.StatusCode)
		}

		// 檢查資料庫中的群組是否被更新
		var updatedGroup models.Group
		db.First(&updatedGroup, group.ID)
		if updatedGroup.Name != "更新後的群組名" {
			t.Errorf("群組名稱未被更新")
		}
		if updatedGroup.Description != "更新後的描述" {
			t.Errorf("群組描述未被更新")
		}
	})

	t.Run("一般成員無法更新群組", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups/:id", func(c *fiber.Ctx) error {
			c.Locals("user_id", user2.ID)
			return c.Next()
		})
		app.Put("/groups/:id", handler.UpdateGroup)

		reqBody := map[string]interface{}{
			"name":        "成員嘗試更新",
			"description": "這不應該成功",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", fmt.Sprintf("/groups/%d", group.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusForbidden, resp.StatusCode)
		}
	})

	t.Run("空的群組名稱更新失敗", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups/:id", func(c *fiber.Ctx) error {
			c.Locals("user_id", user1.ID)
			return c.Next()
		})
		app.Put("/groups/:id", handler.UpdateGroup)

		reqBody := map[string]interface{}{
			"name":        "",
			"description": "有效描述",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", fmt.Sprintf("/groups/%d", group.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusBadRequest, resp.StatusCode)
		}
	})
}

// 測試刪除群組
func TestDeleteGroup(t *testing.T) {
	db := setupGroupTestDB()
	handler := handlers.NewGroupHandler(db)

	user1 := createTestUser(db, "creator@example.com", "creator")
	user2 := createTestUser(db, "admin@example.com", "admin")
	user3 := createTestUser(db, "member@example.com", "member")

	t.Run("創建者成功刪除群組", func(t *testing.T) {
		// 為每個測試創建新的群組，避免測試間影響
		group := createTestGroup(db, "要刪除的群組", "將被刪除", user1.ID)
		addGroupMember(db, group.ID, user2.ID, "admin")
		addGroupMember(db, group.ID, user3.ID, "member")

		app := fiber.New()
		app.Use("/groups/:id", func(c *fiber.Ctx) error {
			c.Locals("user_id", user1.ID)
			return c.Next()
		})
		app.Delete("/groups/:id", handler.DeleteGroup)

		req := httptest.NewRequest("DELETE", fmt.Sprintf("/groups/%d", group.ID), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusOK, resp.StatusCode)
		}

		// 檢查群組是否被軟刪除
		var deletedGroup models.Group
		err = db.Unscoped().First(&deletedGroup, group.ID).Error
		if err != nil {
			t.Errorf("無法找到被刪除的群組")
		}
		if deletedGroup.DeletedAt.Time.IsZero() {
			t.Errorf("群組未被軟刪除")
		}

		// 檢查群組成員是否被清理
		var memberCount int64
		db.Model(&models.GroupMember{}).Where("group_id = ?", group.ID).Count(&memberCount)
		if memberCount != 0 {
			t.Errorf("群組成員未被清理，剩餘 %d 個", memberCount)
		}
	})

	t.Run("管理員無法刪除群組", func(t *testing.T) {
		// 創建新群組用於測試
		group2 := createTestGroup(db, "另一個群組", "用於測試", user1.ID)
		addGroupMember(db, group2.ID, user2.ID, "admin")

		app := fiber.New()
		app.Use("/groups/:id", func(c *fiber.Ctx) error {
			c.Locals("user_id", user2.ID)
			return c.Next()
		})
		app.Delete("/groups/:id", handler.DeleteGroup)

		req := httptest.NewRequest("DELETE", fmt.Sprintf("/groups/%d", group2.ID), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusForbidden, resp.StatusCode)
		}
	})

	t.Run("刪除不存在的群組", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups/:id", func(c *fiber.Ctx) error {
			c.Locals("user_id", user1.ID)
			return c.Next()
		})
		app.Delete("/groups/:id", handler.DeleteGroup)

		req := httptest.NewRequest("DELETE", "/groups/99999", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusNotFound, resp.StatusCode)
		}
	})
}

// 測試添加群組成員
func TestAddMember(t *testing.T) {
	db := setupGroupTestDB()
	handler := handlers.NewGroupHandler(db)

	admin := createTestUser(db, "admin@example.com", "admin")
	member := createTestUser(db, "member@example.com", "member")
	newUser := createTestUser(db, "new@example.com", "newuser")

	group := createTestGroup(db, "測試群組", "用於成員管理測試", admin.ID)
	addGroupMember(db, group.ID, member.ID, "member")

	t.Run("管理員成功添加成員", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups/:id/members", func(c *fiber.Ctx) error {
			c.Locals("user_id", admin.ID)
			return c.Next()
		})
		app.Post("/groups/:id/members", handler.AddMember)

		reqBody := map[string]interface{}{
			"user_id": newUser.ID,
			"role":    "member",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", fmt.Sprintf("/groups/%d/members", group.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusCreated, resp.StatusCode)
		}

		// 檢查成員是否被添加到資料庫
		var groupMember models.GroupMember
		err = db.Where("group_id = ? AND user_id = ?", group.ID, newUser.ID).First(&groupMember).Error
		if err != nil {
			t.Errorf("成員未被添加到資料庫")
		}
		if groupMember.Role != "member" {
			t.Errorf("期望角色 'member'，得到 '%s'", groupMember.Role)
		}
	})

	t.Run("一般成員無法添加成員", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups/:id/members", func(c *fiber.Ctx) error {
			c.Locals("user_id", member.ID)
			return c.Next()
		})
		app.Post("/groups/:id/members", handler.AddMember)

		anotherUser := createTestUser(db, "another@example.com", "another")

		reqBody := map[string]interface{}{
			"user_id": anotherUser.ID,
			"role":    "member",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", fmt.Sprintf("/groups/%d/members", group.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusForbidden, resp.StatusCode)
		}
	})

	t.Run("添加已存在的成員失敗", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups/:id/members", func(c *fiber.Ctx) error {
			c.Locals("user_id", admin.ID)
			return c.Next()
		})
		app.Post("/groups/:id/members", handler.AddMember)

		reqBody := map[string]interface{}{
			"user_id": member.ID,
			"role":    "member",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", fmt.Sprintf("/groups/%d/members", group.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusConflict, resp.StatusCode)
		}
	})

	t.Run("添加不存在的用戶失敗", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups/:id/members", func(c *fiber.Ctx) error {
			c.Locals("user_id", admin.ID)
			return c.Next()
		})
		app.Post("/groups/:id/members", handler.AddMember)

		reqBody := map[string]interface{}{
			"user_id": 99999,
			"role":    "member",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", fmt.Sprintf("/groups/%d/members", group.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusNotFound, resp.StatusCode)
		}
	})
}

// 測試移除群組成員
func TestRemoveMember(t *testing.T) {
	db := setupGroupTestDB()
	handler := handlers.NewGroupHandler(db)

	creator := createTestUser(db, "creator@example.com", "creator")
	admin := createTestUser(db, "admin@example.com", "admin")
	member1 := createTestUser(db, "member1@example.com", "member1")
	member2 := createTestUser(db, "member2@example.com", "member2")

	group := createTestGroup(db, "測試群組", "用於成員移除測試", creator.ID)
	addGroupMember(db, group.ID, admin.ID, "admin")
	addGroupMember(db, group.ID, member1.ID, "member")
	addGroupMember(db, group.ID, member2.ID, "member")

	t.Run("管理員成功移除成員", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups/:id/members/:userId", func(c *fiber.Ctx) error {
			c.Locals("user_id", admin.ID)
			return c.Next()
		})
		app.Delete("/groups/:id/members/:userId", handler.RemoveMember)

		req := httptest.NewRequest("DELETE", fmt.Sprintf("/groups/%d/members/%d", group.ID, member1.ID), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusOK, resp.StatusCode)
		}

		// 檢查成員是否被移除
		var count int64
		db.Model(&models.GroupMember{}).Where("group_id = ? AND user_id = ?", group.ID, member1.ID).Count(&count)
		if count != 0 {
			t.Errorf("成員未被移除")
		}
	})

	t.Run("一般成員無法移除其他成員", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups/:id/members/:userId", func(c *fiber.Ctx) error {
			c.Locals("user_id", member2.ID)
			return c.Next()
		})
		app.Delete("/groups/:id/members/:userId", handler.RemoveMember)

		req := httptest.NewRequest("DELETE", fmt.Sprintf("/groups/%d/members/%d", group.ID, admin.ID), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusForbidden, resp.StatusCode)
		}
	})

	t.Run("無法移除自己", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups/:id/members/:userId", func(c *fiber.Ctx) error {
			c.Locals("user_id", admin.ID)
			return c.Next()
		})
		app.Delete("/groups/:id/members/:userId", handler.RemoveMember)

		req := httptest.NewRequest("DELETE", fmt.Sprintf("/groups/%d/members/%d", group.ID, admin.ID), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("無法移除群組創建者", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups/:id/members/:userId", func(c *fiber.Ctx) error {
			c.Locals("user_id", admin.ID)
			return c.Next()
		})
		app.Delete("/groups/:id/members/:userId", handler.RemoveMember)

		req := httptest.NewRequest("DELETE", fmt.Sprintf("/groups/%d/members/%d", group.ID, creator.ID), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("移除不存在的成員", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups/:id/members/:userId", func(c *fiber.Ctx) error {
			c.Locals("user_id", admin.ID)
			return c.Next()
		})
		app.Delete("/groups/:id/members/:userId", handler.RemoveMember)

		req := httptest.NewRequest("DELETE", fmt.Sprintf("/groups/%d/members/99999", group.ID), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusNotFound, resp.StatusCode)
		}
	})

	t.Run("無效的用戶ID", func(t *testing.T) {
		app := fiber.New()
		app.Use("/groups/:id/members/:userId", func(c *fiber.Ctx) error {
			c.Locals("user_id", admin.ID)
			return c.Next()
		})
		app.Delete("/groups/:id/members/:userId", handler.RemoveMember)

		req := httptest.NewRequest("DELETE", fmt.Sprintf("/groups/%d/members/invalid", group.ID), nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("無法執行請求: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("期望狀態碼 %d，得到 %d", http.StatusBadRequest, resp.StatusCode)
		}
	})
}
