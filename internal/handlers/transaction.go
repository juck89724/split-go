package handlers

import (
	"errors"
	"math"
	"split-go/internal/middleware"
	"split-go/internal/models"
	"split-go/internal/responses"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type TransactionHandler struct {
	db *gorm.DB
}

func NewTransactionHandler(db *gorm.DB) *TransactionHandler {
	return &TransactionHandler{db: db}
}

func (h *TransactionHandler) GetTransactions(c *fiber.Ctx) error {
	user, err := middleware.GetCurrentUser(c, h.db)
	if err != nil {
		return err
	}

	// 查詢用戶參與的所有交易（作為付款者或分帳參與者）
	var transactions []models.Transaction

	// 使用子查詢找出用戶參與的交易 ID
	subQuery := h.db.Model(&models.TransactionSplit{}).
		Select("transaction_id").
		Where("user_id = ?", user.UserID)

	// 查詢條件：用戶是付款者 OR 用戶參與分帳
	if err := h.db.Where("paid_by = ? OR id IN (?)", user.UserID, subQuery).
		Preload("Payer").
		Preload("Creator").
		Preload("Category").
		Preload("Group").
		Preload("Splits.User").
		Order("created_at DESC").
		Find(&transactions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("查詢交易記錄失敗"),
		)
	}

	// 使用簡化的回應格式（列表頁面不需要太詳細的資訊）
	transactionResponses := responses.NewTransactionSimpleResponseList(transactions, user.UserID)

	return c.JSON(responses.SuccessResponse(transactionResponses))
}

func (h *TransactionHandler) CreateTransaction(c *fiber.Ctx) error {
	// 1. 取得當前用戶
	user, err := middleware.GetCurrentUser(c, h.db)
	if err != nil {
		return err
	}

	// 2. 解析請求資料
	var req models.CreateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse("無效的請求格式"),
		)
	}

	// 3. 驗證用戶是群組成員
	_, err = middleware.RequireGroupMember(c, h.db, req.GroupID)
	if err != nil {
		return err
	}

	// 4. 驗證付款者是群組成員
	if err := h.validateGroupMember(h.db, req.GroupID, req.PaidBy); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse(err.Error()),
		)
	}

	// 5. 驗證所有分帳用戶都是群組成員
	var splitUserIDs []uint
	for _, split := range req.Splits {
		splitUserIDs = append(splitUserIDs, split.UserID)
	}

	if err := h.validateSplitUsers(h.db, req.GroupID, splitUserIDs); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse(err.Error()),
		)
	}

	// 6. 計算分帳金額
	// 轉換為通用結構
	calculations := make([]SplitCalculation, len(req.Splits))
	for i, split := range req.Splits {
		calculations[i] = SplitCalculation{
			UserID:     split.UserID,
			Amount:     split.Amount,
			Percentage: split.Percentage,
		}
	}

	calculated, err := h.calculateSplits(req.SplitType, req.Amount, calculations)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse(err.Error()),
		)
	}

	// 轉換回原始結構
	calculatedSplits := req.Splits
	for i, calc := range calculated {
		calculatedSplits[i].Amount = calc.Amount
		calculatedSplits[i].Percentage = calc.Percentage
	}

	// 7. 設定預設值
	if req.Currency == "" {
		req.Currency = "TWD"
	}

	// 8. 使用資料庫交易確保資料一致性
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 9. 創建交易記錄
	transaction := models.Transaction{
		GroupID:     req.GroupID,
		Description: req.Description,
		Amount:      req.Amount,
		Currency:    req.Currency,
		CategoryID:  req.CategoryID,
		PaidBy:      req.PaidBy,
		Receipt:     req.Receipt,
		Notes:       req.Notes,
		CreatedBy:   user.UserID,
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("創建交易失敗"),
		)
	}

	// 10. 創建分帳記錄
	if err := h.createSplitRecords(tx, transaction.ID, req.SplitType, calculatedSplits); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("創建分帳記錄失敗"),
		)
	}

	// 11. 提交交易
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("保存交易失敗"),
		)
	}

	// 12. 載入完整的交易資料回傳
	if err := h.db.Preload("Group").Preload("Payer").Preload("Creator").
		Preload("Category").Preload("Splits.User").
		First(&transaction, transaction.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("載入交易資料失敗"),
		)
	}

	// 轉換為回應格式
	transactionResponse := responses.NewTransactionResponse(transaction, user.UserID)

	return c.Status(fiber.StatusCreated).JSON(
		responses.SuccessWithMessageResponse("交易創建成功", transactionResponse),
	)
}

func (h *TransactionHandler) UpdateTransaction(c *fiber.Ctx) error {
	// 1. 解析交易 ID
	transactionID, err := middleware.ParseTransactionIDFromParams(c)
	if err != nil {
		return err
	}

	// 2. 取得當前用戶
	user, err := middleware.GetCurrentUser(c, h.db)
	if err != nil {
		return err
	}

	// 3. 查詢現有交易
	var existingTransaction models.Transaction
	if err := h.db.Preload("Group").Preload("Payer").Preload("Creator").
		Preload("Category").Preload("Splits.User").
		First(&existingTransaction, transactionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				responses.ErrorResponse("交易不存在"),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("查詢交易失敗"),
		)
	}

	// 4. 檢查用戶權限（只有創建者或付款者可以編輯）
	if existingTransaction.CreatedBy != user.UserID && existingTransaction.PaidBy != user.UserID {
		return c.Status(fiber.StatusForbidden).JSON(
			responses.ErrorResponse("您沒有權限編輯此交易"),
		)
	}

	// 5. 解析更新請求
	var req models.UpdateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse("無效的請求格式"),
		)
	}

	// 6. 開始資料庫交易
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 7. 更新基本資訊
	if err := h.updateBasicFields(tx, &existingTransaction, req); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse(err.Error()),
		)
	}

	// 8. 處理分帳更新
	shouldUpdateSplits := len(req.Splits) > 0 || req.SplitType != "" || req.Amount > 0
	if shouldUpdateSplits {
		if err := h.updateSplits(tx, transactionID, &existingTransaction, req); err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusBadRequest).JSON(
				responses.ErrorResponse(err.Error()),
			)
		}
	}

	// 9. 提交交易
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("保存更新失敗"),
		)
	}

	// 10. 載入更新後的完整資料
	var updatedTransaction models.Transaction
	if err := h.db.Preload("Group").Preload("Payer").Preload("Creator").
		Preload("Category").Preload("Splits.User").
		First(&updatedTransaction, transactionID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("載入更新後的交易資料失敗"),
		)
	}

	// 11. 轉換為回應格式並回傳
	transactionResponse := responses.NewTransactionResponse(updatedTransaction, user.UserID)
	return c.JSON(
		responses.SuccessWithMessageResponse("交易更新成功", transactionResponse),
	)
}

// GetTransaction 使用新的共用函數來示範用法
func (h *TransactionHandler) GetTransaction(c *fiber.Ctx) error {
	// 1. 解析交易 ID
	transactionID, err := middleware.ParseTransactionIDFromParams(c)
	if err != nil {
		return err
	}

	// 2. 取得當前認證用戶
	user, err := middleware.GetCurrentUser(c, h.db)
	if err != nil {
		return err
	}

	// 3. 查詢交易並檢查權限
	var transaction models.Transaction
	if err := h.db.Preload("Group").Preload("Payer").Preload("Creator").
		Preload("Category").Preload("Splits.User").
		First(&transaction, transactionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				responses.ErrorResponse("交易不存在"),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("查詢交易失敗"),
		)
	}

	// 4. 檢查用戶是否有權限查看此交易（必須是群組成員）
	_, err = middleware.RequireGroupMember(c, h.db, transaction.GroupID)
	if err != nil {
		return err
	}

	// 5. 轉換為回應格式並回傳
	transactionResponse := responses.NewTransactionResponse(transaction, user.UserID)
	return c.JSON(responses.SuccessResponse(transactionResponse))
}

func (h *TransactionHandler) DeleteTransaction(c *fiber.Ctx) error {
	return c.Status(fiber.StatusMethodNotAllowed).JSON(
		responses.ErrorResponse("刪除交易功能未提供"),
	)
}

// GetGroupTransactions 示範完整的群組交易查詢邏輯
func (h *TransactionHandler) GetGroupTransactions(c *fiber.Ctx) error {
	// 1. 解析群組 ID
	groupID, err := middleware.ParseGroupIDFromParams(c)
	if err != nil {
		return err
	}

	// 2. 驗證用戶是群組成員並取得認證資訊
	authUser, err := middleware.RequireGroupMember(c, h.db, groupID)
	if err != nil {
		return err
	}

	// 3. 解析查詢參數
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	if limit > 100 {
		limit = 100 // 限制最大頁面大小
	}
	offset := (page - 1) * limit

	// 4. 查詢群組交易
	var transactions []models.Transaction
	query := h.db.Where("group_id = ?", groupID).
		Preload("Payer").
		Preload("Creator").
		Preload("Category").
		Preload("Splits.User").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit)

	if err := query.Find(&transactions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("查詢交易記錄失敗"),
		)
	}

	// 5. 計算總筆數（用於分頁）
	var total int64
	if err := h.db.Model(&models.Transaction{}).Where("group_id = ?", groupID).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("計算總筆數失敗"),
		)
	}

	// 6. 轉換為回應格式
	transactionResponses := responses.NewTransactionSimpleResponseList(transactions, authUser.UserID)
	paginatedResponse := responses.NewPaginatedResponse(transactionResponses, page, limit, total)

	return c.JSON(responses.SuccessResponse(paginatedResponse))
}

func (h *TransactionHandler) GetGroupBalance(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(
		responses.ErrorResponse("功能尚未實現"),
	)
}

// ============ 共用函數 ============

// SplitCalculation 通用的分帳計算結構
type SplitCalculation struct {
	UserID     uint
	Amount     float64
	Percentage float64
}

// validateGroupMember 驗證用戶是否為群組成員
func (h *TransactionHandler) validateGroupMember(db *gorm.DB, groupID, userID uint) error {
	var member models.GroupMember
	if err := db.Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&member).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("付款者不是群組成員")
		}
		return errors.New("驗證群組成員失敗")
	}
	return nil
}

// validateSplitUsers 驗證分帳用戶是否都是群組成員
func (h *TransactionHandler) validateSplitUsers(db *gorm.DB, groupID uint, userIDs []uint) error {
	var memberCount int64
	if err := db.Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id IN ?", groupID, userIDs).
		Count(&memberCount).Error; err != nil {
		return errors.New("驗證分帳用戶失敗")
	}

	if int(memberCount) != len(userIDs) {
		return errors.New("部分分帳用戶不是群組成員")
	}

	return nil
}

// calculateSplits 通用的分帳計算函數
func (h *TransactionHandler) calculateSplits(splitType models.SplitType, totalAmount float64, splits []SplitCalculation) ([]SplitCalculation, error) {
	switch splitType {
	case models.SplitEqual:
		amountPerPerson := totalAmount / float64(len(splits))
		for i := range splits {
			splits[i].Amount = amountPerPerson
			splits[i].Percentage = math.Round((100.0/float64(len(splits)))*100) / 100.0
		}

	case models.SplitPercentage:
		totalPercentage := 0.0
		for _, split := range splits {
			totalPercentage += split.Percentage
		}
		if totalPercentage != 100.0 {
			return nil, errors.New("分帳比例總和必須等於 100%")
		}
		for i := range splits {
			splits[i].Amount = totalAmount * splits[i].Percentage / 100.0
		}

	case models.SplitFixed:
		totalSplitAmount := 0.0
		for _, split := range splits {
			if split.Amount <= 0 {
				return nil, errors.New("固定金額必須大於 0")
			}
			totalSplitAmount += split.Amount
		}
		if totalSplitAmount != totalAmount {
			return nil, errors.New("分帳金額總和必須等於交易金額")
		}
		for i := range splits {
			splits[i].Percentage = splits[i].Amount / totalAmount * 100.0
		}

	default:
		return nil, errors.New("無效的分帳類型")
	}

	return splits, nil
}

// createSplitRecords 創建分帳記錄
func (h *TransactionHandler) createSplitRecords(tx *gorm.DB, transactionID uint, splitType models.SplitType, splits []models.TransactionSplitRequest) error {
	var splitRecords []models.TransactionSplit
	for _, split := range splits {
		splitRecords = append(splitRecords, models.TransactionSplit{
			TransactionID: transactionID,
			UserID:        split.UserID,
			Amount:        split.Amount,
			Percentage:    split.Percentage,
			SplitType:     splitType,
		})
	}

	return tx.Create(&splitRecords).Error
}

// updateBasicFields 更新交易基本欄位
func (h *TransactionHandler) updateBasicFields(tx *gorm.DB, existingTransaction *models.Transaction, req models.UpdateTransactionRequest) error {
	updateData := make(map[string]interface{})

	if req.Description != "" {
		updateData["description"] = req.Description
	}
	if req.Amount > 0 {
		updateData["amount"] = req.Amount
	}
	if req.Currency != "" {
		updateData["currency"] = req.Currency
	}
	if req.CategoryID > 0 {
		updateData["category_id"] = req.CategoryID
	}
	if req.PaidBy > 0 {
		// 驗證新的付款者是群組成員
		if err := h.validateGroupMember(tx, existingTransaction.GroupID, req.PaidBy); err != nil {
			return err
		}
		updateData["paid_by"] = req.PaidBy
	}
	if req.Receipt != "" {
		updateData["receipt"] = req.Receipt
	}
	if req.Notes != "" {
		updateData["notes"] = req.Notes
	}

	if len(updateData) > 0 {
		if err := tx.Model(existingTransaction).Updates(updateData).Error; err != nil {
			return errors.New("更新交易基本資訊失敗")
		}
	}

	return nil
}

// updateSplits 更新分帳記錄
func (h *TransactionHandler) updateSplits(tx *gorm.DB, transactionID uint, existingTransaction *models.Transaction, req models.UpdateTransactionRequest) error {
	// 確定使用的分帳類型和金額
	splitType := req.SplitType
	if splitType == "" {
		splitType = existingTransaction.Splits[0].SplitType
	}

	amount := req.Amount
	if amount == 0 {
		amount = existingTransaction.Amount
	}

	// 準備分帳資料
	var splitsToCalculate []models.TransactionSplitRequest
	if len(req.Splits) > 0 {
		splitsToCalculate = req.Splits

		// 驗證分帳用戶
		var splitUserIDs []uint
		for _, split := range splitsToCalculate {
			splitUserIDs = append(splitUserIDs, split.UserID)
		}

		if err := h.validateSplitUsers(tx, existingTransaction.GroupID, splitUserIDs); err != nil {
			return err
		}
	} else {
		// 保持現有的分帳用戶
		for _, existingSplit := range existingTransaction.Splits {
			splitsToCalculate = append(splitsToCalculate, models.TransactionSplitRequest{
				UserID:     existingSplit.UserID,
				Amount:     existingSplit.Amount,
				Percentage: existingSplit.Percentage,
			})
		}
	}

	// 重新計算分帳
	// 轉換為通用結構
	calculations := make([]SplitCalculation, len(splitsToCalculate))
	for i, split := range splitsToCalculate {
		calculations[i] = SplitCalculation{
			UserID:     split.UserID,
			Amount:     split.Amount,
			Percentage: split.Percentage,
		}
	}

	calculated, err := h.calculateSplits(splitType, amount, calculations)
	if err != nil {
		return err
	}

	// 轉換回原始結構
	for i, calc := range calculated {
		splitsToCalculate[i].Amount = calc.Amount
		splitsToCalculate[i].Percentage = calc.Percentage
	}

	// 刪除舊的分帳記錄
	if err := tx.Where("transaction_id = ?", transactionID).Delete(&models.TransactionSplit{}).Error; err != nil {
		return errors.New("刪除舊分帳記錄失敗")
	}

	// 創建新的分帳記錄
	var newSplits []models.TransactionSplit
	for _, split := range splitsToCalculate {
		newSplits = append(newSplits, models.TransactionSplit{
			TransactionID: transactionID,
			UserID:        split.UserID,
			Amount:        split.Amount,
			Percentage:    split.Percentage,
			SplitType:     splitType,
		})
	}

	if err := tx.Create(&newSplits).Error; err != nil {
		return errors.New("創建新分帳記錄失敗")
	}

	return nil
}
