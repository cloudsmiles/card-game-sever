package room

import (
	"card-game-server/internal/game"
	"card-game-server/internal/game/interfaces"
	"card-game-server/internal/types"
	"fmt"
	"log"
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
		log.Printf("创建游戏失败 [房间：%s, 游戏类型：%s]: %v", id, gameType, err)
		return nil
	}
	r.Game = g

	log.Printf("房间创建成功 [房间：%s, 游戏类型：%s]", id, gameType)
	go r.runBroadcast()
	return r
}

func (r *Room) AddPlayer(playerID string, sendChan chan types.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Printf("尝试添加玩家 [房间：%s, 玩家：%s], 当前人数：%d/%d", r.ID, playerID, len(r.Players), r.Game.MaxPlayers())

	if len(r.Players) >= r.Game.MaxPlayers() {
		log.Printf("房间已满 [房间：%s]", r.ID)
		return ErrRoomFull
	}
	if _, exists := r.Players[playerID]; exists {
		log.Printf("玩家已在房间 [房间：%s, 玩家：%s]", r.ID, playerID)
		return nil
	}

	r.Players[playerID] = &Client{PlayerID: playerID, Send: sendChan}
	log.Printf("玩家加入成功 [房间：%s, 玩家：%s]", r.ID, playerID)

	// 广播玩家加入事件
	r.Broadcast(types.Message{
		Type: types.Broadcast,
		Data: types.BroadcastData{
			Event: types.PlayerJoined,
			Content: types.JoinRoomContent{
				RoomID:   r.ID,
				PlayerID: playerID,
				Message:  fmt.Sprintf("玩家 %s 加入了房间", playerID),
			},
		},
	})

	// 人数够自动开始游戏
	if len(r.Players) >= r.Game.MinPlayers() {
		log.Printf("达到最小玩家数，准备开始游戏 [房间：%s, 玩家数：%d]", r.ID, len(r.Players))
		if err := r.startGame(); err != nil {
			log.Printf("开始游戏失败 [房间：%s]: %v", r.ID, err)
			return err
		}
		log.Printf("游戏启动成功 [房间：%s]", r.ID)
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
	r.Broadcast(types.Message{
		Type: types.Broadcast,
		Data: types.BroadcastData{
			Event: types.PlayerLeft,
			Content: types.LeftRoomContent{
				RoomID:   r.ID,
				PlayerID: playerID,
				Message:  fmt.Sprintf("玩家 %s 离开了房间", playerID),
			},
		},
	})

	// 如果房间空了，清理资源
	if len(r.Players) == 0 {
		close(r.broadcast)
	}
}

// 处理卡牌指令
func (r *Room) ProcessGameAction(playerID string, data types.GameActionData) error {
	r.mu.Lock()
	defer r.mu.Unlock()

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
	log.Printf("初始化游戏 [房间：%s]", r.ID)
	playerIDs := make([]string, 0, len(r.Players))
	for id := range r.Players {
		playerIDs = append(playerIDs, id)
	}
	if err := r.Game.Init(playerIDs); err != nil {
		log.Printf("游戏初始化失败 [房间：%s]: %v", r.ID, err)
		return err
	}
	log.Printf("游戏初始化成功 [房间：%s], 玩家：%v", r.ID, playerIDs)

	// 广播游戏开始（包含初始状态）
	state := r.Game.GetState()
	log.Printf("广播游戏开始 [房间：%s], 状态：%+v", r.ID, state)
	r.Broadcast(types.Message{
		Type: types.Broadcast,
		Data: types.BroadcastData{Event: types.GameStarted, Content: state},
	})
	log.Printf("游戏开始广播完成 [房间：%s]", r.ID)
	return nil
}

func (r *Room) runBroadcast() {
	for msg := range r.broadcast {
		r.mu.RLock()
		sentCount := 0
		for _, c := range r.Players {
			select {
			case c.Send <- msg:
				sentCount++
				log.Printf("广播消息发送给玩家 [%s] [房间：%s], 事件：%v", c.PlayerID, r.ID, msg.Data.(types.BroadcastData).Event)
			default: // 防止单个客户端卡住影响他人
				log.Printf("警告：玩家 [%s] 的消息队列已满，丢弃消息 [房间：%s]", c.PlayerID, r.ID)
			}
		}
		r.mu.RUnlock()
		if sentCount == 0 {
			log.Printf("警告：房间 [%s] 没有成功发送任何消息，当前玩家数：%d", r.ID, len(r.Players))
		}
	}
}
