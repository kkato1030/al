.PHONY: build install clean test fmt vet run help

# 変数定義
BINARY_NAME=al
CMD_PATH=./cmd/al
BUILD_DIR=./bin
VERSION?=0.1.0
BUILD_TIME=$(shell date +%Y-%m-%dT%H:%M:%S)
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# デフォルトターゲット
.DEFAULT_GOAL := help

# ビルド
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# インストール
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)" \
		$(CMD_PATH)
	@echo "Installation complete"

# クリーンアップ
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "Clean complete"

# テスト実行
test:
	@echo "Running tests..."
	@go test -v ./...

# テストカバレッジ
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# フォーマット
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete"

# 静的解析
vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "Vet complete"

# リンター（golangci-lint がインストールされている場合）
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint is not installed. Skipping..."; \
	fi

# 実行
run:
	@go run $(CMD_PATH) $(ARGS)

# 開発用ビルド（デバッグ情報付き）
build-dev: fmt vet
	@echo "Building $(BINARY_NAME) (dev mode)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# リリース用ビルド（最適化）
build-release: fmt vet test
	@echo "Building $(BINARY_NAME) (release mode)..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# クロスコンパイル（macOS用）
build-darwin:
	@echo "Building $(BINARY_NAME) for darwin/amd64..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)" \
		-o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_PATH)
	@echo "Building $(BINARY_NAME) for darwin/arm64..."
	@GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)" \
		-o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_PATH)
	@echo "Build complete"

# ヘルプ
help:
	@echo "Available targets:"
	@echo "  make build          - Build the binary"
	@echo "  make install        - Install the binary to GOPATH/bin"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make fmt            - Format code"
	@echo "  make vet            - Run go vet"
	@echo "  make lint           - Run golangci-lint (if installed)"
	@echo "  make run ARGS=...   - Run the application"
	@echo "  make build-dev      - Build in development mode"
	@echo "  make build-release  - Build optimized release binary"
	@echo "  make build-darwin   - Build for macOS (amd64 and arm64)"
	@echo "  make help           - Show this help message"
