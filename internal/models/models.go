package models

import (
	"time"

	"gorm.io/gorm"
)

// 用戶模型
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"`
	Name      string         `json:"name"`
	Avatar    string         `json:"avatar"`
	FCMToken  string         `json:"-"` // Firebase Cloud Messaging token
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// 群組模型
type Group struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	CreatedBy   uint           `json:"created_by"`
	Creator     User           `json:"creator" gorm:"foreignKey:CreatedBy"`
	Members     []User         `json:"members" gorm:"many2many:group_members;"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// 群組成員關聯表
type GroupMember struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	GroupID  uint      `json:"group_id"`
	UserID   uint      `json:"user_id"`
	Role     string    `json:"role" gorm:"default:'member'"` // member, admin
	JoinedAt time.Time `json:"joined_at"`
}

// 交易分類
type Category struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Name  string `json:"name" gorm:"not null"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

// 交易記錄
type Transaction struct {
	ID          uint               `json:"id" gorm:"primaryKey"`
	GroupID     uint               `json:"group_id" gorm:"not null"`
	Group       Group              `json:"group" gorm:"foreignKey:GroupID"`
	Description string             `json:"description" gorm:"not null"`
	Amount      float64            `json:"amount" gorm:"not null"`
	Currency    string             `json:"currency" gorm:"default:'TWD'"`
	CategoryID  uint               `json:"category_id"`
	Category    Category           `json:"category" gorm:"foreignKey:CategoryID"`
	PaidBy      uint               `json:"paid_by" gorm:"not null"`
	Payer       User               `json:"payer" gorm:"foreignKey:PaidBy"`
	Splits      []TransactionSplit `json:"splits" gorm:"foreignKey:TransactionID"`
	Receipt     string             `json:"receipt"` // 收據圖片 URL
	Notes       string             `json:"notes"`
	CreatedBy   uint               `json:"created_by"`
	Creator     User               `json:"creator" gorm:"foreignKey:CreatedBy"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	DeletedAt   gorm.DeletedAt     `json:"-" gorm:"index"`
}

// 分帳類型枚舉
type SplitType string

const (
	SplitEqual      SplitType = "equal"      // 平均分攤
	SplitPercentage SplitType = "percentage" // 按比例分攤
	SplitFixed      SplitType = "fixed"      // 固定金額
)

// 交易分攤記錄
type TransactionSplit struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	TransactionID uint      `json:"transaction_id" gorm:"not null"`
	UserID        uint      `json:"user_id" gorm:"not null"`
	User          User      `json:"user" gorm:"foreignKey:UserID"`
	Amount        float64   `json:"amount" gorm:"not null"`
	Percentage    float64   `json:"percentage"` // 百分比 (0-100)
	SplitType     SplitType `json:"split_type" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// 結算記錄
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

// 平衡計算結果 (用於 API 回應)
type Balance struct {
	UserID  uint    `json:"user_id"`
	User    User    `json:"user"`
	Balance float64 `json:"balance"` // 正數表示應收，負數表示應付
	Paid    float64 `json:"paid"`    // 總共支付金額
	Owed    float64 `json:"owed"`    // 總共應付金額
}

// 結算建議 (用於 API 回應)
type SettlementSuggestion struct {
	FromUserID uint    `json:"from_user_id"`
	FromUser   User    `json:"from_user"`
	ToUserID   uint    `json:"to_user_id"`
	ToUser     User    `json:"to_user"`
	Amount     float64 `json:"amount"`
}
