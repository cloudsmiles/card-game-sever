# QCard Frontend

QCard 卡牌游戏平台前端应用

## 技术栈

- React 18+ with TypeScript
- Vite 5+
- Material-UI (MUI) v5
- Zustand (状态管理)
- React Router v6
- Framer Motion (动画)
- WebSocket (实时通信)

## 快速开始

### 安装依赖

```bash
npm install
```

### 开发模式

```bash
npm run dev
```

应用将在 http://localhost:3000 启动

### 构建生产版本

```bash
npm run build
```

### 预览生产构建

```bash
npm run preview
```

## 项目结构

```
src/
├── pages/          # 页面组件
├── components/     # 可复用组件
│   ├── common/     # 通用组件
│   ├── layout/     # 布局组件
│   ├── room/       # 房间相关组件
│   ├── chat/       # 聊天组件
│   └── game/       # 游戏通用组件
├── games/          # 游戏特定组件
│   ├── ddz/        # 斗地主
│   ├── mahjong/    # 麻将
│   └── durian/     # 榴莲忘返
├── services/       # 服务层
├── stores/         # Zustand stores
├── hooks/          # 自定义 hooks
├── utils/          # 工具函数
├── types/          # TypeScript 类型定义
├── bots/           # 机器人实现
└── i18n/           # 国际化
```

## 开发规范

### 代码风格

- 使用 ESLint Standard 配置
- 使用 Prettier 格式化代码
- 使用 TypeScript 进行类型检查

### 运行 Lint

```bash
npm run lint
```

### 格式化代码

```bash
npm run format
```

### 运行测试

```bash
npm run test
```

## 环境变量

复制 `.env.example` 到 `.env.development` 或 `.env.production` 并配置：

- `VITE_WS_URL`: WebSocket 服务器地址
- `VITE_API_URL`: HTTP API 地址
- `VITE_APP_NAME`: 应用名称
- `VITE_LOG_LEVEL`: 日志级别

## 浏览器支持

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## License

Private
