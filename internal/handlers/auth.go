package handlers

import (
	"split-go/internal/config"
	"split-go/internal/middleware"
	"split-go/internal/models"
	"split-go/internal/services"
	"split-go/internal/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db         *gorm.DB
	cfg        *config.Config
	jwtService *services.JWTService
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		db:         db,
		cfg:        cfg,
		jwtService: services.NewJWTService(db, cfg),
	}
}

// 登入請求結構
type EnterpriseLoginRequest struct {
	Email             string                     `json:"email" validate:"required,email"`
	Password          string                     `json:"password" validate:"required"`
	DeviceFingerprint services.DeviceFingerprint `json:"device_fingerprint"`
	DeviceName        string                     `json:"device_name"`
	DeviceType        string                     `json:"device_type"` // mobile/desktop/tablet
}

// Register 用戶註冊
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Username string `json:"username" validate:"required,min=3,max=20"`
		Password string `json:"password" validate:"required,min=6"`
		Name     string `json:"name" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "無效的請求數據",
		})
	}

	// 驗證輸入
	if err := utils.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	// 檢查 email 是否已存在
	var existingUser models.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":   true,
			"message": "Email 已經被使用",
		})
	}

	// 檢查用戶名是否已存在
	if err := h.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":   true,
			"message": "用戶名已經被使用",
		})
	}

	// 加密密碼
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "密碼加密失敗",
		})
	}

	// 創建用戶
	user := models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: string(hashedPassword),
		Name:     req.Name,
	}

	if err := h.db.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "創建用戶失敗",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error":   false,
		"message": "註冊成功",
		"data":    user,
	})
}

// Login 企業級用戶登入
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req EnterpriseLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "無效的請求數據",
		})
	}

	// 設置IP地址到設備指紋
	req.DeviceFingerprint.IPAddress = c.IP()

	// 驗證輸入
	if err := utils.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	// 查找用戶
	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		h.jwtService.LogSecurityEvent(0, "", "failed_login", c.IP(), map[string]interface{}{
			"email":  req.Email,
			"reason": "user_not_found",
		})
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Email 或密碼錯誤",
		})
	}

	// 驗證密碼
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		h.jwtService.LogSecurityEvent(user.ID, "", "failed_login", c.IP(), map[string]interface{}{
			"reason": "wrong_password",
		})
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Email 或密碼錯誤",
		})
	}

	// 使用 JWT 服務處理登入邏輯
	tokens, err := h.jwtService.HandleLogin(user, req.DeviceFingerprint, req.DeviceName, req.DeviceType, c.IP())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "登入處理失敗",
		})
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "登入成功",
		"data":    tokens,
	})
}

// RefreshToken 智能刷新 token
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "無效的請求數據",
		})
	}

	// 驗證輸入
	if err := utils.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	// 使用 JWT 服務處理刷新邏輯
	tokens, err := h.jwtService.HandleRefresh(req.RefreshToken, c.IP())
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "Token 刷新成功",
		"data":    tokens,
	})
}

// DeviceRefresh 設備認證降級刷新
func (h *AuthHandler) DeviceRefresh(c *fiber.Ctx) error {
	type DeviceRefreshRequest struct {
		DeviceFingerprint services.DeviceFingerprint `json:"device_fingerprint" validate:"required"`
	}

	var req DeviceRefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "無效的請求數據",
		})
	}

	// 從 Authorization header 獲取 device token
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "缺少 device token",
		})
	}

	deviceTokenString := strings.TrimPrefix(authHeader, "Bearer ")
	req.DeviceFingerprint.IPAddress = c.IP()

	// 使用 JWT 服務處理設備刷新邏輯
	tokens, err := h.jwtService.HandleDeviceRefresh(deviceTokenString, req.DeviceFingerprint, c.IP())
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "設備認證刷新成功",
		"data":    tokens,
	})
}

// Logout 用戶登出
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	sessionID := middleware.GetSessionIDFromContext(c)
	userID := middleware.GetUserIDFromContext(c)

	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "無效的會話",
		})
	}

	// 使用 JWT 服務處理登出邏輯
	if err := h.jwtService.HandleLogout(sessionID, userID, c.IP()); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "登出成功",
	})
}

// GetUserDevices 獲取用戶所有設備
func (h *AuthHandler) GetUserDevices(c *fiber.Ctx) error {
	userID := middleware.GetUserIDFromContext(c)
	currentSessionID := middleware.GetSessionIDFromContext(c)

	devices, err := h.jwtService.GetUserDevices(userID, currentSessionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "獲取設備列表失敗",
		})
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "獲取設備列表成功",
		"data":    devices,
	})
}

// RevokeDevice 撤銷指定設備
func (h *AuthHandler) RevokeDevice(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	userID := middleware.GetUserIDFromContext(c)
	currentSessionID := middleware.GetSessionIDFromContext(c)

	if deviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "設備 ID 不能為空",
		})
	}

	// 使用 JWT 服務處理設備撤銷邏輯
	if err := h.jwtService.RevokeDevice(deviceID, userID, currentSessionID, c.IP()); err != nil {
		if err.Error() == "設備不存在" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "設備已撤銷",
	})
}

// GetSecurityEvents 獲取安全事件記錄
func (h *AuthHandler) GetSecurityEvents(c *fiber.Ctx) error {
	userID := middleware.GetUserIDFromContext(c)

	events, err := h.jwtService.GetSecurityEvents(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "獲取安全事件失敗",
		})
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "獲取安全事件成功",
		"data":    events,
	})
}
