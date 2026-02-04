package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/wilenwang/just_play/Texas-Holdem/internal/card"
)

// 颜色定义
var (
	// 牌面颜色
	suitRed    = lipgloss.Color("196")    // 红桃、方块 - 亮红色
	suitBlack  = lipgloss.Color("15")     // 黑桃、梅花 - 亮白色
	backColor  = lipgloss.Color("239")    // 牌背 - 深灰色
	bgColor    = lipgloss.Color("236")    // 背景色

	// 边框颜色
	borderColor     = lipgloss.Color("240") // 边框
	highlightColor  = lipgloss.Color("214")  // 高亮
	selectedColor   = lipgloss.Color("39")  // 选中状态
)

// CardStyle 扑克牌渲染样式
type CardStyle struct {
	Width       int           // 牌宽度
	Height      int           // 牌高度
	Border      lipgloss.Border // 边框样式
	CornerSize  int           // 圆角大小
	ShowRank    bool          // 显示点数
	ShowSuit    bool          // 显示花色
	Compact     bool          // 紧凑模式
}

// 默认样式
var defaultCardStyle = CardStyle{
	Width:      5,
	Height:     4,
	Border:     lipgloss.RoundedBorder(),
	CornerSize: 1,
	ShowRank:   true,
	ShowSuit:   true,
	Compact:    false,
}

// RenderCard 渲染单张扑克牌（带颜色）
func RenderCard(c card.Card, faceUp bool) string {
	if !faceUp || c.Rank == 0 {
		return RenderCardBack()
	}

	// 根据花色选择颜色
	suitColor := suitBlack
	if c.Suit == card.Hearts || c.Suit == card.Diamonds {
		suitColor = suitRed
	}

	// 点数和花色符号
	rankStr := RankToString(c.Rank)
	suitStr := SuitToChar(c.Suit)

	// 构建牌面内容
	var content strings.Builder

	if defaultCardStyle.ShowRank {
		content.WriteString(rankStr)
	}
	if defaultCardStyle.ShowRank && defaultCardStyle.ShowSuit {
		content.WriteString("\n")
	}
	if defaultCardStyle.ShowSuit {
		content.WriteString(suitStr)
	}

	// 应用样式
	style := lipgloss.NewStyle().
		Border(defaultCardStyle.Border).
		BorderForeground(borderColor).
		Width(defaultCardStyle.Width).
		Height(defaultCardStyle.Height).
		Padding(0, 1)

	// 内部样式（点数和花色）
	innerStyle := lipgloss.NewStyle().
		Foreground(suitColor).
		Bold(true).
		Align(lipgloss.Center)

	return style.Render(innerStyle.Render(content.String()))
}

// RenderCardBack 渲染牌背
func RenderCardBack() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(backColor).
		Width(defaultCardStyle.Width).
		Height(defaultCardStyle.Height).
		Align(lipgloss.Center).
		Bold(true)

	return style.Render("??")
}

// RenderCards 渲染多张扑克牌（水平排列）
func RenderCards(cards []card.Card, faceUp bool) string {
	if len(cards) == 0 {
		return ""
	}

	var rendered []string
	for _, c := range cards {
		rendered = append(rendered, RenderCard(c, faceUp))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

// RenderCardsCompact 紧凑模式渲染多张牌
func RenderCardsCompact(cards []card.Card, faceUp bool) string {
	if len(cards) == 0 {
		return ""
	}

	var rendered []string
	for _, c := range cards {
		rendered = append(rendered, RenderCardCompact(c, faceUp))
	}
	return strings.Join(rendered, " ")
}

// RenderCardCompact 紧凑模式渲染单张牌
func RenderCardCompact(c card.Card, faceUp bool) string {
	if !faceUp || c.Rank == 0 {
		return "[??]"
	}

	suitColor := suitBlack
	if c.Suit == card.Hearts || c.Suit == card.Diamonds {
		suitColor = suitRed
	}

	rankStr := RankToString(c.Rank)
	suitStr := SuitToChar(c.Suit)

	return lipgloss.NewStyle().
		Foreground(suitColor).
		Render(fmt.Sprintf("[%s%s]", rankStr, suitStr))
}

// RenderCardASCII ASCII版本渲染（无颜色，适合不支持颜色的终端）
func RenderCardASCII(c card.Card, faceUp bool) string {
	if !faceUp || c.Rank == 0 {
		return "[??]"
	}
	return fmt.Sprintf("[%s%s]", RankToString(c.Rank), SuitToChar(c.Suit))
}

// RenderCardsASCII ASCII版本渲染多张牌
func RenderCardsASCII(cards []card.Card, faceUp bool) string {
	if len(cards) == 0 {
		return ""
	}

	var parts []string
	for _, c := range cards {
		parts = append(parts, RenderCardASCII(c, faceUp))
	}
	return strings.Join(parts, " ")
}

// RenderCardLarge 大尺寸渲染单张牌（带更多细节）
func RenderCardLarge(c card.Card, faceUp bool) string {
	if !faceUp || c.Rank == 0 {
		return RenderCardBackLarge()
	}

	suitColor := suitBlack
	if c.Suit == card.Hearts || c.Suit == card.Diamonds {
		suitColor = suitRed
	}

	rankStr := RankToString(c.Rank)
	suitStr := SuitToChar(c.Suit)
	suitName := SuitToName(c.Suit)

	// 牌面样式
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(8).
		Height(6).
		Padding(0, 1)

	// 内部内容
	var content strings.Builder

	// 左上角
	content.WriteString(rankStr + suitStr)
	// 右上角
	content.WriteString(fmt.Sprintf("%*s", 8-len(rankStr)-len(suitStr), rankStr+suitStr))
	// 中间大符号
	content.WriteString(fmt.Sprintf("\n\n%*s", 5, ""))
	content.WriteString(lipgloss.NewStyle().
		Foreground(suitColor).
		Bold(true).
		Render(suitStr + suitName[0:1]))
	// 右下角
	content.WriteString(fmt.Sprintf("\n%*s", 8-len(rankStr)-len(suitStr), rankStr+suitStr))

	return style.Render(lipgloss.NewStyle().
		Foreground(suitColor).
		Bold(true).
		Align(lipgloss.Center).
		Render(content.String()))
}

// RenderCardBackLarge 大尺寸牌背
func RenderCardBackLarge() string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(backColor).
		Width(8).
		Height(6).
		Align(lipgloss.Center).
		Bold(true).
		Render("牌背")
}

// RankToString 将点数转换为字符串
func RankToString(r card.Rank) string {
	names := map[card.Rank]string{
		card.Two:   "2",
		card.Three: "3",
		card.Four:  "4",
		card.Five:  "5",
		card.Six:   "6",
		card.Seven: "7",
		card.Eight: "8",
		card.Nine:  "9",
		card.Ten:   "10",
		card.Jack:  "J",
		card.Queen: "Q",
		card.King:  "K",
		card.Ace:   "A",
	}
	return names[r]
}

// SuitToChar 将花色转换为单个字符
func SuitToChar(s card.Suit) string {
	chars := map[card.Suit]string{
		card.Clubs:    "♣",
		card.Diamonds: "♦",
		card.Hearts:   "♥",
		card.Spades:   "♠",
	}
	return chars[s]
}

// SuitToName 将花色转换为完整名称
func SuitToName(s card.Suit) string {
	names := map[card.Suit]string{
		card.Clubs:    "梅花",
		card.Diamonds: "方块",
		card.Hearts:   "红桃",
		card.Spades:   "黑桃",
	}
	return names[s]
}

// RenderCommunityCards 渲染公共牌区域
func RenderCommunityCards(cards [5]card.Card, faceUp bool) string {
	var visibleCards []card.Card
	for _, c := range cards {
		if c.Rank != 0 {
			visibleCards = append(visibleCards, c)
		}
	}

	if len(visibleCards) == 0 {
		return "公共牌: [等待发牌]"
	}

	// 根据阶段显示不同数量的牌
	var displayCards []card.Card
	stageNames := []string{"翻牌前", "翻牌圈", "转牌圈", "河牌圈", "摊牌"}

	for i, c := range cards {
		if c.Rank != 0 || i < len(visibleCards) {
			displayCards = append(displayCards, c)
		}
	}

	return fmt.Sprintf("公共牌 (%s): %s",
		stageNames[len(visibleCards)],
		RenderCardsCompact(displayCards, faceUp))
}

// RenderPot 渲染底池信息
func RenderPot(pot int) string {
	return fmt.Sprintf("底池: %s%d", highlightStyle().Render(""), pot)
}

// RenderPotLarge 大尺寸底池显示
func RenderPotLarge(pot int) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(highlightColor).
		Padding(0, 2).
		Render(fmt.Sprintf("底池: %d", pot))
}

// RenderPlayerInfo 渲染玩家信息行
func RenderPlayerInfo(name string, chips int, status string) string {
	return fmt.Sprintf("%s | 筹码: %d | %s",
		lipgloss.NewStyle().Bold(true).Render(name),
		chips,
		status)
}

// highlightStyle 返回高亮样式
func highlightStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(highlightColor).
		Bold(true)
}

// selectedStyle 返回选中样式
func selectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderForeground(selectedColor).
		Bold(true)
}

// GetCardColor 获取牌面的颜色
func GetCardColor(c card.Card) lipgloss.Color {
	if c.Suit == card.Hearts || c.Suit == card.Diamonds {
		return suitRed
	}
	return suitBlack
}

// RenderCardWithStyle 使用自定义样式渲染牌
func RenderCardWithStyle(c card.Card, faceUp bool, style CardStyle) string {
	if !faceUp || c.Rank == 0 {
		return renderCardBackWithStyle(style)
	}

	suitColor := suitBlack
	if c.Suit == card.Hearts || c.Suit == card.Diamonds {
		suitColor = suitRed
	}

	rankStr := RankToString(c.Rank)
	suitStr := SuitToChar(c.Suit)

	var content strings.Builder
	if style.ShowRank {
		content.WriteString(rankStr)
	}
	if style.ShowRank && style.ShowSuit {
		content.WriteString("\n")
	}
	if style.ShowSuit {
		content.WriteString(suitStr)
	}

	borderStyle := lipgloss.NewStyle().
		Border(style.Border).
		BorderForeground(borderColor).
		Width(style.Width).
		Height(style.Height).
		Padding(0, 1)

	innerStyle := lipgloss.NewStyle().
		Foreground(suitColor).
		Bold(true)

	if style.Compact {
		innerStyle = innerStyle.Align(lipgloss.Left)
	} else {
		innerStyle = innerStyle.Align(lipgloss.Center)
	}

	return borderStyle.Render(innerStyle.Render(content.String()))
}

// renderCardBackWithStyle 使用自定义样式渲染牌背
func renderCardBackWithStyle(style CardStyle) string {
	return lipgloss.NewStyle().
		Border(style.Border).
		BorderForeground(borderColor).
		Background(backColor).
		Width(style.Width).
		Height(style.Height).
		Align(lipgloss.Center).
		Bold(true).
		Render("??")
}
