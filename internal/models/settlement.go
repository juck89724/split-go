package models

import (
	"time"

	"gorm.io/gorm"
)

// Settlement 結算記錄
type Settlement struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	GroupID    uint           `json:"group_id" gorm:"not null"`
	Group      Group          `json:"group" gorm:"foreignKey:GroupID"`
	FromUserID uint           `json:"from_user_id" gorm:"not null"`
	FromUser   User           `json:"from_user" gorm:"foreignKey:FromUserID"`
	ToUserID   uint           `json:"to_user_id" gorm:"not null"`
	ToUser     User           `json:"to_user" gorm:"foreignKey:ToUserID"`
	Amount     float64        `json:"amount" gorm:"not null"`
	Currency   string         `json:"currency" gorm:"default:'TWD'"`
	Status     string         `json:"status" gorm:"default:'pending'"` // pending, paid, cancelled
	SettledAt  *time.Time     `json:"settled_at"`
	Notes      string         `json:"notes"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// Balance 平衡計算結果 (用於 API 回應)
type Balance struct {
	UserID  uint    `json:"user_id"`
	User    User    `json:"user"`
	Balance float64 `json:"balance"` // 正數表示應收，負數表示應付
	Paid    float64 `json:"paid"`    // 總共支付金額
	Owed    float64 `json:"owed"`    // 總共應付金額
}

// SettlementSuggestion 結算建議 (用於 API 回應)
type SettlementSuggestion struct {
	FromUserID uint    `json:"from_user_id"`
	FromUser   User    `json:"from_user"`
	ToUserID   uint    `json:"to_user_id"`
	ToUser     User    `json:"to_user"`
	Amount     float64 `json:"amount"`
}
