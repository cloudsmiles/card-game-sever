package room

import (
	"card-game-server/backend/internal/game"
	"card-game-server/backend/internal/game/interfaces"
	"card-game-server/backend/internal/types"
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
	Bots            map[string]bool      // botID -> true (tracks which players are bots)
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
		Bots:           make(map[string]bool),
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
	log.Printf("玩家加入成功 [房间：%s, 玩家：%s, 座位：%d, 当前人数：%d/%d]", r.ID, playerID, seatNumber, len(r.Players), r.Game.MaxPlayers())

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

// 标记玩家断线（游戏进行中或暂停时调用）
func (r *Room) MarkPlayerOffline(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 只有游戏进行中才处理断线
	if r.State != types.RoomPlaying {
		return
	}

	// 如果已经标记为断线，不再重复处理
	if _, alreadyOffline := r.OfflinePlayers[playerID]; alreadyOffline {
		return
	}

	// 记录断线时间
	r.OfflinePlayers[playerID] = time.Now()

	// 广播断线信息（但不改变房间状态）
	r.broadcastRoomStateWithMessage(fmt.Sprintf("玩家 %s 断线，等待重连...", playerID))

	// 如果是第一个断线的玩家，启动全局超时定时器
	if len(r.OfflinePlayers) == 1 {
		if r.disconnectTimer != nil {
			r.disconnectTimer.Stop()
		}
		r.disconnectTimer = time.AfterFunc(30*time.Second, func() {
			r.handleDisconnectTimeout()
		})
	}
}

// 处理断线超时（全局）
func (r *Room) handleDisconnectTimeout() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查是否还有断线玩家
	if len(r.OfflinePlayers) == 0 {
		return
	}

	// 获取第一个超时的玩家（按时间顺序）
	var timeoutPlayerID string
	var earliestTime time.Time
	for pid, t := range r.OfflinePlayers {
		if earliestTime.IsZero() || t.Before(earliestTime) {
			earliestTime = t
			timeoutPlayerID = pid
		}
	}

	if timeoutPlayerID == "" {
		return
	}

	log.Printf("玩家 [%s] 断线超时，游戏结束 [房间：%s]", timeoutPlayerID, r.ID)

	r.State = types.RoomWaiting

	// 清除所有玩家的准备状态
	for playerID := range r.Ready {
		r.Ready[playerID] = false
	}

	// 广播游戏结束（断线方输）
	r.Broadcast(types.Message{
		Type: types.Broadcast,
		Data: types.BroadcastData{
			Event:   types.GameFinished,
			Content: map[string]string{"winner": "对方获胜（玩家断线）"},
		},
	})

	r.broadcastRoomStateWithMessage(fmt.Sprintf("玩家 %s 断线超时，游戏结束", timeoutPlayerID))

	// 清理所有断线玩家（不只是当前超时的玩家）
	for offlinePlayerID := range r.OfflinePlayers {
		if client, exists := r.Players[offlinePlayerID]; exists && client.SeatNumber >= 0 {
			delete(r.Seats, client.SeatNumber)
		}
		delete(r.Players, offlinePlayerID)
		delete(r.Ready, offlinePlayerID)
		delete(r.lastAction, offlinePlayerID)
		GlobalPlayerTracker.RemovePlayer(offlinePlayerID)
		log.Printf("玩家 [%s] 已从房间移除 [房间：%s]", offlinePlayerID, r.ID)
	}
	// 清空 OfflinePlayers 映射
	r.OfflinePlayers = make(map[string]time.Time)

	// 如果只剩机器人，清理所有机器人
	r.cleanupBotsIfNoHumans()

	// 如果房间空了，清理资源
	if len(r.Players) == 0 {
		if r.disconnectTimer != nil {
			r.disconnectTimer.Stop()
		}
		close(r.broadcast)
		log.Printf("房间 [%s] 资源已清理", r.ID)
		// 通知Manager移除房间
		go GlobalManager.RemoveRoom(r.ID)
	}
}

// UpdatePlayerSend 更新玩家的发送通道（用于同一playerID重新连接时）
func (r *Room) UpdatePlayerSend(playerID string, sendChan chan types.Message) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if client, exists := r.Players[playerID]; exists {
		client.Send = sendChan
		log.Printf("玩家 [%s] 连接通道已更新 [房间：%s]", playerID, r.ID)
		// 广播房间状态让新连接获取最新状态
		r.broadcastRoomStateWithMessage(fmt.Sprintf("玩家 %s 已重新连接", GlobalNicknameTracker.GetNickname(playerID)))
	}
}

// 玩家重连
func (r *Room) ReconnectPlayer(playerID string, sendChan chan types.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查是否是断线玩家（包括被踢标记的）
	if _, wasOffline := r.OfflinePlayers[playerID]; !wasOffline {
		return fmt.Errorf("玩家未断线或不在房间中")
	}

	// 检查玩家是否在房间中
	client, exists := r.Players[playerID]
	if !exists {
		return fmt.Errorf("玩家不在房间中")
	}

	// 更新客户端连接
	client.Send = sendChan

	// 移除断线标记（如果有的话）
	delete(r.OfflinePlayers, playerID)

	// 主动发送当前游戏状态给重连玩家
	if r.Game != nil {
		gameState := r.Game.GetStateForPlayer(playerID)
		if gameState != nil {
			log.Printf("准备向重连玩家 [%s] 发送游戏状态 [房间：%s]", playerID, r.ID)
			select {
			case sendChan <- types.Message{
				Type: types.Broadcast,
				Data: types.BroadcastData{
					Event:   types.StateUpdate,
					Content: gameState,
				},
			}:
				log.Printf("已向重连玩家 [%s] 发送游戏状态 [房间：%s]", playerID, r.ID)
			default:
				log.Printf("警告：无法向重连玩家 [%s] 发送游戏状态，channel已满 [房间：%s]", playerID, r.ID)
			}
		} else {
			log.Printf("警告：无法获取玩家 [%s] 的游戏状态 [房间：%s]", playerID, r.ID)
		}
	} else {
		log.Printf("警告：房间 [%s] 没有关联的游戏对象", r.ID)
	}

	// 检查是否所有断线玩家都已重连
	if len(r.OfflinePlayers) == 0 {
		// 所有断线玩家都已重连，取消超时定时器
		if r.disconnectTimer != nil {
			r.disconnectTimer.Stop()
			r.disconnectTimer = nil
		}

		// 恢复游戏状态（但房间状态始终保持 RoomPlaying）
		log.Printf("所有玩家重连成功，游戏继续 [房间：%s]", r.ID)
		r.broadcastRoomStateWithMessage(fmt.Sprintf("玩家 %s 重连成功，游戏继续", playerID))
	} else {
		// 还有其他玩家在等待重连
		log.Printf("玩家 [%s] 重连成功，等待其他 %d 位玩家重连 [房间：%s]", playerID, len(r.OfflinePlayers), r.ID)
		r.broadcastRoomStateWithMessage(fmt.Sprintf("玩家 %s 重连成功，等待其他玩家... (%d 位)", playerID, len(r.OfflinePlayers)))
	}

	return nil
}

// 移除玩家（用于正常离开房间）
func (r *Room) RemovePlayer(playerID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	client, exists := r.Players[playerID]
	if !exists {
		return len(r.Players) == 0
	}

	// 如果游戏进行中，玩家离开则立即结束游戏
	if r.State == types.RoomPlaying {
		nickname := GlobalNicknameTracker.GetNickname(playerID)
		log.Printf("玩家 [%s] 在游戏中离开房间，游戏立即结束 [房间：%s]", playerID, r.ID)

		// 停止所有定时器
		if r.disconnectTimer != nil {
			r.disconnectTimer.Stop()
			r.disconnectTimer = nil
		}

		// 广播游戏结束
		r.Broadcast(types.Message{
			Type: types.Broadcast,
			Data: types.BroadcastData{
				Event:   types.GameFinished,
				Content: map[string]string{"winner": fmt.Sprintf("玩家 %s 离开，游戏结束", nickname)},
			},
		})

		// 重置房间状态
		r.State = types.RoomWaiting
		for pid := range r.Ready {
			r.Ready[pid] = false
		}
		for botID := range r.Bots {
			r.Ready[botID] = true
		}

		// 清理所有断线玩家（游戏已结束，不再需要等待重连）
		for offlinePlayerID := range r.OfflinePlayers {
			if offClient, exists := r.Players[offlinePlayerID]; exists && offClient.SeatNumber >= 0 {
				delete(r.Seats, offClient.SeatNumber)
			}
			delete(r.Players, offlinePlayerID)
			delete(r.Ready, offlinePlayerID)
			delete(r.lastAction, offlinePlayerID)
			GlobalPlayerTracker.RemovePlayer(offlinePlayerID)
			GlobalNicknameTracker.RemoveNickname(offlinePlayerID)
			log.Printf("断线玩家 [%s] 已从房间移除 [房间：%s]", offlinePlayerID, r.ID)
		}
		r.OfflinePlayers = make(map[string]time.Time)
	}

	// 释放座位
	if client.SeatNumber >= 0 {
		delete(r.Seats, client.SeatNumber)
	}
	delete(r.Players, playerID)
	delete(r.Ready, playerID)
	delete(r.lastAction, playerID)
	delete(r.OfflinePlayers, playerID)

	// 从 GlobalPlayerTracker 移除
	GlobalPlayerTracker.RemovePlayer(playerID)
	delete(r.Bots, playerID)

	// 如果只剩机器人，清理所有机器人
	r.cleanupBotsIfNoHumans()

	// 广播房间状态变更（包含玩家离开）
	r.broadcastRoomStateWithMessage(fmt.Sprintf("玩家 %s 离开了房间", playerID))

	// 如果房间空了，清理资源
	isEmpty := len(r.Players) == 0
	if isEmpty {
		if r.disconnectTimer != nil {
			r.disconnectTimer.Stop()
		}
		close(r.broadcast)
		log.Printf("房间 [%s] 资源已清理", r.ID)
		// 通知Manager移除房间
		go GlobalManager.RemoveRoom(r.ID)
	}
	return isEmpty
}

// cleanupBotsIfNoHumans 如果房间内只剩机器人，清理所有机器人（调用者必须持有 r.mu 写锁）
func (r *Room) cleanupBotsIfNoHumans() {
	if len(r.Players) == 0 {
		return
	}
	hasHuman := false
	for pid := range r.Players {
		if !r.Bots[pid] {
			hasHuman = true
			break
		}
	}
	if !hasHuman {
		log.Printf("房间 [%s] 只剩机器人，清理所有机器人", r.ID)
		for botID := range r.Bots {
			if client, exists := r.Players[botID]; exists {
				if client.SeatNumber >= 0 {
					delete(r.Seats, client.SeatNumber)
				}
				close(client.Send)
			}
			delete(r.Players, botID)
			delete(r.Ready, botID)
			delete(r.Bots, botID)
			GlobalPlayerTracker.RemovePlayer(botID)
			GlobalNicknameTracker.RemoveNickname(botID)
		}
	}
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

	case types.RoomActionAddBot:
		return r.AddBot()

	case types.RoomActionKickBot:
		// 解析要踢的机器人ID
		kickDataBytes, _ := json.Marshal(action.Data)
		var kickData struct {
			BotID string `json:"bot_id"`
		}
		json.Unmarshal(kickDataBytes, &kickData)
		return r.RemoveBot(kickData.BotID)

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
	// 安全网：等待状态下确保所有机器人始终准备
	if r.State == types.RoomWaiting {
		for botID := range r.Bots {
			r.Ready[botID] = true
		}
	}

	// 计算在线玩家数（不包括断线玩家）并打印详细信息
	onlineCount := 0
	onlinePlayers := []string{}
	offlinePlayers := []string{}
	for playerID := range r.Players {
		if _, isOffline := r.OfflinePlayers[playerID]; isOffline {
			offlinePlayers = append(offlinePlayers, playerID)
		} else {
			onlineCount++
			onlinePlayers = append(onlinePlayers, playerID)
		}
	}
	if r.Game == nil {
		log.Printf("错误：r.Game 为 nil")
		return
	}

	players := make([]types.PlayerSeatInfo, 0, len(r.Players))
	for seatNum := 0; seatNum < r.Game.MaxPlayers(); seatNum++ {
		if playerID, exists := r.Seats[seatNum]; exists {
			// 检查是否是断线玩家
			isOffline := false
			if _, ok := r.OfflinePlayers[playerID]; ok {
				isOffline = true
			}
			players = append(players, types.PlayerSeatInfo{
				PlayerID:   playerID,
				Nickname:   GlobalNicknameTracker.GetNickname(playerID),
				SeatNumber: seatNum,
				Ready:      r.Ready[playerID],
				IsOffline:  isOffline,
				IsBot:      r.Bots[playerID],
			})
		}
	}

	// log.Printf("房间 [%s] 广播状态: state=%s, 总玩家数=%d, 在线玩家数=%d, message=%s", r.ID, r.State, len(r.Players), onlineCount, message)

	stateMsg := types.Message{
		Type: types.Broadcast,
		Data: types.BroadcastData{
			Event: types.RoomStateChanged,
			Content: types.RoomStateContent{
				RoomID:   r.ID,
				GameType: r.GameType,
				State:    r.State,
				Players:  players,
				Message:  message,
			},
		},
	}
	r.Broadcast(stateMsg)
}

// 处理卡牌指令
func (r *Room) ProcessGameAction(playerID string, data types.GameActionData) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查游戏是否进行中
	if r.State != types.RoomPlaying {
		return fmt.Errorf("游戏非进行中，无法执行操作")
	}
	// 处于断线暂停
	if len(r.OfflinePlayers) > 0 {
		return fmt.Errorf("游戏暂停中，等待断线玩家重连")
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

	log.Printf("处理卡牌指令 [房间：%s, 玩家：%s, action：%s, 回合是否结束：%v, 游戏是否结束：%v]", r.ID, playerID, action.Type, turnEnded, r.Game.IsGameOver())

	r.broadcastState() // 广播最新的游戏状态

	// 检查游戏是否结束
	if r.Game.IsGameOver() {
		if err := r.handleGameEnd(); err != nil {
			log.Printf("游戏结束处理失败 [房间：%s]: %v", r.ID, err)
			return fmt.Errorf("游戏结束处理失败: %w", err)
		}
		return nil
	}

	// 通过接口触发机器人检查（游戏无关）
	r.triggerBotCheck()

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

	// 设置游戏超时回调（游戏层会自己管理定时器）
	r.setupGameTimers()

	// 检查第一个回合是否是机器人
	firstPlayer := r.Game.CurrentTurn()
	if r.Bots[firstPlayer] {
		go r.checkBotTurn()
	}

	return nil
}

// handleGameEnd 处理游戏结束逻辑（统一游戏结束处理）
func (r *Room) handleGameEnd() error {
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
	// 机器人自动重新准备
	for botID := range r.Bots {
		r.Ready[botID] = true
	}

	// 3. 构建并广播最终的 room_state_changed 事件
	r.broadcastRoomStateWithMessage(fmt.Sprintf("游戏结束！获胜者：%s", winner))

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
					// if broadcastData, ok := msg.Data.(types.BroadcastData); ok {
					// log.Printf("广播消息发送给玩家 [%s] [房间：%s], 事件：%v,", c.PlayerID, r.ID, broadcastData.Event)
					// }
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
				// log.Printf("个性化广播消息发送给玩家 [%s] [房间：%s], 事件：%v", playerID, r.ID, data.Event)
			default:
				log.Printf("警告：玩家 [%s] 的消息队列已满，丢弃个性化消息 [房间：%s]", playerID, r.ID)
			}
		}()
	}
}
