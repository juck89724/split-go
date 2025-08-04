package handlers

import (
	"split-go/internal/models"
	"split-go/internal/responses"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type CategoryHandler struct {
	db *gorm.DB
}

func NewCategoryHandler(db *gorm.DB) *CategoryHandler {
	return &CategoryHandler{db: db}
}

func (h *CategoryHandler) GetCategories(c *fiber.Ctx) error {
	var categories []models.Category
	if err := h.db.Order("name ASC").Find(&categories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("查詢分類失敗"),
		)
	}

	// 轉換為回應格式
	categoryResponses := responses.NewCategoryResponseList(categories)

	return c.JSON(responses.SuccessResponse(categoryResponses))
}
