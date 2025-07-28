package handlers_test

import (
	"split-go/internal/models"
	"split-go/internal/responses"
	"testing"
	"time"
)

func TestNewTransactionResponse(t *testing.T) {
	// 模擬測試資料
	now := time.Now()

	// 創建測試用的 Transaction
	transaction := models.Transaction{
		ID:          123,
		Description: "測試聚餐",
		Amount:      1200.0,
		Currency:    "TWD",
		PaidBy:      1, // Alice 付款
		CreatedBy:   2, // Bob 創建
		CreatedAt:   now,
		UpdatedAt:   now,

		// 關聯資料
		Payer: models.User{
			ID:   1,
			Name: "Alice",
		},
		Creator: models.User{
			ID:   2,
			Name: "Bob",
		},
		Group: models.Group{
			ID:   1,
			Name: "室友分帳群",
		},
		Category: models.Category{
			ID:    1,
			Name:  "餐飲",
			Icon:  "🍽️",
			Color: "#FF6B6B",
		},
		Splits: []models.TransactionSplit{
			{
				ID:         1,
				UserID:     1, // Alice
				Amount:     400.0,
				Percentage: 33.33,
				SplitType:  models.SplitEqual,
				User: models.User{
					ID:   1,
					Name: "Alice",
				},
			},
			{
				ID:         2,
				UserID:     2, // Bob
				Amount:     400.0,
				Percentage: 33.33,
				SplitType:  models.SplitEqual,
				User: models.User{
					ID:   2,
					Name: "Bob",
				},
			},
			{
				ID:         3,
				UserID:     3, // Charlie
				Amount:     400.0,
				Percentage: 33.34,
				SplitType:  models.SplitEqual,
				User: models.User{
					ID:   3,
					Name: "Charlie",
				},
			},
		},
	}

	// 測試案例：當前用戶是付款者 (Alice)
	t.Run("Current user is payer", func(t *testing.T) {
		currentUserID := uint(1) // Alice
		response := responses.NewTransactionResponse(transaction, currentUserID)

		// 驗證基本資訊
		if response.ID != 123 {
			t.Errorf("Expected ID 123, got %d", response.ID)
		}
		if response.Description != "測試聚餐" {
			t.Errorf("Expected description '測試聚餐', got %s", response.Description)
		}
		if response.Amount != 1200.0 {
			t.Errorf("Expected amount 1200.0, got %f", response.Amount)
		}

		// 驗證個人化欄位
		if response.MyAmount != 400.0 {
			t.Errorf("Expected MyAmount 400.0, got %f", response.MyAmount)
		}
		if response.MyBalance != 800.0 { // 1200 - 400 = 800 (我付了1200，但只需要付400)
			t.Errorf("Expected MyBalance 800.0, got %f", response.MyBalance)
		}
		if !response.AmIPayer {
			t.Error("Expected AmIPayer to be true")
		}

		// 驗證權限欄位
		if !response.CanEdit { // 付款者可以編輯
			t.Error("Expected CanEdit to be true for payer")
		}
		if response.CanDelete { // 付款者不是創建者，不能刪除
			t.Error("Expected CanDelete to be false for payer who is not creator")
		}

		// 驗證嵌套資料
		if response.Payer.Name != "Alice" {
			t.Errorf("Expected payer name 'Alice', got %s", response.Payer.Name)
		}
		if response.Group.Name != "室友分帳群" {
			t.Errorf("Expected group name '室友分帳群', got %s", response.Group.Name)
		}
		if len(response.Splits) != 3 {
			t.Errorf("Expected 3 splits, got %d", len(response.Splits))
		}
	})

	// 測試案例：當前用戶不是付款者 (Bob)
	t.Run("Current user is not payer", func(t *testing.T) {
		currentUserID := uint(2) // Bob
		response := responses.NewTransactionResponse(transaction, currentUserID)

		// 驗證個人化欄位
		if response.MyAmount != 400.0 {
			t.Errorf("Expected MyAmount 400.0, got %f", response.MyAmount)
		}
		if response.MyBalance != -400.0 { // 0 - 400 = -400 (我欠400)
			t.Errorf("Expected MyBalance -400.0, got %f", response.MyBalance)
		}
		if response.AmIPayer {
			t.Error("Expected AmIPayer to be false")
		}

		// 驗證權限欄位
		if !response.CanEdit { // 創建者可以編輯
			t.Error("Expected CanEdit to be true for creator")
		}
		if !response.CanDelete { // 創建者可以刪除
			t.Error("Expected CanDelete to be true for creator")
		}
	})

	// 測試案例：當前用戶既不是付款者也不是創建者 (Charlie)
	t.Run("Current user is neither payer nor creator", func(t *testing.T) {
		currentUserID := uint(3) // Charlie
		response := responses.NewTransactionResponse(transaction, currentUserID)

		// 驗證個人化欄位
		if response.MyAmount != 400.0 {
			t.Errorf("Expected MyAmount 400.0, got %f", response.MyAmount)
		}
		if response.MyBalance != -400.0 { // 0 - 400 = -400 (我欠400)
			t.Errorf("Expected MyBalance -400.0, got %f", response.MyBalance)
		}
		if response.AmIPayer {
			t.Error("Expected AmIPayer to be false")
		}

		// 驗證權限欄位
		if response.CanEdit { // 既不是付款者也不是創建者，不能編輯
			t.Error("Expected CanEdit to be false")
		}
		if response.CanDelete { // 既不是付款者也不是創建者，不能刪除
			t.Error("Expected CanDelete to be false")
		}
	})
}

func TestNewTransactionSimpleResponse(t *testing.T) {
	// 創建簡化的測試資料
	transaction := models.Transaction{
		ID:          456,
		Description: "簡化測試",
		Amount:      500.0,
		Currency:    "TWD",
		PaidBy:      1,
		CreatedAt:   time.Now(),

		Payer: models.User{
			ID:   1,
			Name: "Alice",
		},
		Group: models.Group{
			ID:   1,
			Name: "測試群組",
		},
		Category: models.Category{
			ID:    2,
			Name:  "交通",
			Icon:  "🚗",
			Color: "#4ECDC4",
		},
		Splits: []models.TransactionSplit{
			{
				UserID: 2,
				Amount: 250.0,
			},
			{
				UserID: 3,
				Amount: 250.0,
			},
		},
	}

	currentUserID := uint(2)
	response := responses.NewTransactionSimpleResponse(transaction, currentUserID)

	// 驗證基本資訊
	if response.ID != 456 {
		t.Errorf("Expected ID 456, got %d", response.ID)
	}
	if response.Description != "簡化測試" {
		t.Errorf("Expected description '簡化測試', got %s", response.Description)
	}

	// 驗證簡化的個人化欄位
	if response.MyAmount != 250.0 {
		t.Errorf("Expected MyAmount 250.0, got %f", response.MyAmount)
	}
	if response.AmIPayer {
		t.Error("Expected AmIPayer to be false")
	}

	// 驗證嵌套資料存在
	if response.Payer.Name != "Alice" {
		t.Errorf("Expected payer name 'Alice', got %s", response.Payer.Name)
	}
	if response.Group.Name != "測試群組" {
		t.Errorf("Expected group name '測試群組', got %s", response.Group.Name)
	}
}
