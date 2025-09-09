BINARY_NAME=sse
BUILD_DIR=build
DIST_DIR=dist
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-s -w -X main.version=$(VERSION)

.PHONY: build clean install deps test run build-all release

# 本地构建
build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/

# 跨平台构建
build-all:
	@echo "🚀 Building for multiple platforms..."
	@mkdir -p $(DIST_DIR)
	@echo "📦 Version: $(VERSION)"
	
	# Linux
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/
	
	# Windows
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/
	GOOS=windows GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-windows-arm64.exe ./cmd/
	
	# macOS
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/
	
	# FreeBSD
	GOOS=freebsd GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-freebsd-amd64 ./cmd/
	GOOS=freebsd GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME)-freebsd-arm64 ./cmd/
	
	@echo "✅ Cross-platform build completed!"
	@ls -la $(DIST_DIR)/

# 发布版本（压缩）
release: build-all
	@echo "📦 Creating release packages..."
	@echo "📋 Copying configuration files..."
	@cp -r configs $(DIST_DIR)/
	@cd $(DIST_DIR) && \
	for file in sse-*; do \
		if [[ $$file == *.exe ]]; then \
			zip -r "$${file%.exe}.zip" "$$file" configs/; \
		else \
			tar -czf "$$file.tar.gz" "$$file" configs/; \
		fi; \
	done
	@echo "🎉 Release packages created!"
	@ls -la $(DIST_DIR)/*.{zip,tar.gz} 2>/dev/null || true

# 安装到系统
install: build-all
	@echo "🚀 Installing SSE Client..."
	@./install.sh

# 本地安装（开发用）
install-local: build
	@echo "📦 Installing locally..."
	@mkdir -p $(HOME)/.local/bin
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(HOME)/.local/bin/
	@echo "✅ Installed to $(HOME)/.local/bin/$(BINARY_NAME)"
	@echo "💡 Make sure $(HOME)/.local/bin is in your PATH"

# 卸载
uninstall:
	@echo "🗑️  Uninstalling SSE Client..."
	@./uninstall.sh

deps:
	go mod tidy
	go mod download

clean:
	rm -rf $(BUILD_DIR) $(DIST_DIR)

test:
	go test ./...

run:
	go run . $(ARGS)

# 显示支持的平台
platforms:
	@echo "Supported platforms:"
	@go tool dist list | grep -E "(linux|windows|darwin|freebsd)/(amd64|arm64)"

.DEFAULT_GOAL := build