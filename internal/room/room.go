package room

import (
	"card-game-server/internal/game"
	"card-game-server/internal/game/interfaces"
	"card-game-server/internal/types"
	"fmt"
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
	g, err := game.NewGame(r.GameType)
	if err != nil {
		return nil
	}
	r.Game = g

	go r.runBroadcast()
	return r
}

func (r *Room) AddPlayer(playerID string, sendChan chan types.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) >= r.Game.MaxPlayers() {
		return ErrRoomFull
	}
	if _, exists := r.Players[playerID]; exists {
		return nil
	}

	r.Players[playerID] = &Client{PlayerID: playerID, Send: sendChan}

	// 广播玩家加入事件
	r.Broadcast(types.Message{
		Type: types.Broadcast,
		Data: types.BroadcastData{
			Event: types.PlayerJoined,
			Content: types.JoinRoomContent{
				RoomID:  r.ID,
				Message: fmt.Sprintf("玩家 %s 加入了房间", playerID),
			},
		},
	})

	// 人数够自动开始游戏
	if len(r.Players) >= r.Game.MinPlayers() && r.Game == nil {
		if err := r.startGame(); err != nil {
			return err
		}
	}

	return nil
}

func (r *Room) Broadcast(msg types.Message) {
	defer func() {
		if recover() != nil {
			// channel 已关闭，忽略
		}
	}()
	r.broadcast <- msg
}

// 移除玩家（用于断开连接或主动离开）
func (r *Room) RemovePlayer(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Players, playerID)

	// 广播玩家离开事件
	r.broadcast <- types.Message{
		Type: types.Broadcast,
		Data: types.BroadcastData{
			Event: types.PlayerLeft,
			Content: map[string]string{
				"player_id": playerID,
			},
		},
	}

	// 如果房间空了，清理资源
	if len(r.Players) == 0 {
		close(r.broadcast)
	}
	r.broadcastState() // 玩家离开后广播最新状态
}

// 处理卡牌指令
func (r *Room) ProcessGameAction(playerID string, data types.GameActionData) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Game == nil {
		return ErrNoGame
	}

	action := interfaces.Action{Type: string(data.Action), Data: data.Card} // 使用 Action 类型
	turnEnded, err := r.Game.ProcessAction(playerID, action)
	if err != nil {
		return err
	}

	r.broadcastState() // 出牌后立即全房间广播新状态

	if turnEnded {
		r.Game.AdvanceTurn()
		r.broadcastState() // 回合切换后再广播一次
	}

	if r.Game.IsGameOver() {
		r.Broadcast(types.Message{
			Type: types.Broadcast,
			Data: types.BroadcastData{Event: types.GameFinished, Content: map[string]string{"winner": r.Game.Winner()}},
		})
	}
	return nil
}

// 全房间广播最新游戏状态
func (r *Room) broadcastState() {
	if r.Game == nil {
		return
	}
	stateMsg := types.Message{
		Type: types.Broadcast,
		Data: types.BroadcastData{
			Event:   types.StateUpdate,
			Content: r.Game.GetState(),
		},
	}
	r.Broadcast(stateMsg)
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

	// 广播游戏开始（包含初始状态）
	r.Broadcast(types.Message{
		Type: types.Broadcast,
		Data: types.BroadcastData{Event: types.GameStarted, Content: g.GetState()},
	})
	return nil
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
