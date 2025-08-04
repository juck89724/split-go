package services

import (
	"split-go/internal/models"

	"gorm.io/gorm"
)

// BalanceService 平衡計算服務
type BalanceService struct {
	db *gorm.DB
}

// NewBalanceService 創建平衡計算服務
func NewBalanceService(db *gorm.DB) *BalanceService {
	return &BalanceService{db: db}
}

// CalculateGroupBalances 計算群組內每個用戶的平衡
func (s *BalanceService) CalculateGroupBalances(groupID uint) ([]models.Balance, error) {
	// 獲取群組所有成員
	var members []models.GroupMember
	if err := s.db.Where("group_id = ?", groupID).
		Find(&members).Error; err != nil {
		return nil, err
	}

	// 獲取用戶資訊
	var userIDs []uint
	for _, member := range members {
		userIDs = append(userIDs, member.UserID)
	}

	var users []models.User
	if err := s.db.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, err
	}

	// 建立用戶 ID 到用戶對象的映射
	userMap := make(map[uint]models.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// 初始化每個用戶的平衡記錄
	balanceMap := make(map[uint]*models.Balance)
	for _, member := range members {
		if user, exists := userMap[member.UserID]; exists {
			balanceMap[member.UserID] = &models.Balance{
				UserID:  member.UserID,
				User:    user,
				Balance: 0,
				Paid:    0,
				Owed:    0,
			}
		}
	}

	// 獲取群組所有交易
	var transactions []models.Transaction
	if err := s.db.Where("group_id = ?", groupID).
		Preload("Splits").
		Find(&transactions).Error; err != nil {
		return nil, err
	}

	// 計算每筆交易對用戶平衡的影響
	for _, transaction := range transactions {
		// 付款者增加應收金額
		if balance, exists := balanceMap[transaction.PaidBy]; exists {
			balance.Paid += transaction.Amount
			balance.Balance += transaction.Amount
		}

		// 分帳參與者增加應付金額
		for _, split := range transaction.Splits {
			if balance, exists := balanceMap[split.UserID]; exists {
				balance.Owed += split.Amount
				balance.Balance -= split.Amount
			}
		}
	}

	// 考慮已經完成的結算記錄
	var settlements []models.Settlement
	if err := s.db.Where("group_id = ? AND status = 'paid'", groupID).
		Find(&settlements).Error; err != nil {
		return nil, err
	}

	for _, settlement := range settlements {
		// 付款者減少應付
		if balance, exists := balanceMap[settlement.FromUserID]; exists {
			balance.Balance += settlement.Amount
		}
		// 收款者減少應收
		if balance, exists := balanceMap[settlement.ToUserID]; exists {
			balance.Balance -= settlement.Amount
		}
	}

	// 轉換為切片
	var balances []models.Balance
	for _, balance := range balanceMap {
		balances = append(balances, *balance)
	}

	return balances, nil
}
