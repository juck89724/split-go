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
// @Summary 獲取個人資料
// @Description 獲取當前登入用戶的個人資料
// @Tags 用戶
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{error=bool,data=object{id=int,name=string,email=string,username=string,avatar=string}} "用戶資料"
// @Failure 401 {object} object{error=bool,message=string} "未授權"
// @Failure 404 {object} object{error=bool,message=string} "用戶不存在"
// @Router /users/me [get]
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
// @Summary 更新個人資料
// @Description 更新當前登入用戶的個人資料
// @Tags 用戶
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{name=string,avatar=string} true "更新資料"
// @Success 200 {object} object{error=bool,message=string,data=object{id=int,name=string,email=string,username=string,avatar=string}} "更新成功"
// @Failure 400 {object} object{error=bool,message=string} "請求格式錯誤"
// @Failure 401 {object} object{error=bool,message=string} "未授權"
// @Failure 500 {object} object{error=bool,message=string} "服務器內部錯誤"
// @Router /users/me [put]
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
	FCMToken string `json:"fcm_token"`
}

// UpdateFCMToken 更新 FCM Token
// @Summary 更新 FCM 推播令牌
// @Description 更新用戶的 Firebase Cloud Messaging 推播令牌
// @Tags 用戶
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{fcm_token=string} true "FCM Token 資料"
// @Success 200 {object} object{error=bool,message=string} "FCM Token 更新成功"
// @Failure 400 {object} object{error=bool,message=string} "請求格式錯誤或 FCM Token 為空"
// @Failure 401 {object} object{error=bool,message=string} "未授權"
// @Failure 500 {object} object{error=bool,message=string} "更新失敗"
// @Router /users/fcm-token [post]
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

	// 驗證 FCM Token 不為空
	if req.FCMToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse("FCM Token 不能為空"),
		)
	}

	// 更新 FCM Token
	if err := h.db.Model(&models.User{}).Where("id = ?", user.UserID).Update("fcm_token", req.FCMToken).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("更新 FCM Token 失敗"),
		)
	}

	return c.JSON(
		responses.SuccessWithMessageResponse("FCM Token 更新成功", nil),
	)
}
