package mahjong

// MeldType 副露类型
type MeldType int

const (
	MeldChow          MeldType = 0 // 吃（顺子）
	MeldPong          MeldType = 1 // 碰（刻子）
	MeldKongExposed   MeldType = 2 // 明杠（点杠）
	MeldKongConcealed MeldType = 3 // 暗杠
	MeldKongExtended  MeldType = 4 // 补杠（加杠）
)

var meldTypeNames = [5]string{"吃", "碰", "明杠", "暗杠", "补杠"}

// Meld 一组副露
type Meld struct {
	Type       MeldType `json:"type"`
	Tiles      Tiles    `json:"tiles"`
	FromPlayer int      `json:"from_player"` // 来源玩家座位号（暗杠为-1）
}

// IsKong 是否杠
func (m *Meld) IsKong() bool {
	return m.Type == MeldKongExposed || m.Type == MeldKongConcealed || m.Type == MeldKongExtended
}

// IsConcealed 是否暗副露（暗杠）
func (m *Meld) IsConcealed() bool { return m.Type == MeldKongConcealed }

// BaseTile 返回面子的基准牌（碰杠返回牌本身，吃返回最小牌）
func (m *Meld) BaseTile() Tile {
	if len(m.Tiles) == 0 {
		return Tile{}
	}
	if m.Type == MeldChow {
		min := m.Tiles[0]
		for _, t := range m.Tiles[1:] {
			if t.Less(min) {
				min = t
			}
		}
		return min
	}
	return m.Tiles[0]
}

// ToMap 转换为JSON友好的map
func (m *Meld) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"type":        meldTypeNames[m.Type],
		"tiles":       m.Tiles.ToMaps(),
		"from_player": m.FromPlayer,
	}
}

// MeldsToMaps 批量转换
func MeldsToMaps(melds []Meld) []map[string]interface{} {
	result := make([]map[string]interface{}, len(melds))
	for i := range melds {
		result[i] = melds[i].ToMap()
	}
	return result
}
