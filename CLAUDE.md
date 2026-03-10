# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A multiplayer online card game server with Go backend (WebSocket) and React + TypeScript frontend. Supports multiple game types: Dou Dizhu (ddz), Mahjong (mahjong), Durian (durian), and Simple (simple).

## Common Commands

### Backend (Go)

```bash
cd backend
go run main.go                    # Start server on :8080
go test ./internal/game/...       # Run all game tests
go test ./internal/game/durian    # Run specific game tests
```

### Frontend (React + Vite)

```bash
cd frontend
npm run dev      # Start dev server on :3000 (proxies /ws and /api to backend)
npm run build    # Build for production
npm run lint     # Run ESLint
```

## Architecture

### Backend (Go)

- **Entry**: `backend/main.go` - Initializes Gin HTTP server and WebSocket handler
- **WebSocket**: `backend/internal/connection/connection.go` - Manages WebSocket connections
- **Message Handling**: `backend/internal/connection/handler.go` - Routes messages to appropriate handlers
- **Room Management**: `backend/internal/room/room.go` and `manager.go` - Handles room lifecycle
- **Game Logic**: `backend/internal/game/` - Pluggable game implementations

### Game Interface Pattern

All games implement `interfaces.Game` interface (`backend/internal/game/interfaces/game.go`):
- `Init(players []string)` - Initialize game with player IDs
- `ProcessAction(playerID string, action interface{}) (bool, error)` - Handle player actions
- `CurrentTurn() string` - Get current player's turn
- `GetState() / GetStateForPlayer(playerID string)` - Get game state
- `IsGameOver() / Winner()` - Game end detection
- `MaxPlayers() / MinPlayers()` - Player count limits

To add a new game:
1. Create `backend/internal/game/newgame/`
2. Implement the `Game` interface
3. Register in `backend/internal/game/factory.go`

### Supported Games

| Game | Players | Directory |
|------|---------|-----------|
| simple | 2 | `backend/internal/game/simple/` |
| ddz | 3 | `backend/internal/game/ddz/` |
| mahjong | 4 | `backend/internal/game/mahjong/` |
| durian | 4 | `backend/internal/game/durian/` |

### Frontend (React + TypeScript)

- **Routing**: React Router in `frontend/src/App.tsx`
- **WebSocket**: `frontend/src/services/WebSocketService.ts` - Manages WS connection
- **HTTP API**: `frontend/src/services/ApiService.ts` - REST calls for room list
- **Pages**: `frontend/src/pages/` - Lobby, GameRoom, MagicalLobbyPage
- **3D**: Uses Three.js for potential 3D game elements

### Communication Protocol

**WebSocket** (`ws://localhost:8080/ws?player=PLAYER_ID`):
- Client messages: `room.create`, `room.join`, `room.action`, `game.action`, `chat`
- Server messages: `broadcast`, `error`

**HTTP API** (`http://localhost:8080/api/`):
- `GET /api/rooms` - List available rooms

## Development Notes

- Frontend proxies `/ws` and `/api` to backend in dev mode (`vite.config.ts`)
- Backend uses Gin for HTTP, standard library for WebSocket
- Room states: waiting, playing, paused (disconnect), gameover
- Reconnection supported within 30 seconds