# WebSocket 连接调试指南

## 问题诊断步骤

### 1. 检查后端服务器状态

```bash
# 检查后端是否运行
ps aux | grep "go run main.go"

# 检查端口是否监听
lsof -i :8080

# 测试 HTTP 端点
curl http://localhost:8080
```

### 2. 测试 WebSocket 连接

打开测试页面：http://localhost:3000/ws-test.html

点击 "Connect" 按钮，查看连接状态。

### 3. 检查浏览器控制台

1. 打开 Chrome DevTools (Cmd + Option + I)
2. 切换到 Console 标签
3. 查找 `[WebSocket]` 开头的日志
4. 应该看到：
   - `[WebSocket] Connecting to: ws://localhost:8080/ws?player=xxx`
   - `[WebSocket] Connected` (成功) 或错误信息

### 4. 检查 Network 标签

1. 打开 Chrome DevTools
2. 切换到 Network 标签
3. 筛选 WS (WebSocket)
4. 尝试连接
5. 查看 WebSocket 连接状态：
   - Status: 101 Switching Protocols (成功)
   - Status: 其他 (失败)

### 5. 常见问题和解决方案

#### 问题 1: 连接被拒绝 (Connection Refused)
**原因**: 后端服务器未启动
**解决**:
```bash
cd card-game-server/backend
go run main.go
```

#### 问题 2: 404 Not Found
**原因**: WebSocket 路由不正确
**检查**: 
- 前端 URL: `ws://localhost:8080/ws?player=xxx`
- 后端路由: `/ws`

#### 问题 3: CORS 错误
**原因**: 跨域问题
**解决**: 后端需要配置 CORS

#### 问题 4: 连接超时
**原因**: 防火墙或网络问题
**检查**: 
```bash
telnet localhost 8080
```

#### 问题 5: 环境变量未加载
**检查**: 
- 文件: `.env.development`
- 内容: `VITE_WS_URL=ws://localhost:8080/ws`
- 重启开发服务器

### 6. 手动测试 WebSocket

使用 wscat 工具：
```bash
# 安装 wscat
npm install -g wscat

# 测试连接
wscat -c "ws://localhost:8080/ws?player=test123"
```

### 7. 查看后端日志

后端应该输出连接日志，检查是否有错误信息。

### 8. 前端代码检查点

#### WebSocket Service (src/services/websocket.ts)
```typescript
// 应该输出正确的 URL
console.log('[WebSocket] Connecting to:', this.url);
// 预期: ws://localhost:8080/ws?player=xxx
```

#### WebSocket Context (src/contexts/WebSocketContext.tsx)
```typescript
// 检查连接状态更新
connectionStore.setStatus('connecting');
// 然后应该变为 'connected'
```

### 9. 使用 Redux DevTools 查看状态

1. 安装 Redux DevTools 扩展
2. 打开扩展
3. 查看 ConnectionStore 的状态：
   - status: 'connecting' → 'connected'
   - error: null

### 10. 完整测试流程

1. 启动后端：
```bash
cd card-game-server/backend
go run main.go
```

2. 启动前端：
```bash
cd card-game-server/frontend
npm run dev
```

3. 打开浏览器：http://localhost:3000

4. 打开 DevTools Console

5. 输入昵称并点击"进入大厅"

6. 查看 Console 输出：
```
[WebSocket] Connecting to: ws://localhost:8080/ws?player=player_xxx
[WebSocket] Connected
```

7. 如果成功，应该跳转到大厅页面

## 当前配置

### 前端配置
- 开发服务器: http://localhost:3000
- WebSocket URL: ws://localhost:8080/ws
- API URL: http://localhost:8080/api

### 后端配置
- HTTP 服务器: http://localhost:8080
- WebSocket 端点: /ws

## 快速修复

如果连接仍然失败，尝试以下步骤：

1. 重启后端服务器
2. 重启前端开发服务器
3. 清除浏览器缓存 (Cmd + Shift + R)
4. 检查是否有其他程序占用 8080 端口
5. 尝试使用测试页面 http://localhost:3000/ws-test.html

## 获取帮助

如果问题仍未解决，请提供：
1. 浏览器 Console 的完整输出
2. Network 标签中 WebSocket 的状态
3. 后端服务器的日志输出
4. 测试页面的连接结果
