package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type TransactionHandler struct {
	db *gorm.DB
}

func NewTransactionHandler(db *gorm.DB) *TransactionHandler {
	return &TransactionHandler{db: db}
}

func (h *TransactionHandler) GetTransactions(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}

func (h *TransactionHandler) CreateTransaction(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}

func (h *TransactionHandler) GetTransaction(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}

func (h *TransactionHandler) UpdateTransaction(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}

func (h *TransactionHandler) DeleteTransaction(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}

func (h *TransactionHandler) GetGroupTransactions(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}

func (h *TransactionHandler) GetGroupBalance(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}
