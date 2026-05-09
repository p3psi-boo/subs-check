# 使用 bash 作为默认 shell
SHELL=/usr/bin/env bash

# 定义变量
BINARY := subs-check
COMMIT := $(shell git rev-parse --short HEAD)
COMMIT_TIMESTAMP := $(shell git log -1 --format=%ct)
VERSION := $(shell git describe --tags --abbrev=0)
GO_BIN := go

# 构建标志
CGO_ENABLED := 0
FLAGS := -trimpath
LDFLAGS := -s -w -X main.Version=$(VERSION) -X main.CurrentCommit=$(COMMIT)

# 声明伪目标
.PHONY: all build run gotool clean help \
		linux-amd64 linux-arm64 linux-arm linux-386 \
		build-all

# 默认目标：整理代码并编译当前环境
all: build

# 默认构建：当前环境
build:
	$(GO_BIN) generate ./...
	$(GO_BIN) build -o $(BINARY) $(FLAGS) -ldflags "$(LDFLAGS)"

# 清理
clean:
	@if [ -f $(BINARY) ]; then rm -f $(BINARY); fi
	@rm -rf build/

# Linux 平台 (4 个)
linux-amd64:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 $(GO_BIN) build -o $(BINARY)_linux_amd64 $(FLAGS) -ldflags "$(LDFLAGS)"

linux-arm64:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm64 $(GO_BIN) build -o $(BINARY)_linux_arm64 $(FLAGS) -ldflags "$(LDFLAGS)"

linux-arm:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm GOARM=7 $(GO_BIN) build -o $(BINARY)_linux_armv7 $(FLAGS) -ldflags "$(LDFLAGS)"

linux-386:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=386 $(GO_BIN) build -o $(BINARY)_linux_386 $(FLAGS) -ldflags "$(LDFLAGS)"

# 构建所有 Linux 平台
build-all:
	@mkdir -p build
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 $(GO_BIN) build -o build/$(BINARY)_linux_amd64 $(FLAGS) -ldflags "$(LDFLAGS)"; \
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm64 $(GO_BIN) build -o build/$(BINARY)_linux_arm64 $(FLAGS) -ldflags "$(LDFLAGS)"; \
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm GOARM=7 $(GO_BIN) build -o build/$(BINARY)_linux_armv7 $(FLAGS) -ldflags "$(LDFLAGS)"; \
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=386 $(GO_BIN) build -o build/$(BINARY)_linux_386 $(FLAGS) -ldflags "$(LDFLAGS)"

# 帮助信息
help:
# 	@chcp 65001 >nul
	@echo "make              - 整理 Go 代码并编译当前环境的二进制文件"
	@echo "make build        - 编译当前环境的二进制文件"
	@echo "make run          - 直接运行 Go 代码"
	@echo "make gotool       - 运行 Go 工具 'mod tidy', 'fmt' 和 'vet'"
	@echo "make clean        - 移除二进制文件和构建目录"
	@echo "make linux-amd64  - 编译 Linux/amd64 二进制文件"
	@echo "make linux-arm64  - 编译 Linux/arm64 二进制文件"
	@echo "make linux-arm    - 编译 Linux/armv7 二进制文件"
	@echo "make linux-386    - 编译 Linux/386 二进制文件"
	@echo "make build-all    - 编译所有 Linux 平台的二进制文件到 build/ 目录"
	@echo "make help         - 显示此帮助信息"
