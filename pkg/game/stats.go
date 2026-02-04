package game

import (
	"fmt"
	"sync"
	"time"

	"github.com/wilenwang/just_play/Texas-Holdem/internal/common/models"
)

// PlayerStats 玩家统计信息
type PlayerStats struct {
	PlayerID      string    `json:"player_id"`       // 玩家ID
	Name          string    `json:"name"`            // 玩家名称
	HandsPlayed   int       `json:"hands_played"`    // 参与的手牌数
	HandsWon      int       `json:"hands_won"`       // 获胜的手牌数
	TotalWinnings int       `json:"total_winnings"`  // 总盈利
	TotalLosses  int       `json:"total_losses"`    // 总亏损
	BiggestPot    int       `json:"biggest_pot"`     // 最大底池
	TotalBets     int       `json:"total_bets"`     // 总下注金额
	TotalCalls    int       `json:"total_calls"`    // 跟注次数
	TotalRaises   int       `json:"total_raises"`   // 加注次数
	TotalFolds    int       `json:"total_folds"`    // 弃牌次数
	TotalChecks   int       `json:"total_checks"`   // 看牌次数
	WinRate       float64   `json:"win_rate"`       // 胜率 (获胜手牌数/参与手牌数)
	ProfitPerHand float64   `json:"profit_per_hand"` // 每手平均盈利
	FirstActions  int      `json:"first_actions"`  // 首位行动次数
	LastActions   int      `json:"last_actions"`  // 最后行动次数
	CreatedAt     time.Time `json:"created_at"`    // 统计开始时间
	UpdatedAt     time.Time `json:"updated_at"`    // 最后更新时间
}

// StatsManager 管理所有玩家的统计数据
type StatsManager struct {
	mu          sync.RWMutex             // 读写锁
	playerStats map[string]*PlayerStats   // 玩家ID到统计信息的映射
}

// NewStatsManager 创建统计管理器
func NewStatsManager() *StatsManager {
	return &StatsManager{
		playerStats: make(map[string]*PlayerStats),
	}
}

// GetOrCreateStats 获取或创建玩家统计信息
func (s *StatsManager) GetOrCreateStats(playerID, name string) *PlayerStats {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats, exists := s.playerStats[playerID]
	if !exists {
		now := time.Now()
		stats = &PlayerStats{
			PlayerID:  playerID,
			Name:      name,
			CreatedAt: now,
			UpdatedAt: now,
		}
		s.playerStats[playerID] = stats
	}

	return stats
}

// UpdateHandPlayed 更新玩家参与的手牌数
func (s *StatsManager) UpdateHandPlayed(playerID, name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats, exists := s.playerStats[playerID]
	if !exists {
		now := time.Now()
		stats = &PlayerStats{
			PlayerID:  playerID,
			Name:      name,
			CreatedAt: now,
			UpdatedAt: now,
		}
		s.playerStats[playerID] = stats
	}

	stats.HandsPlayed++
	stats.UpdatedAt = time.Now()
}

// UpdateHandWon 更新玩家获胜信息
func (s *StatsManager) UpdateHandWon(playerID, name string, wonChips int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats, exists := s.playerStats[playerID]
	if !exists {
		now := time.Now()
		stats = &PlayerStats{
			PlayerID:  playerID,
			Name:      name,
			CreatedAt: now,
			UpdatedAt: now,
		}
		s.playerStats[playerID] = stats
	}

	stats.HandsWon++
	stats.TotalWinnings += wonChips

	// 更新最大底池
	if wonChips > stats.BiggestPot {
		stats.BiggestPot = wonChips
	}

	// 更新胜率和每手平均盈利
	if stats.HandsPlayed > 0 {
		stats.WinRate = float64(stats.HandsWon) / float64(stats.HandsPlayed)
		stats.ProfitPerHand = float64(stats.TotalWinnings-stats.TotalLosses) / float64(stats.HandsPlayed)
	}

	stats.UpdatedAt = time.Now()
}

// UpdateHandLost 更新玩家亏损信息
func (s *StatsManager) UpdateHandLost(playerID, name string, lostChips int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats, exists := s.playerStats[playerID]
	if !exists {
		now := time.Now()
		stats = &PlayerStats{
			PlayerID:  playerID,
			Name:      name,
			CreatedAt: now,
			UpdatedAt: now,
		}
		s.playerStats[playerID] = stats
	}

	stats.TotalLosses += lostChips

	// 更新每手平均盈利
	if stats.HandsPlayed > 0 {
		stats.ProfitPerHand = float64(stats.TotalWinnings-stats.TotalLosses) / float64(stats.HandsPlayed)
	}

	stats.UpdatedAt = time.Now()
}

// UpdateAction 记录玩家动作
func (s *StatsManager) UpdateAction(playerID, name string, action models.ActionType, amount int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats, exists := s.playerStats[playerID]
	if !exists {
		now := time.Now()
		stats = &PlayerStats{
			PlayerID:  playerID,
			Name:      name,
			CreatedAt: now,
			UpdatedAt: now,
		}
		s.playerStats[playerID] = stats
	}

	switch action {
	case models.ActionFold:
		stats.TotalFolds++
	case models.ActionCheck:
		stats.TotalChecks++
	case models.ActionCall:
		stats.TotalCalls++
		stats.TotalBets += amount
	case models.ActionRaise:
		stats.TotalRaises++
		stats.TotalBets += amount
	case models.ActionAllIn:
		stats.TotalBets += amount
	}

	stats.UpdatedAt = time.Now()
}

// GetStats 获取玩家统计信息（线程安全）
func (s *StatsManager) GetStats(playerID string) *PlayerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if stats, exists := s.playerStats[playerID]; exists {
		return stats
	}
	return nil
}

// GetAllStats 获取所有玩家的统计信息
func (s *StatsManager) GetAllStats() []*PlayerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make([]*PlayerStats, 0, len(s.playerStats))
	for _, s := range s.playerStats {
		stats = append(stats, s)
	}
	return stats
}

// GetLeaderboard 获取排行榜（按盈利排序）
func (s *StatsManager) GetLeaderboard(limit int) []*PlayerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make([]*PlayerStats, 0, len(s.playerStats))
	for _, s := range s.playerStats {
		stats = append(stats, s)
	}

	// 按总盈利排序（盈利-亏损）
	for i := 0; i < len(stats); i++ {
		for j := i + 1; j < len(stats); j++ {
			profitI := stats[i].TotalWinnings - stats[i].TotalLosses
			profitJ := stats[j].TotalWinnings - stats[j].TotalLosses
			if profitJ > profitI {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}

	if limit > 0 && limit < len(stats) {
		return stats[:limit]
	}
	return stats
}

// GetWinRateLeaderboard 获取胜率排行榜（至少参与10手牌）
func (s *StatsManager) GetWinRateLeaderboard(limit int) []*PlayerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make([]*PlayerStats, 0, len(s.playerStats))
	for _, s := range s.playerStats {
		if s.HandsPlayed >= 10 { // 至少参与10手牌才有资格
			stats = append(stats, s)
		}
	}

	// 按胜率排序
	for i := 0; i < len(stats); i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[j].WinRate > stats[i].WinRate {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}

	if limit > 0 && limit < len(stats) {
		return stats[:limit]
	}
	return stats
}

// GetTopWinner 获取最大赢家
func (s *StatsManager) GetTopWinner() *PlayerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var topWinner *PlayerStats
	var topProfit int

	for _, stats := range s.playerStats {
		profit := stats.TotalWinnings - stats.TotalLosses
		if topWinner == nil || profit > topProfit {
			topWinner = stats
			topProfit = profit
		}
	}

	return topWinner
}

// GetTopLoser 获取最大输家
func (s *StatsManager) GetTopLoser() *PlayerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var topLoser *PlayerStats
	var lowestProfit int

	for _, stats := range s.playerStats {
		profit := stats.TotalWinnings - stats.TotalLosses
		if topLoser == nil || profit < lowestProfit {
			topLoser = stats
			lowestProfit = profit
		}
	}

	return topLoser
}

// RemoveStats 删除玩家统计信息
func (s *StatsManager) RemoveStats(playerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.playerStats, playerID)
}

// Clear 清空所有统计信息
func (s *StatsManager) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.playerStats = make(map[string]*PlayerStats)
}

// GetTotalHandsPlayed 获取所有玩家参与的手牌总数
func (s *StatsManager) GetTotalHandsPlayed() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := 0
	for _, stats := range s.playerStats {
		total += stats.HandsPlayed
	}
	return total
}

// GetTotalPotSize 计算所有底池总金额
func (s *StatsManager) GetTotalPotSize() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := 0
	for _, stats := range s.playerStats {
		total += stats.TotalBets
	}
	return total / 2 // 底池是所有下注的一半（两个玩家）
}

// Report 生成玩家统计报告
func (p *PlayerStats) Report() string {
	profit := p.TotalWinnings - p.TotalLosses
	profitSign := ""
	if profit > 0 {
		profitSign = "+"
	}

	return fmt.Sprintf(`=== %s 统计 ===
参与手牌: %d
获胜手牌: %d (%.1f%%)
总盈利: %s%d
每手平均: %s%.2f
最大底池: %d
总下注: %d
跟注次数: %d
加注次数: %d
弃牌次数: %d
看牌次数: %d
统计时间: %s - %s`,
		p.Name,
		p.HandsPlayed,
		p.HandsWon,
		p.WinRate*100,
		profitSign, profit,
		profitSign, p.ProfitPerHand,
		p.BiggestPot,
		p.TotalBets,
		p.TotalCalls,
		p.TotalRaises,
		p.TotalFolds,
		p.TotalChecks,
		p.CreatedAt.Format("2006-01-02"),
		p.UpdatedAt.Format("2006-01-02 15:04"))
}

// ShortReport 生成简短的玩家统计报告
func (p *PlayerStats) ShortReport() string {
	profit := p.TotalWinnings - p.TotalLosses
	profitSign := ""
	if profit > 0 {
		profitSign = "+"
	}

	return fmt.Sprintf(`%s: %d手 %d胜 %.1f%% %s%d`,
		p.Name,
		p.HandsPlayed,
		p.HandsWon,
		p.WinRate*100,
		profitSign, profit)
}
