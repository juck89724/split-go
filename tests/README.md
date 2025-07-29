# 測試說明

這個資料夾包含了 Split Go 專案的所有測試文件。

## 測試結構

```
tests/
├── handlers/
│   ├── auth_test.go         # 認證相關測試
│   ├── user_test.go         # 用戶相關測試
│   ├── group_test.go        # 群組相關測試
│   └── transaction_test.go  # 交易相關測試
└── README.md                # 測試說明文件
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

# 執行群組測試
go test ./tests/handlers/group_test.go

# 執行交易測試
go test ./tests/handlers/transaction_test.go
```

### 執行特定測試函數
```bash
# 執行註冊功能測試
go test ./tests/handlers -run TestRegister

# 執行登入功能測試
go test ./tests/handlers -run TestLogin

# 執行用戶資料相關測試
go test ./tests/handlers -run TestGetProfile

# 執行群組相關測試
go test ./tests/handlers -run TestGetUserGroups
go test ./tests/handlers -run TestCreateGroup
go test ./tests/handlers -run TestGetGroup
go test ./tests/handlers -run TestUpdateGroup
go test ./tests/handlers -run TestDeleteGroup
go test ./tests/handlers -run TestAddMember
go test ./tests/handlers -run TestRemoveMember
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

### 群組測試 (group_test.go)
- ✅ **獲取用戶群組列表 (TestGetUserGroups)**
  - 成功獲取群組列表
  - 未認證用戶處理
  - 多群組成員關係驗證

- ✅ **創建群組 (TestCreateGroup)**
  - 成功創建群組
  - 群組名稱為空驗證
  - 無效JSON格式處理
  - 創建者自動設為管理員驗證

- ✅ **獲取群組詳情 (TestGetGroup)**
  - 群組成員成功獲取詳情
  - 非群組成員權限檢查
  - 群組不存在處理
  - 無效群組ID處理
  - 成員列表和權限驗證

- ✅ **更新群組 (TestUpdateGroup)**
  - 管理員成功更新群組
  - 一般成員權限限制
  - 空群組名稱驗證
  - 資料庫更新確認

- ✅ **刪除群組 (TestDeleteGroup)**
  - 創建者成功刪除群組
  - 管理員權限限制（無法刪除）
  - 不存在群組處理
  - 軟刪除和成員清理驗證

- ✅ **添加群組成員 (TestAddMember)**
  - 管理員成功添加成員
  - 一般成員權限限制
  - 重複成員檢查
  - 不存在用戶處理
  - 角色設定驗證

- ✅ **移除群組成員 (TestRemoveMember)**
  - 管理員成功移除成員
  - 一般成員權限限制
  - 無法移除自己
  - 無法移除群組創建者
  - 不存在成員處理
  - 無效用戶ID處理

## 測試設置

每個測試都使用 SQLite 記憶體資料庫進行隔離測試，確保：
- 測試之間不會互相影響
- 快速執行
- 不需要外部資料庫依賴

### 測試輔助函數
- `setupTestDB()` - 建立測試用的記憶體資料庫
- `setupGroupTestDB()` - 建立群組測試專用資料庫（含群組表）
- `setupTestConfig()` - 建立測試配置
- `createTestUser()` - 創建測試用戶
- `createTestGroup()` - 創建測試群組
- `addGroupMember()` - 添加群組成員

## 注意事項

1. **測試隔離**: 每個測試都使用獨立的記憶體資料庫
2. **Mock 中間件**: 使用 Fiber 中間件模擬認證狀態
3. **錯誤處理**: 充分測試各種錯誤情況
4. **權限驗證**: 詳細測試各種權限組合
5. **繁體中文**: 所有錯誤訊息和測試名稱使用繁體中文

## 群組測試特色

群組測試涵蓋了完整的群組生命週期：
- **權限層級**: 測試創建者、管理員、一般成員的不同權限
- **資料完整性**: 驗證群組和成員資料的一致性
- **邊界條件**: 測試各種異常輸入和邊界情況
- **關聯清理**: 確認刪除操作正確清理相關資料

## 新增測試

在新增測試時，請遵循以下原則：

1. 使用描述性的測試名稱（繁體中文）
2. 包含正常情況和異常情況的測試
3. 使用表格驅動測試方式組織測試案例
4. 確保測試完全隔離，不依賴外部狀態
5. 添加適當的註釋說明測試目的
6. 驗證資料庫狀態變化
7. 測試權限控制邏輯 