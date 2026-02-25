package room

import (
	"card-game-server/internal/game"
	"card-game-server/internal/game/interfaces"
	"card-game-server/internal/types"
	"sync"
)

type Room struct {
	ID        string
	GameType  string
	Players   map[string]*Client
	Game      interfaces.Game
	mu        sync.RWMutex
	broadcast chan types.Message
}

type Client struct {
	PlayerID string
	Send     chan types.Message
}

func NewRoom(id, gameType string) *Room {
	r := &Room{
		ID:        id,
		GameType:  gameType,
		Players:   make(map[string]*Client),
		broadcast: make(chan types.Message, 512), // 加大缓冲
	}
	go r.runBroadcast()
	return r
}

// 【新增】统一全房间广播最新游戏状态（核心修复）
func (r *Room) broadcastState() {
	if r.Game == nil {
		return
	}
	stateMsg := types.Message{
		Type: types.MsgBroadcast,
		Data: types.BroadcastData{
			Event:   "state_update",
			Content: r.Game.GetState(),
		},
	}
	r.Broadcast(stateMsg)
}

func (r *Room) AddPlayer(playerID string, sendChan chan types.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) >= 4 {
		return ErrRoomFull
	}
	if _, exists := r.Players[playerID]; exists {
		return nil
	}

	r.Players[playerID] = &Client{PlayerID: playerID, Send: sendChan}

	// 人数够自动开始游戏
	if len(r.Players) >= 2 && r.Game == nil {
		if err := r.startGame(); err != nil {
			return err
		}
	}

	// 【关键】加入后全房间广播最新状态（包含新玩家列表）
	r.broadcastState()
	return nil
}

func (r *Room) startGame() error {
	g, err := game.NewGame(r.GameType)
	if err != nil {
		return err
	}
	playerIDs := make([]string, 0, len(r.Players))
	for id := range r.Players {
		playerIDs = append(playerIDs, id)
	}
	if err := g.Init(playerIDs); err != nil {
		return err
	}
	r.Game = g

	// 广播游戏开始 + 立即推送状态
	r.Broadcast(types.Message{
		Type: types.MsgBroadcast,
		Data: types.BroadcastData{Event: "game_started", Content: g.GetState()},
	})
	r.broadcastState() // 统一状态
	return nil
}

func (r *Room) Broadcast(msg types.Message) {
	r.broadcast <- msg
}

func (r *Room) runBroadcast() {
	for msg := range r.broadcast {
		r.mu.RLock()
		for _, c := range r.Players {
			select {
			case c.Send <- msg:
			default: // 防止单个客户端卡住影响他人
			}
		}
		r.mu.RUnlock()
	}
}

// 处理卡牌指令
func (r *Room) ProcessGameAction(playerID string, data types.GameActionData) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Game == nil {
		return ErrNoGame
	}

	action := interfaces.Action{Type: data.Action, Data: data.Card}
	turnEnded, err := r.Game.ProcessAction(playerID, action)
	if err != nil {
		return err
	}

	r.broadcastState() // 【关键】出牌后立即全房间广播新状态

	if turnEnded {
		r.Game.AdvanceTurn()
		r.broadcastState() // 回合切换后再广播一次
	}

	if r.Game.IsGameOver() {
		r.Broadcast(types.Message{
			Type: types.MsgBroadcast,
			Data: types.BroadcastData{Event: "game_over", Content: map[string]string{"winner": r.Game.Winner()}},
		})
	}
	return nil
}
