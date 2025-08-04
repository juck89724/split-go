package responses

import (
	"split-go/internal/models"
	"time"
)

// SettlementResponse 結算記錄回應格式
type SettlementResponse struct {
	ID         uint          `json:"id"`
	GroupID    uint          `json:"group_id"`
	Group      GroupResponse `json:"group"`
	FromUserID uint          `json:"from_user_id"`
	FromUser   UserResponse  `json:"from_user"`
	ToUserID   uint          `json:"to_user_id"`
	ToUser     UserResponse  `json:"to_user"`
	Amount     float64       `json:"amount"`
	Currency   string        `json:"currency"`
	Status     string        `json:"status"`
	SettledAt  *time.Time    `json:"settled_at"`
	Notes      string        `json:"notes"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// NewSettlementResponse 創建結算回應
func NewSettlementResponse(settlement models.Settlement) SettlementResponse {
	return SettlementResponse{
		ID:         settlement.ID,
		GroupID:    settlement.GroupID,
		Group:      NewGroupResponse(settlement.Group),
		FromUserID: settlement.FromUserID,
		FromUser:   NewUserResponse(settlement.FromUser),
		ToUserID:   settlement.ToUserID,
		ToUser:     NewUserResponse(settlement.ToUser),
		Amount:     settlement.Amount,
		Currency:   settlement.Currency,
		Status:     settlement.Status,
		SettledAt:  settlement.SettledAt,
		Notes:      settlement.Notes,
		CreatedAt:  settlement.CreatedAt,
		UpdatedAt:  settlement.UpdatedAt,
	}
}

// BalanceResponse 平衡計算回應格式
type BalanceResponse struct {
	UserID  uint         `json:"user_id"`
	User    UserResponse `json:"user"`
	Balance float64      `json:"balance"` // 正數應收，負數應付
	Paid    float64      `json:"paid"`    // 總支付金額
	Owed    float64      `json:"owed"`    // 總應付金額
}

// NewBalanceResponse 創建平衡回應
func NewBalanceResponse(balance models.Balance) BalanceResponse {
	return BalanceResponse{
		UserID:  balance.UserID,
		User:    NewUserResponse(balance.User),
		Balance: balance.Balance,
		Paid:    balance.Paid,
		Owed:    balance.Owed,
	}
}

// SettlementSuggestionResponse 結算建議回應格式
type SettlementSuggestionResponse struct {
	FromUserID uint         `json:"from_user_id"`
	FromUser   UserResponse `json:"from_user"`
	ToUserID   uint         `json:"to_user_id"`
	ToUser     UserResponse `json:"to_user"`
	Amount     float64      `json:"amount"`
}

// NewSettlementSuggestionResponse 創建結算建議回應
func NewSettlementSuggestionResponse(suggestion models.SettlementSuggestion) SettlementSuggestionResponse {
	return SettlementSuggestionResponse{
		FromUserID: suggestion.FromUserID,
		FromUser:   NewUserResponse(suggestion.FromUser),
		ToUserID:   suggestion.ToUserID,
		ToUser:     NewUserResponse(suggestion.ToUser),
		Amount:     suggestion.Amount,
	}
}
