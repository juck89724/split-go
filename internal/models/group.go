package models

import (
	"time"

	"gorm.io/gorm"
)

// Group 群組模型
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

// GroupMember 群組成員關聯表
type GroupMember struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	GroupID  uint      `json:"group_id"`
	UserID   uint      `json:"user_id"`
	Role     string    `json:"role" gorm:"default:'member'"` // member, admin
	JoinedAt time.Time `json:"joined_at"`
}
