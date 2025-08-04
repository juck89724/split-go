package routes

import (
	"split-go/internal/config"
	"split-go/internal/handlers"
	"split-go/internal/middleware"

	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"gorm.io/gorm"
)

// Setup 設定所有路由
func Setup(app *fiber.App, db *gorm.DB, cfg *config.Config) {
	// 健康檢查
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Split Go API is running",
		})
	})

	// Swagger 文檔路由
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// API 路由群組
	api := app.Group("/api/v1")

	// 初始化處理器
	authHandler := handlers.NewAuthHandler(db, cfg)
	userHandler := handlers.NewUserHandler(db)
	groupHandler := handlers.NewGroupHandler(db)
	transactionHandler := handlers.NewTransactionHandler(db)
	categoryHandler := handlers.NewCategoryHandler(db)
	settlementHandler := handlers.NewSettlementHandler(db)

	// 認証相關路由 (不需要驗證)
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)
	auth.Post("/device-refresh", authHandler.DeviceRefresh)
	auth.Post("/logout", authHandler.Logout)

	// 需要認證的路由 - 使用企業級中間件
	protected := api.Group("/", middleware.EnterpriseJWTMiddleware(cfg.AccessTokenSecret, cfg.RefreshTokenSecret))

	// 用戶相關路由
	users := protected.Group("/users")
	users.Get("/me", userHandler.GetProfile)
	users.Put("/me", userHandler.UpdateProfile)
	users.Post("/fcm-token", userHandler.UpdateFCMToken)

	// 企業級認證管理路由
	devices := protected.Group("/devices")
	devices.Get("/", authHandler.GetUserDevices)
	devices.Delete("/:deviceId", authHandler.RevokeDevice)

	// 安全事件路由
	security := protected.Group("/security")
	security.Get("/events", authHandler.GetSecurityEvents)

	// 群組相關路由
	groups := protected.Group("/groups")
	groups.Get("/", groupHandler.GetUserGroups)
	groups.Post("/", groupHandler.CreateGroup)
	groups.Get("/:id", groupHandler.GetGroup)
	groups.Put("/:id", groupHandler.UpdateGroup)
	groups.Delete("/:id", groupHandler.DeleteGroup)
	groups.Post("/:id/members", groupHandler.AddMember)
	groups.Delete("/:id/members/:userId", groupHandler.RemoveMember)

	// 交易相關路由
	transactions := protected.Group("/transactions")
	transactions.Get("/", transactionHandler.GetTransactions)
	transactions.Post("/", transactionHandler.CreateTransaction)
	transactions.Get("/:id", transactionHandler.GetTransaction)
	transactions.Put("/:id", transactionHandler.UpdateTransaction)
	transactions.Delete("/:id", transactionHandler.DeleteTransaction)

	// 群組交易路由
	groups.Get("/:id/transactions", transactionHandler.GetGroupTransactions)
	groups.Get("/:id/balance", transactionHandler.GetGroupBalance)

	// 分類相關路由
	categories := protected.Group("/categories")
	categories.Get("/", categoryHandler.GetCategories)

	// 結算相關路由
	settlements := protected.Group("/settlements")
	settlements.Get("/", settlementHandler.GetSettlements)
	settlements.Post("/", settlementHandler.CreateSettlement)
	settlements.Put("/:id/paid", settlementHandler.MarkAsPaid)
	settlements.Delete("/:id", settlementHandler.CancelSettlement)

	// 群組結算路由
	groups.Get("/:id/settlement-suggestions", settlementHandler.GetSettlementSuggestions)
}
