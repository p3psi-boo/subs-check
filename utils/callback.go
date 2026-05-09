package utils

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sinspired/subs-check/config"
)

// ExecuteCallback 执行回调脚本
func ExecuteCallback(successCount int) {
	callbackScript := config.GlobalConfig.CallbackScript
	if callbackScript == "" {
		return
	}

	slog.Info(fmt.Sprintf("执行回调脚本: %s", callbackScript))

	// 检查脚本文件是否存在
	if _, err := os.Stat(callbackScript); os.IsNotExist(err) {
		slog.Error(fmt.Sprintf("回调脚本不存在: %s", callbackScript))
		return
	}

	if err := os.Chmod(callbackScript, 0755); err != nil {
		slog.Warn(fmt.Sprintf("设置脚本执行权限失败: %v", err))
	}

	// 检查脚本是否有shebang
	content, err := os.ReadFile(callbackScript)
	if err == nil && len(content) > 0 {
		hasShebang := len(content) >= 2 && content[0] == '#' && content[1] == '!'
		if !hasShebang {
			slog.Warn("脚本缺少shebang行，请在脚本开头添加对应的：#!/bin/bash、#!/bin/sh、#!/usr/bin/env bash 等")
		}
	}

	absPath, err := filepath.Abs(callbackScript)
	if err != nil {
		slog.Error(fmt.Sprintf("获取脚本绝对路径失败: %v", err))
		return
	}

	cmd := exec.Command(absPath)
	cmd.Dir = filepath.Dir(absPath)

	// 设置环境变量，传递成功节点数量
	cmd.Env = append(os.Environ(), fmt.Sprintf("SUCCESS_COUNT=%d", successCount))

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error(fmt.Sprintf("执行回调脚本失败: %v, 输出: %s", err, string(output)))
		return
	}
	slog.Info("回调脚本执行成功")
}
