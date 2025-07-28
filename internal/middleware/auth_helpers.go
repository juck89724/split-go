package middleware

import (
	"split-go/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// AuthenticatedUser 包含認證用戶資訊
type AuthenticatedUser struct {
	User      models.User `json:"user"`
	UserID    uint        `json:"user_id"`
	Email     string      `json:"email"`
	SessionID string      `json:"session_id"`
	DeviceID  string      `json:"device_id"`
}

// GetCurrentUser 從 JWT 上下文中取得完整的用戶資訊
// 這是一個共用函數，所有需要驗證用戶的 handler 都可以使用
func GetCurrentUser(c *fiber.Ctx, db *gorm.DB) (*AuthenticatedUser, error) {
	// 從 context 中取得 JWT 解析出的用戶資訊
	userID := GetUserIDFromContext(c)
	if userID == 0 {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "未找到用戶資訊")
	}

	email := GetUserEmailFromContext(c)
	sessionID := GetSessionIDFromContext(c)
	deviceID := GetDeviceIDFromContext(c)

	// 從資料庫查詢完整的用戶資料
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.NewError(fiber.StatusNotFound, "用戶不存在")
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, "查詢用戶資料失敗")
	}

	return &AuthenticatedUser{
		User:      user,
		UserID:    userID,
		Email:     email,
		SessionID: sessionID,
		DeviceID:  deviceID,
	}, nil
}

// GetCurrentUserID 快速取得當前用戶 ID（不查詢資料庫）
// 適用於只需要用戶 ID 的簡單場景
func GetCurrentUserID(c *fiber.Ctx) (uint, error) {
	userID := GetUserIDFromContext(c)
	if userID == 0 {
		return 0, fiber.NewError(fiber.StatusUnauthorized, "未找到用戶資訊")
	}
	return userID, nil
}

// RequireGroupMember 驗證用戶是否為指定群組成員
// 這是一個常用的權限檢查，可以在群組相關的 handler 中使用
func RequireGroupMember(c *fiber.Ctx, db *gorm.DB, groupID uint) (*AuthenticatedUser, error) {
	// 取得當前用戶
	authUser, err := GetCurrentUser(c, db)
	if err != nil {
		return nil, err
	}

	// 檢查用戶是否為群組成員
	var member models.GroupMember
	if err := db.Where("group_id = ? AND user_id = ?", groupID, authUser.UserID).
		First(&member).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.NewError(fiber.StatusForbidden, "您不是此群組的成員")
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, "檢查群組權限失敗")
	}

	return authUser, nil
}

// RequireGroupAdmin 驗證用戶是否為指定群組管理員
// 適用於需要管理員權限的操作
func RequireGroupAdmin(c *fiber.Ctx, db *gorm.DB, groupID uint) (*AuthenticatedUser, error) {
	// 取得當前用戶
	authUser, err := GetCurrentUser(c, db)
	if err != nil {
		return nil, err
	}

	// 檢查用戶是否為群組管理員
	var member models.GroupMember
	if err := db.Where("group_id = ? AND user_id = ? AND role = ?", groupID, authUser.UserID, "admin").
		First(&member).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.NewError(fiber.StatusForbidden, "您沒有此群組的管理員權限")
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, "檢查管理員權限失敗")
	}

	return authUser, nil
}

// ParseGroupIDFromParams 從 URL 參數中安全地解析群組 ID
func ParseGroupIDFromParams(c *fiber.Ctx) (uint, error) {
	groupID, err := c.ParamsInt("id")
	if err != nil || groupID <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "無效的群組 ID")
	}
	return uint(groupID), nil
}

// ParseTransactionIDFromParams 從 URL 參數中安全地解析交易 ID
func ParseTransactionIDFromParams(c *fiber.Ctx) (uint, error) {
	transactionID, err := c.ParamsInt("id")
	if err != nil || transactionID <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "無效的交易 ID")
	}
	return uint(transactionID), nil
}
