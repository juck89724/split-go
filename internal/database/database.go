package database

import (
	"split-go/internal/models"
	"time"

	"golang.org/x/crypto/bcrypt"
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
		&models.UserSession{},
		&models.SecurityEvent{},
	)
}

// SeedAll 執行所有 seed 操作
func SeedAll(db *gorm.DB) error {
	if err := SeedCategories(db); err != nil {
		return err
	}

	if err := SeedUsers(db); err != nil {
		return err
	}

	if err := SeedGroups(db); err != nil {
		return err
	}

	if err := SeedGroupMembers(db); err != nil {
		return err
	}

	if err := SeedTransactions(db); err != nil {
		return err
	}

	return nil
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

// SeedUsers 創建測試用戶
func SeedUsers(db *gorm.DB) error {
	// 預設密碼：password123
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	users := []models.User{
		{
			Email:    "alice@example.com",
			Username: "alice",
			Password: string(hashedPassword),
			Name:     "張愛莉絲",
			Avatar:   "https://ui-avatars.com/api/?name=Alice&background=ff6b6b&color=fff",
		},
		{
			Email:    "bob@example.com",
			Username: "bob",
			Password: string(hashedPassword),
			Name:     "李小明",
			Avatar:   "https://ui-avatars.com/api/?name=Bob&background=4ecdc4&color=fff",
		},
		{
			Email:    "charlie@example.com",
			Username: "charlie",
			Password: string(hashedPassword),
			Name:     "王大華",
			Avatar:   "https://ui-avatars.com/api/?name=Charlie&background=45b7d1&color=fff",
		},
		{
			Email:    "diana@example.com",
			Username: "diana",
			Password: string(hashedPassword),
			Name:     "陳美玲",
			Avatar:   "https://ui-avatars.com/api/?name=Diana&background=96ceb4&color=fff",
		},
		{
			Email:    "eve@example.com",
			Username: "eve",
			Password: string(hashedPassword),
			Name:     "林小雨",
			Avatar:   "https://ui-avatars.com/api/?name=Eve&background=ffeaa7&color=fff",
		},
	}

	for _, user := range users {
		// 檢查用戶是否已存在
		var existingUser models.User
		result := db.Where("email = ?", user.Email).First(&existingUser)
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&user).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// SeedGroups 創建測試群組
func SeedGroups(db *gorm.DB) error {
	// 先確保用戶存在
	var alice, bob models.User
	if err := db.Where("email = ?", "alice@example.com").First(&alice).Error; err != nil {
		return err
	}
	if err := db.Where("email = ?", "bob@example.com").First(&bob).Error; err != nil {
		return err
	}

	groups := []models.Group{
		{
			Name:        "室友分帳群",
			Description: "跟室友一起分攤生活費用",
			CreatedBy:   alice.ID,
		},
		{
			Name:        "日本旅遊",
			Description: "東京五日遊費用分攤",
			CreatedBy:   bob.ID,
		},
		{
			Name:        "公司聚餐",
			Description: "部門聚餐費用",
			CreatedBy:   alice.ID,
		},
	}

	for _, group := range groups {
		// 檢查群組是否已存在
		var existingGroup models.Group
		result := db.Where("name = ? AND created_by = ?", group.Name, group.CreatedBy).First(&existingGroup)
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&group).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// SeedGroupMembers 創建群組成員關聯
func SeedGroupMembers(db *gorm.DB) error {
	// 取得用戶和群組
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		return err
	}

	var groups []models.Group
	if err := db.Find(&groups).Error; err != nil {
		return err
	}

	if len(users) < 3 || len(groups) < 1 {
		return nil // 沒有足夠的資料來創建關聯
	}

	// 室友分帳群成員 (Alice, Bob, Charlie)
	if len(groups) >= 1 {
		roomMateMembers := []models.GroupMember{
			{GroupID: groups[0].ID, UserID: users[0].ID, Role: "admin", JoinedAt: time.Now()},
			{GroupID: groups[0].ID, UserID: users[1].ID, Role: "member", JoinedAt: time.Now()},
			{GroupID: groups[0].ID, UserID: users[2].ID, Role: "member", JoinedAt: time.Now()},
		}

		for _, member := range roomMateMembers {
			var existingMember models.GroupMember
			result := db.Where("group_id = ? AND user_id = ?", member.GroupID, member.UserID).First(&existingMember)
			if result.Error == gorm.ErrRecordNotFound {
				if err := db.Create(&member).Error; err != nil {
					return err
				}
			}
		}
	}

	// 日本旅遊群成員 (Bob, Diana, Eve)
	if len(groups) >= 2 && len(users) >= 5 {
		travelMembers := []models.GroupMember{
			{GroupID: groups[1].ID, UserID: users[1].ID, Role: "admin", JoinedAt: time.Now()},
			{GroupID: groups[1].ID, UserID: users[3].ID, Role: "member", JoinedAt: time.Now()},
			{GroupID: groups[1].ID, UserID: users[4].ID, Role: "member", JoinedAt: time.Now()},
		}

		for _, member := range travelMembers {
			var existingMember models.GroupMember
			result := db.Where("group_id = ? AND user_id = ?", member.GroupID, member.UserID).First(&existingMember)
			if result.Error == gorm.ErrRecordNotFound {
				if err := db.Create(&member).Error; err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// SeedTransactions 創建測試交易
func SeedTransactions(db *gorm.DB) error {
	// 取得必要的資料
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		return err
	}

	var groups []models.Group
	if err := db.Find(&groups).Error; err != nil {
		return err
	}

	var categories []models.Category
	if err := db.Find(&categories).Error; err != nil {
		return err
	}

	if len(users) < 3 || len(groups) < 1 || len(categories) < 1 {
		return nil // 沒有足夠的資料
	}

	// 室友分帳群的交易
	transactions := []models.Transaction{
		{
			GroupID:     groups[0].ID,
			Description: "便利商店買日用品",
			Amount:      450.0,
			Currency:    "TWD",
			CategoryID:  categories[4].ID, // 購物
			PaidBy:      users[0].ID,      // Alice
			CreatedBy:   users[0].ID,
			Notes:       "衛生紙、洗髮精等",
		},
		{
			GroupID:     groups[0].ID,
			Description: "週末聚餐",
			Amount:      1200.0,
			Currency:    "TWD",
			CategoryID:  categories[0].ID, // 餐飲
			PaidBy:      users[1].ID,      // Bob
			CreatedBy:   users[1].ID,
			Notes:       "天母牛排館",
		},
		{
			GroupID:     groups[0].ID,
			Description: "電費分攤",
			Amount:      2400.0,
			Currency:    "TWD",
			CategoryID:  categories[7].ID, // 其他
			PaidBy:      users[2].ID,      // Charlie
			CreatedBy:   users[2].ID,
			Notes:       "9月電費帳單",
		},
	}

	for _, transaction := range transactions {
		// 檢查交易是否已存在
		var existingTransaction models.Transaction
		result := db.Where("description = ? AND group_id = ? AND paid_by = ?",
			transaction.Description, transaction.GroupID, transaction.PaidBy).First(&existingTransaction)
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&transaction).Error; err != nil {
				return err
			}

			// 創建對應的分帳記錄
			if err := createTransactionSplits(db, transaction); err != nil {
				return err
			}
		}
	}

	return nil
}

// createTransactionSplits 為交易創建分帳記錄
func createTransactionSplits(db *gorm.DB, transaction models.Transaction) error {
	// 取得群組成員
	var members []models.GroupMember
	if err := db.Where("group_id = ?", transaction.GroupID).Find(&members).Error; err != nil {
		return err
	}

	if len(members) == 0 {
		return nil
	}

	// 平均分攤
	amountPerPerson := transaction.Amount / float64(len(members))
	percentagePerPerson := 100.0 / float64(len(members))

	for _, member := range members {
		split := models.TransactionSplit{
			TransactionID: transaction.ID,
			UserID:        member.UserID,
			Amount:        amountPerPerson,
			Percentage:    percentagePerPerson,
			SplitType:     models.SplitEqual,
		}

		if err := db.Create(&split).Error; err != nil {
			return err
		}
	}

	return nil
}
