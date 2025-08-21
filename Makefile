# Ora2Pg-Admin Makefile
# 用于构建、测试和发布的自动化脚本

# 项目信息
PROJECT_NAME := ora2pg-admin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go 相关配置
GO := go
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOVERSION := $(shell go version | awk '{print $$3}')

# 构建配置
BUILD_DIR := build
DIST_DIR := dist
BINARY_NAME := ora2pg-admin
ifeq ($(GOOS),windows)
    BINARY_EXT := .exe
else
    BINARY_EXT :=
endif

# 编译标志
LDFLAGS := -X main.Version=$(VERSION) \
           -X main.BuildTime=$(BUILD_TIME) \
           -X main.GitCommit=$(GIT_COMMIT) \
           -X main.GoVersion=$(GOVERSION)

# 测试配置
TEST_TIMEOUT := 30m
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "Ora2Pg-Admin 构建工具"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 清理
.PHONY: clean
clean: ## 清理构建文件
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -f $(COVERAGE_FILE)
	@rm -f $(COVERAGE_HTML)
	@echo "清理完成"

# 创建构建目录
$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)

$(DIST_DIR):
	@mkdir -p $(DIST_DIR)

# 依赖管理
.PHONY: deps
deps: ## 下载依赖
	@echo "下载Go模块依赖..."
	@$(GO) mod download
	@$(GO) mod tidy
	@echo "依赖下载完成"

.PHONY: deps-update
deps-update: ## 更新依赖
	@echo "更新Go模块依赖..."
	@$(GO) get -u ./...
	@$(GO) mod tidy
	@echo "依赖更新完成"

# 代码检查
.PHONY: fmt
fmt: ## 格式化代码
	@echo "格式化Go代码..."
	@$(GO) fmt ./...
	@echo "代码格式化完成"

.PHONY: vet
vet: ## 静态分析
	@echo "执行Go vet检查..."
	@$(GO) vet ./...
	@echo "静态分析完成"

.PHONY: lint
lint: ## 代码规范检查
	@echo "执行golangci-lint检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint未安装，跳过检查"; \
	fi

# 测试
.PHONY: test
test: ## 运行单元测试
	@echo "运行单元测试..."
	@$(GO) test -v -timeout $(TEST_TIMEOUT) ./...
	@echo "单元测试完成"

.PHONY: test-short
test-short: ## 运行快速测试（跳过集成测试）
	@echo "运行快速测试..."
	@$(GO) test -v -short -timeout 5m ./...
	@echo "快速测试完成"

.PHONY: test-integration
test-integration: ## 运行集成测试
	@echo "运行集成测试..."
	@$(GO) test -v -timeout $(TEST_TIMEOUT) ./tests/...
	@echo "集成测试完成"

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "运行测试并生成覆盖率报告..."
	@$(GO) test -v -timeout $(TEST_TIMEOUT) -coverprofile=$(COVERAGE_FILE) ./...
	@$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "覆盖率报告已生成: $(COVERAGE_HTML)"

.PHONY: test-race
test-race: ## 运行竞态检测测试
	@echo "运行竞态检测测试..."
	@$(GO) test -v -race -timeout $(TEST_TIMEOUT) ./...
	@echo "竞态检测测试完成"

# 构建
.PHONY: build
build: $(BUILD_DIR) ## 构建当前平台的二进制文件
	@echo "构建 $(GOOS)/$(GOARCH) 版本..."
	@$(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT) .
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT)"

.PHONY: build-all
build-all: $(DIST_DIR) ## 构建所有平台的二进制文件
	@echo "构建所有平台版本..."
	
	@echo "构建 Linux amd64..."
	@GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
	
	@echo "构建 Linux arm64..."
	@GOOS=linux GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .
	
	@echo "构建 Windows amd64..."
	@GOOS=windows GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	
	@echo "构建 macOS amd64..."
	@GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	
	@echo "构建 macOS arm64..."
	@GOOS=darwin GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .
	
	@echo "所有平台构建完成"

.PHONY: install
install: ## 安装到系统路径
	@echo "安装到系统路径..."
	@$(GO) install -ldflags "$(LDFLAGS)" .
	@echo "安装完成"

# 发布
.PHONY: release
release: clean deps fmt vet test build-all ## 完整的发布流程
	@echo "创建发布包..."
	@cd $(DIST_DIR) && \
	for binary in *; do \
		if [[ "$$binary" == *".exe" ]]; then \
			zip "$${binary%.exe}.zip" "$$binary"; \
		else \
			tar -czf "$$binary.tar.gz" "$$binary"; \
		fi; \
	done
	@echo "发布包创建完成"

# 开发工具
.PHONY: dev
dev: deps fmt vet test-short build ## 开发模式构建
	@echo "开发构建完成"

.PHONY: watch
watch: ## 监视文件变化并自动构建
	@echo "开始监视文件变化..."
	@if command -v fswatch >/dev/null 2>&1; then \
		fswatch -o . | xargs -n1 -I{} make dev; \
	elif command -v inotifywait >/dev/null 2>&1; then \
		while inotifywait -r -e modify,create,delete .; do make dev; done; \
	else \
		echo "需要安装 fswatch 或 inotifywait 来监视文件变化"; \
	fi

.PHONY: run
run: build ## 构建并运行
	@echo "运行程序..."
	@$(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT) --help

# Docker 相关
.PHONY: docker-build
docker-build: ## 构建Docker镜像
	@echo "构建Docker镜像..."
	@docker build -t $(PROJECT_NAME):$(VERSION) .
	@docker tag $(PROJECT_NAME):$(VERSION) $(PROJECT_NAME):latest
	@echo "Docker镜像构建完成"

.PHONY: docker-run
docker-run: ## 运行Docker容器
	@echo "运行Docker容器..."
	@docker run --rm -it $(PROJECT_NAME):latest

# 文档
.PHONY: docs
docs: ## 生成文档
	@echo "生成文档..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "启动文档服务器: http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "请安装 godoc: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# 信息
.PHONY: info
info: ## 显示构建信息
	@echo "项目信息:"
	@echo "  名称: $(PROJECT_NAME)"
	@echo "  版本: $(VERSION)"
	@echo "  构建时间: $(BUILD_TIME)"
	@echo "  Git提交: $(GIT_COMMIT)"
	@echo "  Go版本: $(GOVERSION)"
	@echo "  目标平台: $(GOOS)/$(GOARCH)"

# 检查工具
.PHONY: check-tools
check-tools: ## 检查开发工具
	@echo "检查开发工具..."
	@echo -n "Go: "; $(GO) version
	@echo -n "Git: "; git --version 2>/dev/null || echo "未安装"
	@echo -n "Make: "; make --version | head -1
	@echo -n "golangci-lint: "; golangci-lint --version 2>/dev/null || echo "未安装"
	@echo -n "Docker: "; docker --version 2>/dev/null || echo "未安装"

# 性能测试
.PHONY: bench
bench: ## 运行性能测试
	@echo "运行性能测试..."
	@$(GO) test -bench=. -benchmem ./...
	@echo "性能测试完成"

# 安全检查
.PHONY: security
security: ## 运行安全检查
	@echo "运行安全检查..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec未安装，跳过安全检查"; \
		echo "安装命令: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi
