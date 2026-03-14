# API 集成说明

## 后端 API 接口

### 1. 房间列表接口

**端点**: `GET /api/rooms`

**请求参数** (可选):
- `game_type`: 游戏类型筛选 (ddz, mahjong, durian)
- `status`: 状态筛选 (waiting, playing)

**响应格式**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "room_xxx",
      "game_type": "ddz",
      "state": "waiting",
      "players": 2,
      "max_players": 3
    }
  ]
}
```

**前端实现**:
- 文件: `src/services/roomApi.ts`
- 方法: `roomApi.getRooms(params)`
- 自动轮询: 每 5 秒刷新一次

### 2. WebSocket 接口

**端点**: `ws://localhost:8080/ws?player={playerId}`

**消息格式**: 参考 `PROTOCOL.md`

**前端实现**:
- 文件: `src/services/websocket.ts`
- Context: `src/contexts/WebSocketContext.tsx`
- Hook: `src/hooks/useWebSocket.ts`

## 数据格式转换

### 后端 → 前端

后端返回的房间信息需要转换为前端格式：

```typescript
// 后端格式
interface ApiRoomInfo {
  id: string;
  game_type: string;
  state: string;
  players: number;
  max_players: number;
}

// 前端格式
interface RoomInfo {
  room_id: string;
  game_type: GameType;
  status: 'waiting' | 'playing';
  players: SeatPlayerInfo[];
  max_players: number;
}
```

转换逻辑在 `RoomList.tsx` 的 `convertApiRoom` 函数中。

## 环境配置

### 开发环境 (.env.development)
```env
VITE_WS_URL=ws://localhost:8080/ws
VITE_API_URL=http://localhost:8080/api
```

### 生产环境 (.env.production)
```env
VITE_WS_URL=ws://your-domain.com/ws
VITE_API_URL=https://your-domain.com/api
```

## 错误处理

### API 错误
- 网络错误: 显示通知 "获取房间列表失败"
- 后端错误: 根据 `code` 和 `message` 显示相应错误

### WebSocket 错误
- 连接失败: 自动重连（最多 5 次）
- 断线: 显示重连状态
- 消息错误: 在控制台输出日志

## 测试

### 测试房间列表 API
```bash
# 获取所有房间
curl http://localhost:8080/api/rooms

# 筛选斗地主房间
curl "http://localhost:8080/api/rooms?game_type=ddz"

# 筛选等待中的房间
curl "http://localhost:8080/api/rooms?status=waiting"
```

### 测试 WebSocket
打开测试页面: http://localhost:3000/ws-test.html

## 已实现功能

✅ 房间列表获取
✅ 房间筛选（游戏类型、状态）
✅ 自动刷新（5秒轮询）
✅ 前端分页
✅ WebSocket 连接
✅ WebSocket 自动重连
✅ 房间状态实时更新
✅ 聊天消息实时推送

## 待实现功能

⏳ 后端分页支持
⏳ 房间搜索功能
⏳ 房间排序功能
⏳ 更多筛选条件

## 注意事项

1. **CORS**: 后端已配置 CORS，允许前端跨域请求
2. **轮询频率**: 当前为 5 秒，可根据需要调整
3. **错误重试**: API 请求失败不会自动重试，需要手动刷新
4. **WebSocket 重连**: 最多重连 5 次，使用指数退避策略
5. **数据同步**: 房间列表通过 HTTP 轮询，房间详情通过 WebSocket 实时更新

## 调试

### 查看 API 请求
1. 打开 Chrome DevTools
2. 切换到 Network 标签
3. 筛选 Fetch/XHR
4. 查看 `/api/rooms` 请求

### 查看 WebSocket 消息
1. 打开 Chrome DevTools
2. 切换到 Network 标签
3. 筛选 WS
4. 选择 WebSocket 连接
5. 查看 Messages 标签

### 查看状态
使用左下角的 ConnectionDebug 组件查看：
- WebSocket 连接状态
- Player ID
- 错误信息
- WebSocket URL
