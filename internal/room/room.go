package room

import (
	"card-game-server/internal/game"
	"card-game-server/internal/game/interfaces"
	"card-game-server/internal/types"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

type Room struct {
	ID         string
	GameType   string
	State      types.RoomState
	Players    map[string]*Client // playerID -> Client
	Seats      map[int]string     // seatNumber -> playerID
	Ready      map[string]bool    // playerID -> ready status
	Game       interfaces.Game
	mu         sync.RWMutex
	broadcast  chan types.Message
	lastAction map[string]int64 // playerID -> 上次操作时间戳（毫秒）
}

type Client struct {
	PlayerID   string
	SeatNumber int // 座位号
	Send       chan types.Message
}

func NewRoom(id, gameType string) *Room {
	r := &Room{
		ID:         id,
		GameType:   gameType,
		State:      types.RoomWaiting,
		Players:    make(map[string]*Client),
		Seats:      make(map[int]string),
		Ready:      make(map[string]bool),
		broadcast:  make(chan types.Message, 512), // 加大缓冲
		lastAction: make(map[string]int64),
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

	// 游戏中不能加入
	if r.State == types.RoomPlaying {
		return fmt.Errorf("游戏进行中，无法加入")
	}

	if len(r.Players) >= r.Game.MaxPlayers() {
		log.Printf("房间已满 [房间：%s]", r.ID)
		return ErrRoomFull
	}
	if _, exists := r.Players[playerID]; exists {
		log.Printf("玩家已在房间 [房间：%s, 玩家：%s]", r.ID, playerID)
		return nil
	}

	// 分配座位号（第一个空座位）
	seatNumber := -1
	for i := 0; i < r.Game.MaxPlayers(); i++ {
		if _, taken := r.Seats[i]; !taken {
			seatNumber = i
			r.Seats[i] = playerID
			break
		}
	}

	r.Players[playerID] = &Client{PlayerID: playerID, SeatNumber: seatNumber, Send: sendChan}
	log.Printf("玩家加入成功 [房间：%s, 玩家：%s, 座位：%d]", r.ID, playerID, seatNumber)

	// 广播房间状态变更（包含玩家加入）
	r.broadcastRoomStateWithMessage(fmt.Sprintf("玩家 %s 加入了房间", playerID))

	return nil
}

func (r *Room) Broadcast(msg types.Message) {
	defer func() {
		if recover() != nil {
			// channel 已关闭，忽略
			log.Printf("警告：向已关闭的广播 channel 发送消息被忽略 [房间：%s]", r.ID)
		}
	}()
	r.broadcast <- msg
}

// 移除玩家（用于断开连接或主动离开）
func (r *Room) RemovePlayer(playerID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 释放座位
	client, exists := r.Players[playerID]
	if exists && client.SeatNumber >= 0 {
		delete(r.Seats, client.SeatNumber)
	}
	delete(r.Players, playerID)
	delete(r.Ready, playerID)
	delete(r.lastAction, playerID)

	// 广播房间状态变更（包含玩家离开）
	r.broadcastRoomStateWithMessage(fmt.Sprintf("玩家 %s 离开了房间", playerID))

	// 如果房间空了，清理资源
	isEmpty := len(r.Players) == 0
	if isEmpty {
		close(r.broadcast)
		log.Printf("房间 [%s] 已空，资源已清理", r.ID)
	}
	return isEmpty
}

// ProcessRoomAction 统一处理房间内操作
func (r *Room) ProcessRoomAction(playerID string, action types.RoomActionData) error {
	switch action.Action {
	case types.RoomActionReady:
		// 解析准备操作数据
		readyDataBytes, _ := json.Marshal(action.Data)
		var readyData types.RoomActionReadyData
		json.Unmarshal(readyDataBytes, &readyData)
		return r.setPlayerReady(playerID, readyData.Ready)

	case types.RoomActionSit:
		// 解析选座操作数据
		sitDataBytes, _ := json.Marshal(action.Data)
		var sitData types.RoomActionSitData
		json.Unmarshal(sitDataBytes, &sitData)
		return r.setPlayerSeat(playerID, sitData.SeatNumber)

	default:
		return fmt.Errorf("未知的房间操作: %s", action.Action)
	}
}

// 玩家准备/取消准备
func (r *Room) setPlayerReady(playerID string, ready bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State == types.RoomPlaying {
		return fmt.Errorf("游戏进行中，无法更改准备状态")
	}

	if _, exists := r.Players[playerID]; !exists {
		return fmt.Errorf("玩家不在房间中")
	}

	r.Ready[playerID] = ready
	log.Printf("玩家准备状态变更 [房间：%s, 玩家：%s, 准备：%v]", r.ID, playerID, ready)

	// 广播房间状态变更（包含准备状态变化）
	r.broadcastRoomStateWithMessage(fmt.Sprintf("玩家 %s %s", playerID, map[bool]string{true: "已准备", false: "取消准备"}[ready]))

	// 检查是否所有玩家都准备好了
	if r.canStartGame() {
		log.Printf("所有玩家已准备，开始游戏 [房间：%s]", r.ID)
		// 自动开始游戏
		if err := r.startGame(); err != nil {
			log.Printf("开始游戏失败 [房间：%s]: %v", r.ID, err)
			return err
		}
	}

	return nil
}

// 玩家选择座位
func (r *Room) setPlayerSeat(playerID string, seatNumber int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State == types.RoomPlaying {
		return fmt.Errorf("游戏进行中，无法更换座位")
	}

	if seatNumber < 0 || seatNumber >= r.Game.MaxPlayers() {
		return fmt.Errorf("无效的座位号: %d", seatNumber)
	}

	client, exists := r.Players[playerID]
	if !exists {
		return fmt.Errorf("玩家不在房间中")
	}

	// 检查目标座位是否被占用
	if occupiedBy, taken := r.Seats[seatNumber]; taken && occupiedBy != playerID {
		return fmt.Errorf("座位 %d 已被占用", seatNumber)
	}

	// 释放原座位
	if client.SeatNumber >= 0 {
		delete(r.Seats, client.SeatNumber)
	}

	// 占用新座位
	r.Seats[seatNumber] = playerID
	client.SeatNumber = seatNumber

	log.Printf("玩家更换座位 [房间：%s, 玩家：%s, 座位：%d]", r.ID, playerID, seatNumber)

	// 广播房间状态变更（包含座位变化）
	r.broadcastRoomStateWithMessage(fmt.Sprintf("玩家 %s 选择了座位 %d", playerID, seatNumber))

	return nil
}

// 检查是否可以开始游戏
func (r *Room) canStartGame() bool {
	if len(r.Players) < r.Game.MinPlayers() {
		return false
	}
	// 所有玩家都必须准备
	for playerID := range r.Players {
		if !r.Ready[playerID] {
			return false
		}
	}
	return true
}

// 广播房间状态（带自定义消息）
func (r *Room) broadcastRoomStateWithMessage(message string) {
	r.broadcastRoomStateInternal(message)
}

// 内部方法：广播房间状态
func (r *Room) broadcastRoomStateInternal(message string) {
	players := make([]types.PlayerSeatInfo, 0, len(r.Players))
	for seatNum := 0; seatNum < r.Game.MaxPlayers(); seatNum++ {
		if playerID, exists := r.Seats[seatNum]; exists {
			players = append(players, types.PlayerSeatInfo{
				PlayerID:   playerID,
				SeatNumber: seatNum,
				Ready:      r.Ready[playerID],
			})
		}
	}

	stateMsg := types.Message{
		Type: types.Broadcast,
		Data: types.BroadcastData{
			Event: types.RoomStateChanged,
			Content: types.RoomStateContent{
				RoomID:  r.ID,
				State:   r.State,
				Players: players,
				Message: message,
			},
		},
	}
	r.Broadcast(stateMsg)
}

// 处理卡牌指令
func (r *Room) ProcessGameAction(playerID string, data types.GameActionData) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 防抖检查：100ms内同一玩家的重复操作将被忽略
	now := time.Now().UnixMilli()
	if lastTime, exists := r.lastAction[playerID]; exists && now-lastTime < 100 {
		log.Printf("忽略玩家 [%s] 的重复操作", playerID)
		return nil
	}
	r.lastAction[playerID] = now

	action := interfaces.Action{Type: string(data.Action), Data: data.Card} // 使用 Action 类型
	turnEnded, err := r.Game.ProcessAction(playerID, action)
	if err != nil {
		return err
	}

	log.Printf("ProcessAction 完成: turnEnded=%v, IsGameOver=%v", turnEnded, r.Game.IsGameOver())

	r.broadcastState() // 出牌后立即全房间广播新状态

	// 检查游戏是否结束（玩家出完牌）
	log.Printf("检查游戏结束: IsGameOver=%v", r.Game.IsGameOver())
	if r.Game.IsGameOver() {
		winner := r.Game.Winner()
		r.Broadcast(types.Message{
			Type: types.Broadcast,
			Data: types.BroadcastData{Event: types.GameFinished, Content: map[string]string{"winner": winner}},
		})
		// 游戏结束，重置房间状态
		// r.resetRoomAfterGame()
		return nil
	}

	if turnEnded {
		r.Game.AdvanceTurn()
		r.broadcastState() // 回合切换后再广播一次
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

	// 按座位号排序玩家ID
	playerIDs := make([]string, r.Game.MaxPlayers())
	for seatNum := 0; seatNum < r.Game.MaxPlayers(); seatNum++ {
		if playerID, exists := r.Seats[seatNum]; exists {
			playerIDs[seatNum] = playerID
		}
	}

	if err := r.Game.Init(playerIDs); err != nil {
		log.Printf("游戏初始化失败 [房间：%s]: %v", r.ID, err)
		return err
	}

	r.State = types.RoomPlaying
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

// 游戏结束后重置房间状态
func (r *Room) resetRoomAfterGame() {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Printf("游戏结束，重置房间状态 [房间：%s]", r.ID)

	// 重置房间状态
	r.State = types.RoomWaiting

	// 清除所有玩家的准备状态
	for playerID := range r.Ready {
		r.Ready[playerID] = false
	}

	r.broadcastRoomStateWithMessage("游戏结束，等待下一局")
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
