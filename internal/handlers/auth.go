package handlers

import (
	"split-go/internal/config"
	"split-go/internal/middleware"
	"split-go/internal/models"
	"split-go/internal/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		db:  db,
		cfg: cfg,
	}
}

// 註冊請求結構
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=20"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required"`
}

// 登入請求結構
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Token 回應結構
type TokenResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
	User         models.User `json:"user"`
}

// Register 用戶註冊
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
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

	// 生成 tokens
	tokens, err := h.generateTokens(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "生成 token 失敗",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error":   false,
		"message": "註冊成功",
		"data":    tokens,
	})
}

// Login 用戶登入
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
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

	// 查找用戶
	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Email 或密碼錯誤",
		})
	}

	// 驗證密碼
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Email 或密碼錯誤",
		})
	}

	// 生成 tokens
	tokens, err := h.generateTokens(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "生成 token 失敗",
		})
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "登入成功",
		"data":    tokens,
	})
}

// RefreshToken 刷新 token
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	// 這裡可以實現 refresh token 邏輯
	// 暫時返回未實現
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error":   true,
		"message": "功能尚未實現",
	})
}

// generateTokens 生成 access token 和 refresh token
func (h *AuthHandler) generateTokens(user models.User) (*TokenResponse, error) {
	// 設定 token 過期時間
	expirationTime := time.Now().Add(24 * time.Hour)

	// 創建 JWT claims
	claims := &middleware.JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// 創建 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	// 暫時使用相同的 token 作為 refresh token
	// 實際應用中應該使用不同的密鑰和過期時間
	refreshToken := accessToken

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expirationTime.Unix(),
		User:         user,
	}, nil
}
