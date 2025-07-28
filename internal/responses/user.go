package responses

import (
	"split-go/internal/models"
)

// UserResponse 用戶回應結構
type UserResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Avatar   string `json:"avatar,omitempty"`
}

// NewUserResponse 創建用戶回應
func NewUserResponse(user models.User) UserResponse {
	return UserResponse{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Username: user.Username,
		Avatar:   user.Avatar,
	}
}

// NewUserResponseList 批量轉換用戶列表
func NewUserResponseList(users []models.User) []UserResponse {
	responses := make([]UserResponse, len(users))
	for i, user := range users {
		responses[i] = NewUserResponse(user)
	}
	return responses
}

// UserSimpleResponse 簡化的用戶回應（用於嵌套）
type UserSimpleResponse struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar,omitempty"`
}

// NewUserSimpleResponse 創建簡化的用戶回應
func NewUserSimpleResponse(user models.User) UserSimpleResponse {
	return UserSimpleResponse{
		ID:     user.ID,
		Name:   user.Name,
		Avatar: user.Avatar,
	}
}
