package handlers

import (
	"time"

	"split-go/internal/middleware"
	"split-go/internal/models"
	"split-go/internal/responses"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type GroupHandler struct {
	db *gorm.DB
}

func NewGroupHandler(db *gorm.DB) *GroupHandler {
	return &GroupHandler{db: db}
}

// GetUserGroups 獲取用戶加入的所有群組
func (h *GroupHandler) GetUserGroups(c *fiber.Ctx) error {
	// 驗證用戶身份
	authUser, err := middleware.GetCurrentUser(c, h.db)
	if err != nil {
		return err
	}

	var groups []models.Group
	// 查詢用戶加入的所有群組，包含創建者資訊
	err = h.db.Joins("Creator").
		Joins("JOIN group_members ON groups.id = group_members.group_id").
		Where("group_members.user_id = ?", authUser.UserID).
		Order("groups.updated_at DESC").
		Find(&groups).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("獲取群組列表失敗"),
		)
	}

	// 轉換為回應格式
	groupResponses := make([]responses.GroupResponse, len(groups))
	for i, group := range groups {
		groupResponses[i] = responses.NewGroupResponse(group)
	}

	return c.JSON(responses.SuccessWithMessageResponse("成功獲取群組列表", groupResponses))
}

// CreateGroupRequest 創建群組請求結構
type CreateGroupRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=50"`
	Description string `json:"description" validate:"max=200"`
}

// CreateGroup 創建新群組
func (h *GroupHandler) CreateGroup(c *fiber.Ctx) error {
	// 驗證用戶身份
	authUser, err := middleware.GetCurrentUser(c, h.db)
	if err != nil {
		return err
	}

	// 解析請求資料
	var req CreateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse("請求格式錯誤"),
		)
	}

	// 驗證輸入
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse("群組名稱不能為空"),
		)
	}

	// 開始資料庫交易
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 創建群組
	group := models.Group{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   authUser.UserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := tx.Create(&group).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("創建群組失敗"),
		)
	}

	// 將創建者加為管理員成員
	member := models.GroupMember{
		GroupID:  group.ID,
		UserID:   authUser.UserID,
		Role:     "admin",
		JoinedAt: time.Now(),
	}

	if err := tx.Create(&member).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("添加創建者為管理員失敗"),
		)
	}

	// 提交交易
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("創建群組失敗"),
		)
	}

	// 重新查詢群組資料，包含創建者資訊
	if err := h.db.Preload("Creator").First(&group, group.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("獲取群組資訊失敗"),
		)
	}

	return c.Status(fiber.StatusCreated).JSON(
		responses.SuccessWithMessageResponse("群組創建成功", responses.NewGroupResponse(group)),
	)
}

// GetGroup 獲取群組詳細資訊
func (h *GroupHandler) GetGroup(c *fiber.Ctx) error {
	// 解析群組ID
	groupID, err := middleware.ParseGroupIDFromParams(c)
	if err != nil {
		return err
	}

	// 驗證用戶是否為群組成員
	authUser, err := middleware.RequireGroupMember(c, h.db, groupID)
	if err != nil {
		return err
	}

	// 查詢群組資訊
	var group models.Group
	if err := h.db.Preload("Creator").First(&group, groupID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				responses.ErrorResponse("群組不存在"),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("獲取群組資訊失敗"),
		)
	}

	// 查詢群組成員
	var members []models.User
	err = h.db.Joins("JOIN group_members ON users.id = group_members.user_id").
		Where("group_members.group_id = ?", groupID).
		Order("group_members.role DESC, users.username").
		Find(&members).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("獲取群組成員失敗"),
		)
	}

	// 獲取當前用戶在群組中的角色
	var currentMember models.GroupMember
	if err := h.db.Where("group_id = ? AND user_id = ?", groupID, authUser.UserID).
		First(&currentMember).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("獲取用戶權限失敗"),
		)
	}

	// 創建詳細回應
	groupDetail := responses.NewGroupDetailResponse(group, members, authUser.UserID, currentMember.Role)

	return c.JSON(responses.SuccessWithMessageResponse("成功獲取群組資訊", groupDetail))
}

// UpdateGroupRequest 更新群組請求結構
type UpdateGroupRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=50"`
	Description string `json:"description" validate:"max=200"`
}

// UpdateGroup 更新群組資訊（需要管理員權限）
func (h *GroupHandler) UpdateGroup(c *fiber.Ctx) error {
	// 解析群組ID
	groupID, err := middleware.ParseGroupIDFromParams(c)
	if err != nil {
		return err
	}

	// 驗證用戶是否為群組管理員
	_, err = middleware.RequireGroupAdmin(c, h.db, groupID)
	if err != nil {
		return err
	}

	// 解析請求資料
	var req UpdateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse("請求格式錯誤"),
		)
	}

	// 驗證輸入
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse("群組名稱不能為空"),
		)
	}

	// 查詢群組
	var group models.Group
	if err := h.db.First(&group, groupID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				responses.ErrorResponse("群組不存在"),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("獲取群組資訊失敗"),
		)
	}

	// 更新群組資訊
	group.Name = req.Name
	group.Description = req.Description
	group.UpdatedAt = time.Now()

	if err := h.db.Save(&group).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("更新群組失敗"),
		)
	}

	// 重新查詢包含創建者資訊
	if err := h.db.Preload("Creator").First(&group, group.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("獲取更新後的群組資訊失敗"),
		)
	}

	return c.JSON(responses.SuccessWithMessageResponse("群組更新成功", responses.NewGroupResponse(group)))
}

// DeleteGroup 刪除群組（只有創建者可以刪除）
func (h *GroupHandler) DeleteGroup(c *fiber.Ctx) error {
	// 解析群組ID
	groupID, err := middleware.ParseGroupIDFromParams(c)
	if err != nil {
		return err
	}

	// 驗證用戶身份
	authUser, err := middleware.GetCurrentUser(c, h.db)
	if err != nil {
		return err
	}

	// 查詢群組
	var group models.Group
	if err := h.db.First(&group, groupID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				responses.ErrorResponse("群組不存在"),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("獲取群組資訊失敗"),
		)
	}

	// 只有創建者可以刪除群組
	if group.CreatedBy != authUser.UserID {
		return c.Status(fiber.StatusForbidden).JSON(
			responses.ErrorResponse("只有群組創建者可以刪除群組"),
		)
	}

	// 開始資料庫交易
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 刪除所有群組成員關聯
	if err := tx.Where("group_id = ?", groupID).Delete(&models.GroupMember{}).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("刪除群組成員關聯失敗"),
		)
	}

	// 軟刪除群組
	if err := tx.Delete(&group).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("刪除群組失敗"),
		)
	}

	// 提交交易
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("刪除群組失敗"),
		)
	}

	return c.JSON(responses.SuccessWithMessageResponse("群組刪除成功", nil))
}

// AddMemberRequest 添加成員請求結構
type AddMemberRequest struct {
	UserID uint   `json:"user_id" validate:"required"`
	Role   string `json:"role" validate:"oneof=member admin"`
}

// AddMember 添加群組成員
func (h *GroupHandler) AddMember(c *fiber.Ctx) error {
	// 解析群組ID
	groupID, err := middleware.ParseGroupIDFromParams(c)
	if err != nil {
		return err
	}

	// 驗證用戶是否為群組管理員
	_, err = middleware.RequireGroupAdmin(c, h.db, groupID)
	if err != nil {
		return err
	}

	// 解析請求資料
	var req AddMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse("請求格式錯誤"),
		)
	}

	// 設置預設角色
	if req.Role == "" {
		req.Role = "member"
	}

	// 驗證用戶ID
	if req.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse("用戶ID不能為空"),
		)
	}

	// 檢查要添加的用戶是否存在
	var targetUser models.User
	if err := h.db.First(&targetUser, req.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				responses.ErrorResponse("用戶不存在"),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("查詢用戶失敗"),
		)
	}

	// 檢查用戶是否已經是群組成員
	var existingMember models.GroupMember
	err = h.db.Where("group_id = ? AND user_id = ?", groupID, req.UserID).First(&existingMember).Error
	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(
			responses.ErrorResponse("用戶已經是群組成員"),
		)
	} else if err != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("檢查成員狀態失敗"),
		)
	}

	// 添加新成員
	member := models.GroupMember{
		GroupID:  groupID,
		UserID:   req.UserID,
		Role:     req.Role,
		JoinedAt: time.Now(),
	}

	if err := h.db.Create(&member).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("添加成員失敗"),
		)
	}

	return c.Status(fiber.StatusCreated).JSON(responses.SuccessWithMessageResponse("成員添加成功", fiber.Map{
		"user_id": req.UserID,
		"role":    req.Role,
	}))
}

// RemoveMember 移除群組成員（需要管理員權限）
func (h *GroupHandler) RemoveMember(c *fiber.Ctx) error {
	// 解析群組ID
	groupID, err := middleware.ParseGroupIDFromParams(c)
	if err != nil {
		return err
	}

	// 解析要移除的用戶 ID
	userID, err := middleware.ParseUserIDFromParams(c)
	if err != nil {
		return err
	}

	// 驗證當前用戶是否為群組管理員
	authUser, err := middleware.RequireGroupAdmin(c, h.db, groupID)
	if err != nil {
		return err
	}

	// 檢查要移除的成員是否存在
	var member models.GroupMember
	err = h.db.Where("group_id = ? AND user_id = ?", groupID, uint(userID)).First(&member).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				responses.ErrorResponse("用戶不是群組成員"),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("查詢成員失敗"),
		)
	}

	// 不能移除自己
	if uint(userID) == authUser.UserID {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse("不能移除自己"),
		)
	}

	// 查詢群組資訊以檢查是否要移除創建者
	var group models.Group
	if err := h.db.First(&group, groupID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("獲取群組資訊失敗"),
		)
	}

	// 不能移除群組創建者
	if uint(userID) == group.CreatedBy {
		return c.Status(fiber.StatusBadRequest).JSON(
			responses.ErrorResponse("不能移除群組創建者"),
		)
	}

	// 移除成員
	if err := h.db.Delete(&member).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			responses.ErrorResponse("移除成員失敗"),
		)
	}

	return c.JSON(responses.SuccessWithMessageResponse("成員移除成功", nil))
}
