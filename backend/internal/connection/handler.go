package connection

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"card-game-server/backend/internal/auth"
	"card-game-server/backend/internal/room"
	"card-game-server/backend/internal/types"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	var playerID string
	var currentRoomID string = ""
	send := make(chan types.Message, 256)

	// 优先使用 token 认证
	tokenStr := r.URL.Query().Get("token")
	if tokenStr != "" {
		claims, err := auth.ParseToken(tokenStr)
		if err != nil {
			conn.WriteJSON(types.Message{
				Type: types.Error,
				Data: types.ErrorData{Code: 401, Message: "token 无效或已过期"},
			})
			conn.Close()
			return
		}
		playerID = "user_" + strconv.FormatUint(uint64(claims.UserID), 10)
		nickname := claims.Nickname
		room.GlobalNicknameTracker.SetNickname(playerID, nickname)
	} else {
		// 向后兼容：使用 player 参数
		playerID = r.URL.Query().Get("player")
		if playerID == "" {
			playerID = "player_" + randomString(6)
		}
		nickname := r.URL.Query().Get("nickname")
		if nickname == "" {
			nickname = playerID
		}
		room.GlobalNicknameTracker.SetNickname(playerID, nickname)
	}

	// 如果玩家已有连接，踢掉旧连接
	if kickExistingConnection(playerID) {
		log.Printf("玩家 [%s] 重新登录，已踢掉旧连接", playerID)
		// 等一小段时间让旧连接清理完成
		time.Sleep(100 * time.Millisecond)
	}

	// 注册新连接，获取被踢通知channel
	kickedChan := registerPlayerConnection(playerID, conn)
	log.Printf("玩家已连接 [玩家: %s]", playerID)

	// 检查是否是断线重连
	if existingRoomID, exists := room.GlobalPlayerTracker.GetPlayerRoom(playerID); exists {
		roomObj, err := room.GlobalManager.GetRoom(existingRoomID)
		if err == nil && roomObj.State == types.RoomPlaying {
			// 尝试重连
			if reconnectErr := roomObj.ReconnectPlayer(playerID, send); reconnectErr == nil {
				currentRoomID = existingRoomID
				log.Printf("玩家 [%s] 断线重连成功 [房间：%s]", playerID, existingRoomID)
			}
		}
	}

	// 连接验证通过后才注册清理 defer
	defer func() {
		// 只有当前连接仍是该玩家的注册连接时，才清理资源
		// 避免旧连接的 defer 清理掉新连接的状态
		connectedPlayersMu.RLock()
		existing := connectedPlayers[playerID]
		isCurrentConn := existing != nil && existing.conn == conn
		connectedPlayersMu.RUnlock()

		if isCurrentConn {
			// 连接断开时清理资源
			if playerID != "" && currentRoomID != "" {
				roomObj, err := room.GlobalManager.GetRoom(currentRoomID)
				if err == nil {
					// 游戏进行中标记为断线，其他状态直接移除
					if roomObj.State == types.RoomPlaying {
						roomObj.MarkPlayerOffline(playerID)
						log.Printf("玩家 [%s] 游戏中断线，已标记 [房间：%s]", playerID, currentRoomID)
					} else {
						isEmpty := roomObj.RemovePlayer(playerID)
						log.Printf("玩家 [%s] 离开房间 [%s]", playerID, currentRoomID)
						// 如果房间空了，从管理器移除
						if isEmpty {
							room.GlobalManager.RemoveRoom(currentRoomID)
						}
					}
				}
			}
			// 标记玩家为已断开连接
			unregisterPlayerConnection(playerID, conn)
			room.GlobalNicknameTracker.RemoveNickname(playerID)
			log.Printf("玩家已断开连接 [玩家: %s]", playerID)
		} else {
			log.Printf("旧连接清理跳过（已被新连接替代）[玩家: %s]", playerID)
		}
		conn.Close()
	}()

	// 启动心跳检测
	heartbeatDone := heartbeat(conn, playerID)

	go writePump(conn, send)

	for {
		select {
		case <-heartbeatDone:
			log.Printf("心跳超时，连接断开 [%s]", playerID)
			return
		case <-kickedChan:
			log.Printf("被新连接踢掉 [%s]", playerID)
			return
		default:
		}

		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf("连接读取错误 [%s]: %v", playerID, err)
			break
		}

		var rawMsg map[string]interface{}
		if err := json.Unmarshal(msgBytes, &rawMsg); err != nil {
			sendErrorMessage(conn, types.ErrorData{Code: 400, Message: "消息格式错误"})
			continue
		}

		// 解析消息类型
		msgType, ok := rawMsg["type"].(string)
		if !ok {
			sendErrorMessage(conn, types.ErrorData{Code: 400, Message: "缺少消息类型"})
			continue
		}

		msg := types.Message{
			Type:     types.MsgType(msgType),
			RoomID:   getString(rawMsg, "room_id"),
			PlayerID: playerID,
		}

		// 根据消息类型解析 data 字段
		dataBytes, _ := json.Marshal(rawMsg["data"])
		switch msg.Type {
		case types.RoomCreate:
			var createData types.CreateRoomData
			json.Unmarshal(dataBytes, &createData)
			msg.Data = createData
			handleCreateRoom(conn, send, msg, &currentRoomID)
		case types.RoomJoin:
			var joinData types.JoinRoomData
			json.Unmarshal(dataBytes, &joinData)
			msg.Data = joinData
			handleJoinRoom(conn, send, msg, &currentRoomID)
		case types.RoomLeave:
			msg.RoomID = getString(rawMsg, "room_id")
			handleLeaveRoom(conn, send, msg, &currentRoomID)
		case types.Chat:
			var chatData types.ChatData
			json.Unmarshal(dataBytes, &chatData)
			msg.Data = chatData
			handleChat(conn, send, msg)
		case types.GameAction:
			var actionData types.GameActionData
			json.Unmarshal(dataBytes, &actionData)
			msg.Data = actionData
			handleGameAction(conn, send, msg)
		case types.RoomAction:
			var roomActionData types.RoomActionData
			json.Unmarshal(dataBytes, &roomActionData)
			msg.Data = roomActionData
			handleRoomAction(conn, send, msg)
		}
	}
}

func handleCreateRoom(conn *websocket.Conn, send chan types.Message, msg types.Message, roomIDRef *string) {
	// 0. 检查玩家是否已在其他房间中
	if existingRoomID, exists := room.GlobalPlayerTracker.GetPlayerRoom(msg.PlayerID); exists {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrAlreadyInRoomCode,
			Message: fmt.Sprintf("玩家已在房间 [%s] 中，请先离开当前房间", existingRoomID),
		})
		return
	}

	// 1. 类型断言和数据提取
	createData, ok := msg.Data.(types.CreateRoomData)
	if !ok {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrInvalidDataCode,
			Message: "创建房间数据格式错误",
		})
		return
	}

	// 2. 参数验证
	if createData.GameType == "" {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrInvalidDataCode,
			Message: "缺少游戏类型参数",
		})
		return
	}

	// 3. 通过RoomManager创建房间
	roomID := room.GlobalManager.CreateRoom(createData.GameType)
	if roomID == "" {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrInternalCode,
			Message: "创建房间失败",
		})
		return
	}

	// 4. 获取房间对象
	roomObj, err := room.GlobalManager.GetRoom(roomID)
	if err != nil {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrInternalCode,
			Message: fmt.Sprintf("获取房间失败: %v", err),
		})
		return
	}

	// 5. 将玩家加入房间
	if err := roomObj.AddPlayer(msg.PlayerID, send); err != nil {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrJoinFailedCode,
			Message: fmt.Sprintf("加入房间失败：%v", err),
		})
		return
	}
	// 6. 记录玩家所在房间
	room.GlobalPlayerTracker.SetPlayerRoom(msg.PlayerID, roomID)
	// 设置当前房间 ID（用于断开时清理）
	*roomIDRef = roomID
}

func handleJoinRoom(conn *websocket.Conn, send chan types.Message, msg types.Message, roomIDRef *string) {
	// 0. 检查玩家是否已在其他房间中
	if existingRoomID, exists := room.GlobalPlayerTracker.GetPlayerRoom(msg.PlayerID); exists {
		// 如果已经在目标房间，提示已在房间中
		if existingRoomID == msg.RoomID {
			sendErrorMessage(conn, types.ErrorData{
				Code:    types.ErrAlreadyInRoomCode,
				Message: "玩家已在该房间中",
			})
			return
		}
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrAlreadyInRoomCode,
			Message: fmt.Sprintf("玩家已在房间 [%s] 中，请先离开当前房间", existingRoomID),
		})
		return
	}

	// 1. 类型断言和数据提取
	joinData, ok := msg.Data.(types.JoinRoomData)
	if !ok {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrInvalidDataCode,
			Message: "加入房间数据格式错误",
		})
		return
	}

	// 2. 参数验证
	if joinData.RoomID == "" {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrInvalidDataCode,
			Message: "缺少房间号参数",
		})
		return
	}

	// 3. 获取房间对象
	roomObj, err := room.GlobalManager.GetRoom(joinData.RoomID)
	if err != nil {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrRoomNotFoundCode,
			Message: fmt.Sprintf("房间不存在: %v", err),
		})
		return
	}

	// 4. 将玩家加入房间
	if err := roomObj.AddPlayer(msg.PlayerID, send); err != nil {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrJoinFailedCode,
			Message: fmt.Sprintf("加入房间失败：%v", err),
		})
		return
	}
	// 5. 记录玩家所在房间
	room.GlobalPlayerTracker.SetPlayerRoom(msg.PlayerID, joinData.RoomID)
	// 设置当前房间 ID（用于断开时清理）
	*roomIDRef = joinData.RoomID
}

func handleLeaveRoom(conn *websocket.Conn, send chan types.Message, msg types.Message, roomIDRef *string) {
	log.Printf("收到离开房间请求 [玩家: %s, room_id: %s]", msg.PlayerID, msg.RoomID)

	roomID := msg.RoomID
	if roomID == "" {
		// 尝试从 tracker 获取
		if rid, exists := room.GlobalPlayerTracker.GetPlayerRoom(msg.PlayerID); exists {
			roomID = rid
			log.Printf("从 tracker 获取房间ID [玩家: %s, 房间: %s]", msg.PlayerID, roomID)
		} else {
			sendErrorMessage(conn, types.ErrorData{
				Code:    types.ErrPlayerNotInRoomCode,
				Message: "玩家不在任何房间中",
			})
			return
		}
	}

	roomObj, err := room.GlobalManager.GetRoom(roomID)
	if err != nil {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrRoomNotFoundCode,
			Message: fmt.Sprintf("房间不存在: %v", err),
		})
		return
	}

	isEmpty := roomObj.RemovePlayer(msg.PlayerID)
	log.Printf("玩家 [%s] 主动离开房间 [%s], 房间是否为空: %v", msg.PlayerID, roomID, isEmpty)

	if isEmpty {
		room.GlobalManager.RemoveRoom(roomID)
	}

	// 通知客户端离开成功
	conn.WriteJSON(types.Message{
		Type: types.Broadcast,
		Data: types.BroadcastData{
			Event:   "room_left",
			Content: map[string]string{"room_id": roomID},
		},
	})

	// 清除当前房间ID
	*roomIDRef = ""
}

func handleChat(conn *websocket.Conn, send chan types.Message, msg types.Message) {
	// 0. 验证玩家是否在房间中
	if err := validatePlayerInRoom(msg.PlayerID, msg.RoomID); err != nil {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrPlayerNotInRoomCode,
			Message: err.Error(),
		})
		return
	}

	// 1. 类型断言和数据提取
	chatData, ok := msg.Data.(types.ChatData)
	if !ok {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrInvalidDataCode,
			Message: "聊天数据格式错误",
		})
		return
	}

	// 2. 参数验证
	if chatData.Content == "" {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrInvalidDataCode,
			Message: "缺少聊天内容参数",
		})
		return
	}

	// 3. 广播聊天消息
	roomObj, err := room.GlobalManager.GetRoom(msg.RoomID)
	if err != nil {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrRoomNotFoundCode,
			Message: fmt.Sprintf("房间不存在: %v", err),
		})
		return
	}
	roomObj.Broadcast(types.Message{
		Type: types.Broadcast,
		Data: types.BroadcastData{
			Event: types.ChatMessage,
			Content: types.ChatContent{
				RoomID:   msg.RoomID,
				PlayerID: msg.PlayerID,
				Nickname: room.GlobalNicknameTracker.GetNickname(msg.PlayerID),
				Content:  chatData.Content,
			},
		},
	})
}

func handleGameAction(conn *websocket.Conn, send chan types.Message, msg types.Message) {
	// 0. 验证玩家是否在房间中
	if err := validatePlayerInRoom(msg.PlayerID, msg.RoomID); err != nil {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrPlayerNotInRoomCode,
			Message: err.Error(),
		})
		return
	}

	// 1. 类型断言和数据提取
	gameActionData, ok := msg.Data.(types.GameActionData)
	if !ok {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrInvalidDataCode,
			Message: "游戏动作数据格式错误",
		})
		return
	}

	// 2. 参数验证
	if gameActionData.Action == "" {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrInvalidDataCode,
			Message: "缺少游戏动作参数",
		})
		return
	}

	// 3. 处理游戏动作
	roomObj, err := room.GlobalManager.GetRoom(msg.RoomID)
	if err != nil {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrRoomNotFoundCode,
			Message: fmt.Sprintf("房间不存在: %v", err),
		})
		return
	}

	if err := roomObj.ProcessGameAction(msg.PlayerID, gameActionData); err != nil {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrInternalCode,
			Message: fmt.Sprintf("处理游戏动作失败: %v", err),
		})
		return
	}
}

func handleRoomAction(conn *websocket.Conn, send chan types.Message, msg types.Message) {
	// 0. 验证玩家是否在房间中
	if err := validatePlayerInRoom(msg.PlayerID, msg.RoomID); err != nil {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrPlayerNotInRoomCode,
			Message: err.Error(),
		})
		return
	}

	// 1. 类型断言和数据提取
	roomActionData, ok := msg.Data.(types.RoomActionData)
	if !ok {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrInvalidDataCode,
			Message: "房间操作数据格式错误",
		})
		return
	}

	// 3. 获取房间对象
	roomObj, err := room.GlobalManager.GetRoom(msg.RoomID)
	if err != nil {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrRoomNotFoundCode,
			Message: fmt.Sprintf("房间不存在: %v", err),
		})
		return
	}

	// 4. 交给房间处理具体操作
	if err := roomObj.ProcessRoomAction(msg.PlayerID, roomActionData); err != nil {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrInternalCode,
			Message: fmt.Sprintf("房间操作失败: %v", err),
		})
		return
	}
}

// 统一的错误发送函数
func sendErrorMessage(conn *websocket.Conn, errorData types.ErrorData) {
	conn.WriteJSON(types.Message{
		Type: types.Error,
		Data: errorData,
	})
}

func writePump(conn *websocket.Conn, send chan types.Message) {
	for msg := range send {
		conn.WriteJSON(msg)
	}
}

func randomString(n int) string {
	// 简单实现，生产环境用 uuid
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}

// getString 从 map 中安全获取字符串值
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// validatePlayerInRoom 验证玩家是否在指定的房间中
func validatePlayerInRoom(playerID, roomID string) error {
	if roomID == "" {
		return fmt.Errorf("缺少房间号参数")
	}

	// 检查玩家是否在任何房间中
	playerRoomID, exists := room.GlobalPlayerTracker.GetPlayerRoom(playerID)
	if !exists {
		return fmt.Errorf("玩家不在任何房间中")
	}

	// 检查玩家是否在指定的房间中
	if playerRoomID != roomID {
		return fmt.Errorf("玩家不在该房间中，当前在房间 [%s]", playerRoomID)
	}

	return nil
}
