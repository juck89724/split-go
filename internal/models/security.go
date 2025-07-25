package models

import (
	"time"

	"gorm.io/datatypes"
)

// SecurityEvent 安全事件記錄
type SecurityEvent struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	SessionID string         `json:"session_id"`
	EventType string         `json:"event_type" gorm:"not null"` // login, refresh, logout, suspicious
	EventData datatypes.JSON `json:"event_data"`
	IPAddress string         `json:"ip_address"`
	CreatedAt time.Time      `json:"created_at"`

	// 關聯
	User    User        `json:"user" gorm:"foreignKey:UserID"`
	Session UserSession `json:"session" gorm:"foreignKey:SessionID"`
}
