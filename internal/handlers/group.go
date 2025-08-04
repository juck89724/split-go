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
// @Summary 獲取用戶群組列表
// @Description 獲取當前用戶加入的所有群組，包含創建者資訊
// @Tags 群組
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{error=bool,message=string,data=[]object{id=int,name=string,description=string,created_by=int,creator=object{id=int,name=string,username=string},created_at=string,updated_at=string}} "群組列表"
// @Failure 401 {object} object{error=bool,message=string} "未授權"
// @Failure 500 {object} object{error=bool,message=string} "服務器內部錯誤"
// @Router /groups [get]
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
// @Summary 創建新群組
// @Description 創建一個新的群組，創建者自動成為群組管理員
// @Tags 群組
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{name=string,description=string} true "群組資料"
// @Success 201 {object} object{error=bool,message=string,data=object{id=int,name=string,description=string,created_by=int,creator=object{id=int,name=string,username=string},created_at=string,updated_at=string}} "群組創建成功"
// @Failure 400 {object} object{error=bool,message=string} "請求格式錯誤或群組名稱為空"
// @Failure 401 {object} object{error=bool,message=string} "未授權"
// @Failure 500 {object} object{error=bool,message=string} "服務器內部錯誤"
// @Router /groups [post]
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
// @Summary 獲取群組詳細資訊
// @Description 獲取群組的詳細信息，包含成員列表和當前用戶角色
// @Tags 群組
// @Produce json
// @Security BearerAuth
// @Param id path int true "群組ID"
// @Success 200 {object} object{error=bool,data=object{group=object{id=int,name=string,description=string,created_by=int,creator=object{id=int,name=string,username=string},created_at=string,updated_at=string},members=[]object{id=int,name=string,username=string,avatar=string},current_user_role=string}} "群組詳細資訊"
// @Failure 401 {object} object{error=bool,message=string} "未授權"
// @Failure 403 {object} object{error=bool,message=string} "不是群組成員"
// @Failure 404 {object} object{error=bool,message=string} "群組不存在"
// @Failure 500 {object} object{error=bool,message=string} "服務器內部錯誤"
// @Router /groups/{id} [get]
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
// @Summary 更新群組資訊
// @Description 更新群組的名稱和描述，需要管理員權限
// @Tags 群組
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "群組ID"
// @Param request body object{name=string,description=string} true "更新資料"
// @Success 200 {object} object{error=bool,message=string,data=object{id=int,name=string,description=string,created_by=int,creator=object{id=int,name=string,username=string},created_at=string,updated_at=string}} "群組更新成功"
// @Failure 400 {object} object{error=bool,message=string} "請求格式錯誤或群組名稱為空"
// @Failure 401 {object} object{error=bool,message=string} "未授權"
// @Failure 403 {object} object{error=bool,message=string} "沒有管理員權限"
// @Failure 404 {object} object{error=bool,message=string} "群組不存在"
// @Failure 500 {object} object{error=bool,message=string} "服務器內部錯誤"
// @Router /groups/{id} [put]
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
// @Summary 刪除群組
// @Description 刪除群組及其所有相關資料，只有創建者可以執行此操作
// @Tags 群組
// @Produce json
// @Security BearerAuth
// @Param id path int true "群組ID"
// @Success 200 {object} object{error=bool,message=string} "群組刪除成功"
// @Failure 401 {object} object{error=bool,message=string} "未授權"
// @Failure 403 {object} object{error=bool,message=string} "只有群組創建者可以刪除群組"
// @Failure 404 {object} object{error=bool,message=string} "群組不存在"
// @Failure 500 {object} object{error=bool,message=string} "服務器內部錯誤"
// @Router /groups/{id} [delete]
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
// @Summary 添加群組成員
// @Description 向群組添加新成員，需要管理員權限
// @Tags 群組
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "群組ID"
// @Param request body object{user_id=int,role=string} true "成員資料，role可選值：member, admin"
// @Success 201 {object} object{error=bool,message=string,data=object{user_id=int,role=string}} "成員添加成功"
// @Failure 400 {object} object{error=bool,message=string} "請求格式錯誤或用戶ID為空"
// @Failure 401 {object} object{error=bool,message=string} "未授權"
// @Failure 403 {object} object{error=bool,message=string} "沒有管理員權限"
// @Failure 404 {object} object{error=bool,message=string} "群組或用戶不存在"
// @Failure 409 {object} object{error=bool,message=string} "用戶已是群組成員"
// @Failure 500 {object} object{error=bool,message=string} "服務器內部錯誤"
// @Router /groups/{id}/members [post]
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
// @Summary 移除群組成員
// @Description 從群組中移除成員，需要管理員權限，不能移除創建者
// @Tags 群組
// @Produce json
// @Security BearerAuth
// @Param id path int true "群組ID"
// @Param userId path int true "要移除的用戶ID"
// @Success 200 {object} object{error=bool,message=string} "成員移除成功"
// @Failure 401 {object} object{error=bool,message=string} "未授權"
// @Failure 403 {object} object{error=bool,message=string} "沒有管理員權限或不能移除創建者"
// @Failure 404 {object} object{error=bool,message=string} "群組或成員不存在"
// @Failure 500 {object} object{error=bool,message=string} "服務器內部錯誤"
// @Router /groups/{id}/members/{userId} [delete]
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
