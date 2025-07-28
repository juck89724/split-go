package models

import (
	"time"

	"gorm.io/gorm"
)

// Transaction 交易記錄
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

// SplitType 分帳類型枚舉
type SplitType string

const (
	SplitEqual      SplitType = "equal"      // 平均分攤
	SplitPercentage SplitType = "percentage" // 按比例分攤
	SplitFixed      SplitType = "fixed"      // 固定金額
)

// TransactionSplit 交易分攤記錄
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

// CreateTransactionRequest 創建交易的請求結構
type CreateTransactionRequest struct {
	GroupID     uint                      `json:"group_id" validate:"required"`
	Description string                    `json:"description" validate:"required,min=1,max=255"`
	Amount      float64                   `json:"amount" validate:"required,gt=0"`
	Currency    string                    `json:"currency"`
	CategoryID  uint                      `json:"category_id"`
	PaidBy      uint                      `json:"paid_by" validate:"required"`
	SplitType   SplitType                 `json:"split_type" validate:"required,oneof=equal percentage fixed"`
	Splits      []TransactionSplitRequest `json:"splits" validate:"required,min=1"`
	Receipt     string                    `json:"receipt"`
	Notes       string                    `json:"notes" validate:"max=500"`
}

// CreateTransactionSplit 創建分帳的請求結構
type TransactionSplitRequest struct {
	UserID     uint    `json:"user_id" validate:"required"`
	Amount     float64 `json:"amount"`     // 固定金額模式使用
	Percentage float64 `json:"percentage"` // 百分比模式使用 (0-100)
}

// UpdateTransactionRequest 更新交易的請求結構
type UpdateTransactionRequest struct {
	Description string                    `json:"description" validate:"omitempty,min=1,max=255"`
	Amount      float64                   `json:"amount" validate:"omitempty,gt=0"`
	Currency    string                    `json:"currency"`
	CategoryID  uint                      `json:"category_id"`
	PaidBy      uint                      `json:"paid_by"`
	SplitType   SplitType                 `json:"split_type" validate:"omitempty,oneof=equal percentage fixed"`
	Splits      []TransactionSplitRequest `json:"splits" validate:"omitempty,min=1"`
	Receipt     string                    `json:"receipt"`
	Notes       string                    `json:"notes" validate:"max=500"`
}
