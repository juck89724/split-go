package responses

import (
	"split-go/internal/models"
	"time"
)

// TransactionSplitResponse 分帳回應結構
type TransactionSplitResponse struct {
	ID         uint               `json:"id"`
	User       UserSimpleResponse `json:"user"`
	Amount     float64            `json:"amount"`
	Percentage float64            `json:"percentage"`
	SplitType  models.SplitType   `json:"split_type"`
}

// NewTransactionSplitResponse 創建分帳回應
func NewTransactionSplitResponse(split models.TransactionSplit) TransactionSplitResponse {
	return TransactionSplitResponse{
		ID:         split.ID,
		User:       NewUserSimpleResponse(split.User),
		Amount:     split.Amount,
		Percentage: split.Percentage,
		SplitType:  split.SplitType,
	}
}

// TransactionResponse 交易回應結構
type TransactionResponse struct {
	ID          uint                       `json:"id"`
	Description string                     `json:"description"`
	Amount      float64                    `json:"amount"`
	Currency    string                     `json:"currency"`
	Group       GroupSimpleResponse        `json:"group"`
	Category    *CategoryResponse          `json:"category,omitempty"`
	Payer       UserSimpleResponse         `json:"payer"`
	Creator     UserSimpleResponse         `json:"creator"`
	Splits      []TransactionSplitResponse `json:"splits"`
	Receipt     string                     `json:"receipt,omitempty"`
	Notes       string                     `json:"notes,omitempty"`
	CreatedAt   time.Time                  `json:"created_at"`
	UpdatedAt   time.Time                  `json:"updated_at"`

	// 計算欄位（基於當前用戶）
	MyAmount  float64 `json:"my_amount"`  // 我需要付的金額
	MyBalance float64 `json:"my_balance"` // 我的平衡狀況 (付出 - 應付)
	AmIPayer  bool    `json:"am_i_payer"` // 我是否為付款者

	// 權限欄位
	CanEdit   bool `json:"can_edit"`   // 是否可以編輯
	CanDelete bool `json:"can_delete"` // 是否可以刪除
}

// NewTransactionResponse 創建交易回應
func NewTransactionResponse(tx models.Transaction, currentUserID uint) TransactionResponse {
	splits := make([]TransactionSplitResponse, len(tx.Splits))
	for i, split := range tx.Splits {
		splits[i] = NewTransactionSplitResponse(split)
	}

	// 計算當前用戶的應付金額
	myAmount := 0.0
	for _, split := range tx.Splits {
		if split.UserID == currentUserID {
			myAmount = split.Amount
			break
		}
	}

	// 計算平衡狀況
	var myBalance float64
	if tx.PaidBy == currentUserID {
		// 我是付款者：我付的錢 - 我應該付的錢
		myBalance = tx.Amount - myAmount
	} else {
		// 我不是付款者：0 - 我應該付的錢 = 負數（我欠錢）
		myBalance = -myAmount
	}

	// 權限檢查
	canEdit := tx.CreatedBy == currentUserID || tx.PaidBy == currentUserID
	canDelete := tx.CreatedBy == currentUserID

	// 處理 Category（可能為空）
	var categoryResponse *CategoryResponse
	if tx.CategoryID != 0 && tx.Category.ID != 0 {
		cat := NewCategoryResponse(tx.Category)
		categoryResponse = &cat
	}

	return TransactionResponse{
		ID:          tx.ID,
		Description: tx.Description,
		Amount:      tx.Amount,
		Currency:    tx.Currency,
		Group:       NewGroupSimpleResponse(tx.Group),
		Category:    categoryResponse,
		Payer:       NewUserSimpleResponse(tx.Payer),
		Creator:     NewUserSimpleResponse(tx.Creator),
		Splits:      splits,
		Receipt:     tx.Receipt,
		Notes:       tx.Notes,
		CreatedAt:   tx.CreatedAt,
		UpdatedAt:   tx.UpdatedAt,
		MyAmount:    myAmount,
		MyBalance:   myBalance,
		AmIPayer:    tx.PaidBy == currentUserID,
		CanEdit:     canEdit,
		CanDelete:   canDelete,
	}
}

// NewTransactionResponseList 批量轉換交易列表
func NewTransactionResponseList(transactions []models.Transaction, currentUserID uint) []TransactionResponse {
	responses := make([]TransactionResponse, len(transactions))
	for i, tx := range transactions {
		responses[i] = NewTransactionResponse(tx, currentUserID)
	}
	return responses
}

// TransactionSimpleResponse 簡化的交易回應（用於列表頁面）
type TransactionSimpleResponse struct {
	ID          uint                `json:"id"`
	Description string              `json:"description"`
	Amount      float64             `json:"amount"`
	Currency    string              `json:"currency"`
	Group       GroupSimpleResponse `json:"group"`
	Category    *CategoryResponse   `json:"category,omitempty"`
	Payer       UserSimpleResponse  `json:"payer"`
	CreatedAt   time.Time           `json:"created_at"`

	// 簡化的計算欄位
	MyAmount float64 `json:"my_amount"`
	AmIPayer bool    `json:"am_i_payer"`
}

// NewTransactionSimpleResponse 創建簡化的交易回應
func NewTransactionSimpleResponse(tx models.Transaction, currentUserID uint) TransactionSimpleResponse {
	// 計算當前用戶的應付金額
	myAmount := 0.0
	for _, split := range tx.Splits {
		if split.UserID == currentUserID {
			myAmount = split.Amount
			break
		}
	}

	// 處理 Category
	var categoryResponse *CategoryResponse
	if tx.CategoryID != 0 && tx.Category.ID != 0 {
		cat := NewCategoryResponse(tx.Category)
		categoryResponse = &cat
	}

	return TransactionSimpleResponse{
		ID:          tx.ID,
		Description: tx.Description,
		Amount:      tx.Amount,
		Currency:    tx.Currency,
		Group:       NewGroupSimpleResponse(tx.Group),
		Category:    categoryResponse,
		Payer:       NewUserSimpleResponse(tx.Payer),
		CreatedAt:   tx.CreatedAt,
		MyAmount:    myAmount,
		AmIPayer:    tx.PaidBy == currentUserID,
	}
}

// NewTransactionSimpleResponseList 批量轉換簡化交易列表
func NewTransactionSimpleResponseList(transactions []models.Transaction, currentUserID uint) []TransactionSimpleResponse {
	responses := make([]TransactionSimpleResponse, len(transactions))
	for i, tx := range transactions {
		responses[i] = NewTransactionSimpleResponse(tx, currentUserID)
	}
	return responses
}
