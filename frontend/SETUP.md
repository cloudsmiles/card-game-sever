# 项目设置说明

## ⚠️ 重要：Node.js 版本要求

当前系统 Node.js 版本：v14.21.3
项目要求：Node.js >= 18.0.0

### 升级 Node.js

#### 使用 nvm (推荐)

```bash
# 安装 Node.js 18 或更高版本
nvm install 18
nvm use 18

# 或者安装最新的 LTS 版本
nvm install --lts
nvm use --lts
```

#### 使用 Homebrew (macOS)

```bash
brew install node@18
```

#### 直接下载

访问 https://nodejs.org/ 下载并安装 LTS 版本

## 安装步骤

1. 确保 Node.js 版本 >= 18.0.0

```bash
node --version  # 应该显示 v18.x.x 或更高
```

2. 安装依赖

```bash
cd card-game-server/frontend
npm install
```

3. 启动开发服务器

```bash
npm run dev
```

4. 在浏览器中打开 http://localhost:3000

## 如果遇到问题

### 清除缓存

```bash
rm -rf node_modules package-lock.json
npm install
```

### 使用 yarn 代替 npm

```bash
yarn install
yarn dev
```

## 项目已创建的文件

✅ package.json - 项目配置和依赖
✅ tsconfig.json - TypeScript 配置
✅ vite.config.ts - Vite 构建配置
✅ index.html - HTML 入口文件
✅ src/ - 源代码目录
✅ .env.* - 环境变量文件
✅ README.md - 项目文档

## 下一步

升级 Node.js 后，运行：

```bash
cd card-game-server/frontend
npm install
npm run dev
```

项目将在 http://localhost:3000 启动！
