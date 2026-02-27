package connection

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"card-game-server/internal/room"
	"card-game-server/internal/types"

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

	defer func() {
		// 连接断开时清理资源
		if playerID != "" && currentRoomID != "" {
			roomObj, err := room.GlobalManager.GetRoom(currentRoomID)
			if err == nil {
				isEmpty := roomObj.RemovePlayer(playerID)
				log.Printf("玩家 [%s] 离开房间 [%s]", playerID, currentRoomID)
				// 如果房间空了，从管理器移除
				if isEmpty {
					room.GlobalManager.RemoveRoom(currentRoomID)
				}
			}
		}
		close(send)
		conn.Close()
	}()

	playerID = r.URL.Query().Get("player") // 客户端传 player=xxx
	if playerID == "" {
		playerID = "player_" + randomString(6)
	}

	go writePump(conn, send)

	for {
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
	// 设置当前房间 ID（用于断开时清理）
	*roomIDRef = roomID
}

func handleJoinRoom(conn *websocket.Conn, send chan types.Message, msg types.Message, roomIDRef *string) {
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
	// 设置当前房间 ID（用于断开时清理）
	*roomIDRef = joinData.RoomID
}

func handleChat(conn *websocket.Conn, send chan types.Message, msg types.Message) {
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

	if msg.RoomID == "" {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrInvalidDataCode,
			Message: "缺少房间号参数",
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
				Content:  chatData.Content,
			},
		},
	})
}

func handleGameAction(conn *websocket.Conn, send chan types.Message, msg types.Message) {
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

	if msg.RoomID == "" {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrInvalidDataCode,
			Message: "缺少房间号参数",
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
	// 1. 类型断言和数据提取
	roomActionData, ok := msg.Data.(types.RoomActionData)
	if !ok {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrInvalidDataCode,
			Message: "房间操作数据格式错误",
		})
		return
	}

	// 2. 参数验证
	if msg.RoomID == "" {
		sendErrorMessage(conn, types.ErrorData{
			Code:    types.ErrInvalidDataCode,
			Message: "缺少房间号参数",
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
