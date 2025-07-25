package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// User 用戶模型
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

// UserSession 用戶會話表（企業級核心）
type UserSession struct {
	ID                string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID            uint           `json:"user_id" gorm:"not null"`
	DeviceID          string         `json:"device_id" gorm:"not null"`
	DeviceFingerprint datatypes.JSON `json:"device_fingerprint" gorm:"not null"`

	// Token 管理
	RefreshTokenHash   string `json:"-" gorm:"not null"`
	AccessTokenVersion int    `json:"access_token_version" gorm:"default:1"`

	// 設備信息
	DeviceName string `json:"device_name"`
	DeviceType string `json:"device_type"` // mobile/desktop/tablet
	UserAgent  string `json:"-"`

	// 地理位置
	IPAddress string `json:"ip_address"`
	Country   string `json:"country"`
	City      string `json:"city"`

	// 狀態管理
	TrustLevel   int        `json:"trust_level" gorm:"default:0"` // 0:新設備 1:可信 2:高度可信
	LastActivity time.Time  `json:"last_activity" gorm:"default:CURRENT_TIMESTAMP"`
	ExpiresAt    time.Time  `json:"expires_at" gorm:"not null"`
	RevokedAt    *time.Time `json:"revoked_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 關聯
	User User `json:"user" gorm:"foreignKey:UserID"`
}
