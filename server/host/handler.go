package host

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/wilenwang/just_play/Texas-Holdem/internal/common/models"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/protocol"
	gamepkg "github.com/wilenwang/just_play/Texas-Holdem/pkg/game"
)

// handleJoin 处理玩家加入游戏
func (s *Server) handleJoin(client *Client, data []byte) {
	var req protocol.JoinRequest
	if err := json.Unmarshal(data, &req); err != nil {
		log.Printf("[加入] 解析失败 | 客户端=%s | 错误=%v", client.ID, err)
		s.sendError(client.ID, "Invalid join request format", 1001)
		return
	}

	client.Name = req.PlayerName
	log.Printf("[加入] 收到请求 | 玩家=%s | 客户端ID=%s | 请求座位=%d", req.PlayerName, client.ID, req.Seat)

	// 检查座位号
	seat := req.Seat
	if seat < 0 {
		// 自动分配座位
		seat = s.findAvailableSeat()
		log.Printf("[加入] 自动分配座位 | 玩家=%s | 分配座位=%d", req.PlayerName, seat)
	}

	// 尝试添加玩家到游戏
	player, err := s.gameEngine.AddPlayer(client.ID, req.PlayerName, seat)
	if err != nil {
		log.Printf("[加入] 失败 | 玩家=%s | 座位=%d | 错误=%v", req.PlayerName, seat, err)
		switch err {
		case gamepkg.ErrGameFull:
			s.sendError(client.ID, "Game is full", 2001)
		case gamepkg.ErrInvalidSeat:
			s.sendError(client.ID, "Invalid seat number", 2002)
		case gamepkg.ErrSeatOccupied:
			s.sendError(client.ID, "Seat is already occupied", 2003)
		default:
			s.sendError(client.ID, "Failed to join game", 2000)
		}
		return
	}

	client.Seat = seat

	// 发送加入确认
	ack := &protocol.JoinAck{
		BaseMessage: protocol.NewBaseMessage(protocol.MsgTypeJoinAck),
		Success:     true,
		PlayerID:    client.ID,
		Seat:        seat,
		Message:     "Successfully joined the game!",
		GameState:   s.getGameStateInfo(client.ID),
	}
	s.sendToClient(client.ID, ack)

	// 广播新玩家加入
	joinedMsg := &protocol.PlayerJoined{
		BaseMessage: protocol.NewBaseMessage(protocol.MsgTypePlayerJoined),
		Player: protocol.PlayerInfo{
			ID:       player.ID,
			Name:     player.Name,
			Seat:     player.Seat,
			Chips:    player.Chips,
			Status:   player.Status,
			IsDealer: player.IsDealer,
			IsSelf:   false,
		},
	}
	data, _ = json.Marshal(joinedMsg)
	s.broadcastToOthers(client.ID, data)

	// 打印当前牌桌玩家列表
	state := s.gameEngine.GetState()
	log.Printf("[加入] 成功 | 玩家=%s | 座位=%d | 筹码=%d | 当前玩家数=%d",
		req.PlayerName, seat, player.Chips, len(state.Players))
	s.logPlayerList(state)

	// 如果游戏正在进行中（不是等待/结算阶段），将新玩家标记为弃牌，等下一局参与
	if state.Stage != gamepkg.StageWaiting && state.Stage != gamepkg.StageShowdown && state.Stage != gamepkg.StageEnd {
		s.gameEngine.SetPlayerStatus(client.ID, models.PlayerStatusFolded)
		log.Printf("[加入] 游戏进行中，玩家 %s 标记为弃牌，等待下一局参与", req.PlayerName)
	}
	// 不再自动开局，改为大厅准备制：所有玩家按准备后才开始
}

// handleLeave 处理玩家离开游戏
func (s *Server) handleLeave(client *Client) {
	log.Printf("[离开] 收到请求 | 玩家=%s | 客户端ID=%s | 座位=%d", client.Name, client.ID, client.Seat)

	if err := s.gameEngine.RemovePlayer(client.ID); err != nil {
		log.Printf("[离开] 失败 | 玩家=%s | 错误=%v", client.Name, err)
		s.sendError(client.ID, "Failed to leave game", 2004)
		return
	}

	state := s.gameEngine.GetState()
	log.Printf("[离开] 成功 | 玩家=%s | 剩余玩家数=%d | 当前阶段=%s",
		client.Name, len(state.Players), state.Stage)
}

// handlePlayerAction 处理玩家动作
func (s *Server) handlePlayerAction(client *Client, data []byte) {
	var req protocol.PlayerActionRequest
	if err := json.Unmarshal(data, &req); err != nil {
		log.Printf("[动作] 解析失败 | 玩家=%s | 错误=%v", client.Name, err)
		s.sendError(client.ID, "Invalid action request format", 1001)
		return
	}

	// 获取动作前的状态快照
	beforeState := s.gameEngine.GetState()
	log.Printf("[动作] 收到请求 | 玩家=%s | 动作=%s | 金额=%d | 当前阶段=%s | 底池=%d | 当前下注=%d",
		client.Name, actionName(req.Action), req.Amount, beforeState.Stage, beforeState.Pot, beforeState.CurrentBet)

	// 检查游戏是否在下注阶段（等待/摊牌/结束阶段不接受玩家动作）
	if beforeState.Stage == gamepkg.StageWaiting || beforeState.Stage == gamepkg.StageShowdown || beforeState.Stage == gamepkg.StageEnd {
		log.Printf("[动作] 拒绝 | 玩家=%s | 原因=当前阶段(%s)不是下注阶段", client.Name, beforeState.Stage)
		s.sendError(client.ID, "Game is not in a betting stage", 3003)
		return
	}

	// 验证是否是该玩家的回合
	if beforeState.CurrentPlayer >= len(beforeState.Players) {
		log.Printf("[动作] 拒绝 | 玩家=%s | 原因=CurrentPlayer(%d) >= 玩家数(%d)",
			client.Name, beforeState.CurrentPlayer, len(beforeState.Players))
		s.sendError(client.ID, "Not your turn", 3001)
		return
	}

	currentPlayer := beforeState.Players[beforeState.CurrentPlayer]
	if currentPlayer.ID != client.ID {
		log.Printf("[动作] 拒绝 | 玩家=%s | 原因=不是你的回合 | 当前行动玩家=%s(idx=%d)",
			client.Name, currentPlayer.Name, beforeState.CurrentPlayer)
		s.sendError(client.ID, "Not your turn", 3001)
		return
	}

	// 打印动作前的玩家状态
	log.Printf("[动作] 执行前状态 | 玩家=%s | 筹码=%d | 已下注=%d | 状态=%s",
		client.Name, currentPlayer.Chips, currentPlayer.CurrentBet, playerStatusName(currentPlayer.Status))

	// 执行动作
	if err := s.gameEngine.PlayerAction(client.ID, req.Action, req.Amount); err != nil {
		log.Printf("[动作] 执行失败 | 玩家=%s | 动作=%s | 金额=%d | 错误=%v",
			client.Name, actionName(req.Action), req.Amount, err)
		s.sendError(client.ID, err.Error(), 3002)

		// 动作被拒绝，仅在游戏处于下注阶段时重新发送 YourTurn
		// 非下注阶段（等待/摊牌/结束）不应重发，避免客户端误以为轮到自己
		afterRejectState := s.gameEngine.GetState()
		if afterRejectState.Stage == gamepkg.StagePreFlop || afterRejectState.Stage == gamepkg.StageFlop ||
			afterRejectState.Stage == gamepkg.StageTurn || afterRejectState.Stage == gamepkg.StageRiver {
			minAction := s.getMinAction(client.ID)
			maxAction := s.getMaxAction(client.ID)
			turnMsg := &protocol.YourTurn{
				BaseMessage: protocol.NewBaseMessage(protocol.MsgTypeYourTurn),
				PlayerID:    client.ID,
				MinAction:   minAction,
				MaxAction:   maxAction,
				CurrentBet:  beforeState.CurrentBet,
				TimeLeft:    30,
			}
			s.sendToClient(client.ID, turnMsg)
			log.Printf("[动作] 重发行动通知 | 玩家=%s | 需补=%d | 最大=%d", client.Name, minAction, maxAction)
		}
		return
	}

	// 获取动作后的状态
	afterState := s.gameEngine.GetState()

	// 打印详细的动作结果
	log.Printf("[动作] 执行成功 | 玩家=%s | 动作=%s | 金额=%d", client.Name, actionName(req.Action), req.Amount)
	log.Printf("[动作] 状态变化 | 阶段: %s→%s | 底池: %d→%d | 当前下注: %d→%d",
		beforeState.Stage, afterState.Stage, beforeState.Pot, afterState.Pot, beforeState.CurrentBet, afterState.CurrentBet)

	// 打印动作后所有玩家状态
	s.logPlayerList(afterState)

	// 广播玩家动作
	actedMsg := &protocol.PlayerActed{
		BaseMessage: protocol.NewBaseMessage(protocol.MsgTypePlayerActed),
		PlayerID:    client.ID,
		PlayerName:  client.Name,
		Action:      req.Action,
		Amount:      req.Amount,
	}
	data, _ = json.Marshal(actedMsg)
	s.broadcast <- data

	// 检查游戏状态
	if afterState.Stage == gamepkg.StageEnd || afterState.Stage == gamepkg.StageShowdown {
		log.Printf("[状态机] 本局结束 | 阶段=%s | 底池=%d", afterState.Stage, afterState.Pot)
		s.logFinalResult(afterState)

		// 广播结算详情给所有玩家
		s.broadcastShowdownResult(afterState)

		// 重置准备状态，等待玩家确认下一局
		s.resetReadyState()
		log.Printf("[状态机] 等待所有玩家确认下一局...")
		return
	}

	// 如果游戏仍在进行，通知下一个行动玩家
	if afterState.CurrentPlayer < len(afterState.Players) {
		nextPlayer := afterState.Players[afterState.CurrentPlayer]
		minAction := s.getMinAction(nextPlayer.ID)
		maxAction := s.getMaxAction(nextPlayer.ID)

		stageChanged := beforeState.Stage != afterState.Stage
		log.Printf("[轮转] 下一个行动 | 玩家=%s(idx=%d) | 筹码=%d | 已下注=%d | 需补=%d | 最大=%d | 换轮=%v",
			nextPlayer.Name, afterState.CurrentPlayer, nextPlayer.Chips, nextPlayer.CurrentBet, minAction, maxAction, stageChanged)

		// 发送行动通知：换轮时无论是否同一人都要通知，同一轮内只通知不同玩家
		if stageChanged || nextPlayer.ID != client.ID {
			turnMsg := &protocol.YourTurn{
				BaseMessage: protocol.NewBaseMessage(protocol.MsgTypeYourTurn),
				PlayerID:    nextPlayer.ID,
				MinAction:   minAction,
				MaxAction:   maxAction,
				CurrentBet:  afterState.CurrentBet,
				TimeLeft:    30,
			}
			s.sendToClient(nextPlayer.ID, turnMsg)
		}
	}
}

// handleChat 处理聊天消息
func (s *Server) handleChat(client *Client, data []byte) {
	var req protocol.ChatRequest
	if err := json.Unmarshal(data, &req); err != nil {
		s.sendError(client.ID, "Invalid chat format", 1001)
		return
	}

	log.Printf("[聊天] 玩家=%s | 内容=%s", client.Name, req.Content)

	chatMsg := &protocol.ChatMessage{
		BaseMessage: protocol.NewBaseMessage(protocol.MsgTypeChat),
		PlayerID:    client.ID,
		PlayerName:  client.Name,
		Content:     req.Content,
		IsSystem:    false,
	}

	data, _ = json.Marshal(chatMsg)
	s.broadcast <- data
}

// handlePing 处理心跳检测
func (s *Server) handlePing(client *Client) {
	pong := &protocol.Pong{
		BaseMessage: protocol.NewBaseMessage(protocol.MsgTypePong),
		ServerTime:  time.Now().UnixMilli(),
	}
	s.sendToClient(client.ID, pong)
}

// findAvailableSeat 查找可用座位
func (s *Server) findAvailableSeat() int {
	state := s.gameEngine.GetState()
	usedSeats := make(map[int]bool)
	for _, p := range state.Players {
		usedSeats[p.Seat] = true
	}

	for i := 0; i < 9; i++ {
		if !usedSeats[i] {
			return i
		}
	}
	return -1
}

// getMinAction 获取最小可执行下注金额
func (s *Server) getMinAction(playerID string) int {
	state := s.gameEngine.GetState()
	for _, p := range state.Players {
		if p.ID == playerID {
			return state.CurrentBet - p.CurrentBet
		}
	}
	return 0
}

// getMaxAction 获取最大可执行下注金额
func (s *Server) getMaxAction(playerID string) int {
	state := s.gameEngine.GetState()
	for _, p := range state.Players {
		if p.ID == playerID {
			return state.CurrentBet + p.Chips
		}
	}
	return 0
}

// tryAutoStartHand 尝试自动开始新的一局（当玩家人数满足最低要求且当前没有进行中的游戏时）
func (s *Server) tryAutoStartHand() {
	state := s.gameEngine.GetState()
	config := s.gameEngine.GetConfig()
	minPlayers := config.MinPlayers

	log.Printf("[自动开局] 检查条件 | 当前阶段=%s | 玩家数=%d | 最少=%d", state.Stage, len(state.Players), minPlayers)

	// 仅在等待、摊牌、局结束状态下才能开始新的一局
	if state.Stage != gamepkg.StageWaiting && state.Stage != gamepkg.StageShowdown && state.Stage != gamepkg.StageEnd {
		log.Printf("[自动开局] 跳过 | 原因=当前阶段(%s)不允许开始新局", state.Stage)
		return
	}

	// 检查玩家数量是否满足最低要求（使用配置的最少玩家数）
	if len(state.Players) < minPlayers {
		log.Printf("[自动开局] 跳过 | 原因=玩家数不足(需要>=%d, 当前=%d)", minPlayers, len(state.Players))
		return
	}

	// 统计有筹码的活跃玩家
	activeWithChips := 0
	for _, p := range state.Players {
		if p.Chips > 0 {
			activeWithChips++
		}
	}
	if activeWithChips < minPlayers {
		log.Printf("[自动开局] 跳过 | 原因=有筹码的玩家不足(需要>=%d, 当前=%d)", minPlayers, activeWithChips)
		return
	}

	// 开始新的一局
	if err := s.gameEngine.StartHand(); err != nil {
		log.Printf("[自动开局] 失败 | 错误=%v", err)
		return
	}

	newState := s.gameEngine.GetState()
	log.Printf("[自动开局] 成功! | 玩家数=%d | 庄家位置=%d | 底池=%d | 当前下注=%d",
		len(newState.Players), newState.DealerButton, newState.Pot, newState.CurrentBet)
	s.logPlayerList(newState)

	// 打印公共牌信息
	s.logCommunityCards(newState)

	// 通知当前行动玩家
	if newState.CurrentPlayer < len(newState.Players) {
		nextPlayer := newState.Players[newState.CurrentPlayer]
		minAction := s.getMinAction(nextPlayer.ID)
		maxAction := s.getMaxAction(nextPlayer.ID)

		log.Printf("[自动开局] 第一个行动 | 玩家=%s(idx=%d) | 筹码=%d | 已下注=%d | 需补=%d",
			nextPlayer.Name, newState.CurrentPlayer, nextPlayer.Chips, nextPlayer.CurrentBet, minAction)

		turnMsg := &protocol.YourTurn{
			BaseMessage: protocol.NewBaseMessage(protocol.MsgTypeYourTurn),
			PlayerID:    nextPlayer.ID,
			MinAction:   minAction,
			MaxAction:   maxAction,
			CurrentBet:  newState.CurrentBet,
			TimeLeft:    30,
		}
		s.sendToClient(nextPlayer.ID, turnMsg)
	}
}

// broadcastShowdownResult 广播结算结果给所有客户端
func (s *Server) broadcastShowdownResult(state *gamepkg.GameState) {
	if state.LastShowdown == nil {
		log.Printf("[结算广播] 无结算数据，跳过")
		return
	}

	sd := state.LastShowdown

	// 构建获胜者列表
	var winners []protocol.WinnerInfo
	for _, pr := range sd.Players {
		if pr.IsWinner {
			winners = append(winners, protocol.WinnerInfo{
				PlayerID:   state.Players[pr.PlayerIdx].ID,
				PlayerName: pr.PlayerName,
				HandRank:   pr.HandRank,
				HandName:   pr.HandName,
				WonChips:   pr.WonAmount,
				RawCards:   pr.BestCards,
			})
		}
	}

	// 构建摊牌消息（包含所有玩家详情，方便客户端展示）
	showdownMsg := &protocol.Showdown{
		BaseMessage: protocol.NewBaseMessage(protocol.MsgTypeShowdown),
		Winners:     winners,
		Pot:         sd.TotalPot,
		IsEarlyEnd:  sd.IsEarlyEnd,
		AllPlayers:  make([]protocol.ShowdownPlayerDetail, 0, len(sd.Players)),
		CommunityCards: state.CommunityCards,
	}

	for _, pr := range sd.Players {
		detail := protocol.ShowdownPlayerDetail{
			PlayerName:  pr.PlayerName,
			HoleCards:   pr.HoleCards,
			HandName:    pr.HandName,
			WonAmount:   pr.WonAmount,
			IsWinner:    pr.IsWinner,
			IsFolded:    pr.IsFolded,
			ChipsAfter:  pr.ChipsAfter,
		}
		showdownMsg.AllPlayers = append(showdownMsg.AllPlayers, detail)
	}

	data, err := json.Marshal(showdownMsg)
	if err != nil {
		log.Printf("[结算广播] 序列化失败: %v", err)
		return
	}

	log.Printf("[结算广播] 广播结算结果 | 总底池=%d | 赢家数=%d | 提前结束=%v",
		sd.TotalPot, len(winners), sd.IsEarlyEnd)
	s.broadcast <- data
}

// ==================== 准备下一局相关方法 ====================

// handleReadyForNext 处理玩家准备下一局请求
func (s *Server) handleReadyForNext(client *Client, data []byte) {
	log.Printf("[准备] 收到请求 | 玩家=%s | 客户端ID=%s", client.Name, client.ID)

	// 检查当前是否处于等待准备状态
	state := s.gameEngine.GetState()
	if state.Stage != gamepkg.StageEnd && state.Stage != gamepkg.StageShowdown && state.Stage != gamepkg.StageWaiting {
		log.Printf("[准备] 拒绝 | 玩家=%s | 原因=当前阶段(%s)不在结算状态", client.Name, state.Stage)
		s.sendError(client.ID, "当前不在结算阶段", 4001)
		return
	}

	// 标记该玩家已准备
	s.readyMu.Lock()
	s.readyPlayers[client.ID] = true
	s.waitingReady = true
	s.readyMu.Unlock()

	// 构建已准备玩家名称列表
	readyNames := s.getReadyPlayerNames()
	totalPlayers := len(state.Players)
	allReady := len(readyNames) >= totalPlayers

	log.Printf("[准备] 玩家 %s 已准备 | 已准备=%d/%d | 全部准备=%v",
		client.Name, len(readyNames), totalPlayers, allReady)

	// 广播准备状态给所有玩家
	readyMsg := &protocol.PlayerReadyNotify{
		BaseMessage:  protocol.NewBaseMessage(protocol.MsgTypePlayerReady),
		PlayerID:     client.ID,
		PlayerName:   client.Name,
		ReadyPlayers: readyNames,
		TotalPlayers: totalPlayers,
		AllReady:     allReady,
	}
	msgData, _ := json.Marshal(readyMsg)
	s.broadcast <- msgData

	// 如果所有玩家都准备好了，开始下一局
	if allReady {
		log.Printf("[准备] 所有玩家已准备，开始下一局!")
		s.resetReadyState()
		s.tryAutoStartHand()
	}
}

// resetReadyState 重置所有玩家的准备状态
func (s *Server) resetReadyState() {
	s.readyMu.Lock()
	defer s.readyMu.Unlock()

	s.readyPlayers = make(map[string]bool)
	s.waitingReady = true
	log.Printf("[准备] 已重置所有玩家准备状态")
}

// getReadyPlayerNames 获取已准备玩家的名称列表
func (s *Server) getReadyPlayerNames() []string {
	s.readyMu.RLock()
	defer s.readyMu.RUnlock()

	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	var names []string
	for playerID := range s.readyPlayers {
		if client, ok := s.clients[playerID]; ok {
			names = append(names, client.Name)
		}
	}
	return names
}

// ==================== 日志辅助方法 ====================

// logPlayerList 打印当前所有玩家信息列表
func (s *Server) logPlayerList(state *gamepkg.GameState) {
	if len(state.Players) == 0 {
		log.Printf("[牌桌] 无玩家")
		return
	}
	var lines []string
	for i, p := range state.Players {
		dealer := ""
		if p.IsDealer {
			dealer = " [庄]"
		}
		current := ""
		if i == state.CurrentPlayer {
			current = " ←行动"
		}
		lines = append(lines, fmt.Sprintf("  座位%d: %s | 筹码=%d | 下注=%d | 状态=%s | HasActed=%v%s%s",
			p.Seat, p.Name, p.Chips, p.CurrentBet, playerStatusName(p.Status), p.HasActed, dealer, current))
	}
	log.Printf("[牌桌] 玩家列表:\n%s", strings.Join(lines, "\n"))
}

// logCommunityCards 打印公共牌信息
func (s *Server) logCommunityCards(state *gamepkg.GameState) {
	var cards []string
	for _, c := range state.CommunityCards {
		if c.Rank == 0 {
			cards = append(cards, "**")
		} else {
			cards = append(cards, c.String())
		}
	}
	log.Printf("[牌桌] 公共牌: [%s]", strings.Join(cards, " "))
}

// logFinalResult 打印本局最终结果
func (s *Server) logFinalResult(state *gamepkg.GameState) {
	log.Printf("[结算] ====== 本局结果 ======")
	s.logCommunityCards(state)
	for _, p := range state.Players {
		log.Printf("[结算] %s | 筹码=%d | 状态=%s", p.Name, p.Chips, playerStatusName(p.Status))
	}
	log.Printf("[结算] ========================")
}

// actionName 返回动作的中文名称
func actionName(action models.ActionType) string {
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
		return fmt.Sprintf("未知(%v)", action)
	}
}

// playerStatusName 返回玩家状态的中文名称
func playerStatusName(status models.PlayerStatus) string {
	switch status {
	case models.PlayerStatusInactive:
		return "未入座"
	case models.PlayerStatusActive:
		return "活跃"
	case models.PlayerStatusFolded:
		return "已弃牌"
	case models.PlayerStatusAllIn:
		return "全下"
	default:
		return fmt.Sprintf("未知(%d)", status)
	}
}
