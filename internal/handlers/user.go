package handlers

import (
	"split-go/internal/middleware"
	"split-go/internal/models"
	"split-go/internal/responses"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// GetProfile 獲取用戶資料
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID := middleware.GetUserIDFromContext(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(
			responses.ErrorResponse("未找到用戶資訊"),
		)
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			responses.ErrorResponse("用戶不存在"),
		)
	}

	// 使用 response 轉換
	userResponse := responses.NewUserResponse(user)
	return c.JSON(responses.SuccessResponse(userResponse))
}

// UpdateProfile 更新用戶資料
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := middleware.GetUserIDFromContext(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(
			responses.ErrorResponse("未找到用戶資訊"),
		)
	}

	// 檢查用戶是否存在
	var existingUser models.User
	if err := h.db.First(&existingUser, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			responses.ErrorResponse("用戶不存在"),
		)
	}

	// 直接解析到更新 DTO
	var updateData models.UserUpdateRequest
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse("無效的請求數據"),
		)
	}

	if err := h.db.Model(&models.User{}).Where("id = ?", userID).
		Updates(updateData).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("更新用戶資料失敗"),
		)
	}

	// 獲取更新後的用戶資料
	var updatedUser models.User
	if err := h.db.First(&updatedUser, userID).Error; err != nil {
		// 如果查詢失敗，回傳更新成功但沒有最新資料
		return c.Status(fiber.StatusOK).JSON(
			responses.SuccessWithMessageResponse("更新用戶資料成功", nil),
		)
	}

	// 轉換為回應格式
	userResponse := responses.NewUserResponse(updatedUser)
	return c.Status(fiber.StatusOK).JSON(
		responses.SuccessWithMessageResponse("更新用戶資料成功", userResponse),
	)
}

type UpdateFCMTokenRequest struct {
	FCMToken string `json:"fcm_token" validate:"required"`
}

// UpdateFCMToken 更新 FCM Token
func (h *UserHandler) UpdateFCMToken(c *fiber.Ctx) error {
	// 驗證用戶身份
	user, err := middleware.GetCurrentUser(c, h.db)
	if err != nil {
		return err
	}

	// 解析請求資料
	var req UpdateFCMTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse("無效的請求格式"),
		)
	}

	// 更新 FCM Token
	if err := h.db.Model(&user).Update("fcm_token", req.FCMToken).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("更新 FCM Token 失敗"),
		)
	}

	return c.JSON(
		responses.SuccessWithMessageResponse("FCM Token 更新成功", nil),
	)
}
