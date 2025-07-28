# 🌱 Split Go 測試資料 Seed 指南

這份指南說明如何使用 Split Go 專案的測試資料 seed 功能。

## 🚀 快速開始

### 1. 基本資料庫設置
```bash
# 執行資料庫遷移 + 基本分類
make migrate
```

### 2. 建立完整測試資料
```bash
# 建立用戶、群組、交易等完整測試資料
make migrate-seed
```

### 3. 一鍵開發環境設置
```bash
# 初始化 + 測試資料 + 啟動開發伺服器
make quick-start
```

## 📋 可用命令

### Migration 命令
```bash
make migrate              # 資料庫遷移 + 基本分類
make migrate-reset        # 重置資料庫 (刪除所有資料)
make migrate-seed     # 建立完整測試資料
```

### 開發環境命令
```bash
make setup-dev           # 初始化開發環境 (包含測試資料)
make reset-dev           # 重置開發環境 (包含測試資料)
make quick-start         # 一鍵啟動完整開發環境
```

### 直接使用 Go 命令
```bash
go run cmd/migrate/main.go -action=migrate     # 資料庫遷移
go run cmd/migrate/main.go -action=seed        # 基本分類
go run cmd/migrate/main.go -action=seed-all    # 完整測試資料
go run cmd/migrate/main.go -action=reset       # 重置資料庫
```

## 👥 測試用戶帳號

執行 `make migrate-seed-all` 後，您將獲得以下測試用戶：

| 姓名 | Email | 用戶名 | 密碼 |
|------|-------|--------|------|
| 張愛莉絲 | alice@example.com | alice | password123 |
| 李小明 | bob@example.com | bob | password123 |
| 王大華 | charlie@example.com | charlie | password123 |
| 陳美玲 | diana@example.com | diana | password123 |
| 林小雨 | eve@example.com | eve | password123 |

## 🏠 測試群組

### 1. 室友分帳群
- **成員**: Alice (管理員), Bob, Charlie
- **用途**: 生活費分攤
- **測試交易**: 
  - 便利商店買日用品 (450元, Alice付款)
  - 週末聚餐 (1200元, Bob付款)
  - 電費分攤 (2400元, Charlie付款)

### 2. 日本旅遊
- **成員**: Bob (管理員), Diana, Eve
- **用途**: 東京五日遊費用分攤

### 3. 公司聚餐
- **成員**: Alice (創建者)
- **用途**: 部門聚餐費用

## 📂 預設分類

系統會自動建立以下分類：

| 分類 | 圖示 | 顏色 |
|------|------|------|
| 餐飲 | 🍽️ | #FF6B6B |
| 交通 | 🚗 | #4ECDC4 |
| 住宿 | 🏠 | #45B7D1 |
| 娛樂 | 🎬 | #96CEB4 |
| 購物 | 🛍️ | #FFEAA7 |
| 醫療 | 🏥 | #DDA0DD |
| 教育 | 📚 | #98D8C8 |
| 其他 | 💡 | #F7DC6F |

## 💰 測試交易範例

### 室友分帳群的交易
1. **便利商店買日用品**
   - 金額: 450 元
   - 付款者: Alice
   - 分攤: 3人平均分攤 (每人150元)
   - 分類: 購物

2. **週末聚餐**
   - 金額: 1200 元
   - 付款者: Bob
   - 分攤: 3人平均分攤 (每人400元)
   - 分類: 餐飲

3. **電費分攤**
   - 金額: 2400 元
   - 付款者: Charlie
   - 分攤: 3人平均分攤 (每人800元)
   - 分類: 其他

## 🔧 開發建議

### 1. 初次設置
```bash
# 複製專案後首次設置
make setup-dev
```

### 2. 重置開發環境
```bash
# 當需要清理所有資料重新開始時
make reset-dev
```

### 3. 只需要基本分類
```bash
# 如果只需要分類，不需要測試用戶和交易
make migrate-seed
```

## 🧪 測試 API

使用 seed 資料後，您可以測試以下 API：

### 1. 登入測試
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "password123",
    "device_fingerprint": {"ip_address": "127.0.0.1"},
    "device_name": "Test Device",
    "device_type": "desktop"
  }'
```

### 2. 取得用戶交易
```bash
# 先登入取得 JWT token，然後：
curl -X GET http://localhost:8080/api/v1/transactions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 3. 取得群組交易
```bash
# 取得室友分帳群 (群組ID: 1) 的交易
curl -X GET http://localhost:8080/api/v1/groups/1/transactions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 💡 注意事項

1. **密碼安全**: 所有測試用戶的密碼都是 `password123`，僅供開發測試使用
2. **資料重複**: seed 函數會檢查資料是否已存在，避免重複建立
3. **關聯完整性**: 所有群組成員、交易分帳都有正確的關聯關係
4. **頭像設置**: 使用 UI Avatars 服務生成測試頭像

## 🐛 疑難排解

### 1. 權限錯誤
```bash
# 確保資料庫連接正常
make migrate
```

### 2. 依賴缺失
```bash
# 重新下載依賴
make mod-download
```

### 3. 資料不一致
```bash
# 完全重置後重新建立
make reset-dev
```

---

🎉 現在您已經有了完整的測試環境，可以開始開發和測試 Split Go 的分帳功能了！ 