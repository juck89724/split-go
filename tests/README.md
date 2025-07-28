# 測試說明

這個資料夾包含了 Split Go 專案的所有測試文件。

## 測試結構

```
tests/
├── handlers/
│   ├── auth_test.go    # 認證相關測試
│   └── user_test.go    # 用戶相關測試
└── README.md           # 測試說明文件
```

## 執行測試

### 執行所有測試
```bash
go test ./tests/...
```

### 執行特定測試文件
```bash
# 執行認證測試
go test ./tests/handlers/auth_test.go

# 執行用戶測試
go test ./tests/handlers/user_test.go
```

### 執行特定測試函數
```bash
# 執行註冊功能測試
go test ./tests/handlers -run TestRegister

# 執行登入功能測試
go test ./tests/handlers -run TestLogin

# 執行用戶資料相關測試
go test ./tests/handlers -run TestGetProfile
```

### 詳細輸出測試
```bash
# 顯示詳細測試過程
go test -v ./tests/...

# 顯示測試覆蓋率
go test -cover ./tests/...

# 生成覆蓋率報告
go test -coverprofile=coverage.out ./tests/...
go tool cover -html=coverage.out
```

## 測試涵蓋範圍

### 認證測試 (auth_test.go)
- ✅ **註冊功能 (TestRegister)**
  - 成功註冊
  - Email 格式驗證
  - 密碼長度驗證
  - 用戶名長度驗證
  - 必要欄位驗證
  - 重複 Email 檢查
  - 重複用戶名檢查

- ✅ **登入功能 (TestLogin)**
  - 成功登入
  - 錯誤密碼處理
  - 不存在用戶處理
  - 無效請求格式處理
  - Token 生成驗證

- ✅ **Token 刷新功能 (TestRefreshToken)**
  - 成功刷新 Token
  - 無效 Refresh Token 處理
  - 空 Refresh Token 處理

- ✅ **登出功能 (TestLogout)**
  - 成功登出
  - 無效會話處理

### 用戶測試 (user_test.go)
- ✅ **獲取用戶資料 (TestGetProfile)**
  - 成功獲取用戶資料
  - 未認證用戶處理
  - 用戶不存在處理

- ✅ **更新用戶資料 (TestUpdateProfile)**
  - 成功更新完整資料
  - 部分欄位更新
  - 空更新請求
  - 無效用戶 ID 處理
  - 用戶不存在處理
  - 無效 JSON 請求處理

- ✅ **更新 FCM Token (TestUpdateFCMToken)**
  - 未實現功能檢查

## 測試設置

每個測試都使用 SQLite 記憶體資料庫進行隔離測試，確保：
- 測試之間不會互相影響
- 快速執行
- 不需要外部資料庫依賴

### 測試輔助函數
- `setupTestDB()` - 建立測試用的記憶體資料庫
- `setupTestConfig()` - 建立測試配置
- `createTestUser()` - 創建測試用戶

## 注意事項

1. **測試隔離**: 每個測試都使用獨立的記憶體資料庫
2. **Mock 中間件**: 使用 Fiber 中間件模擬認證狀態
3. **錯誤處理**: 充分測試各種錯誤情況
4. **繁體中文**: 所有錯誤訊息和測試名稱使用繁體中文

## 新增測試

在新增測試時，請遵循以下原則：

1. 使用描述性的測試名稱（繁體中文）
2. 包含正常情況和異常情況的測試
3. 使用表格驅動測試方式組織測試案例
4. 確保測試完全隔離，不依賴外部狀態
5. 添加適當的註釋說明測試目的 