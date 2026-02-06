package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/wilenwang/just_play/Texas-Holdem/pkg/game"
	"github.com/wilenwang/just_play/Texas-Holdem/server/host"
)

// 命令行参数
var port = flag.Int("port", 8080, "服务器端口")
var sb = flag.Int("sb", 10, "小盲注金额")
var bb = flag.Int("bb", 20, "大盲注金额")
var ante = flag.Int("ante", 0, "前注金额（0表示禁用）")
var chips = flag.Int("chips", 1000, "初始筹码")

func main() {
	flag.Parse()

	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║      德州扑克游戏服务器 v1.0           ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	// 创建游戏配置
	config := &game.Config{
		MinPlayers:    2,
		MaxPlayers:    9,
		SmallBlind:    *sb,
		BigBlind:      *bb,
		Ante:          *ante,
		StartingChips: *chips,
		ActionTimeout: 30,
	}

	// 创建完整的游戏服务器（包含消息处理和游戏引擎）
	server := host.NewServer(config)

	// 启动服务器主循环（处理注册、注销、消息路由、广播）
	go server.Run()

	// 注册 WebSocket 路由，使用 host.Server 的 ServeHTTP 处理连接
	http.Handle("/", server)

	fmt.Printf("游戏配置:\n")
	fmt.Printf("  盲注: %d/%d\n", *sb, *bb)
	fmt.Printf("  前注: %d\n", *ante)
	fmt.Printf("  初始筹码: %d\n", *chips)
	fmt.Printf("  服务器端口: %d\n", *port)
	fmt.Println()

	// 启动信号处理
	go handleSignals()

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("服务器启动成功!\n")
	fmt.Printf("连接地址: ws://localhost:%d\n", *port)
	fmt.Println()
	fmt.Println("等待玩家连接...")
	fmt.Println("按 Ctrl+C 停止服务器")
	fmt.Println()

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("服务器错误:", err)
	}
}

// handleSignals 处理系统信号，优雅关闭服务器
func handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n正在关闭服务器...")
	os.Exit(0)
}
