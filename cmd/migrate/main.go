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
		if err := database.SeedCategories(db); err != nil {
			log.Fatal("建立預設分類失敗:", err)
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
		if err := database.SeedCategories(db); err != nil {
			log.Fatal("重置後建立預設分類失敗:", err)
		}

		fmt.Println("✅ 資料庫重置完成")

	case "seed":
		fmt.Println("開始建立預設資料...")
		if err := database.SeedCategories(db); err != nil {
			log.Fatal("建立預設分類失敗:", err)
		}
		fmt.Println("✅ 預設資料建立完成")

	default:
		fmt.Printf("未知動作: %s\n", *action)
		fmt.Println("可用動作: migrate, reset, seed")
	}
}
