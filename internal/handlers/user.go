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
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error":   true,
		"message": "功能尚未實現",
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
