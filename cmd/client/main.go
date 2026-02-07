package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	uiclient "github.com/wilenwang/just_play/Texas-Holdem/ui/client"
)

// main 程序入口
// 启动德州扑克 TUI 客户端
func main() {
	// 创建 TUI 模型
	model := uiclient.NewModel()

	// 创建 Bubble Tea 程序
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(), // 使用替代屏幕（全屏模式）
	)

	// 运行程序
	if _, err := p.Run(); err != nil {
		log.Fatalf("TUI 运行错误: %v", err)
	}
}
