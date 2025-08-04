# Split Go - 分帳記帳系統

一個使用 Go Fiber 框架開發的分帳記帳系統後端 API。

## ✨ 功能特色

- 🔐 用戶註冊/登入 (JWT 認證 + 設備管理)
- 👥 群組管理 (建立、加入、管理分帳群組)
- 💰 交易記錄 (新增、修改、刪除支出記錄)
- 📊 複雜分帳邏輯 (平均分、按比例分、固定金額分)
- ⚖️ 自動平衡計算與結算建議
- 🔔 Firebase 推播通知
- 📖 完整 Swagger API 文檔

## 🛠 技術棧

- **Go Fiber** - Web 框架
- **GORM** - ORM 資料庫操作
- **PostgreSQL** - 主要資料庫
- **JWT** - 用戶認證
- **Firebase** - 推播通知
- **Docker** - 容器化開發

## 🚀 快速開始

### 1. 克隆專案

```bash
git clone <repository-url>
cd split-go
```

### 2. 設置環境變數

創建 `.env` 文件：

```bash
# 資料庫設定
DATABASE_URL=postgres://postgres:postgres@db:5432/split_go_db?sslmode=disable

# 應用程式設定
APP_PORT=3000
JWT_SECRET=your_jwt_secret_key_here_please_change_this_in_production
APP_ENV=development

# Firebase 設定
FIREBASE_PROJECT_ID=your_firebase_project_id
FIREBASE_CREDENTIALS_PATH=./firebase-credentials.json
```

### 3. 一鍵啟動開發環境

```bash
# 初始化開發環境 (包含測試資料)
make setup-dev

# 啟動開發服務器 (熱重載)
make dev
```

## 📖 API 文檔

### Swagger 文檔

```bash
# 生成 API 文檔
make docs

# 啟動服務器
make run

# 訪問 Swagger UI
http://localhost:3000/swagger/index.html
```

### 主要 API 端點

| 分類 | 端點 | 說明 |
|------|------|------|
| **認證** | `POST /auth/register` | 用戶註冊 |
| | `POST /auth/login` | 用戶登入 |
| | `POST /auth/refresh` | 刷新令牌 |
| **用戶** | `GET /users/me` | 獲取個人資料 |
| | `PUT /users/me` | 更新個人資料 |
| **群組** | `GET /groups` | 獲取群組列表 |
| | `POST /groups` | 創建群組 |
| | `GET /groups/:id` | 獲取群組詳情 |
| **交易** | `GET /transactions` | 獲取交易列表 |
| | `POST /transactions` | 創建交易 |
| | `GET /groups/:id/balance` | 獲取群組平衡 |
| **結算** | `GET /settlements` | 獲取結算記錄 |
| | `POST /settlements` | 創建結算 |
| | `GET /groups/:id/settlement-suggestions` | 獲取結算建議 |

> 完整 API 文檔請查看 Swagger UI

## 🔧 開發指令

```bash
# 查看所有可用指令
make help

# 開發相關
make dev                    # 熱重載開發模式
make test                   # 運行測試
make build                  # 編譯應用

# 資料庫相關
make migrate                # 執行資料庫遷移
make migrate-seed           # 建立測試資料
make migrate-reset          # 重置資料庫

# 文檔相關
make docs                   # 生成 API 文檔
make docs-clean             # 清理文檔

# 環境管理
make setup-dev              # 初始化開發環境
make reset-dev              # 重置開發環境
make quick-start            # 一鍵啟動完整環境
```

## 🗂️ 專案結構

```
split-go/
├── cmd/api/                # 應用程式入口
├── internal/
│   ├── handlers/           # HTTP 處理器
│   ├── middleware/         # 中介軟體
│   ├── models/             # 資料模型
│   ├── services/           # 業務邏輯服務
│   ├── responses/          # API 回應格式
│   └── routes/             # 路由配置
├── tests/                  # 測試文件
├── docs/                   # API 文檔
├── .devcontainer/          # 開發容器配置
└── Makefile               # 開發指令
```

## 🔍 資料庫管理

pgAdmin 已配置在開發環境：

- **URL**: http://localhost:5050
- **帳號**: admin@example.com / admin
- **資料庫連接**: db:5432 / postgres / postgres

## ⚠️ 生產環境注意事項

1. **JWT Secret**: 使用強密碼替換 `JWT_SECRET`
2. **資料庫密碼**: 修改預設的資料庫密碼
3. **Firebase**: 配置正式的 Firebase 專案憑證
4. **HTTPS**: 生產環境請使用 HTTPS

## 🤝 貢獻

歡迎提交 Issue 和 Pull Request！

---

更多詳細資訊請參考專案內的文檔或 Swagger API 文檔。