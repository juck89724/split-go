# Split Go 測試文件

本文件夾包含 Split Go 分帳系統的完整測試套件，涵蓋所有核心功能的單元測試和整合測試。

## 📁 測試結構

```
tests/
├── handlers/                    # Handler 層測試
│   ├── auth_test.go            # 🔐 認證相關測試
│   ├── user_test.go            # 👤 用戶相關測試  
│   ├── group_test.go           # 👥 群組管理測試
│   ├── transaction_test.go     # 💰 交易相關測試
│   ├── settlement_test.go      # ⚖️ 結算相關測試
│   └── category_test.go        # 📁 分類相關測試
└── README.md                   # 📖 測試說明文件
```

## 🚀 快速開始

### 執行所有測試
```bash
# 執行所有測試（簡潔輸出）
go test ./tests/...

# 執行所有測試（詳細輸出）
go test -v ./tests/...

# 顯示測試覆蓋率
go test -cover ./tests/...

# 生成 HTML 覆蓋率報告
go test -coverprofile=coverage.out ./tests/...
go tool cover -html=coverage.out -o coverage.html
```

### 執行特定模組測試
```bash
# 認證模組
go test -v ./tests/handlers -run TestRegister
go test -v ./tests/handlers -run TestLogin

# 用戶模組  
go test -v ./tests/handlers -run TestGetProfile
go test -v ./tests/handlers -run TestUpdateProfile
go test -v ./tests/handlers -run TestUpdateFCMToken

# 群組模組
go test -v ./tests/handlers -run TestCreateGroup
go test -v ./tests/handlers -run TestGetGroup
go test -v ./tests/handlers -run TestAddMember

# 交易模組
go test -v ./tests/handlers -run TestTransaction

# 結算模組
go test -v ./tests/handlers -run TestSettlement

# 分類模組
go test -v ./tests/handlers -run TestGetCategories
```

### 執行單個測試文件
```bash
# 注意：需要指定完整路徑或使用目錄執行
go test -v ./tests/handlers/ -run TestSpecificFunction
```

## 📊 測試覆蓋範圍

### 🔐 認證測試 (auth_test.go)
| 功能 | 測試案例 | 涵蓋範圍 |
|------|---------|---------|
| **用戶註冊** | `TestRegister` | ✅ 成功註冊<br>✅ Email 格式驗證<br>✅ 密碼長度驗證<br>✅ 重複檢查 |
| **用戶登入** | `TestLogin` | ✅ 成功登入<br>✅ 錯誤密碼<br>✅ 不存在用戶<br>✅ Token 生成 |
| **Token 刷新** | `TestRefreshToken` | ✅ 成功刷新<br>✅ 無效 Token<br>✅ 過期處理 |
| **用戶登出** | `TestLogout` | ✅ 成功登出<br>✅ 會話清理 |

### 👤 用戶測試 (user_test.go) 
| 功能 | 測試案例 | 涵蓋範圍 |
|------|---------|---------|
| **獲取資料** | `TestGetProfile` | ✅ 成功獲取<br>✅ 未認證處理<br>✅ 用戶不存在 |
| **更新資料** | `TestUpdateProfile` | ✅ 完整更新<br>✅ 部分更新<br>✅ 權限驗證 |
| **FCM Token** | `TestUpdateFCMToken` | ✅ 成功更新<br>✅ 空值驗證<br>✅ 資料庫更新確認 |

### 👥 群組測試 (group_test.go)
| 功能 | 測試案例 | 涵蓋範圍 |
|------|---------|---------|
| **群組管理** | `TestCreateGroup`<br>`TestUpdateGroup`<br>`TestDeleteGroup` | ✅ CRUD 操作<br>✅ 權限控制<br>✅ 資料驗證 |
| **成員管理** | `TestAddMember`<br>`TestRemoveMember` | ✅ 添加/移除成員<br>✅ 角色管理<br>✅ 權限層級 |
| **查詢功能** | `TestGetUserGroups`<br>`TestGetGroup` | ✅ 列表查詢<br>✅ 詳情查詢<br>✅ 權限過濾 |

### 💰 交易測試 (transaction_test.go)
| 功能 | 測試案例 | 涵蓋範圍 |
|------|---------|---------|
| **交易記錄** | `TestNewTransactionResponse` | ✅ 回應格式<br>✅ 分帳計算<br>✅ 用戶角色 |
| **群組平衡** | `TestGetGroupBalance` | ✅ 平衡計算<br>✅ 多用戶場景<br>✅ 權限驗證 |

### ⚖️ 結算測試 (settlement_test.go)
| 功能 | 測試案例 | 涵蓋範圍 |
|------|---------|---------|
| **結算記錄** | `TestGetSettlements`<br>`TestCreateSettlement` | ✅ 記錄查詢<br>✅ 創建驗證<br>✅ 業務規則 |
| **狀態管理** | `TestMarkAsPaid`<br>`TestCancelSettlement` | ✅ 狀態轉換<br>✅ 權限控制<br>✅ 資料一致性 |
| **智能建議** | `TestGetSettlementSuggestions` | ✅ 平衡計算<br>✅ 最優化演算法<br>✅ 複雜場景 |

### 📁 分類測試 (category_test.go)
| 功能 | 測試案例 | 涵蓋範圍 |
|------|---------|---------|
| **基本功能** | `TestGetCategories` | ✅ 列表查詢<br>✅ 排序驗證<br>✅ 空資料處理 |
| **資料驗證** | `TestCategoryContentValidation` | ✅ 欄位完整性<br>✅ 格式驗證<br>✅ 邊界條件 |
| **性能測試** | `TestCategoryPerformance` | ✅ 大量資料<br>✅ 查詢效能<br>✅ 記憶體使用 |

## 🛠️ 測試架構

### 資料庫設置
每個測試使用獨立的 **SQLite 記憶體資料庫**，確保：
- ⚡ **高速執行** - 記憶體操作，無 I/O 延遲
- 🔒 **完全隔離** - 測試間互不影響
- 🚫 **零依賴** - 不需要外部資料庫

### 輔助函數
```go
// 基礎設置
setupTestDB()              // 基本記憶體資料庫
setupTestConfig()          // 測試配置（JWT密鑰等）

// 專用設置  
setupGroupTestDB()         // 群組功能專用DB
setupSettlementTestDB()    // 結算功能專用DB
setupCategoryTestDB()      // 分類功能專用DB

// 測試資料創建
createTestUser()           // 創建測試用戶
createTestGroup()          // 創建測試群組
createTestTransaction()    // 創建測試交易
createTestSettlement()     // 創建測試結算
createTestCategory()       // 創建測試分類

// 關聯操作
addGroupMember()           // 添加群組成員
```

### Mock 中間件
```go
// 模擬用戶認證
app.Use("/protected", func(c *fiber.Ctx) error {
    c.Locals("user_id", testUserID)
    return c.Next()
})

// 模擬權限檢查
app.Use("/admin", func(c *fiber.Ctx) error {
    c.Locals("user_role", "admin")
    return c.Next()
})
```

## 📈 測試品質標準

### ✅ 測試完整性檢查清單
- [ ] **正常流程** - 成功案例測試
- [ ] **異常處理** - 錯誤情況測試  
- [ ] **邊界條件** - 極值和特殊輸入
- [ ] **權限控制** - 各種角色權限
- [ ] **資料驗證** - 輸入格式檢查
- [ ] **狀態檢查** - 資料庫狀態變化
- [ ] **併發安全** - 多用戶操作

### 🎯 測試原則
1. **描述性命名** - 使用繁體中文清楚描述測試目的
2. **表格驅動** - 使用結構化測試案例
3. **獨立執行** - 每個測試可單獨運行
4. **快速回饋** - 測試執行時間控制在秒級
5. **易於維護** - 清晰的測試結構和註釋

## 🔍 調試和故障排除

### 常見問題

**Q: 測試執行失敗，提示資料庫錯誤？**
```bash
# 確保在專案根目錄執行
cd /path/to/split-go

# 檢查依賴是否正確安裝
go mod tidy

# 清理並重新運行
go clean -testcache
go test -v ./tests/handlers/
```

**Q: 部分測試通過，部分失敗？**
```bash
# 單獨執行失敗的測試，查看詳細錯誤
go test -v ./tests/handlers/ -run TestSpecificFailingTest

# 檢查測試資料是否正確清理
# 查看測試代碼中的清理邏輯 (defer db.Delete(...))
```

**Q: 如何查看測試覆蓋率？**
```bash
# 生成覆蓋率報告
go test -coverprofile=coverage.out ./tests/handlers/

# 查看總體覆蓋率
go tool cover -func=coverage.out

# 生成 HTML 報告（推薦）
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

## 📝 貢獻指南

### 新增測試
1. **選擇對應文件** - 根據功能選擇 `*_test.go` 文件
2. **遵循命名規範** - `TestFunctionName` 格式
3. **使用測試模板**：
```go
func TestNewFeature(t *testing.T) {
    // 設置
    db := setupTestDB()
    handler := handlers.NewHandler(db)
    
    // 測試案例
    tests := []struct {
        name           string
        input          interface{}
        expectedStatus int
        expectedError  string
    }{
        // 案例定義...
    }
    
    // 執行測試
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 測試邏輯...
        })
    }
    
    // 清理
    // db.Delete(...)
}
```