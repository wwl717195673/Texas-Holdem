# 德州扑克游戏 - 技术方案

## 文档信息

| 项目 | 内容 |
|-----|------|
| 项目名称 | Texas Hold'em Online |
| 框架 | Bubble Tea (charmbracelet/bubbletea) |
| 协议 | WebSocket |
| 语言 | Go 1.21+ |
| 文档版本 | v1.0 |

## 文档结构

| 文档 | 描述 |
|-----|------|
| [README.md](README.md) | 整体架构概览、Phase 1-2 详细设计 |
| [Phase_1.md](Phase_1.md) | Phase 1: 核心游戏引擎 |
| [Phase_2.md](Phase_2.md) | Phase 2: WebSocket 通信层 |
| [Phase_3.md](Phase_3.md) | Phase 3: 高级功能 |

## 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                     德州扑克游戏架构                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌──────────────┐          WebSocket         ┌──────────────┐ │
│   │   HOST       │◄──────────────────────────►│   CLIENT     │ │
│   │  (服务端)    │                              │  (玩家端)    │ │
│   │              │                              │              │ │
│   │ • 游戏引擎    │                              │ • TUI界面    │ │
│   │ • 状态管理    │                              │ • 输入处理    │ │
│   │ • 牌型评估    │                              │ • 状态显示    │ │
│   │ • WebSocket  │                              │              │ │
│   └──────────────┘                              └──────────────┘ │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## 实现阶段

### Phase 1: 核心游戏引擎
- 扑克牌表示与管理
- 牌型评估器 (10种牌型)
- 游戏引擎 (状态机、下注逻辑)
- 单元测试

### Phase 2: 通信层与UI
- WebSocket 服务器与客户端
- HOST TUI 界面
- CLIENT TUI 界面
- 实时状态同步

### Phase 3: 高级功能
- AI 玩家 (4种策略)
- 历史记录
- 聊天系统
- 游戏统计

## 快速开始

```bash
# Phase 1: 运行单机游戏
go run ./cmd/solo/main.go

# Phase 2: 启动服务器
go run ./cmd/host/main.go --port 8080

# Phase 2: 启动客户端
go run ./cmd/client/main.go --server localhost:8080
```

## 技术栈

- **UI框架**: Bubble Tea + Lipgloss
- **通信**: gorilla/websocket
- **配置**: Viper / Flag

## 依赖

```
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
go get github.com/gorilla/websocket
```

---

*文档版本: v1.0*
*创建日期: 2026-02-04*
