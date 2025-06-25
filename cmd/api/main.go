package main

import (
	"log"
	"os"

	"split-go/internal/config"
	"split-go/internal/database"
	"split-go/internal/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// 載入環境變數
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 檔案")
	}

	// 初始化配置
	cfg := config.Load()

	// 初始化資料庫
	db, err := database.Init(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("無法連接資料庫:", err)
	}

	// 建立 Fiber 應用程式
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		},
	})

	// 中介軟體
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// 路由設定
	routes.Setup(app, db, cfg)

	// 啟動伺服器
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("伺服器運行在埠口 %s", port)
	log.Fatal(app.Listen(":" + port))
}
