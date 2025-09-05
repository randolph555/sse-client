package internal

import (
	"strings"
)

// 提取AI响应中的命令
func extractCommands(aiResponse string) []string {
	var commands []string
	lines := strings.Split(aiResponse, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 跳过空行
		if line == "" {
			continue
		}

		// 跳过明显的解释文字（中文）
		if strings.Contains(line, "命令") || strings.Contains(line, "可以") ||
			strings.Contains(line, "这个") || strings.Contains(line, "使用") ||
			strings.Contains(line, "以下") || strings.Contains(line, "执行") ||
			strings.Contains(line, "运行") || strings.Contains(line, "查看") ||
			strings.Contains(line, "显示") || strings.Contains(line, "获取") {
			continue
		}

		// 跳过明显的解释文字（英文）
		lineLower := strings.ToLower(line)
		if strings.Contains(lineLower, "using") ||
			strings.Contains(lineLower, "command") ||
			strings.Contains(lineLower, "this will") ||
			strings.Contains(lineLower, "you can") ||
			strings.Contains(lineLower, "to view") ||
			strings.Contains(lineLower, "to check") ||
			strings.Contains(lineLower, "to show") ||
			strings.Contains(lineLower, "here is") ||
			strings.Contains(lineLower, "here are") {
			continue
		}

		// 跳过注释和代码块标记
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "```") ||
			strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") {
			continue
		}

		// 检查是否看起来像命令
		if isLikelyCommand(line) {
			commands = append(commands, line)
		}
	}

	return commands
}

// 判断是否像命令
func isLikelyCommand(line string) bool {
	lineLower := strings.ToLower(line)

	// 如果包含明显的非命令文字，直接排除
	excludeWords := []string{
		"以下是", "这是", "可以使用", "建议", "推荐", "注意", "提示",
		"here is", "this is", "you can use", "recommend", "suggest", "note", "tip",
	}

	for _, word := range excludeWords {
		if strings.Contains(lineLower, word) {
			return false
		}
	}

	// 常见的命令开头
	commandPrefixes := []string{
		"ls", "ps", "top", "htop", "df", "du", "free", "uptime", "w", "who", "whoami", "date",
		"cat", "grep", "find", "which", "whereis", "locate", "head", "tail", "less", "more",
		"docker", "kubectl", "systemctl", "service", "journalctl",
		"git", "npm", "yarn", "pip", "go", "make", "mvn", "gradle",
		"curl", "wget", "ping", "netstat", "ss", "nslookup", "dig",
		"rm", "cp", "mv", "mkdir", "chmod", "chown", "ln", "touch",
		"sudo", "su", "kill", "killall", "pkill", "pgrep",
		"awk", "sed", "sort", "uniq", "wc", "tr", "cut",
		"tar", "zip", "unzip", "gzip", "gunzip",
		"mount", "umount", "lsblk", "fdisk", "parted",
		"iptables", "ufw", "firewall-cmd",
		"crontab", "at", "nohup", "screen", "tmux",
		"history", "alias", "export", "env", "printenv",
	}

	// 检查是否以命令开头
	for _, prefix := range commandPrefixes {
		if strings.HasPrefix(lineLower, prefix+" ") || lineLower == prefix {
			return true
		}
	}

	// 检查是否是管道命令（包含 | 符号）
	if strings.Contains(line, "|") && !strings.Contains(lineLower, "或者") && !strings.Contains(lineLower, "or") {
		return true
	}

	// 检查是否是重定向命令（包含 > 或 >> 符号）
	if strings.Contains(line, ">") && !strings.Contains(lineLower, "大于") && !strings.Contains(lineLower, "greater") {
		return true
	}

	return false
}
