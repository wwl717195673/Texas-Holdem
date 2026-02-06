package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/card"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/common/models"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/protocol"
)

// 命令行参数
var server = flag.String("server", "localhost:8080", "服务器地址")
var playerName = flag.String("name", "Player", "玩家名称")

// WebSocket 连接
var conn *websocket.Conn
var done chan struct{}
var playerID string

// 扑克牌花色和点数映射
var suitSymbols = map[card.Suit]string{
	card.Hearts:   "♥",
	card.Diamonds: "♦",
	card.Clubs:    "♣",
	card.Spades:   "♠",
}

var rankSymbols = []string{"", "A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}

func main() {
	flag.Parse()

	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║      德州扑克游戏客户端 v1.0           ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	// 连接服务器
	fmt.Printf("正在连接到服务器: ws://%s\n", *server)
	var err error
	conn, _, err = websocket.DefaultDialer.Dial("ws://"+*server, nil)
	if err != nil {
		log.Fatal("连接服务器失败:", err)
	}
	defer conn.Close()

	fmt.Println("连接成功!")
	fmt.Println()

	// 设置中断信号处理
	done = make(chan struct{})
	go handleInterrupt()

	// 读取玩家输入
	reader := bufio.NewReader(os.Stdin)

	// 主菜单循环
	menuLoop(reader)
}

// handleInterrupt 处理中断信号
func handleInterrupt() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n正在断开连接...")
	conn.Close()
	os.Exit(0)
}

// menuLoop 主菜单循环
func menuLoop(reader *bufio.Reader) {
	for {
		printMenu()
		fmt.Print("请选择: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			joinGame(reader)
		case "2":
			showHelp()
		case "3":
			fmt.Println("再见!")
			os.Exit(0)
		default:
			fmt.Println("无效选择，请重试")
		}
		fmt.Println()
	}
}

// printMenu 打印菜单
func printMenu() {
	fmt.Println("┌─────────────────────────────────┐")
	fmt.Println("│         主菜单                  │")
	fmt.Println("├─────────────────────────────────┤")
	fmt.Println("│  1. 加入游戏                    │")
	fmt.Println("│  2. 帮助                        │")
	fmt.Println("│  3. 退出                        │")
	fmt.Println("└─────────────────────────────────┘")
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║            帮助信息                 ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("操作指令:")
	fmt.Println("  F / fold    - 弃牌")
	fmt.Println("  X / check   - 看牌")
	fmt.Println("  C / call    - 跟注")
	fmt.Println("  R / raise   - 加注 (格式: R 100)")
	fmt.Println("  A / allin   - 全下")
	fmt.Println("  S / status  - 查看当前状态")
	fmt.Println("  Q / quit    - 退出游戏")
	fmt.Println()
	fmt.Println("按回车键刷新游戏状态...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

// joinGame 加入游戏
func joinGame(reader *bufio.Reader) {
	// 发送加入请求
	joinReq := protocol.NewJoinRequest(*playerName, -1)
	sendMessage(joinReq)

	// 等待加入确认
	for {
		select {
		case <-done:
			return
		default:
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("读取消息失败:", err)
			return
		}

		var msg protocol.BaseMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("解析消息失败:", err)
			continue
		}

		switch msg.Type {
		case protocol.MsgTypeJoinAck:
			var ack protocol.JoinAck
			json.Unmarshal(message, &ack)
			if ack.Success {
				playerID = ack.PlayerID
				fmt.Printf("\n加入游戏成功! 你的玩家ID: %s\n", playerID)
				fmt.Printf("座位号: %d\n", ack.Seat)
				fmt.Println()
				gameLoop(reader)
			} else {
				fmt.Println("加入游戏失败:", ack.Message)
			}
			return

		case protocol.MsgTypeError:
			var errMsg protocol.Error
			json.Unmarshal(message, &errMsg)
			fmt.Println("错误:", errMsg.Message)
			return
		}
	}
}

// gameLoop 游戏主循环（使用独立协程分别处理服务器消息和用户输入）
func gameLoop(reader *bufio.Reader) {
	// 服务器消息通道
	msgChan := make(chan []byte, 256)
	go receiveMessages(msgChan)

	// 用户输入通道
	inputChan := make(chan string, 16)
	go readUserInput(reader, inputChan)

	// 显示初始提示
	fmt.Print("游戏> ")

	for {
		select {
		case <-done:
			return

		case message, ok := <-msgChan:
			if !ok {
				fmt.Println("\n与服务器的连接已断开")
				return
			}
			// 处理服务器消息，处理完后重新显示提示符
			handleGameMessage(message)
			fmt.Print("游戏> ")

		case text, ok := <-inputChan:
			if !ok {
				return
			}
			// 处理用户输入
			handleUserInput(text)
			fmt.Print("游戏> ")
		}
	}
}

// readUserInput 在独立协程中持续读取用户输入
func readUserInput(reader *bufio.Reader, inputChan chan<- string) {
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			close(inputChan)
			return
		}
		text = strings.TrimSpace(text)
		// 忽略空输入，不发送到通道
		if text != "" {
			inputChan <- text
		}
	}
}

// receiveMessages 持续接收服务器消息
func receiveMessages(msgChan chan []byte) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			close(msgChan)
			return
		}
		msgChan <- message
	}
}

// handleGameMessage 处理服务器推送的游戏消息
func handleGameMessage(message []byte) {
	var baseMsg protocol.BaseMessage
	if err := json.Unmarshal(message, &baseMsg); err != nil {
		log.Println("解析消息失败:", err)
		return
	}

	switch baseMsg.Type {
	case protocol.MsgTypeGameState:
		var state protocol.GameState
		json.Unmarshal(message, &state)
		displayGameState(&state)

	case protocol.MsgTypeYourTurn:
		var turn protocol.YourTurn
		json.Unmarshal(message, &turn)
		fmt.Println()
		fmt.Println("╔══════════════════════════════════╗")
		fmt.Println("║         轮到你行动了!            ║")
		fmt.Println("╠══════════════════════════════════╣")
		fmt.Printf("║  当前下注: %-6d                ║\n", turn.CurrentBet)
		fmt.Printf("║  最小操作: %-6d                ║\n", turn.MinAction)
		fmt.Printf("║  剩余时间: %-2d 秒                 ║\n", turn.TimeLeft)
		fmt.Println("╠══════════════════════════════════╣")
		fmt.Println("║  F-弃牌 X-看牌 C-跟注           ║")
		fmt.Println("║  R <金额>-加注  A-全下           ║")
		fmt.Println("╚══════════════════════════════════╝")

	case protocol.MsgTypeShowdown:
		var showdown protocol.Showdown
		json.Unmarshal(message, &showdown)
		displayShowdown(&showdown)

	case protocol.MsgTypePlayerJoined:
		var joined protocol.PlayerJoined
		json.Unmarshal(message, &joined)
		fmt.Printf("\n[通知] 玩家 %s 加入游戏 (座位 %d)\n", joined.Player.Name, joined.Player.Seat+1)

	case protocol.MsgTypePlayerLeft:
		var left protocol.PlayerLeft
		json.Unmarshal(message, &left)
		fmt.Printf("\n[通知] 玩家 %s 离开游戏\n", left.PlayerName)

	case protocol.MsgTypePlayerActed:
		var acted protocol.PlayerActed
		json.Unmarshal(message, &acted)
		fmt.Printf("\n[动作] %s %s", acted.PlayerName, getActionText(acted.Action))
		if acted.Amount > 0 {
			fmt.Printf(" %d", acted.Amount)
		}
		fmt.Println()

	case protocol.MsgTypeError:
		var errMsg protocol.Error
		json.Unmarshal(message, &errMsg)
		fmt.Printf("\n[错误] %s\n", errMsg.Message)
	}
}

// handleUserInput 处理用户输入的命令
func handleUserInput(text string) {
	switch strings.ToLower(text) {
	case "f", "fold":
		sendAction(models.ActionFold, 0)
	case "x", "check":
		sendAction(models.ActionCheck, 0)
	case "c", "call":
		sendAction(models.ActionCall, 0)
	case "a", "allin":
		sendAction(models.ActionAllIn, 0)
	case "s", "status":
		requestState()
	case "h", "help":
		showHelp()
	case "q", "quit":
		fmt.Println("再见!")
		os.Exit(0)
	default:
		// 检查是否是加注命令
		parts := strings.SplitN(text, " ", 2)
		if len(parts) == 2 && (strings.ToLower(parts[0]) == "r" || strings.ToLower(parts[0]) == "raise") {
			var amount int
			fmt.Sscanf(parts[1], "%d", &amount)
			if amount > 0 {
				sendAction(models.ActionRaise, amount)
			} else {
				fmt.Println("无效的加注金额，格式: R 100")
			}
		} else {
			fmt.Println("未知指令，输入 'h' 查看帮助")
		}
	}
}

// getActionText 获取动作描述
func getActionText(action models.ActionType) string {
	switch action {
	case models.ActionFold:
		return "弃牌"
	case models.ActionCheck:
		return "看牌"
	case models.ActionCall:
		return "跟注"
	case models.ActionRaise:
		return "加注"
	case models.ActionAllIn:
		return "全下"
	default:
		return "未知动作"
	}
}

// sendAction 发送动作
func sendAction(action models.ActionType, amount int) {
	req := protocol.NewPlayerActionRequest(playerID, action, amount)
	sendMessage(req)
}

// requestState 请求游戏状态
func requestState() {
	// 这里可以发送状态请求消息
	fmt.Println("正在刷新状态...")
}

// sendMessage 发送消息
func sendMessage(msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("序列化消息失败:", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Println("发送消息失败:", err)
	}
}

// displayGameState 显示游戏状态
func displayGameState(state *protocol.GameState) {
	fmt.Println()
	fmt.Println("┌─────────────────────────────────────────────────────────┐")
	fmt.Printf("│ 阶段: %-8s │ 底池: %-5d │ 庄家位置: %d           │\n",
		state.Stage, state.Pot, state.DealerButton+1)
	fmt.Println("├─────────────────────────────────────────────────────────┤")

	// 显示公共牌
	fmt.Print("│ 公共牌: ")
	for _, c := range state.CommunityCards {
		if c.Rank == 0 {
			fmt.Print("   ")
		} else {
			fmt.Printf("%s%s ", rankSymbols[c.Rank], suitSymbols[c.Suit])
		}
	}
	fmt.Println("                                   │")

	fmt.Println("├─────────────────────────────────────────────────────────┤")
	fmt.Println("│ 玩家列表:                                               │")

	for idx, p := range state.Players {
		status := ""
		switch p.Status {
		case models.PlayerStatusActive: // 1 = 游戏中
			if p.IsSelf {
				status = "【你】"
			} else {
				status = "    "
			}
		case models.PlayerStatusFolded: // 2 = 已弃牌
			status = "已弃牌"
		case models.PlayerStatusAllIn: // 3 = 全下
			status = "全下  "
		default: // 0 = 未入座或其他
			status = "    "
		}
		fmt.Printf("│   [%d] %-10s 筹码:%-4d 下注:%-3d %s",
			idx+1, p.Name, p.Chips, p.CurrentBet, status)
		if p.IsSelf && p.HoleCards[0].Rank != 0 {
			fmt.Printf(" 底牌: %s %s", formatCard(p.HoleCards[0]), formatCard(p.HoleCards[1]))
		}
		fmt.Println("                      │")
	}
	fmt.Println("└─────────────────────────────────────────────────────────┘")
}

// formatCard 格式化单张牌
func formatCard(c card.Card) string {
	if c.Rank == 0 {
		return "??"
	}
	return fmt.Sprintf("%s%s", rankSymbols[c.Rank], suitSymbols[c.Suit])
}

// displayShowdown 显示摊牌结果
func displayShowdown(showdown *protocol.Showdown) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║            摊牌结果                  ║")
	fmt.Println("╠════════════════════════════════════════╣")

	for _, winner := range showdown.Winners {
		fmt.Printf("║  %-10s: %-20s    ║\n", winner.PlayerName, winner.HandName)
		fmt.Printf("║               赢得: %d 筹码              ║\n", winner.WonChips)
		fmt.Print("║               手牌: ")
		for _, c := range winner.RawCards {
			fmt.Printf("%s ", formatCard(c))
		}
		fmt.Println("             ║")
	}

	fmt.Println("╚════════════════════════════════════════╝")
}
