# Swagger 註解快速指南

## 基本註解模板

### GET 端點 (查詢資料)
```go
// FunctionName 功能描述
// @Summary 簡短摘要
// @Description 詳細描述
// @Tags 標籤名稱
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "成功回應"
// @Failure 400 {object} map[string]interface{} "請求錯誤"
// @Failure 401 {object} map[string]interface{} "未授權"
// @Router /api/path [get]
func (h *Handler) FunctionName(c *fiber.Ctx) error {
```

### POST 端點 (創建資料)
```go
// CreateFunction 創建功能
// @Summary 創建資源
// @Description 創建新的資源
// @Tags 資源
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body RequestStruct true "請求資料"
// @Success 201 {object} map[string]interface{} "創建成功"
// @Failure 400 {object} map[string]interface{} "請求錯誤"
// @Router /api/path [post]
func (h *Handler) CreateFunction(c *fiber.Ctx) error {
```

### PUT 端點 (更新資料)
```go
// UpdateFunction 更新功能
// @Summary 更新資源
// @Description 更新指定資源
// @Tags 資源
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "資源 ID"
// @Param request body RequestStruct true "更新資料"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "請求錯誤"
// @Failure 404 {object} map[string]interface{} "資源不存在"
// @Router /api/path/{id} [put]
func (h *Handler) UpdateFunction(c *fiber.Ctx) error {
```

### DELETE 端點 (刪除資料)
```go
// DeleteFunction 刪除功能
// @Summary 刪除資源
// @Description 刪除指定資源
// @Tags 資源
// @Security BearerAuth
// @Param id path int true "資源 ID"
// @Success 200 {object} map[string]interface{} "刪除成功"
// @Failure 404 {object} map[string]interface{} "資源不存在"
// @Router /api/path/{id} [delete]
func (h *Handler) DeleteFunction(c *fiber.Ctx) error {
```

## 常用註解說明

| 註解 | 說明 | 範例 |
|------|------|------|
| `@Summary` | API 簡短摘要 | `@Summary 獲取用戶列表` |
| `@Description` | API 詳細描述 | `@Description 獲取所有活躍用戶的列表` |
| `@Tags` | API 分組標籤 | `@Tags 用戶管理` |
| `@Accept` | 接受的內容類型 | `@Accept json` |
| `@Produce` | 回應的內容類型 | `@Produce json` |
| `@Security` | 安全驗證方式 | `@Security BearerAuth` |
| `@Param` | 參數定義 | `@Param id path int true "用戶ID"` |
| `@Success` | 成功回應 | `@Success 200 {object} UserResponse` |
| `@Failure` | 錯誤回應 | `@Failure 404 {object} ErrorResponse` |
| `@Router` | 路由定義 | `@Router /users/{id} [get]` |

## 參數類型

### 路徑參數 (Path)
```go
// @Param id path int true "用戶 ID"
```

### 查詢參數 (Query)
```go
// @Param page query int false "頁碼" default(1)
// @Param limit query int false "每頁數量" default(10)
```

### 請求體參數 (Body)
```go
// @Param request body CreateUserRequest true "用戶資料"
```

### 表單參數 (Form)
```go
// @Param username formData string true "用戶名"
// @Param password formData string true "密碼"
```

### Header 參數
```go
// @Param Authorization header string true "Bearer token"
```

## 回應對象定義

### 使用結構體
```go
// @Success 200 {object} models.User "用戶資料"
// @Success 200 {array} models.User "用戶列表"
```

### 使用 map
```go
// @Success 200 {object} map[string]interface{} "成功回應"
```

### 內聯對象定義
```go
// @Param request body object{name=string,email=string} true "用戶資料"
```

## 重新生成文檔

每次修改註解後，需要重新生成文檔：

```bash
# 重新生成文檔
swag init -g cmd/api/main.go -o docs

# 或者設置 PATH 後
export PATH=$PATH:$(go env GOPATH)/bin
swag init -g cmd/api/main.go -o docs
```

## 最佳實踐

1. **一致性**: 使用統一的命名和描述風格
2. **完整性**: 為每個公開 API 添加完整註解
3. **準確性**: 確保註解與實際代碼一致
4. **中文支援**: 支援繁體中文描述
5. **版本控制**: 將生成的 docs 文件加入版本控制

## 下一步建議

1. 為所有 handler 添加 Swagger 註解
2. 定義標準的回應結構體
3. 添加更詳細的錯誤碼說明
4. 考慮使用 swag 的 `--parseInternal` 選項解析內部包
5. 設置 CI/CD 自動生成文檔