package responses

import (
	"split-go/internal/models"
)

// CategoryResponse 分類回應結構
type CategoryResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Icon  string `json:"icon,omitempty"`
	Color string `json:"color,omitempty"`
}

// NewCategoryResponse 創建分類回應
func NewCategoryResponse(category models.Category) CategoryResponse {
	return CategoryResponse{
		ID:    category.ID,
		Name:  category.Name,
		Icon:  category.Icon,
		Color: category.Color,
	}
}

// NewCategoryResponseList 批量轉換分類列表
func NewCategoryResponseList(categories []models.Category) []CategoryResponse {
	responses := make([]CategoryResponse, len(categories))
	for i, category := range categories {
		responses[i] = NewCategoryResponse(category)
	}
	return responses
}
