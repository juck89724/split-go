package database

import (
	"split-go/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Init 初始化資料庫連接
func Init(databaseURL string) (*gorm.DB, error) {
	// 設定 GORM 配置
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 連接資料庫
	db, err := gorm.Open(postgres.Open(databaseURL), config)
	if err != nil {
		return nil, err
	}

	// 自動遷移
	err = AutoMigrate(db)
	if err != nil {
		return nil, err
	}

	// 建立預設分類
	err = SeedCategories(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// AutoMigrate 執行資料庫遷移
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.GroupMember{},
		&models.Category{},
		&models.Transaction{},
		&models.TransactionSplit{},
		&models.Settlement{},
	)
}

// SeedCategories 建立預設分類
func SeedCategories(db *gorm.DB) error {
	categories := []models.Category{
		{Name: "餐飲", Icon: "🍽️", Color: "#FF6B6B"},
		{Name: "交通", Icon: "🚗", Color: "#4ECDC4"},
		{Name: "住宿", Icon: "🏠", Color: "#45B7D1"},
		{Name: "娛樂", Icon: "🎬", Color: "#96CEB4"},
		{Name: "購物", Icon: "🛍️", Color: "#FFEAA7"},
		{Name: "醫療", Icon: "🏥", Color: "#DDA0DD"},
		{Name: "教育", Icon: "📚", Color: "#98D8C8"},
		{Name: "其他", Icon: "💡", Color: "#F7DC6F"},
	}

	for _, category := range categories {
		// 檢查分類是否已存在
		var existingCategory models.Category
		result := db.Where("name = ?", category.Name).First(&existingCategory)
		if result.Error == gorm.ErrRecordNotFound {
			// 分類不存在，創建新分類
			if err := db.Create(&category).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
