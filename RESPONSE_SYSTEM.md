# 🎯 Split Go Response 系統架構

這份文檔詳細說明 Split Go 專案中實作的 Response 系統，類似 Laravel Resource 的 Go 版本。

## 📁 **系統架構**

```
internal/
├── responses/
│   ├── common.go          # 通用回應結構
│   ├── user.go           # 用戶相關回應
│   ├── category.go       # 分類相關回應
│   ├── group.go          # 群組相關回應
│   └── transaction.go    # 交易相關回應
└── handlers/
    ├── transaction.go    # 使用 response 系統
    └── user.go          # 使用 response 系統
```

## 🏗️ **設計原則**

### **1. 類型安全**
- 所有回應結構都有明確的類型定義
- 使用 JSON tags 控制序列化行為
- 支援 `omitempty` 來處理可選欄位

### **2. 分層設計**
- **Simple Response**: 簡化版本，用於列表頁面
- **Full Response**: 完整版本，用於詳情頁面
- **Nested Response**: 用於嵌套在其他回應中

### **3. 權限整合**
- 計算欄位基於當前用戶
- 權限檢查欄位 (`can_edit`, `can_delete`)
- 個人化資料 (`my_amount`, `am_i_payer`)

## 🔧 **核心組件**

### **1. 通用回應結構 (`common.go`)**

```go
// 標準 API 回應
type APIResponse struct {
    Error   bool        `json:"error"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
}

// 分頁回應
type PaginatedResponse struct {
    Data       interface{}    `json:"data"`
    Pagination PaginationMeta `json:"pagination"`
}
```

### **2. Transform 函數模式**

```go
// 單一轉換
func NewTransactionResponse(tx models.Transaction, currentUserID uint) TransactionResponse

// 批量轉換
func NewTransactionResponseList(transactions []models.Transaction, currentUserID uint) []TransactionResponse

// 建構器模式
responses.SuccessResponse(data)
responses.ErrorResponse(message)
responses.SuccessWithMessageResponse(message, data)
```

## 📊 **Transaction Response 範例**

### **完整回應結構**
```go
type TransactionResponse struct {
    // 基本資訊
    ID          uint                       `json:"id"`
    Description string                     `json:"description"`
    Amount      float64                    `json:"amount"`
    Currency    string                     `json:"currency"`
    
    // 關聯資料
    Group       GroupSimpleResponse        `json:"group"`
    Category    *CategoryResponse          `json:"category,omitempty"`
    Payer       UserSimpleResponse         `json:"payer"`
    Creator     UserSimpleResponse         `json:"creator"`
    Splits      []TransactionSplitResponse `json:"splits"`
    
    // 計算欄位（基於當前用戶）
    MyAmount    float64 `json:"my_amount"`    // 我需要付的金額
    MyBalance   float64 `json:"my_balance"`   // 我的平衡狀況
    AmIPayer    bool    `json:"am_i_payer"`   // 我是否為付款者
    
    // 權限欄位
    CanEdit     bool    `json:"can_edit"`     // 是否可以編輯
    CanDelete   bool    `json:"can_delete"`   // 是否可以刪除
    
    // 時間戳記
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### **計算邏輯範例**
```go
// 計算當前用戶的應付金額
myAmount := 0.0
for _, split := range tx.Splits {
    if split.UserID == currentUserID {
        myAmount = split.Amount
        break
    }
}

// 計算平衡狀況
var myBalance float64
if tx.PaidBy == currentUserID {
    // 我是付款者：我付的錢 - 我應該付的錢
    myBalance = tx.Amount - myAmount
} else {
    // 我不是付款者：0 - 我應該付的錢 = 負數（我欠錢）
    myBalance = -myAmount
}
```

## 🎨 **使用方式**

### **在 Handler 中使用**

```go
func (h *TransactionHandler) GetTransaction(c *fiber.Ctx) error {
    // 1. 取得資料
    var transaction models.Transaction
    // ... 查詢邏輯
    
    // 2. 取得當前用戶
    user, err := middleware.GetCurrentUser(c, h.db)
    if err != nil {
        return err
    }
    
    // 3. 轉換為回應格式
    response := responses.NewTransactionResponse(transaction, user.UserID)
    
    // 4. 回傳
    return c.JSON(responses.SuccessResponse(response))
}
```

### **分頁回應**

```go
func (h *TransactionHandler) GetGroupTransactions(c *fiber.Ctx) error {
    // ... 查詢邏輯
    
    // 轉換為簡化回應格式
    transactionResponses := responses.NewTransactionSimpleResponseList(transactions, authUser.UserID)
    
    // 包裝為分頁回應
    paginatedResponse := responses.NewPaginatedResponse(transactionResponses, page, limit, total)
    
    return c.JSON(responses.SuccessResponse(paginatedResponse))
}
```

## 📤 **API 回應範例**

### **成功回應**
```json
{
  "error": false,
  "data": {
    "id": 123,
    "description": "聚餐費用",
    "amount": 1200.00,
    "currency": "TWD",
    "group": {
      "id": 1,
      "name": "室友分帳群"
    },
    "category": {
      "id": 1,
      "name": "餐飲",
      "icon": "🍽️",
      "color": "#FF6B6B"
    },
    "payer": {
      "id": 5,
      "name": "張愛莉絲",
      "avatar": "https://..."
    },
    "splits": [
      {
        "id": 456,
        "user": {
          "id": 5,
          "name": "張愛莉絲"
        },
        "amount": 400.00,
        "percentage": 33.33,
        "split_type": "equal"
      }
    ],
    "my_amount": 400.00,
    "my_balance": -400.00,
    "am_i_payer": false,
    "can_edit": false,
    "can_delete": false,
    "created_at": "2024-01-20T10:30:00Z"
  }
}
```

### **分頁回應**
```json
{
  "error": false,
  "data": {
    "data": [
      // ... 交易列表
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 150,
      "total_pages": 8
    }
  }
}
```

### **錯誤回應**
```json
{
  "error": true,
  "message": "交易不存在"
}
```

## ✨ **核心功能**

### **1. 資料隱藏**
- 自動排除敏感欄位（如密碼）
- 使用 `json:"-"` 標籤
- 只暴露需要的資料

### **2. 計算欄位**
- 基於當前用戶的個人化資料
- 動態計算權限狀況
- 前端友善的格式

### **3. 嵌套優化**
- Simple 版本用於嵌套
- 避免 N+1 查詢問題
- 減少資料傳輸量

### **4. 類型安全**
- 編譯時檢查
- IDE 自動完成
- 重構安全

## 🔄 **與 Laravel Resource 對比**

| 功能 | Laravel Resource | Go Response 系統 |
|------|-----------------|------------------|
| 資料轉換 | `toArray()` 方法 | Transform 函數 |
| 權限檢查 | `when()` 條件 | 計算欄位 |
| 嵌套資源 | Resource Collection | Nested Response |
| 分頁 | Resource Collection | PaginatedResponse |
| 錯誤處理 | Exception | ErrorResponse |

## 🚀 **效能優勢**

1. **編譯時優化**: 結構體預定義，無運行時反射
2. **記憶體效率**: 固定結構大小，GC 友善
3. **CPU 效率**: 直接欄位賦值，無動態解析
4. **網路效率**: 精確控制輸出格式

## 🛠️ **擴展指南**

### **添加新的 Response 類型**

1. 在 `internal/responses/` 創建新檔案
2. 定義 Response 結構體
3. 實作 Transform 函數
4. 在 Handler 中使用

### **添加計算欄位**

```go
type MyResponse struct {
    // ... 基本欄位
    
    // 計算欄位
    IsOwner     bool    `json:"is_owner"`
    Permission  string  `json:"permission"`
    CustomData  string  `json:"custom_data"`
}

func NewMyResponse(data models.MyModel, currentUserID uint) MyResponse {
    return MyResponse{
        // ... 基本賦值
        IsOwner:    data.CreatedBy == currentUserID,
        Permission: calculatePermission(data, currentUserID),
        CustomData: formatCustomData(data),
    }
}
```

## 💡 **最佳實踐**

1. **命名一致性**: 使用 `NewXxxResponse` 命名 Transform 函數
2. **批量轉換**: 提供 `NewXxxResponseList` 函數
3. **權限計算**: 在 Transform 中處理權限邏輯
4. **嵌套優化**: 使用 Simple 版本避免過度查詢
5. **錯誤處理**: 統一使用 ErrorResponse
6. **分頁支援**: 使用 PaginatedResponse 包裝列表資料

---

🎉 這個 Response 系統提供了類似 Laravel Resource 的功能，同時保持了 Go 的性能優勢和類型安全！ 