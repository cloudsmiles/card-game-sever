package room

import (
	"card-game-server/backend/internal/game/interfaces"
	"card-game-server/backend/internal/types"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"
)

var botNames = []string{"小明", "小红", "小刚", "小丽", "小强", "阿花", "阿宝", "大壮"}

// AddBot 向房间添加一个机器人
func (r *Room) AddBot() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State == types.RoomPlaying {
		return fmt.Errorf("游戏进行中，无法添加机器人")
	}

	if len(r.Players) >= r.Game.MaxPlayers() {
		return fmt.Errorf("房间已满")
	}

	// 生成机器人ID和昵称
	botIndex := 0
	for {
		botID := fmt.Sprintf("bot_%d", botIndex)
		if _, exists := r.Players[botID]; !exists {
			// 找到一个未使用的bot ID
			nickname := botNames[botIndex%len(botNames)]
			botSend := make(chan types.Message, 256)

			// 分配座位
			seatNumber := -1
			for i := 0; i < r.Game.MaxPlayers(); i++ {
				if _, taken := r.Seats[i]; !taken {
					seatNumber = i
					r.Seats[i] = botID
					break
				}
			}

			r.Players[botID] = &Client{PlayerID: botID, SeatNumber: seatNumber, Send: botSend}
			r.Bots[botID] = true
			r.Ready[botID] = true // 机器人自动准备

			// 设置昵称
			GlobalNicknameTracker.SetNickname(botID, nickname+"(机器人)")
			GlobalPlayerTracker.SetPlayerRoom(botID, r.ID)

			// 启动bot消息消费goroutine（丢弃所有消息）
			go func(ch chan types.Message) {
				for range ch {
					// 丢弃bot收到的消息
				}
			}(botSend)

			log.Printf("机器人加入 [房间：%s, ID：%s, 昵称：%s, 座位：%d]", r.ID, botID, nickname, seatNumber)

			r.broadcastRoomStateWithMessage(fmt.Sprintf("机器人 %s 加入了房间", nickname))

			// 检查是否所有玩家都准备好了
			if r.canStartGame() {
				log.Printf("所有玩家已准备，开始游戏 [房间：%s]", r.ID)
				if err := r.startGame(); err != nil {
					log.Printf("开始游戏失败 [房间：%s]: %v", r.ID, err)
					return err
				}
				// 游戏开始后检查是否轮到机器人
				go r.checkBotTurn()
			}

			return nil
		}
		botIndex++
	}
}

// IsBot 检查玩家是否是机器人
func (r *Room) IsBot(playerID string) bool {
	return r.Bots[playerID]
}

// checkBotTurn 检查当前是否轮到机器人，如果是则自动执行操作
func (r *Room) checkBotTurn() {
	time.Sleep(800 * time.Millisecond) // 模拟思考时间

	r.mu.Lock()
	if r.State != types.RoomPlaying || r.Game == nil || r.Game.IsGameOver() {
		r.mu.Unlock()
		return
	}

	currentPlayer := r.Game.CurrentTurn()
	if !r.Bots[currentPlayer] {
		r.mu.Unlock()
		return
	}

	// 获取游戏状态来决定机器人行为
	state := r.Game.GetStateForPlayer(currentPlayer)
	stateMap, ok := state.(map[string]interface{})
	if !ok {
		r.mu.Unlock()
		return
	}

	phase, _ := stateMap["phase"].(string)

	var action interfaces.Action
	switch phase {
	case "call":
		action = r.botCallAction(stateMap)
	case "play":
		action = r.botPlayAction(currentPlayer, stateMap)
	default:
		r.mu.Unlock()
		return
	}

	log.Printf("机器人操作 [房间：%s, 玩家：%s, action：%s]", r.ID, currentPlayer, action.Type)

	_, err := r.Game.ProcessAction(currentPlayer, action)
	if err != nil {
		log.Printf("机器人操作失败 [房间：%s, 玩家：%s]: %v", r.ID, currentPlayer, err)
		r.mu.Unlock()
		return
	}

	r.broadcastState()

	if r.Game.IsGameOver() {
		r.handleGameEnd()
		r.mu.Unlock()
		return
	}

	r.mu.Unlock()

	// 继续检查下一个是否也是机器人
	go r.checkBotTurn()
}

// botCallAction 机器人叫地主策略
func (r *Room) botCallAction(stateMap map[string]interface{}) interfaces.Action {
	// 简单策略：随机叫0-2分
	score := rand.Intn(3) // 0, 1, 2
	return interfaces.Action{
		Type: "call_landlord",
		Data: float64(score),
	}
}

// botPlayAction 机器人出牌策略
func (r *Room) botPlayAction(playerID string, stateMap map[string]interface{}) interfaces.Action {
	// 获取手牌
	myHandRaw, ok := stateMap["my_hand"].([]int)
	if !ok || len(myHandRaw) == 0 {
		return interfaces.Action{Type: "pass"}
	}
	myHand := myHandRaw

	// 获取上一手牌 - last_play 是 ddz.Play 结构体，通过 JSON 序列化/反序列化获取
	// 由于是同进程调用，直接用反射或类型断言不方便，改用 JSON 中转
	lastPlayRaw := stateMap["last_play"]
	lastType := ""
	var lastCardValues []int

	// 尝试通过 JSON 中转解析
	if lastPlayRaw != nil {
		jsonBytes, _ := json.Marshal(lastPlayRaw)
		var lastPlayMap map[string]interface{}
		if json.Unmarshal(jsonBytes, &lastPlayMap) == nil {
			lastType, _ = lastPlayMap["Type"].(string)
			if cardsRaw, ok := lastPlayMap["Cards"].([]interface{}); ok {
				for _, c := range cardsRaw {
					if cm, ok := c.(map[string]interface{}); ok {
						if v, ok := cm["Value"].(float64); ok {
							lastCardValues = append(lastCardValues, int(v))
						}
					}
				}
			}
		}
	}

	// 如果没有上一手牌（新回合），出最小的单牌
	if lastType == "" {
		return interfaces.Action{
			Type: "play_cards",
			Data: []interface{}{map[string]interface{}{"value": float64(myHand[0])}},
		}
	}

	lastMaxValue := 0
	if len(lastCardValues) > 0 {
		sort.Ints(lastCardValues)
		lastMaxValue = lastCardValues[0] // 排序后最小值用于比较
	}

	// 有上一手牌，尝试压制
	switch lastType {
	case "single":
		for _, v := range myHand {
			if v > lastMaxValue {
				return interfaces.Action{
					Type: "play_cards",
					Data: []interface{}{map[string]interface{}{"value": float64(v)}},
				}
			}
		}
	case "pair":
		pairs := findPairs(myHand)
		for _, pv := range pairs {
			if pv > lastMaxValue {
				return interfaces.Action{
					Type: "play_cards",
					Data: []interface{}{
						map[string]interface{}{"value": float64(pv)},
						map[string]interface{}{"value": float64(pv)},
					},
				}
			}
		}
	case "triple":
		triples := findTriples(myHand)
		for _, tv := range triples {
			if tv > lastMaxValue {
				cards := make([]interface{}, 3)
				for i := range cards {
					cards[i] = map[string]interface{}{"value": float64(tv)}
				}
				return interfaces.Action{Type: "play_cards", Data: cards}
			}
		}
	case "bomb":
		quads := findQuads(myHand)
		for _, qv := range quads {
			if qv > lastMaxValue {
				cards := make([]interface{}, 4)
				for i := range cards {
					cards[i] = map[string]interface{}{"value": float64(qv)}
				}
				return interfaces.Action{Type: "play_cards", Data: cards}
			}
		}
	default:
		// 复杂牌型，尝试用炸弹
		quads := findQuads(myHand)
		if len(quads) > 0 {
			cards := make([]interface{}, 4)
			for i := range cards {
				cards[i] = map[string]interface{}{"value": float64(quads[0])}
			}
			return interfaces.Action{Type: "play_cards", Data: cards}
		}
	}

	// 无法压制，pass
	return interfaces.Action{Type: "pass"}
}

// 辅助函数
func findPairs(hand []int) []int {
	countMap := make(map[int]int)
	for _, v := range hand {
		countMap[v]++
	}
	var pairs []int
	for v, c := range countMap {
		if c >= 2 {
			pairs = append(pairs, v)
		}
	}
	sort.Ints(pairs)
	return pairs
}

func findTriples(hand []int) []int {
	countMap := make(map[int]int)
	for _, v := range hand {
		countMap[v]++
	}
	var triples []int
	for v, c := range countMap {
		if c >= 3 {
			triples = append(triples, v)
		}
	}
	sort.Ints(triples)
	return triples
}

func findQuads(hand []int) []int {
	countMap := make(map[int]int)
	for _, v := range hand {
		countMap[v]++
	}
	var quads []int
	for v, c := range countMap {
		if c >= 4 {
			quads = append(quads, v)
		}
	}
	sort.Ints(quads)
	return quads
}

// RemoveBots 移除房间中所有机器人
func (r *Room) RemoveBots() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State == types.RoomPlaying {
		return
	}

	for botID := range r.Bots {
		if client, exists := r.Players[botID]; exists {
			if client.SeatNumber >= 0 {
				delete(r.Seats, client.SeatNumber)
			}
			// 关闭bot的消息channel
			close(client.Send)
		}
		delete(r.Players, botID)
		delete(r.Ready, botID)
		delete(r.Bots, botID)
		GlobalPlayerTracker.RemovePlayer(botID)
		GlobalNicknameTracker.RemoveNickname(botID)
		log.Printf("机器人移除 [房间：%s, ID：%s]", r.ID, botID)
	}

	r.broadcastRoomStateWithMessage("所有机器人已移除")
}
