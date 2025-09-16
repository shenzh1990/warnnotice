package util

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ScriptConfig 脚本配置结构
type ScriptConfig struct {
	Path       string `json:"path"`
	Parameters string `json:"parameters"`
	Timeout    int    `json:"timeout"`  // 超时时间(秒)
	Interval   int    `json:"interval"` // 执行间隔(分钟)
}

// ExecuteScript 执行脚本
func ExecuteScript(config ScriptConfig) (int, string, error) {
	// 检查脚本路径是否为空
	if config.Path == "" {
		return -1, "", fmt.Errorf("脚本路径不能为空")
	}

	// 检查脚本文件是否存在
	absPath, err := filepath.Abs(config.Path)
	if err != nil {
		return -1, "", fmt.Errorf("获取脚本绝对路径失败: %v", err)
	}

	// 构造命令
	var cmd *exec.Cmd
	if config.Parameters != "" {
		params := strings.Fields(config.Parameters)
		args := make([]string, 0, len(params)+1)
		args = append(args, absPath) // 使用绝对路径而不是原始路径
		args = append(args, params...)
		cmd = exec.Command(args[0], args[1:]...)
	} else {
		cmd = exec.Command(absPath)
	}

	// 设置超时
	if config.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
		defer cancel()
		cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
	}

	// 执行命令
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return -1, outputStr, fmt.Errorf("执行脚本失败: %v, 输出: %s", err, outputStr)
	}

	// 尝试将输出解析为整数
	result := strings.TrimSpace(outputStr)

	// 尝试转换为整数
	var resultInt int
	_, err = fmt.Sscanf(result, "%d", &resultInt)
	if err != nil {
		return -1, outputStr, fmt.Errorf("脚本返回值不是有效整数: %s", result)
	}

	return resultInt, outputStr, nil
}
