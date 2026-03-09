package durian

import "fmt"

// AngerToken 愤怒标记
type AngerToken struct {
	Value int `json:"value"` // 分值（1-7）
}

// AngerTokenPool 公共标记池
type AngerTokenPool struct {
	Remaining []*AngerToken `json:"remaining"` // 剩余标记（按分值升序排列）
}

// NewAngerTokenPool 初始化标记池（7枚，分值1-7）
func NewAngerTokenPool() *AngerTokenPool {
	tokens := make([]*AngerToken, 0, 7)
	for i := 1; i <= 7; i++ {
		tokens = append(tokens, &AngerToken{Value: i})
	}
	return &AngerTokenPool{Remaining: tokens}
}

// TakeSmallest 取出当前剩余分值最小的标记
func (p *AngerTokenPool) TakeSmallest() (*AngerToken, error) {
	if len(p.Remaining) == 0 {
		return nil, fmt.Errorf("标记池已空")
	}
	// 已按升序排列，取第一枚
	token := p.Remaining[0]
	p.Remaining = p.Remaining[1:]
	return token, nil
}

// IsEmpty 标记池是否为空
func (p *AngerTokenPool) IsEmpty() bool {
	return len(p.Remaining) == 0
}

// RemainingValues 返回剩余标记的分值列表（用于状态广播）
func (p *AngerTokenPool) RemainingValues() []int {
	vals := make([]int, len(p.Remaining))
	for i, t := range p.Remaining {
		vals[i] = t.Value
	}
	return vals
}

// PlayerAnger 玩家持有的愤怒标记
type PlayerAnger struct {
	PlayerID string        `json:"player_id"`
	Tokens   []*AngerToken `json:"tokens"` // 持有的标记列表
}

// AddToken 添加一枚愤怒标记
func (pa *PlayerAnger) AddToken(token *AngerToken) {
	pa.Tokens = append(pa.Tokens, token)
}

// TotalScore 计算愤怒总分
func (pa *PlayerAnger) TotalScore() int {
	total := 0
	for _, t := range pa.Tokens {
		total += t.Value
	}
	return total
}

// IsEliminated 是否达到淘汰条件（总分>=7）
func (pa *PlayerAnger) IsEliminated() bool {
	return pa.TotalScore() >= 7
}

// TokenValues 返回持有标记的分值列表（用于状态广播）
func (pa *PlayerAnger) TokenValues() []int {
	vals := make([]int, len(pa.Tokens))
	for i, t := range pa.Tokens {
		vals[i] = t.Value
	}
	return vals
}
