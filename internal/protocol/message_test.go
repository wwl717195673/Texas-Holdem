package protocol

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/wilenwang/just_play/Texas-Holdem/internal/card"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/common/models"
	"github.com/wilenwang/just_play/Texas-Holdem/pkg/evaluator"
	"github.com/wilenwang/just_play/Texas-Holdem/pkg/game"
)

// TestNewBaseMessage 测试创建基本消息
func TestNewBaseMessage(t *testing.T) {
	msg := NewBaseMessage(MsgTypeJoin)

	if msg.Type != MsgTypeJoin {
		t.Errorf("Expected type %s, got %s", MsgTypeJoin, msg.Type)
	}

	if msg.Timestamp == 0 {
		t.Error("Expected timestamp to be set")
	}
}

// TestNewJoinRequest 测试创建加入请求
func TestNewJoinRequest(t *testing.T) {
	req := NewJoinRequest("TestPlayer", 0)

	if req.PlayerName != "TestPlayer" {
		t.Errorf("Expected player name 'TestPlayer', got '%s'", req.PlayerName)
	}

	if req.Seat != 0 {
		t.Errorf("Expected seat 0, got %d", req.Seat)
	}

	if req.Type != MsgTypeJoin {
		t.Errorf("Expected type %s, got %s", MsgTypeJoin, req.Type)
	}
}

// TestNewPlayerActionRequest 测试创建动作请求
func TestNewPlayerActionRequest(t *testing.T) {
	req := NewPlayerActionRequest("player123", models.ActionCall, 100)

	if req.PlayerID != "player123" {
		t.Errorf("Expected player ID 'player123', got '%s'", req.PlayerID)
	}

	if req.Action != models.ActionCall {
		t.Errorf("Expected action %v, got %v", models.ActionCall, req.Action)
	}

	if req.Amount != 100 {
		t.Errorf("Expected amount 100, got %d", req.Amount)
	}
}

// TestNewChatRequest 测试创建聊天请求
func TestNewChatRequest(t *testing.T) {
	req := NewChatRequest("player123", "Hello!")

	if req.PlayerID != "player123" {
		t.Errorf("Expected player ID 'player123', got '%s'", req.PlayerID)
	}

	if req.Content != "Hello!" {
		t.Errorf("Expected content 'Hello!', got '%s'", req.Content)
	}
}

// TestNewPingRequest 测试创建心跳请求
func TestNewPingRequest(t *testing.T) {
	req := NewPingRequest()

	if req.Type != MsgTypePing {
		t.Errorf("Expected type %s, got %s", MsgTypePing, req.Type)
	}
}

// TestJoinAck_JSON 测试 JoinAck 序列化
func TestJoinAck_JSON(t *testing.T) {
	ack := &JoinAck{
		BaseMessage: NewBaseMessage(MsgTypeJoinAck),
		Success:     true,
		PlayerID:   "player123",
		Seat:       3,
		Message:    "Welcome!",
		GameState:  nil,
	}

	data, err := json.Marshal(ack)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded JoinAck
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Success != ack.Success {
		t.Errorf("Expected success %v, got %v", ack.Success, decoded.Success)
	}

	if decoded.PlayerID != ack.PlayerID {
		t.Errorf("Expected player ID '%s', got '%s'", ack.PlayerID, decoded.PlayerID)
	}

	if decoded.Seat != ack.Seat {
		t.Errorf("Expected seat %d, got %d", ack.Seat, decoded.Seat)
	}
}

// TestGameState_JSON 测试 GameState 序列化
func TestGameState_JSON(t *testing.T) {
	state := &GameState{
		BaseMessage:     NewBaseMessage(MsgTypeGameState),
		GameID:         "game123",
		Stage:          game.StageFlop,
		DealerButton:   0,
		CurrentPlayer: 1,
		CurrentBet:    100,
		Pot:           500,
		CommunityCards: [5]card.Card{
			card.NewCard(card.Hearts, card.Ace),
			card.NewCard(card.Hearts, card.King),
			card.NewCard(card.Hearts, card.Queen),
			{},
			{},
		},
		Players: []PlayerInfo{
			{
				ID:        "player1",
				Name:      "Alice",
				Seat:      0,
				Chips:     1000,
				CurrentBet: 50,
				Status:    models.PlayerStatusActive,
				IsDealer:  true,
				IsSelf:    false,
			},
		},
		MinRaise: 200,
		MaxRaise: 1000,
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded GameState
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.GameID != state.GameID {
		t.Errorf("Expected game ID '%s', got '%s'", state.GameID, decoded.GameID)
	}

	if decoded.Pot != state.Pot {
		t.Errorf("Expected pot %d, got %d", state.Pot, decoded.Pot)
	}

	if len(decoded.Players) != len(state.Players) {
		t.Errorf("Expected %d players, got %d", len(state.Players), len(decoded.Players))
	}
}

// TestYourTurn 测试 YourTurn 结构
func TestYourTurn(t *testing.T) {
	turn := &YourTurn{
		BaseMessage: NewBaseMessage(MsgTypeYourTurn),
		PlayerID:   "player1",
		MinAction:  50,
		MaxAction:  500,
		CurrentBet: 100,
		TimeLeft:   30,
	}

	if turn.PlayerID != "player1" {
		t.Errorf("Expected player ID 'player1', got '%s'", turn.PlayerID)
	}

	if turn.MinAction != 50 {
		t.Errorf("Expected min action 50, got %d", turn.MinAction)
	}

	if turn.TimeLeft != 30 {
		t.Errorf("Expected time left 30, got %d", turn.TimeLeft)
	}
}

// TestShowdown 测试 Showdown 结构
func TestShowdown(t *testing.T) {
	showdown := &Showdown{
		BaseMessage: NewBaseMessage(MsgTypeShowdown),
		Winners: []WinnerInfo{
			{
				PlayerID:   "player1",
				PlayerName: "Alice",
				HandRank:   evaluator.RankFlush,
				HandName:   "同花",
				WonChips:   500,
				RawCards: []card.Card{
					card.NewCard(card.Hearts, card.Ace),
					card.NewCard(card.Hearts, card.King),
				},
			},
		},
		Pot: 500,
	}

	if len(showdown.Winners) != 1 {
		t.Errorf("Expected 1 winner, got %d", len(showdown.Winners))
	}

	if showdown.Winners[0].PlayerName != "Alice" {
		t.Errorf("Expected winner name 'Alice', got '%s'", showdown.Winners[0].PlayerName)
	}

	if showdown.Winners[0].WonChips != 500 {
		t.Errorf("Expected won chips 500, got %d", showdown.Winners[0].WonChips)
	}
}

// TestChatMessage 测试 ChatMessage 结构
func TestChatMessage(t *testing.T) {
	msg := &ChatMessage{
		BaseMessage: NewBaseMessage(MsgTypeChat),
		PlayerID:   "player1",
		PlayerName: "Alice",
		Content:    "Hello everyone!",
		IsSystem:   false,
	}

	if msg.PlayerID != "player1" {
		t.Errorf("Expected player ID 'player1', got '%s'", msg.PlayerID)
	}

	if msg.Content != "Hello everyone!" {
		t.Errorf("Expected content 'Hello everyone!', got '%s'", msg.Content)
	}

	if msg.IsSystem {
		t.Error("Expected IsSystem to be false")
	}
}

// TestError 测试 Error 结构
func TestError(t *testing.T) {
	err := &Error{
		BaseMessage: NewBaseMessage(MsgTypeError),
		Code:    1001,
		Message: "Invalid move",
	}

	if err.Code != 1001 {
		t.Errorf("Expected code 1001, got %d", err.Code)
	}

	if err.Message != "Invalid move" {
		t.Errorf("Expected message 'Invalid move', got '%s'", err.Message)
	}
}

// TestPong 测试 Pong 结构
func TestPong(t *testing.T) {
	now := time.Now().UnixMilli()
	pong := &Pong{
		BaseMessage: NewBaseMessage(MsgTypePong),
		ServerTime:  now,
	}

	if pong.ServerTime != now {
		t.Errorf("Expected server time %d, got %d", now, pong.ServerTime)
	}
}
