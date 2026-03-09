package mahjong

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// Suit 花色
type Suit int

const (
	SuitWan  Suit = 0 // 万
	SuitTiao Suit = 1 // 条
	SuitTong Suit = 2 // 筒
	SuitZi   Suit = 3 // 字
)

// 字牌编号
const (
	WindEast    = 1 // 东
	WindSouth   = 2 // 南
	WindWest    = 3 // 西
	WindNorth   = 4 // 北
	DragonRed   = 5 // 中
	DragonGreen = 6 // 发
	DragonWhite = 7 // 白
)

var suitNames = [4]string{"万", "条", "筒", "字"}
var suitCodes = [4]string{"wan", "tiao", "tong", "zi"}
var ziNames = [8]string{"", "东", "南", "西", "北", "中", "发", "白"}

// Tile 麻将牌
type Tile struct {
	Suit Suit `json:"suit"`
	Rank int  `json:"rank"` // 万条筒: 1-9, 字牌: 1-7
}

func (t Tile) String() string {
	if t.Suit == SuitZi && t.Rank >= 1 && t.Rank <= 7 {
		return ziNames[t.Rank]
	}
	return fmt.Sprintf("%d%s", t.Rank, suitNames[t.Suit])
}

func (t Tile) Equal(other Tile) bool {
	return t.Suit == other.Suit && t.Rank == other.Rank
}

func (t Tile) Less(other Tile) bool {
	if t.Suit != other.Suit {
		return t.Suit < other.Suit
	}
	return t.Rank < other.Rank
}

// IsHonor 是否字牌
func (t Tile) IsHonor() bool { return t.Suit == SuitZi }

// IsTerminal 是否老头牌(1或9)
func (t Tile) IsTerminal() bool {
	return !t.IsHonor() && (t.Rank == 1 || t.Rank == 9)
}

// IsTerminalOrHonor 是否幺九牌
func (t Tile) IsTerminalOrHonor() bool { return t.IsHonor() || t.IsTerminal() }

// IsWind 是否风牌
func (t Tile) IsWind() bool { return t.Suit == SuitZi && t.Rank >= 1 && t.Rank <= 4 }

// IsDragon 是否箭牌(中发白)
func (t Tile) IsDragon() bool { return t.Suit == SuitZi && t.Rank >= 5 && t.Rank <= 7 }

// ToIndex 转换为唯一索引(0-33)，用于计数数组
func (t Tile) ToIndex() int { return int(t.Suit)*9 + (t.Rank - 1) }

// TileFromIndex 从索引还原牌
func TileFromIndex(idx int) Tile {
	return Tile{Suit: Suit(idx / 9), Rank: idx%9 + 1}
}

// ToMap 转换为JSON友好的map
func (t Tile) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"suit": suitCodes[t.Suit],
		"rank": t.Rank,
	}
}

// ParseTile 从map解析牌（处理客户端JSON数据）
func ParseTile(data interface{}) (Tile, bool) {
	m, ok := data.(map[string]interface{})
	if !ok {
		return Tile{}, false
	}
	suitStr, _ := m["suit"].(string)

	var rank int
	switch v := m["rank"].(type) {
	case float64:
		rank = int(v)
	case int:
		rank = v
	default:
		return Tile{}, false
	}
	if rank < 1 {
		return Tile{}, false
	}

	var suit Suit
	switch suitStr {
	case "wan":
		suit = SuitWan
	case "tiao":
		suit = SuitTiao
	case "tong":
		suit = SuitTong
	case "zi":
		suit = SuitZi
		if rank > 7 {
			return Tile{}, false
		}
	default:
		return Tile{}, false
	}
	if suit != SuitZi && rank > 9 {
		return Tile{}, false
	}
	return Tile{Suit: suit, Rank: rank}, true
}

// Tiles 牌组（可排序）
type Tiles []Tile

func (ts Tiles) Len() int           { return len(ts) }
func (ts Tiles) Swap(i, j int)      { ts[i], ts[j] = ts[j], ts[i] }
func (ts Tiles) Less(i, j int) bool { return ts[i].Less(ts[j]) }

// Contains 是否包含某张牌
func (ts Tiles) Contains(t Tile) bool {
	for _, tile := range ts {
		if tile.Equal(t) {
			return true
		}
	}
	return false
}

// Count 统计某张牌的数量
func (ts Tiles) Count(t Tile) int {
	n := 0
	for _, tile := range ts {
		if tile.Equal(t) {
			n++
		}
	}
	return n
}

// Remove 移除一张牌，返回新切片
func (ts Tiles) Remove(t Tile) Tiles {
	for i, tile := range ts {
		if tile.Equal(t) {
			result := make(Tiles, 0, len(ts)-1)
			result = append(result, ts[:i]...)
			result = append(result, ts[i+1:]...)
			return result
		}
	}
	return ts
}

// RemoveN 移除N张相同牌
func (ts Tiles) RemoveN(t Tile, n int) Tiles {
	result := make(Tiles, 0, len(ts))
	removed := 0
	for _, tile := range ts {
		if tile.Equal(t) && removed < n {
			removed++
			continue
		}
		result = append(result, tile)
	}
	return result
}

// Copy 复制
func (ts Tiles) Copy() Tiles {
	result := make(Tiles, len(ts))
	copy(result, ts)
	return result
}

// ToMaps 批量转为map
func (ts Tiles) ToMaps() []map[string]interface{} {
	result := make([]map[string]interface{}, len(ts))
	for i, t := range ts {
		result[i] = t.ToMap()
	}
	return result
}

// ToCountArray 转为34元素计数数组
func (ts Tiles) ToCountArray() [34]int {
	var counts [34]int
	for _, t := range ts {
		counts[t.ToIndex()]++
	}
	return counts
}

// NewDeck 创建136张标准麻将牌
func NewDeck() Tiles {
	deck := make(Tiles, 0, 136)
	for suit := SuitWan; suit <= SuitTong; suit++ {
		for rank := 1; rank <= 9; rank++ {
			for i := 0; i < 4; i++ {
				deck = append(deck, Tile{Suit: suit, Rank: rank})
			}
		}
	}
	for rank := 1; rank <= 7; rank++ {
		for i := 0; i < 4; i++ {
			deck = append(deck, Tile{Suit: SuitZi, Rank: rank})
		}
	}
	return deck
}

// ShuffleDeck 洗牌
func ShuffleDeck(deck Tiles) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
}

// SortTiles 排序
func SortTiles(tiles Tiles) { sort.Sort(tiles) }
