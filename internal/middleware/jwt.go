package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// JWTClaims JWT 聲明結構
type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// JWTMiddleware JWT 驗證中介軟體
func JWTMiddleware(secret string) fiber.Handler {
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

		// 解析 token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// 驗證簽名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("意外的簽名方法")
			}
			return []byte(secret), nil
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "無效的 token",
			})
		}

		// 檢查 token 有效性
		if !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "token 已失效",
			})
		}

		// 提取用戶資訊
		if claims, ok := token.Claims.(*JWTClaims); ok {
			c.Locals("user_id", claims.UserID)
			c.Locals("user_email", claims.Email)
		} else {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "無法解析 token 聲明",
			})
		}

		return c.Next()
	}
}

// GetUserIDFromContext 從上下文中獲取用戶 ID
func GetUserIDFromContext(c *fiber.Ctx) uint {
	if userID := c.Locals("user_id"); userID != nil {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	return 0
}

// GetUserEmailFromContext 從上下文中獲取用戶 email
func GetUserEmailFromContext(c *fiber.Ctx) string {
	if email := c.Locals("user_email"); email != nil {
		if e, ok := email.(string); ok {
			return e
		}
	}
	return ""
}
