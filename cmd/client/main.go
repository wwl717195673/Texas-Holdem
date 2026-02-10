package main

import (
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	uiclient "github.com/wilenwang/just_play/Texas-Holdem/ui/client"
)

// main 程序入口
// 启动德州扑克 TUI 客户端
func main() {
	// 将 log 输出重定向到 /dev/null，防止日志破坏 Bubble Tea 的终端渲染
	// 客户端库（server/client）中的 log.Printf 会直接输出到 stderr，
	// 这些输出会混入 Bubble Tea 的 alt screen 导致画面错乱重叠
	logFile, err := os.OpenFile("/tmp/poker-client.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		// 打不开日志文件就直接丢弃
		log.SetOutput(io.Discard)
	} else {
		defer logFile.Close()
		log.SetOutput(logFile)
	}

	// 创建 TUI 模型
	model := uiclient.NewModel()

	// 创建 Bubble Tea 程序
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(), // 使用替代屏幕（全屏模式）
	)

	// 运行程序
	if _, err := p.Run(); err != nil {
		// 恢复 stderr 输出以显示致命错误
		log.SetOutput(os.Stderr)
		log.Fatalf("TUI 运行错误: %v", err)
	}
}
