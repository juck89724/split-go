package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// AccessTokenClaims Access Token 聲明結構
type AccessTokenClaims struct {
	UserID       uint   `json:"user_id"`
	Email        string `json:"email"`
	SessionID    string `json:"session_id"`
	TokenVersion int    `json:"token_version"`
	DeviceID     string `json:"device_id"`
	jwt.RegisteredClaims
}

// RefreshTokenClaims Refresh Token 聲明結構
type RefreshTokenClaims struct {
	UserID    uint   `json:"user_id"`
	SessionID string `json:"session_id"`
	DeviceID  string `json:"device_id"`
	TokenType string `json:"token_type"` // 標記為 "refresh"
	jwt.RegisteredClaims
}

// DeviceTokenClaims Device Token 聲明結構
type DeviceTokenClaims struct {
	UserID    uint   `json:"user_id"`
	DeviceID  string `json:"device_id"`
	TokenType string `json:"token_type"` // 標記為 "device_auth"
	jwt.RegisteredClaims
}

// JWTMiddleware JWT 驗證中介軟體（保持向後兼容）
func JWTMiddleware(secret string) fiber.Handler {
	return EnterpriseJWTMiddleware(secret, "")
}

// EnterpriseJWTMiddleware 企業級 JWT 驗證中介軟體
func EnterpriseJWTMiddleware(accessSecret, refreshSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 從 Header 中獲取 Authorization
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "缺少 Authorization header",
			})
		}

		// 檢查 Bearer token 格式
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "無效的 token 格式",
			})
		}

		// 提取 token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 解析 Access Token
		token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
			// 驗證簽名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("意外的簽名方法")
			}
			return []byte(accessSecret), nil
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "無效的 access token",
			})
		}

		// 檢查 token 有效性
		if !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "access token 已失效",
			})
		}

		// 提取用戶資訊並注入上下文
		if claims, ok := token.Claims.(*AccessTokenClaims); ok {
			c.Locals("user_id", claims.UserID)
			c.Locals("user_email", claims.Email)
			c.Locals("session_id", claims.SessionID)
			c.Locals("device_id", claims.DeviceID)
			c.Locals("token_version", claims.TokenVersion)
		} else {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "無法解析 token 聲明",
			})
		}

		return c.Next()
	}
}

// Helper functions for extracting data from context
func GetUserIDFromContext(c *fiber.Ctx) uint {
	if userID := c.Locals("user_id"); userID != nil {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	return 0
}

func GetUserEmailFromContext(c *fiber.Ctx) string {
	if email := c.Locals("user_email"); email != nil {
		if e, ok := email.(string); ok {
			return e
		}
	}
	return ""
}

func GetSessionIDFromContext(c *fiber.Ctx) string {
	if sessionID := c.Locals("session_id"); sessionID != nil {
		if id, ok := sessionID.(string); ok {
			return id
		}
	}
	return ""
}

func GetDeviceIDFromContext(c *fiber.Ctx) string {
	if deviceID := c.Locals("device_id"); deviceID != nil {
		if id, ok := deviceID.(string); ok {
			return id
		}
	}
	return ""
}

func GetTokenVersionFromContext(c *fiber.Ctx) int {
	if version := c.Locals("token_version"); version != nil {
		if v, ok := version.(int); ok {
			return v
		}
	}
	return 0
}
