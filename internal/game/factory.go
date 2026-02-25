package game

import (
	"card-game-server/internal/game/ddz" // 新增
	"card-game-server/internal/game/interfaces"
	"card-game-server/internal/game/simple"
	"fmt"
)

// NewGame 支持 simple + ddz
func NewGame(gameType string) (interfaces.Game, error) {
	switch gameType {
	case "simple":
		return simple.New(), nil
	case "ddz":
		return ddz.New(), nil // 新增
	default:
		return nil, fmt.Errorf("unknown game type: %s", gameType)
	}
}
