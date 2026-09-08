// Package ai 是電腦玩家的決策層（`docs/spec/017`）。
//
// **兩套並存**：`Original` 照原版——在玩家走得到的路上，原版按「實行」
// 時電腦不下任何命令（`docs/re/79`：兩個局面十八回合，決定計數器一直
// 是 0）。`Focused` 是強化版，會挑目標、集結、包圍。
//
// 預設走 `Original`，強化版由玩家自己開；這樣對拍收據與「與原版逐位元組
// 相同」那個宣稱才站得住。
package ai

import "github.com/wicanr2/logh3/internal/rules"

// Order 是一道命令：叫第 Fleet 支艦隊去 Target 那個星域。
//
// 對應原版的兩個欄位——`+0x07` 寫目的地、`+0x08` 解開移動狀態
// （`docs/internals/33`）。
type Order struct {
	Fleet  int `json:"fleet"`  // 索引，對應 World.Fleets
	Target int `json:"target"` // 目的地星域
}

// World 是 AI 看得到的局面。**只讀**——AI 只回命令，不自己改狀態。
type World struct {
	Fleets []rules.Fleet
	Graph  *rules.StarGraph

	// Owner 回某個星域屬於誰：0 帝國、1 同盟，其餘 ＝ 無主。
	Owner func(region int) int

	// Value 回攻下某個星域值多少（經濟力之類）。nil ＝ 每個都算 1。
	Value func(region int) int

	// Strength 回一支艦隊的戰力。nil ＝ 每支算 1。
	//
	// 原版自己的算式在 `2E54:1876`（`docs/re/79`）：每一組編制
	// `ds:6032[艦種+4] × 一單位幾艘 × 1000 ÷ 艦種+0x0A` 加總。
	// remake 還沒把艦艇池接進來，所以這裡開一個口，接上之後換掉即可。
	Strength func(f rules.Fleet) int
}

func (w World) owner(region int) int {
	if w.Owner == nil {
		return -1
	}
	return w.Owner(region)
}

func (w World) value(region int) int {
	if w.Value == nil {
		return 1
	}
	return w.Value(region)
}

func (w World) strength(f rules.Fleet) int {
	if w.Strength == nil {
		return 1
	}
	return w.Strength(f)
}

// Brain 是一套決策。每回合叫一次，回這一回合要下的命令。
type Brain interface {
	Orders(w World, faction, turn int) []Order
}

// Original 照原版：什麼都不做。
//
// **這不是還沒實作**——是量出來的行為（`docs/re/79`）。
type Original struct{}

// Orders 回 nil。
func (Original) Orders(World, int, int) []Order { return nil }
