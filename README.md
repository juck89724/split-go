# Split Go - 分帳記帳系統

一個使用 Go Fiber 框架開發的分帳記帳系統後端 API。

## 功能特色

- 🔐 用戶註冊/登入 (JWT 認證)
- 👥 群組管理 (建立、加入、管理分帳群組)
- 💰 交易記錄 (新增、修改、刪除支出記錄)
- 📊 複雜分帳邏輯 (平均分、按比例分、固定金額分)
- ⚖️ 自動平衡計算 (計算每個人應付/應收金額)
- 🔔 Firebase 推播通知
- 📱 為 Flutter 前端提供完整 API

## 技術棧

- **Go Fiber** - Web 框架
- **GORM** - ORM 資料庫操作
- **PostgreSQL** - 主要資料庫
- **JWT** - 用戶認證
- **Firebase** - 推播通知
- **Docker** - 容器化開發

## 快速開始

### 1. 複製專案

```bash
git clone <repository-url>
cd split-go
```

### 2. 設置環境變數

創建 `.env` 文件：

```bash
# 資料庫設定
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=split_go_db
POSTGRES_HOST=db
POSTGRES_PORT=5432
DATABASE_URL=postgres://postgres:postgres@db:5432/split_go_db?sslmode=disable

# 應用程式設定
APP_PORT=3000
JWT_SECRET=your_jwt_secret_key_here_please_change_this_in_production
APP_ENV=development

# Firebase 設定 (推播通知用)
FIREBASE_PROJECT_ID=your_firebase_project_id
FIREBASE_CREDENTIALS_PATH=./firebase-credentials.json

# pgAdmin 設定
PGADMIN_DEFAULT_EMAIL=admin@example.com
PGADMIN_DEFAULT_PASSWORD=admin
```

### 3. 開發容器啟動

如果使用 VS Code Dev Container：

1. 在 VS Code 中打開專案
2. 按 `Ctrl+Shift+P` (Windows/Linux) 或 `Cmd+Shift+P` (Mac)
3. 選擇 "Dev Containers: Reopen in Container"

或者使用 Docker Compose：

```bash
docker-compose -f .devcontainer/docker-compose.yml up -d
```

### 4. 安裝依賴

```bash
go mod tidy
```

### 5. 運行應用程式

```bash
go run cmd/api/main.go
```

或使用 Air 進行熱重載：

```bash
air
```

## API 文檔

### 認證相關

- `POST /api/v1/auth/register` - 用戶註冊
- `POST /api/v1/auth/login` - 用戶登入
- `POST /api/v1/auth/refresh` - 刷新 Token

### 用戶相關

- `GET /api/v1/users/me` - 獲取個人資料
- `PUT /api/v1/users/me` - 更新個人資料
- `POST /api/v1/users/fcm-token` - 更新推播 Token

### 群組相關

- `GET /api/v1/groups` - 獲取用戶群組列表
- `POST /api/v1/groups` - 創建新群組
- `GET /api/v1/groups/:id` - 獲取群組詳情
- `PUT /api/v1/groups/:id` - 更新群組資訊
- `DELETE /api/v1/groups/:id` - 刪除群組
- `POST /api/v1/groups/:id/members` - 添加群組成員
- `DELETE /api/v1/groups/:id/members/:userId` - 移除群組成員

### 交易相關

- `GET /api/v1/transactions` - 獲取交易列表
- `POST /api/v1/transactions` - 創建新交易
- `GET /api/v1/transactions/:id` - 獲取交易詳情
- `PUT /api/v1/transactions/:id` - 更新交易
- `DELETE /api/v1/transactions/:id` - 刪除交易
- `GET /api/v1/groups/:id/transactions` - 獲取群組交易
- `GET /api/v1/groups/:id/balance` - 獲取群組平衡

### 結算相關

- `GET /api/v1/settlements` - 獲取結算記錄
- `POST /api/v1/settlements` - 創建結算
- `PUT /api/v1/settlements/:id/paid` - 標記已付款
- `DELETE /api/v1/settlements/:id` - 取消結算
- `GET /api/v1/groups/:id/settlement-suggestions` - 獲取結算建議

## 資料庫管理

pgAdmin 已經配置在開發環境中：

- URL: http://localhost:5050
- Email: admin@example.com
- Password: admin

連接 PostgreSQL：
- Host: db
- Port: 5432
- Database: split_go_db
- Username: postgres
- Password: postgres

## 專案結構

```
split-go/
├── cmd/api/                # 應用程式入口
├── internal/
│   ├── config/             # 配置管理
│   ├── database/           # 資料庫連接與遷移
│   ├── handlers/           # HTTP 處理器
│   ├── middleware/         # 中介軟體
│   ├── models/             # 資料模型
│   ├── routes/             # 路由配置
│   └── utils/              # 工具函數
├── .devcontainer/          # 開發容器配置
├── go.mod                  # Go 模組文件
└── README.md
```

## 開發注意事項

1. **JWT Secret**: 請在生產環境中使用強密碼
2. **資料庫密碼**: 請在生產環境中修改預設密碼
3. **Firebase**: 需要配置 Firebase 項目並下載憑證文件
4. **埠口**: 確保 3000, 5432, 5050 埠口未被占用

## 後續開發

- [ ] 完善所有 API 端點
- [ ] 添加單元測試
- [ ] 實現複雜分帳算法
- [ ] Firebase 推播通知整合
- [ ] API 文檔自動生成
- [ ] 部署配置

## 貢獻

歡迎提交 Issue 和 Pull Request！
