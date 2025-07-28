package main

import (
	"flag"
	"fmt"
	"log"
	"split-go/internal/config"
	"split-go/internal/database"
	"split-go/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	var (
		action = flag.String("action", "migrate", "執行動作: migrate, reset, seed")
		dbURL  = flag.String("db", "", "資料庫連接字符串 (可選，將使用環境變數)")
	)
	flag.Parse()

	// 初始化配置
	cfg := config.Load()

	// 使用指定的資料庫 URL 或配置文件中的 URL
	databaseURL := cfg.DatabaseURL
	if *dbURL != "" {
		databaseURL = *dbURL
	}

	fmt.Printf("連接資料庫: %s\n", databaseURL)

	// 設定 GORM 配置
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 連接資料庫
	db, err := gorm.Open(postgres.Open(databaseURL), gormConfig)
	if err != nil {
		log.Fatal("資料庫連接失敗:", err)
	}

	switch *action {
	case "migrate":
		fmt.Println("開始執行資料庫遷移...")
		if err := database.AutoMigrate(db); err != nil {
			log.Fatal("遷移失敗:", err)
		}
		fmt.Println("✅ 資料庫遷移完成")

		fmt.Println("開始建立預設分類...")
		if err := database.SeedAll(db); err != nil {
			log.Fatal("建立預設資料失敗:", err)
		}
		fmt.Println("✅ 預設分類建立完成")

	case "reset":
		fmt.Println("⚠️  警告: 這將刪除所有資料表!")
		fmt.Print("確定要繼續嗎? (y/N): ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println("操作已取消")
			return
		}

		fmt.Println("開始重置資料庫...")

		// 刪除所有表 (按相反順序)
		tables := []interface{}{
			&models.SecurityEvent{},
			&models.Settlement{},
			&models.TransactionSplit{},
			&models.Transaction{},
			&models.GroupMember{},
			&models.Group{},
			&models.UserSession{},
			&models.Category{},
			&models.User{},
		}

		for _, table := range tables {
			if err := db.Migrator().DropTable(table); err != nil {
				fmt.Printf("警告: 刪除表失敗: %v\n", err)
			}
		}

		fmt.Println("開始重新建立資料表...")
		if err := database.AutoMigrate(db); err != nil {
			log.Fatal("重置後遷移失敗:", err)
		}

		fmt.Println("開始建立預設分類...")
		if err := database.SeedAll(db); err != nil {
			log.Fatal("重置後建立預設資料失敗:", err)
		}

		fmt.Println("✅ 資料庫重置完成")
	case "seed":
		fmt.Println("開始建立完整測試資料...")
		fmt.Println("📂 建立分類...")
		if err := database.SeedCategories(db); err != nil {
			log.Fatal("建立分類失敗:", err)
		}
		fmt.Println("✅ 分類建立完成")

		fmt.Println("👥 建立測試用戶...")
		if err := database.SeedUsers(db); err != nil {
			log.Fatal("建立用戶失敗:", err)
		}
		fmt.Println("✅ 用戶建立完成 (密碼: password123)")

		fmt.Println("🏠 建立測試群組...")
		if err := database.SeedGroups(db); err != nil {
			log.Fatal("建立群組失敗:", err)
		}
		fmt.Println("✅ 群組建立完成")

		fmt.Println("🔗 建立群組成員關聯...")
		if err := database.SeedGroupMembers(db); err != nil {
			log.Fatal("建立群組成員失敗:", err)
		}
		fmt.Println("✅ 群組成員關聯建立完成")

		fmt.Println("💰 建立測試交易...")
		if err := database.SeedTransactions(db); err != nil {
			log.Fatal("建立交易失敗:", err)
		}
		fmt.Println("✅ 測試交易建立完成")

		fmt.Println()
		fmt.Println("🎉 完整測試資料建立完成!")
		fmt.Println()
		fmt.Println("📋 測試用戶帳號:")
		fmt.Println("   📧 alice@example.com (張愛莉絲)")
		fmt.Println("   📧 bob@example.com (李小明)")
		fmt.Println("   📧 charlie@example.com (王大華)")
		fmt.Println("   📧 diana@example.com (陳美玲)")
		fmt.Println("   📧 eve@example.com (林小雨)")
		fmt.Println("   🔑 統一密碼: password123")
		fmt.Println()
		fmt.Println("🏠 測試群組:")
		fmt.Println("   1. 室友分帳群 (Alice, Bob, Charlie)")
		fmt.Println("   2. 日本旅遊 (Bob, Diana, Eve)")
		fmt.Println("   3. 公司聚餐 (創建者: Alice)")

	default:
		fmt.Printf("未知動作: %s\n", *action)
		fmt.Println("可用動作:")
		fmt.Println("  migrate   - 執行資料庫遷移和基本分類")
		fmt.Println("  reset     - 重置資料庫 (刪除所有資料)")
		fmt.Println("  seed      - 建立完整測試資料 (用戶、群組、交易)")
	}
}
