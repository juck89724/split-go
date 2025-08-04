package handlers

import (
	"math"
	"time"

	"split-go/internal/middleware"
	"split-go/internal/models"
	"split-go/internal/responses"
	"split-go/internal/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type SettlementHandler struct {
	db                *gorm.DB
	balanceService    *services.BalanceService
	validationService *services.ValidationService
}

func NewSettlementHandler(db *gorm.DB) *SettlementHandler {
	return &SettlementHandler{
		db:                db,
		balanceService:    services.NewBalanceService(db),
		validationService: services.NewValidationService(db),
	}
}

func (h *SettlementHandler) GetSettlements(c *fiber.Ctx) error {
	// 驗證用戶身份
	user, err := middleware.GetCurrentUser(c, h.db)
	if err != nil {
		return err
	}

	// 查詢用戶相關的結算記錄（作為付款者或收款者）
	var settlements []models.Settlement
	if err := h.db.Where("from_user_id = ? OR to_user_id = ?", user.UserID, user.UserID).
		Preload("Group").
		Preload("FromUser").
		Preload("ToUser").
		Order("created_at DESC").
		Find(&settlements).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("查詢結算記錄失敗"),
		)
	}

	// 轉換為回應格式
	settlementResponses := make([]responses.SettlementResponse, len(settlements))
	for i, settlement := range settlements {
		settlementResponses[i] = responses.NewSettlementResponse(settlement)
	}

	return c.JSON(responses.SuccessResponse(settlementResponses))
}

type CreateSettlementRequest struct {
	GroupID  uint    `json:"group_id" validate:"required"`
	ToUserID uint    `json:"to_user_id" validate:"required"`
	Amount   float64 `json:"amount" validate:"required,gt=0"`
	Currency string  `json:"currency"`
	Notes    string  `json:"notes"`
}

func (h *SettlementHandler) CreateSettlement(c *fiber.Ctx) error {
	// 驗證用戶身份
	user, err := middleware.GetCurrentUser(c, h.db)
	if err != nil {
		return err
	}

	// 解析請求資料
	var req CreateSettlementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse("無效的請求格式"),
		)
	}

	// 驗證用戶是群組成員
	_, err = middleware.RequireGroupMember(c, h.db, req.GroupID)
	if err != nil {
		return err
	}

	// 驗證收款者是群組成員
	if err := h.validationService.ValidateGroupMember(req.GroupID, req.ToUserID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse(err.Error()),
		)
	}

	// 檢查是否試圖向自己結算
	if user.UserID == req.ToUserID {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse("不能向自己結算"),
		)
	}

	// 設定預設幣別
	if req.Currency == "" {
		req.Currency = "TWD"
	}

	// 創建結算記錄
	settlement := models.Settlement{
		GroupID:    req.GroupID,
		FromUserID: user.UserID,
		ToUserID:   req.ToUserID,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Status:     "pending",
		Notes:      req.Notes,
	}

	if err := h.db.Create(&settlement).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("創建結算記錄失敗"),
		)
	}

	// 載入關聯資料並返回
	if err := h.db.Preload("Group").Preload("FromUser").Preload("ToUser").
		First(&settlement, settlement.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("載入結算記錄失敗"),
		)
	}

	return c.Status(fiber.StatusCreated).JSON(
		responses.SuccessWithMessageResponse("結算記錄創建成功", responses.NewSettlementResponse(settlement)),
	)
}

func (h *SettlementHandler) MarkAsPaid(c *fiber.Ctx) error {
	// 驗證用戶身份
	user, err := middleware.GetCurrentUser(c, h.db)
	if err != nil {
		return err
	}

	// 獲取結算 ID
	settlementID, err := middleware.ParseSettlementIDFromParams(c)
	if err != nil {
		return err
	}

	// 查詢結算記錄
	var settlement models.Settlement
	if err := h.db.Preload("Group").Preload("FromUser").Preload("ToUser").
		First(&settlement, uint(settlementID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				responses.ErrorResponse("結算記錄不存在"),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("查詢結算記錄失敗"),
		)
	}

	// 檢查權限（只有收款者可以標記為已付款）
	if settlement.ToUserID != user.UserID {
		return c.Status(fiber.StatusForbidden).JSON(
			responses.ErrorResponse("只有收款者可以標記為已付款"),
		)
	}

	// 檢查狀態
	if settlement.Status != "pending" {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse("只能標記待付款的結算記錄"),
		)
	}

	// 更新狀態
	now := time.Now()
	if err := h.db.Model(&settlement).Updates(map[string]interface{}{
		"status":     "paid",
		"settled_at": &now,
	}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("更新結算狀態失敗"),
		)
	}

	// 更新本地對象
	settlement.Status = "paid"
	settlement.SettledAt = &now

	return c.JSON(
		responses.SuccessWithMessageResponse("結算已標記為已付款", responses.NewSettlementResponse(settlement)),
	)
}

func (h *SettlementHandler) CancelSettlement(c *fiber.Ctx) error {
	// 驗證用戶身份
	user, err := middleware.GetCurrentUser(c, h.db)
	if err != nil {
		return err
	}

	// 獲取結算 ID
	settlementID, err := middleware.ParseSettlementIDFromParams(c)
	if err != nil {
		return err
	}

	// 查詢結算記錄
	var settlement models.Settlement
	if err := h.db.First(&settlement, uint(settlementID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				responses.ErrorResponse("結算記錄不存在"),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("查詢結算記錄失敗"),
		)
	}

	// 檢查權限（付款者或收款者都可以取消）
	if settlement.FromUserID != user.UserID && settlement.ToUserID != user.UserID {
		return c.Status(fiber.StatusForbidden).JSON(
			responses.ErrorResponse("無權限取消此結算記錄"),
		)
	}

	// 檢查狀態（只能取消待付款的記錄）
	if settlement.Status != "pending" {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse("只能取消待付款的結算記錄"),
		)
	}

	// 軟刪除結算記錄
	if err := h.db.Delete(&settlement).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("取消結算記錄失敗"),
		)
	}

	return c.JSON(
		responses.SuccessWithMessageResponse("結算記錄已取消", nil),
	)
}

func (h *SettlementHandler) GetSettlementSuggestions(c *fiber.Ctx) error {
	// 獲取群組 ID
	groupID, err := middleware.ParseGroupIDFromParams(c)
	if err != nil {
		return err
	}

	// 驗證用戶是群組成員
	_, err = middleware.RequireGroupMember(c, h.db, groupID)
	if err != nil {
		return err
	}

	// 計算群組內每個用戶的平衡
	balances, err := h.balanceService.CalculateGroupBalances(groupID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("計算群組平衡失敗"),
		)
	}

	// 生成結算建議
	suggestions := h.generateSettlementSuggestions(balances)

	// 轉換為回應格式
	suggestionResponses := make([]responses.SettlementSuggestionResponse, len(suggestions))
	for i, suggestion := range suggestions {
		suggestionResponses[i] = responses.NewSettlementSuggestionResponse(suggestion)
	}

	return c.JSON(responses.SuccessResponse(suggestionResponses))
}

// generateSettlementSuggestions 生成結算建議（使用貪心算法最小化轉帳次數）
func (h *SettlementHandler) generateSettlementSuggestions(balances []models.Balance) []models.SettlementSuggestion {
	var suggestions []models.SettlementSuggestion
	const tolerance = 0.01 // 容差值，處理浮點數精度問題

	// 分離債權人和債務人
	var creditors []models.Balance // 應收錢的人（balance > 0）
	var debtors []models.Balance   // 應付錢的人（balance < 0）

	for _, balance := range balances {
		if balance.Balance > tolerance {
			creditors = append(creditors, balance)
		} else if balance.Balance < -tolerance {
			debtors = append(debtors, balance)
		}
	}

	// 使用貪心算法生成結算建議
	i, j := 0, 0
	for i < len(creditors) && j < len(debtors) {
		creditor := &creditors[i]
		debtor := &debtors[j]

		// 計算轉帳金額
		amount := math.Min(creditor.Balance, -debtor.Balance)

		// 創建結算建議
		suggestions = append(suggestions, models.SettlementSuggestion{
			FromUserID: debtor.UserID,
			FromUser:   debtor.User,
			ToUserID:   creditor.UserID,
			ToUser:     creditor.User,
			Amount:     amount,
		})

		// 更新餘額
		creditor.Balance -= amount
		debtor.Balance += amount

		// 移動指針
		if math.Abs(creditor.Balance) <= tolerance {
			i++
		}
		if math.Abs(debtor.Balance) <= tolerance {
			j++
		}
	}

	return suggestions
}
