package main

import (
	"fmt"
	"os"

	"github.com/wilenwang/just_play/Texas-Holdem/internal/card"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/common/models"
	"github.com/wilenwang/just_play/Texas-Holdem/pkg/game"
)

// 简单控制台德州扑克游戏示例
func main() {
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║       德州扑克 - 控制台版 v1.0      ║")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()

	// 创建游戏配置
	config := &game.Config{
		MinPlayers:     2,
		MaxPlayers:     9,
		SmallBlind:     10,
		BigBlind:       20,
		Ante:           0,   // 前注（设为0禁用）
		StartingChips:  1000, // 初始筹码
		ActionTimeout:  30,
	}

	// 创建游戏引擎
	engine := game.NewEngine(config)
	fmt.Println("游戏配置:")
	fmt.Printf("  盲注: %d/%d\n", config.SmallBlind, config.BigBlind)
	fmt.Printf("  初始筹码: %d\n", config.StartingChips)
	fmt.Println()

	// 添加玩家
	playerNames := []string{"Alice", "Bob", "Charlie"}
	for i, name := range playerNames {
		engine.AddPlayer(fmt.Sprintf("p%d", i+1), name, i)
	}
	fmt.Printf("已添加 %d 位玩家\n\n", len(playerNames))

	// 开始第一局
	fmt.Println("═══════════════════════════════════════")
	fmt.Println("第 1 局")
	fmt.Println("═══════════════════════════════════════")

	err := engine.StartHand()
	if err != nil {
		fmt.Printf("开始游戏失败: %v\n", err)
		os.Exit(1)
	}

	// 模拟游戏流程
	playGame(engine)

	fmt.Println("\n游戏演示完成！")
}

// 模拟完整游戏流程
func playGame(engine *game.GameEngine) {
	state := engine.GetState()

	fmt.Println("\n【阶段】:", state.Stage)
	fmt.Printf("【底池】: %d\n", state.Pot)
	fmt.Printf("【庄家】: 玩家 %d (%s)\n", state.DealerButton+1,
		state.Players[state.DealerButton].Name)

	// 显示公共牌
	fmt.Print("【公共牌】: ")
	if state.CommunityCards[0].Rank == 0 {
		fmt.Println("（尚未发出）")
	} else {
		showCards(state.CommunityCards[:])
	}

	// 显示玩家状态
	fmt.Println("\n【玩家状态】:")
	for i, p := range state.Players {
		status := ""
		switch p.Status {
		case models.PlayerStatusActive:
			status = "活跃"
		case models.PlayerStatusFolded:
			status = "已弃牌"
		case models.PlayerStatusAllIn:
			status = "全下"
		}
		cards := "[  ?  ][  ?  ]"
		if p.HoleCards[0].Rank != 0 {
			cards = fmt.Sprintf("[%s][%s]", p.HoleCards[0].String(), p.HoleCards[1].String())
		}
		fmt.Printf("  [%d] %-10s 筹码:%-4d 状态:%s %s\n",
			i+1, p.Name, p.Chips, status, cards)
	}

	// 模拟下注轮
	simulateBettingRounds(engine)
}

// 模拟各轮下注
func simulateBettingRounds(engine *game.GameEngine) {
	// 翻牌前下注
	fmt.Println("\n--- 翻牌前下注 ---")
	simulateRound(engine)

	// 翻牌圈
	if engine.GetState().Stage == game.StageFlop {
		fmt.Println("\n--- 翻牌圈 ---")
		showCommunityCards(engine)
		simulateRound(engine)
	}

	// 转牌圈
	if engine.GetState().Stage == game.StageTurn {
		fmt.Println("\n--- 转牌圈 ---")
		showCommunityCards(engine)
		simulateRound(engine)
	}

	// 河牌圈
	if engine.GetState().Stage == game.StageRiver {
		fmt.Println("\n--- 河牌圈 ---")
		showCommunityCards(engine)
		simulateRound(engine)
	}

	// 摊牌
	if engine.GetState().Stage == game.StageShowdown {
		fmt.Println("\n========== 摊牌 ==========")
		state := engine.GetState()

		// 显示所有玩家的底牌
		fmt.Println("【最终手牌】:")
		for _, p := range state.Players {
			if p.Status == models.PlayerStatusActive || p.Status == models.PlayerStatusAllIn {
				fmt.Printf("  %s: %s %s\n",
					p.Name,
					p.HoleCards[0].String(),
					p.HoleCards[1].String())
			}
		}

		// 显示获胜者
		fmt.Println("\n【底池分配】:")
		for i, p := range state.Players {
			if p.Chips > 1000 {
				winnings := p.Chips - 1000
				fmt.Printf("  %s 赢得 %d 筹码\n", p.Name, winnings)
			}
			if i < len(state.Players)-1 {
				_ = i
			}
		}
	}

	// 显示底池清空
	fmt.Printf("\n【本局底池】: %d (已分配)\n", engine.GetState().Pot)
}

// 模拟一轮下注
func simulateRound(engine *game.GameEngine) {
	state := engine.GetState()
	activeCount := 0
	for _, p := range state.Players {
		if p.Status == models.PlayerStatusActive {
			activeCount++
		}
	}

	if activeCount <= 1 {
		fmt.Println("  (活跃玩家不足，跳过下注)")
		return
	}

	// 简单模拟：每位活跃玩家依次行动
	actions := 0
	for actions < activeCount*2 { // 简单限制
		state = engine.GetState()
		if state.Stage != game.StagePreFlop &&
		   state.Stage != game.StageFlop &&
		   state.Stage != game.StageTurn &&
		   state.Stage != game.StageRiver {
			break
		}

		// 找到当前玩家
		currentPlayerIdx := state.CurrentPlayer
		if currentPlayerIdx >= len(state.Players) {
			break
		}

		p := state.Players[currentPlayerIdx]
		if p.Status != models.PlayerStatusActive {
			actions++
			continue
		}

		// 简单AI：随机行动
		fmt.Printf("  %s 行动 (筹码:%d, 下注:%d, 底池:%d)\n",
			p.Name, p.Chips, p.CurrentBet, state.Pot)

		// 模拟行动
		action := models.ActionCheck
		if state.CurrentBet > p.CurrentBet {
			// 需要跟注或弃牌
			if p.Chips <= state.CurrentBet-p.CurrentBet {
				action = models.ActionAllIn
			} else {
				// 50%跟注，50%弃牌
				action = models.ActionCall
			}
		} else {
			// 可以看牌或加注
			if state.CurrentBet == 0 {
				action = models.ActionCheck
			} else {
				// 30%加注，70%看牌
				action = models.ActionCheck
			}
		}

		engine.PlayerAction(p.ID, action, 0)
		actions++

		state = engine.GetState()
		if state.Stage == game.StageShowdown ||
		   state.Stage == game.StageEnd {
			break
		}
	}

	fmt.Printf("  下注完成，底池: %d\n", engine.GetState().Pot)
}

// 显示公共牌
func showCommunityCards(engine *game.GameEngine) {
	state := engine.GetState()
	fmt.Print("【公共牌】: ")
	showCards(state.CommunityCards[:])
}

// 显示牌数组
func showCards(cards []card.Card) {
	if len(cards) == 0 || cards[0].Rank == 0 {
		fmt.Println("（空）")
		return
	}
	for i, c := range cards {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(c.String())
	}
	fmt.Println()
}
