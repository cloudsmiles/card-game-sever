package room

import (
	"card-game-server/backend/internal/game/interfaces"
	"card-game-server/backend/internal/types"
	"log"
)

// setupGameTimers 为游戏设置超时回调
// 游戏层在需要时会调用这些回调来通知房间层处理超时
func (r *Room) setupGameTimers() {
	// 使用类型断言获取支持超时处理的通用游戏实例
	if th, ok := r.Game.(interfaces.TimeoutHandler); ok {
		th.SetTimeoutCallbacks(
			r.onTurnTimeout,    // 回合操作超时回调
			r.onPendingTimeout, // 等待响应超时回调
		)
	}
}

// onTurnTimeout 回合操作超时回调（由游戏层调用）
func (r *Room) onTurnTimeout(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State != types.RoomPlaying {
		return
	}

	log.Printf("[房间：%s] 玩家 %s 回合操作超时，委托游戏层处理", r.ID, playerID)

	// 委托游戏层自己处理超时逻辑
	if th, ok := r.Game.(interfaces.TimeoutHandler); ok {
		if err := th.HandleTurnTimeout(playerID); err != nil {
			log.Printf("[房间：%s] 游戏层超时处理失败：%v", r.ID, err)
			return
		}
	} else {
		log.Printf("[房间：%s] 游戏不支持超时处理接口", r.ID)
		return
	}

	// 超时处理完成后，打印处理日志（模拟正常操作流程）
	log.Printf("处理卡牌指令 [房间：%s, 玩家：%s, action：discard(auto), 回合是否结束：true, 游戏是否结束：%v]", r.ID, playerID, r.Game.IsGameOver())

	r.broadcastState()

	// 检查游戏是否结束
	if r.Game.IsGameOver() {
		if err := r.handleGameEnd(); err != nil {
			log.Printf("[房间：%s] 游戏结束处理失败：%v", r.ID, err)
		}
		return
	}
}

// onPendingTimeout 等待响应超时回调（由游戏层调用）
func (r *Room) onPendingTimeout() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State != types.RoomPlaying {
		return
	}

	log.Printf("[房间：%s] 等待响应超时，委托游戏层处理", r.ID)

	// 委托游戏层自己处理超时逻辑
	if th, ok := r.Game.(interfaces.TimeoutHandler); ok {
		if err := th.HandlePendingTimeout(); err != nil {
			log.Printf("[房间：%s] 游戏层超时处理失败：%v", r.ID, err)
			return
		}
	} else {
		log.Printf("[房间：%s] 游戏不支持超时处理接口", r.ID)
		return
	}

	r.broadcastState()
}
