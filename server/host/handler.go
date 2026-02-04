package host

import (
	"encoding/json"
	"log"
	"time"

	"github.com/wilenwang/just_play/Texas-Holdem/internal/protocol"
	gamepkg "github.com/wilenwang/just_play/Texas-Holdem/pkg/game"
)

// handleJoin 处理玩家加入游戏
func (s *Server) handleJoin(client *Client, data []byte) {
	var req protocol.JoinRequest
	if err := json.Unmarshal(data, &req); err != nil {
		s.sendError(client.ID, "Invalid join request format", 1001)
		return
	}

	client.Name = req.PlayerName

	// 检查座位号
	seat := req.Seat
	if seat < 0 {
		// 自动分配座位
		seat = s.findAvailableSeat()
	}

	// 尝试添加玩家到游戏
	player, err := s.gameEngine.AddPlayer(client.ID, req.PlayerName, seat)
	if err != nil {
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
			ID:         player.ID,
			Name:       player.Name,
			Seat:       player.Seat,
			Chips:      player.Chips,
			Status:     player.Status,
			IsDealer:   player.IsDealer,
			IsSelf:     false,
		},
	}
	data, _ = json.Marshal(joinedMsg)
	s.broadcastToOthers(client.ID, data)

	log.Printf("Player %s joined at seat %d", req.PlayerName, seat)
}

// handleLeave 处理玩家离开游戏
func (s *Server) handleLeave(client *Client) {
	if err := s.gameEngine.RemovePlayer(client.ID); err != nil {
		s.sendError(client.ID, "Failed to leave game", 2004)
		return
	}

	log.Printf("Player %s left the game", client.Name)

	// 发送离开确认
	// 实际游戏逻辑中可能不需要确认，直接关闭连接即可
}

// handlePlayerAction 处理玩家动作
func (s *Server) handlePlayerAction(client *Client, data []byte) {
	var req protocol.PlayerActionRequest
	if err := json.Unmarshal(data, &req); err != nil {
		s.sendError(client.ID, "Invalid action request format", 1001)
		return
	}

	// 验证是否是该玩家的回合
	state := s.gameEngine.GetState()
	if state.CurrentPlayer >= len(state.Players) {
		s.sendError(client.ID, "Not your turn", 3001)
		return
	}

	currentPlayer := state.Players[state.CurrentPlayer]
	if currentPlayer.ID != client.ID {
		s.sendError(client.ID, "Not your turn", 3001)
		return
	}

	// 执行动作
	if err := s.gameEngine.PlayerAction(client.ID, req.Action, req.Amount); err != nil {
		s.sendError(client.ID, err.Error(), 3002)
		return
	}

	log.Printf("Player %s performed action: %s", client.Name, req.Action)

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

	// 检查是否轮到新玩家
	newState := s.gameEngine.GetState()
	if newState.CurrentPlayer < len(newState.Players) {
		nextPlayer := newState.Players[newState.CurrentPlayer]

		// 如果是新玩家的回合，发送通知
		if nextPlayer.ID != client.ID {
			turnMsg := &protocol.YourTurn{
				BaseMessage: protocol.NewBaseMessage(protocol.MsgTypeYourTurn),
				PlayerID:    nextPlayer.ID,
				MinAction:   s.getMinAction(nextPlayer.ID),
				MaxAction:   s.getMaxAction(nextPlayer.ID),
				CurrentBet:  newState.CurrentBet,
				TimeLeft:    30, // 30秒超时
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
