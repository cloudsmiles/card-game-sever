# CORS 配置说明

## 问题

前端从 `http://localhost:3000` 访问后端 API `http://localhost:8080/api/rooms` 时，浏览器报 CORS 错误：

```
Access to fetch at 'http://localhost:8080/api/rooms' from origin 'http://localhost:3000' 
has been blocked by CORS policy: No 'Access-Control-Allow-Origin' header is present on 
the requested resource.
```

## 解决方案

在 `main.go` 中添加 CORS 中间件，允许跨域请求。

### 修改内容

```go
// 配置 CORS 中间件
r.Use(func(c *gin.Context) {
    c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
    c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
    c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
    c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

    if c.Request.Method == "OPTIONS" {
        c.AbortWithStatus(204)
        return
    }

    c.Next()
})
```

### CORS 头说明

- `Access-Control-Allow-Origin: *` - 允许所有域名访问（生产环境应该限制为特定域名）
- `Access-Control-Allow-Credentials: true` - 允许携带凭证
- `Access-Control-Allow-Headers` - 允许的请求头
- `Access-Control-Allow-Methods` - 允许的 HTTP 方法
- OPTIONS 请求返回 204 - 处理预检请求

## 测试

### 测试 OPTIONS 预检请求

```bash
curl -H "Origin: http://localhost:3000" \
     -H "Access-Control-Request-Method: GET" \
     -H "Access-Control-Request-Headers: Content-Type" \
     -X OPTIONS \
     http://localhost:8080/api/rooms \
     -v
```

应该返回：
```
HTTP/1.1 204 No Content
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: POST, OPTIONS, GET, PUT, DELETE
...
```

### 测试实际请求

```bash
curl -H "Origin: http://localhost:3000" \
     http://localhost:8080/api/rooms
```

应该正常返回数据，并包含 CORS 头。

## 生产环境配置

在生产环境中，应该限制允许的域名：

```go
// 只允许特定域名
allowedOrigins := []string{
    "https://your-domain.com",
    "https://www.your-domain.com",
}

r.Use(func(c *gin.Context) {
    origin := c.Request.Header.Get("Origin")
    
    // 检查是否在允许列表中
    for _, allowed := range allowedOrigins {
        if origin == allowed {
            c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
            break
        }
    }
    
    // ... 其他 CORS 头设置
})
```

## 使用第三方 CORS 库

也可以使用 Gin 的 CORS 中间件库：

```bash
go get github.com/gin-contrib/cors
```

```go
import "github.com/gin-contrib/cors"

func main() {
    r := gin.Default()
    
    // 使用默认配置
    r.Use(cors.Default())
    
    // 或自定义配置
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))
    
    // ... 路由设置
}
```

## 注意事项

1. **安全性**: `Access-Control-Allow-Origin: *` 允许所有域名访问，生产环境应该限制
2. **凭证**: 如果设置 `AllowCredentials: true`，不能使用 `*` 作为 `AllowOrigin`
3. **预检请求**: 浏览器会先发送 OPTIONS 请求检查是否允许跨域
4. **缓存**: 可以设置 `Access-Control-Max-Age` 来缓存预检请求结果

## 重启服务器

修改 CORS 配置后，需要重启后端服务器：

```bash
# 停止旧进程
pkill -f "go run main.go"

# 启动新进程
cd card-game-server/backend
go run main.go
```

## 验证

1. 打开浏览器 http://localhost:3000
2. 打开 DevTools Console
3. 应该不再看到 CORS 错误
4. 房间列表应该正常加载
