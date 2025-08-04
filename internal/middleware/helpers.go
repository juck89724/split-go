package middleware

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// ParseSettlementIDFromParams 從 URL 參數中安全地解析結算 ID
func ParseSettlementIDFromParams(c *fiber.Ctx) (uint, error) {
	idStr := c.Params("id")
	if idStr == "" {
		return 0, fiber.NewError(fiber.StatusBadRequest, "缺少結算 ID")
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "無效的結算 ID")
	}

	return uint(id), nil
}

// ParseUserIDFromParams 從 URL 參數中安全地解析用戶 ID
func ParseUserIDFromParams(c *fiber.Ctx) (uint, error) {
	idStr := c.Params("userId") // 注意這裡是 userId 不是 id
	if idStr == "" {
		return 0, fiber.NewError(fiber.StatusBadRequest, "缺少用戶 ID")
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "無效的用戶 ID")
	}

	return uint(id), nil
}
