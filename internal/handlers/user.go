package handlers

import (
	"split-go/internal/middleware"
	"split-go/internal/models"

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
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "未找到用戶資訊",
		})
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "用戶不存在",
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  user,
	})
}

// UpdateProfile 更新用戶資料
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	// 待實現
	userID := middleware.GetUserIDFromContext(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "未找到用戶資訊",
		})
	}

	// 檢查用戶是否存在
	var existingUser models.User
	if err := h.db.First(&existingUser, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "用戶不存在",
		})
	}

	// 直接解析到更新 DTO
	var updateData models.UserUpdateRequest
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "無效的請求數據",
		})
	}

	if err := h.db.Model(&models.User{}).Where("id = ?", userID).
		Updates(updateData).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "更新用戶資料失敗",
		})
	}

	// 獲取更新後的用戶資料
	var updatedUser models.User
	if err := h.db.First(&updatedUser, userID).Error; err != nil {
		// 如果查詢失敗，回傳更新成功但沒有最新資料
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"error":   false,
			"message": "更新用戶資料成功",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":   false,
		"message": "更新用戶資料成功",
		"data":    updatedUser,
	})
}

// UpdateFCMToken 更新 FCM Token
func (h *UserHandler) UpdateFCMToken(c *fiber.Ctx) error {
	// 待實現
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error":   true,
		"message": "功能尚未實現",
	})
}
