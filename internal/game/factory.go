package game

import (
	"card-game-server/internal/game/ddz"
	"card-game-server/internal/game/durian"
	"card-game-server/internal/game/interfaces"
	"card-game-server/internal/game/mahjong"
	"card-game-server/internal/game/simple"
	"fmt"
)

// NewGame 支持 simple + ddz + mahjong + durian
func NewGame(gameType string) (interfaces.Game, error) {
	switch gameType {
	case "simple":
		return simple.New(), nil
	case "ddz":
		return ddz.New(), nil
	case "mahjong":
		return mahjong.New(), nil
	case "durian":
		return durian.New(), nil
	default:
		return nil, fmt.Errorf("unknown game type: %s", gameType)
	}
}
