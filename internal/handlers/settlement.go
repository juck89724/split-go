package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type SettlementHandler struct {
	db *gorm.DB
}

func NewSettlementHandler(db *gorm.DB) *SettlementHandler {
	return &SettlementHandler{db: db}
}

func (h *SettlementHandler) GetSettlements(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}

func (h *SettlementHandler) CreateSettlement(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}

func (h *SettlementHandler) MarkAsPaid(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}

func (h *SettlementHandler) CancelSettlement(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}

func (h *SettlementHandler) GetSettlementSuggestions(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}
