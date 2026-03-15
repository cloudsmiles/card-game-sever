package room

import (
	"card-game-server/backend/internal/game/interfaces"
	"card-game-server/backend/internal/types"
	"fmt"
	"log"
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
			r.Ready[botID] = true

			GlobalNicknameTracker.SetNickname(botID, nickname+"(机器人)")
			GlobalPlayerTracker.SetPlayerRoom(botID, r.ID)

			// 启动bot消息消费goroutine（丢弃所有消息）
			go func(ch chan types.Message) {
				for range ch {
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

// checkBotTurn 检查是否有机器人需要操作，通过 BotPlayer 接口委托给游戏层
func (r *Room) checkBotTurn() {
	time.Sleep(800 * time.Millisecond) // 模拟思考时间

	r.mu.Lock()
	if r.State != types.RoomPlaying || r.Game == nil || r.Game.IsGameOver() {
		r.mu.Unlock()
		return
	}

	// 优先使用 BotPlayer 接口（游戏层自定义机器人逻辑）
	if bp, ok := r.Game.(interfaces.BotPlayer); ok {
		r.checkBotWithInterface(bp)
		r.mu.Unlock()
		return
	}

	// 兜底：没有实现 BotPlayer 的游戏，只检查当前回合玩家
	currentPlayer := r.Game.CurrentTurn()
	if !r.Bots[currentPlayer] {
		r.mu.Unlock()
		return
	}

	log.Printf("警告：游戏 %s 未实现 BotPlayer 接口，机器人无法操作", r.Game.ID())
	r.mu.Unlock()
}

// checkBotWithInterface 通过 BotPlayer 接口检查并执行机器人操作
// 调用者必须持有 r.mu 锁
func (r *Room) checkBotWithInterface(bp interfaces.BotPlayer) {
	for botID := range r.Bots {
		action := bp.GetBotAction(botID)
		if action == nil {
			continue
		}

		log.Printf("机器人操作 [房间：%s, 玩家：%s, action：%s]", r.ID, botID, action.Type)

		_, err := r.Game.ProcessAction(botID, *action)
		if err != nil {
			log.Printf("机器人操作失败 [房间：%s, 玩家：%s]: %v", r.ID, botID, err)
			continue
		}

		r.broadcastState()

		if r.Game.IsGameOver() {
			r.handleGameEnd()
			return
		}

		// 操作成功后，释放锁并重新检查（可能还有其他机器人需要操作）
		r.mu.Unlock()
		go r.checkBotTurn()
		r.mu.Lock()
		return
	}
}

// triggerBotCheck 在操作后触发机器人检查（游戏无关的通用逻辑）
// 调用者必须持有 r.mu 锁
func (r *Room) triggerBotCheck() {
	if bp, ok := r.Game.(interfaces.BotPlayer); ok {
		if bp.NeedsBotCheckAfterAction() {
			// 游戏声明每次操作后都需要检查机器人
			go r.checkBotTurn()
			return
		}
	}
	// 默认行为：只在轮到机器人时检查
	nextPlayer := r.Game.CurrentTurn()
	if r.Bots[nextPlayer] {
		go r.checkBotTurn()
	}
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
