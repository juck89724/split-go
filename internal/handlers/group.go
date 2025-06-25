package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type GroupHandler struct {
	db *gorm.DB
}

func NewGroupHandler(db *gorm.DB) *GroupHandler {
	return &GroupHandler{db: db}
}

func (h *GroupHandler) GetUserGroups(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}

func (h *GroupHandler) CreateGroup(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}

func (h *GroupHandler) GetGroup(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}

func (h *GroupHandler) UpdateGroup(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}

func (h *GroupHandler) DeleteGroup(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}

func (h *GroupHandler) AddMember(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}

func (h *GroupHandler) RemoveMember(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": true, "message": "功能尚未實現",
	})
}
