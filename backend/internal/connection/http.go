package connection

import (
	"net/http"

	"card-game-server/backend/internal/room"
	"github.com/gin-gonic/gin"
)

// RoomListResponse 房间列表响应
type RoomListResponse struct {
	Code    int                   `json:"code"`
	Message string                `json:"message"`
	Data    []room.RoomInfo       `json:"data"`
}

// RoomListHandler 获取房间列表
func RoomListHandler(c *gin.Context) {
	rooms := room.GlobalManager.GetAllRooms()

	roomList := make([]room.RoomInfo, 0, len(rooms))
	for _, r := range rooms {
		roomList = append(roomList, room.RoomInfo{
			ID:          r.ID,
			GameType:    r.GameType,
			State:       string(r.State),
			Players:     len(r.Players),
			MaxPlayers:  r.Game.MaxPlayers(),
		})
	}

	c.JSON(http.StatusOK, RoomListResponse{
		Code:    0,
		Message: "success",
		Data:    roomList,
	})
}

// SetupHTTPHandlers 设置 HTTP 路由
func SetupHTTPHandlers(r *gin.Engine) {
	// 房间相关接口
	rooms := r.Group("/api/rooms")
	{
		rooms.GET("", RoomListHandler) // 获取房间列表
	}
}