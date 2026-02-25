package connection

import (
	"encoding/json"
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
	defer conn.Close()

	send := make(chan types.Message, 256)
	playerID := r.URL.Query().Get("player") // 客户端传 player=xxx
	if playerID == "" {
		playerID = "player_" + randomString(6)
	}

	go writePump(conn, send)

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var msg types.Message
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}
		msg.PlayerID = playerID

		switch msg.Type {
		case types.MsgCreateRoom:
			data := msg.Data.(map[string]interface{})
			gameType := data["game_type"].(string)
			roomID := room.GlobalManager.CreateRoom(gameType)
			roomObj, err := room.GlobalManager.GetRoom(roomID)
			if err != nil {
				continue
			}
			roomObj.AddPlayer(playerID, send) // 自动加入创建者
			conn.WriteJSON(types.Message{Type: types.MsgBroadcast, Data: map[string]string{"room_id": roomID}})

		case types.MsgJoinRoom:
			data := msg.Data.(map[string]interface{})
			roomID := data["room_id"].(string)
			roomObj, err := room.GlobalManager.GetRoom(roomID)
			if err != nil {
				conn.WriteJSON(types.Message{
					Type: types.MsgError,
					Data: map[string]string{"error": "房间不存在"},
				})
				break
			}
			if roomObj != nil {
				err = roomObj.AddPlayer(playerID, send)
				if err != nil {
					conn.WriteJSON(types.Message{
						Type: types.MsgError,
						Data: map[string]string{"error": err.Error()},
					})
				} else {
					// 关键修复：加入成功后返回 room_id 确认（和创建房间一致）
					conn.WriteJSON(types.Message{
						Type: types.MsgBroadcast,
						Data: map[string]string{"room_id": roomID},
					})
					log.Printf("玩家 %s 成功加入房间 %s", playerID, roomID)
				}
			}

		case types.MsgChat:
			chatData := msg.Data.(map[string]interface{})
			roomObj, _ := room.GlobalManager.GetRoom(msg.RoomID)
			if roomObj != nil {
				roomObj.Broadcast(types.Message{
					Type: types.MsgBroadcast,
					Data: types.BroadcastData{Event: "chat", Content: map[string]string{
						"player":  playerID,
						"content": chatData["content"].(string),
					}},
				})
			}

		case types.MsgGameAction:
			if msg.RoomID == "" {
				conn.WriteJSON(types.Message{
					Type: types.MsgError,
					Data: map[string]string{"error": "缺少房间号"},
				})
				break
			}
			actionData, ok := msg.Data.(map[string]interface{})
			if !ok {
				conn.WriteJSON(types.Message{
					Type: types.MsgError,
					Data: map[string]string{"error": "出牌数据格式错误"},
				})
				break
			}
			roomObj, err := room.GlobalManager.GetRoom(msg.RoomID)
			if err != nil {
				conn.WriteJSON(types.Message{
					Type: types.MsgError,
					Data: map[string]string{"error": "房间不存在"},
				})
				break
			}
			gData := types.GameActionData{
				Action: actionData["action"].(string),
				Card:   actionData["card"],
			}
			if err := roomObj.ProcessGameAction(playerID, gData); err != nil {
				conn.WriteJSON(types.Message{
					Type: types.MsgError,
					Data: map[string]string{"error": err.Error()},
				})
				log.Printf("出牌失败: %s", err)
			} else {
				log.Printf("玩家 %s 出牌成功", playerID)
			}
		}
	}
}

func writePump(conn *websocket.Conn, send chan types.Message) {
	for msg := range send {
		conn.WriteJSON(msg)
	}
}

func randomString(n int) string {
	// 简单实现，生产环境用 uuid
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}
