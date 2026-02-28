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
	ID              string
	GameType        string
	State           types.RoomState
	Players         map[string]*Client   // playerID -> Client
	Seats           map[int]string       // seatNumber -> playerID
	Ready           map[string]bool      // playerID -> ready status
	OfflinePlayers  map[string]time.Time // playerID -> 断线时间
	Game            interfaces.Game
	mu              sync.RWMutex
	broadcast       chan types.Message
	lastAction      map[string]int64 // playerID -> 上次操作时间戳（毫秒）
	disconnectTimer *time.Timer      // 断线超时定时器
}

type Client struct {
	PlayerID   string
	SeatNumber int // 座位号
	Send       chan types.Message
}

func NewRoom(id, gameType string) *Room {
	r := &Room{
		ID:             id,
		GameType:       gameType,
		State:          types.RoomWaiting,
		Players:        make(map[string]*Client),
		Seats:          make(map[int]string),
		Ready:          make(map[string]bool),
		OfflinePlayers: make(map[string]time.Time),
		broadcast:      make(chan types.Message, 512), // 加大缓冲
		lastAction:     make(map[string]int64),
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

// Broadcast 广播消息给所有玩家（通过channel异步发送，确保顺序）
func (r *Room) Broadcast(msg types.Message) {
	defer func() {
		if recover() != nil {
			// channel 已关闭，忽略
			log.Printf("警告：向已关闭的广播 channel 发送消息被忽略 [房间：%s]", r.ID)
		}
	}()
	r.broadcast <- msg
}

// BroadcastPersonalized 广播个性化消息给每个玩家（通过channel异步发送）
// contentFunc 根据玩家ID生成个性化内容
func (r *Room) BroadcastPersonalized(event types.Event, contentFunc func(playerID string) interface{}) {
	// 将个性化消息生成函数包装成特殊消息，通过channel发送
	// 使用 BroadcastData 的 Content 字段存储生成函数
	// runBroadcast 会识别并处理这种特殊消息
	defer func() {
		if recover() != nil {
			log.Printf("警告：向已关闭的广播 channel 发送个性化消息被忽略 [房间：%s]", r.ID)
		}
	}()

	// 创建一个包含生成函数的特殊消息
	// 使用特殊的事件类型标记这是个性化消息
	msg := types.Message{
		Type: types.Broadcast,
		Data: types.PersonalizedBroadcastData{
			Event:       event,
			ContentFunc: contentFunc,
		},
	}
	r.broadcast <- msg
}

// 标记玩家断线（游戏进行中时调用）
func (r *Room) MarkPlayerOffline(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 只有游戏进行中才处理断线
	if r.State != types.RoomPlaying {
		return
	}

	// 记录断线时间
	r.OfflinePlayers[playerID] = time.Now()

	// 暂停游戏
	r.State = types.RoomPaused
	log.Printf("玩家 [%s] 断线，游戏暂停 [房间：%s]", playerID, r.ID)

	// 广播暂停状态
	r.broadcastRoomStateWithMessage(fmt.Sprintf("玩家 %s 断线，游戏暂停，等待重连...", playerID))

	// 启动30秒超时定时器
	if r.disconnectTimer != nil {
		r.disconnectTimer.Stop()
	}
	r.disconnectTimer = time.AfterFunc(30*time.Second, func() {
		r.handleDisconnectTimeout(playerID)
	})
}

// 处理断线超时
func (r *Room) handleDisconnectTimeout(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查玩家是否仍然断线
	if _, stillOffline := r.OfflinePlayers[playerID]; !stillOffline {
		return // 玩家已重连
	}

	log.Printf("玩家 [%s] 断线超时，游戏结束 [房间：%s]", playerID, r.ID)

	// 结束游戏
	r.State = types.RoomGameOver

	// 广播游戏结束（断线方输）
	r.Broadcast(types.Message{
		Type: types.Broadcast,
		Data: types.BroadcastData{
			Event:   types.GameFinished,
			Content: map[string]string{"winner": "对方获胜（玩家断线）"},
		},
	})

	r.broadcastRoomStateWithMessage(fmt.Sprintf("玩家 %s 断线超时，游戏结束", playerID))
}

// 玩家重连
func (r *Room) ReconnectPlayer(playerID string, sendChan chan types.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查是否是断线玩家
	if _, wasOffline := r.OfflinePlayers[playerID]; !wasOffline {
		return fmt.Errorf("玩家未断线或不在房间中")
	}

	// 更新客户端连接
	if client, exists := r.Players[playerID]; exists {
		client.Send = sendChan
	}

	// 移除断线标记
	delete(r.OfflinePlayers, playerID)

	// 取消超时定时器
	if r.disconnectTimer != nil {
		r.disconnectTimer.Stop()
		r.disconnectTimer = nil
	}

	// 恢复游戏状态
	if r.State == types.RoomPaused && len(r.OfflinePlayers) == 0 {
		r.State = types.RoomPlaying
		log.Printf("玩家 [%s] 重连成功，游戏恢复 [房间：%s]", playerID, r.ID)
		r.broadcastRoomStateWithMessage(fmt.Sprintf("玩家 %s 重连成功，游戏继续", playerID))
	} else {
		log.Printf("玩家 [%s] 重连成功 [房间：%s]", playerID, r.ID)
		r.broadcastRoomStateWithMessage(fmt.Sprintf("玩家 %s 重连成功", playerID))
	}

	return nil
}

// 移除玩家（用于正常离开房间）
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
	delete(r.OfflinePlayers, playerID)

	// 广播房间状态变更（包含玩家离开）
	r.broadcastRoomStateWithMessage(fmt.Sprintf("玩家 %s 离开了房间", playerID))

	// 如果房间空了，清理资源
	isEmpty := len(r.Players) == 0
	if isEmpty {
		if r.disconnectTimer != nil {
			r.disconnectTimer.Stop()
		}
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

// 内部方法：广播房间状态（调用者必须持有 r.mu 的读锁或写锁）
func (r *Room) broadcastRoomStateInternal(message string) {
	log.Printf("构建 room_state_changed 消息，当前玩家数: %d", len(r.Players))

	if r.Game == nil {
		log.Printf("错误：r.Game 为 nil")
		return
	}

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

	log.Printf("房间 [%s] 广播状态: state=%s, players=%d, message=%s", r.ID, r.State, len(players), message)

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

	// 检查游戏是否暂停
	if r.State == types.RoomPaused {
		return fmt.Errorf("游戏暂停中，等待断线玩家重连")
	}

	// 检查游戏是否已结束
	if r.State == types.RoomGameOver {
		return fmt.Errorf("游戏已结束")
	}

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
		log.Printf("游戏结束，获胜者: %s", winner)

		// 1. 先广播 game_over 事件
		r.Broadcast(types.Message{
			Type: types.Broadcast,
			Data: types.BroadcastData{Event: types.GameFinished, Content: map[string]string{"winner": winner}},
		})

		// 2. 重置房间状态为 waiting
		r.State = types.RoomWaiting
		// 清除准备状态
		for playerID := range r.Ready {
			r.Ready[playerID] = false
		}

		// 3. 构建并广播 room_state_changed 事件
		r.broadcastRoomStateWithMessage(fmt.Sprintf("游戏结束！获胜者：%s", winner))
		return nil
	}

	if turnEnded {
		r.Game.AdvanceTurn()
		r.broadcastState() // 回合切换后再广播一次
	}
	return nil
}

// 全房间广播最新游戏状态
// broadcastState 广播游戏状态（调用者必须持有 r.mu 的读锁或写锁）
// 为每个玩家发送包含其手牌的个性化状态
func (r *Room) broadcastState() {
	if r.Game == nil {
		log.Printf("错误：broadcastState 时 r.Game 为 nil")
		return
	}

	// 使用 BroadcastPersonalized 发送个性化状态更新
	r.BroadcastPersonalized(types.StateUpdate, func(playerID string) interface{} {
		return r.Game.GetStateForPlayer(playerID)
	})
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

	// 使用 BroadcastPersonalized 发送个性化游戏开始消息
	r.BroadcastPersonalized(types.GameStarted, func(playerID string) interface{} {
		return r.Game.GetStateForPlayer(playerID)
	})

	log.Printf("游戏开始广播完成 [房间：%s]", r.ID)
	return nil
}

func (r *Room) runBroadcast() {
	for msg := range r.broadcast {
		// 检查是否为个性化广播消息
		if personalizedData, ok := msg.Data.(types.PersonalizedBroadcastData); ok {
			r.handlePersonalizedBroadcast(personalizedData)
			continue
		}

		// 普通广播消息
		r.mu.RLock()
		sentCount := 0
		for _, c := range r.Players {
			// 跳过断线玩家的发送
			if _, isOffline := r.OfflinePlayers[c.PlayerID]; isOffline {
				continue
			}
			// 跳过没有Send channel的玩家
			if c.Send == nil {
				continue
			}
			// 使用 recover 捕获可能的 panic（如 channel 已关闭）
			func() {
				defer func() {
					if rec := recover(); rec != nil {
						log.Printf("警告：向玩家 [%s] 发送消息时发生 panic: %v [房间：%s]", c.PlayerID, rec, r.ID)
					}
				}()
				select {
				case c.Send <- msg:
					sentCount++
					// 安全类型断言，避免 panic
					if broadcastData, ok := msg.Data.(types.BroadcastData); ok {
						log.Printf("广播消息发送给玩家 [%s] [房间：%s], 事件：%v", c.PlayerID, r.ID, broadcastData.Event)
					}
				default: // 防止单个客户端卡住影响他人
					log.Printf("警告：玩家 [%s] 的消息队列已满，丢弃消息 [房间：%s]", c.PlayerID, r.ID)
				}
			}()
		}
		r.mu.RUnlock()
		if sentCount == 0 {
			log.Printf("警告：房间 [%s] 没有成功发送任何消息，当前玩家数：%d", r.ID, len(r.Players))
		}
	}
}

// handlePersonalizedBroadcast 处理个性化广播
func (r *Room) handlePersonalizedBroadcast(data types.PersonalizedBroadcastData) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for playerID, client := range r.Players {
		if client == nil || client.Send == nil {
			continue
		}

		// 跳过断线玩家
		if _, isOffline := r.OfflinePlayers[playerID]; isOffline {
			continue
		}

		// 生成个性化内容
		content := data.ContentFunc(playerID)

		msg := types.Message{
			Type: types.Broadcast,
			Data: types.BroadcastData{
				Event:   data.Event,
				Content: content,
			},
		}

		// 使用 recover 捕获可能的 panic（如 channel 已关闭）
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					log.Printf("警告：向玩家 [%s] 发送个性化消息时发生 panic: %v [房间：%s]", playerID, rec, r.ID)
				}
			}()
			// 非阻塞发送
			select {
			case client.Send <- msg:
				log.Printf("个性化广播消息发送给玩家 [%s] [房间：%s], 事件：%v", playerID, r.ID, data.Event)
			default:
				log.Printf("警告：玩家 [%s] 的消息队列已满，丢弃个性化消息 [房间：%s]", playerID, r.ID)
			}
		}()
	}
}
