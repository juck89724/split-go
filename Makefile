# Go 相關變數
GO=go
MIGRATE_CMD=cmd/migrate/main.go
API_CMD=cmd/api/main.go

# 資料庫相關
DB_URL ?= postgres://postgres:postgres@db:5432/split_go_db?sslmode=disable

# 顏色輸出
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: help build run dev migrate migrate-reset migrate-seed clean test

# 預設目標
help: ## 顯示幫助信息
	@echo "可用指令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

# 建置相關
build: ## 編譯應用程序
	@echo "$(YELLOW)🔨 編譯應用程序...$(NC)"
	$(GO) build -o bin/api $(API_CMD)
	$(GO) build -o bin/migrate $(MIGRATE_CMD)
	@echo "$(GREEN)✅ 編譯完成$(NC)"

clean: ## 清理編譯檔案
	@echo "$(YELLOW)🧹 清理編譯檔案...$(NC)"
	rm -rf bin/
	@echo "$(GREEN)✅ 清理完成$(NC)"

# 運行相關
run: ## 運行 API 服務器
	@echo "$(YELLOW)🚀 啟動 API 服務器...$(NC)"
	$(GO) run $(API_CMD)

dev: ## 使用 air 運行開發模式 (熱重載)
	@echo "$(YELLOW)🔥 啟動開發模式 (熱重載)...$(NC)"
	air

# Migration 相關
migrate: ## 執行資料庫遷移
	@echo "$(YELLOW)📊 執行資料庫遷移...$(NC)"
	$(GO) run $(MIGRATE_CMD) -action=migrate
	@echo "$(GREEN)✅ 資料庫遷移完成$(NC)"

migrate-reset: ## 重置資料庫 (刪除所有資料)
	@echo "$(RED)⚠️  警告: 這將刪除所有資料!$(NC)"
	@echo "$(YELLOW)🔄 重置資料庫...$(NC)"
	$(GO) run $(MIGRATE_CMD) -action=reset
	@echo "$(GREEN)✅ 資料庫重置完成$(NC)"

migrate-seed: ## 建立完整測試資料 (用戶、群組、交易)
	@echo "$(YELLOW)🌱 建立完整測試資料...$(NC)"
	$(GO) run $(MIGRATE_CMD) -action=seed
	@echo "$(GREEN)✅ 完整測試資料建立完成$(NC)"

migrate-custom: ## 使用自定義資料庫 URL 進行遷移 (使用: make migrate-custom DB_URL="your_url")
	@echo "$(YELLOW)📊 使用自定義 URL 執行遷移...$(NC)"
	$(GO) run $(MIGRATE_CMD) -action=migrate -db="$(DB_URL)"
	@echo "$(GREEN)✅ 遷移完成$(NC)"

# 測試相關
test: ## 運行測試
	@echo "$(YELLOW)🧪 運行測試...$(NC)"
	$(GO) test ./...
	@echo "$(GREEN)✅ 測試完成$(NC)"

test-verbose: ## 運行詳細測試
	@echo "$(YELLOW)🧪 運行詳細測試...$(NC)"
	$(GO) test -v ./...
	@echo "$(GREEN)✅ 詳細測試完成$(NC)"

# 開發工具
mod-tidy: ## 整理 Go modules
	@echo "$(YELLOW)📦 整理 Go modules...$(NC)"
	$(GO) mod tidy
	@echo "$(GREEN)✅ Modules 整理完成$(NC)"

mod-download: ## 下載依賴包
	@echo "$(YELLOW)📥 下載依賴包...$(NC)"
	$(GO) mod download
	@echo "$(GREEN)✅ 依賴包下載完成$(NC)"

# 部署相關
build-prod: ## 編譯生產版本
	@echo "$(YELLOW)🏭 編譯生產版本...$(NC)"
	CGO_ENABLED=0 GOOS=linux $(GO) build -a -installsuffix cgo -o bin/api-prod $(API_CMD)
	CGO_ENABLED=0 GOOS=linux $(GO) build -a -installsuffix cgo -o bin/migrate-prod $(MIGRATE_CMD)
	@echo "$(GREEN)✅ 生產版本編譯完成$(NC)"

# Docker 相關 (如果需要)
docker-build: ## 構建 Docker 映像
	@echo "$(YELLOW)🐳 構建 Docker 映像...$(NC)"
	docker build -t split-go-api .
	@echo "$(GREEN)✅ Docker 映像構建完成$(NC)"

# 一鍵設置
setup: mod-download migrate ## 初始化項目 (下載依賴 + 資料庫遷移)
	@echo "$(GREEN)🎉 項目設置完成!$(NC)"

setup-dev: mod-download migrate migrate-seed ## 初始化開發環境 (包含測試資料)
	@echo "$(GREEN)🎉 開發環境設置完成! 包含完整測試資料$(NC)"

# 重置開發環境
reset-dev: clean migrate-reset migrate-seed ## 重置整個開發環境 (包含測試資料)
	@echo "$(GREEN)🔄 開發環境重置完成! 包含測試資料$(NC)"

# 快速開發環境
quick-start: setup-dev dev ## 一鍵啟動完整開發環境 