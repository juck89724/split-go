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

// GetCategories 獲取所有分類
// @Summary 獲取分類列表
// @Description 獲取系統中所有可用的交易分類
// @Tags 分類
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{error=bool,data=[]object{id=int,name=string,description=string,created_at=string,updated_at=string}} "分類列表"
// @Failure 401 {object} object{error=bool,message=string} "未授權"
// @Failure 500 {object} object{error=bool,message=string} "服務器內部錯誤"
// @Router /categories [get]
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
