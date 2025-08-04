package services

import (
	"errors"
	"split-go/internal/models"

	"gorm.io/gorm"
)

// ValidationService 驗證服務
type ValidationService struct {
	db *gorm.DB
}

// NewValidationService 創建驗證服務
func NewValidationService(db *gorm.DB) *ValidationService {
	return &ValidationService{db: db}
}

// ValidateGroupMember 驗證用戶是否為群組成員
func (s *ValidationService) ValidateGroupMember(groupID, userID uint) error {
	var member models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&member).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("用戶不是群組成員")
		}
		return errors.New("驗證群組成員失敗")
	}
	return nil
}

// ValidateMultipleGroupMembers 驗證多個用戶是否都是群組成員
func (s *ValidationService) ValidateMultipleGroupMembers(groupID uint, userIDs []uint) error {
	var memberCount int64
	if err := s.db.Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id IN ?", groupID, userIDs).
		Count(&memberCount).Error; err != nil {
		return errors.New("驗證群組成員失敗")
	}

	if int(memberCount) != len(userIDs) {
		return errors.New("部分用戶不是群組成員")
	}

	return nil
}
