package responses

import (
	"split-go/internal/models"
	"time"
)

// GroupResponse 群組回應結構
type GroupResponse struct {
	ID          uint               `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Creator     UserSimpleResponse `json:"creator"`
	CreatedAt   time.Time          `json:"created_at"`
}

// NewGroupResponse 創建群組回應
func NewGroupResponse(group models.Group) GroupResponse {
	return GroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Creator:     NewUserSimpleResponse(group.Creator),
		CreatedAt:   group.CreatedAt,
	}
}

// GroupSimpleResponse 簡化的群組回應（用於嵌套）
type GroupSimpleResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// NewGroupSimpleResponse 創建簡化的群組回應
func NewGroupSimpleResponse(group models.Group) GroupSimpleResponse {
	return GroupSimpleResponse{
		ID:   group.ID,
		Name: group.Name,
	}
}

// GroupDetailResponse 群組詳細回應（包含成員）
type GroupDetailResponse struct {
	GroupResponse
	Members   []UserSimpleResponse `json:"members"`
	MyRole    string               `json:"my_role"`    // 當前用戶在群組中的角色
	CanEdit   bool                 `json:"can_edit"`   // 是否可以編輯群組
	CanDelete bool                 `json:"can_delete"` // 是否可以刪除群組
}

// NewGroupDetailResponse 創建群組詳細回應
func NewGroupDetailResponse(group models.Group, members []models.User, currentUserID uint, myRole string) GroupDetailResponse {
	memberResponses := make([]UserSimpleResponse, len(members))
	for i, member := range members {
		memberResponses[i] = NewUserSimpleResponse(member)
	}

	canEdit := myRole == "admin" || group.CreatedBy == currentUserID
	canDelete := group.CreatedBy == currentUserID

	return GroupDetailResponse{
		GroupResponse: NewGroupResponse(group),
		Members:       memberResponses,
		MyRole:        myRole,
		CanEdit:       canEdit,
		CanDelete:     canDelete,
	}
}
