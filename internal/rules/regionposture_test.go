package rules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// 收據 `docs/internals/data/region-posture-s3ai6.json`：從 1 MB 傾印
// `s3ai6.bin` 倒出來的 32 格 `+0x24`／`+0x13` 與當下的擁有者與艦隊
// （`tools/scripts/region_posture.py`、`docs/re/244`）。
type posturePair struct {
	Owners  []int                                  `json:"owners"`
	Fleets  []struct{ Region, Faction, Flags int } `json:"fleets"`
	Regions []struct {
		Owner       int `json:"owner"`
		CapitalCost int `json:"capital_cost"`
		Flags       int `json:"flags"`
	} `json:"regions"`
}

// 第二、三份收據：`region-posture-chain4.json` 是整條歸屬鏈的同狀態驗證
// （新遊戲跑一回合、poke 一顆行星讓星域 4 易主），`region-posture-s3ai6.json`
// 是 S3 的局面。兩份的局面、擁有者與艦隊都不同（`docs/spec/033`）。
func TestPosturesMatchOracleReceipts(t *testing.T) {
	dir := filepath.Join("..", "..", "docs", "internals", "data")
	g, err := LoadStarGraph(filepath.Join(dir, "star-graph.json"))
	if err != nil {
		t.Fatalf("讀星域圖：%v", err)
	}
	for _, name := range []string{"region-posture-s3ai6.json", "region-posture-chain4.json"} {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("讀 %s：%v", name, err)
		}
		var rec posturePair
		if err := json.Unmarshal(b, &rec); err != nil {
			t.Fatalf("解 %s：%v", name, err)
		}
		fleets := make([]Fleet, 0, len(rec.Fleets))
		for _, f := range rec.Fleets {
			fleets = append(fleets, Fleet{Region: f.Region, Faction: f.Faction, Flags: f.Flags})
		}
		got := g.Postures(rec.Owners, fleets)
		occ := g.Occupations(rec.Owners, fleets)
		for i, w := range rec.Regions {
			if w.Owner != RegionOwnerNone && got[i].CapitalCost != w.CapitalCost {
				t.Errorf("%s 星域 %d 的 +0x24 ＝ %d，收據是 %d",
					name, i, got[i].CapitalCost, w.CapitalCost)
			}
			// 兩支常式合起來管 bit0–bit4。
			if bits := got[i].Bits() | occ[i].Bits(); bits != byte(w.Flags)&0x1D {
				t.Errorf("%s 星域 %d 的 +0x13 bit0–bit4 ＝ %02X，收據是 %02X",
					name, i, bits, byte(w.Flags)&0x1D)
			}
		}
	}
}

func TestPosturesMatchOracle(t *testing.T) {
	dir := filepath.Join("..", "..", "docs", "internals", "data")
	g, err := LoadStarGraph(filepath.Join(dir, "star-graph.json"))
	if err != nil {
		t.Fatalf("讀星域圖：%v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "region-posture-s3ai6.json"))
	if err != nil {
		t.Fatalf("讀收據：%v", err)
	}
	var rec posturePair
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatalf("解收據：%v", err)
	}
	fleets := make([]Fleet, 0, len(rec.Fleets))
	for _, f := range rec.Fleets {
		fleets = append(fleets, Fleet{Region: f.Region, Faction: f.Faction, Flags: f.Flags})
	}

	got := g.Postures(rec.Owners, fleets)
	for i, want := range rec.Regions {
		p := got[i]
		if want.Owner != RegionOwnerNone && p.CapitalCost != want.CapitalCost {
			t.Errorf("星域 %d 的 +0x24 ＝ %d，收據是 %d", i, p.CapitalCost, want.CapitalCost)
		}
		if want.Owner == RegionOwnerNone && p.CapitalCost != -1 {
			t.Errorf("星域 %d 無主，+0x24 應該不寫（-1），得到 %d", i, p.CapitalCost)
		}
		// bit0｜bit2｜bit3 逐位元比；其餘位元原版留著不動。
		if gotBits, wantBits := p.Bits(), byte(want.Flags)&0x0D; gotBits != wantBits {
			t.Errorf("星域 %d 的 +0x13 三個位元 ＝ %02X，收據是 %02X（原值 %02X）",
				i, gotBits, wantBits, want.Flags)
		}
		if kept := ApplyPosture(byte(want.Flags), p); kept != byte(want.Flags) {
			t.Errorf("星域 %d 套回去 ＝ %02X，收據是 %02X", i, kept, want.Flags)
		}
	}
}

// 無主星域不算 Far，也不寫 +0x24（原版那條 `cmp al, 2` 就直接跳過）。
func TestPostureSkipsNeutralRegion(t *testing.T) {
	dir := filepath.Join("..", "..", "docs", "internals", "data")
	g, err := LoadStarGraph(filepath.Join(dir, "star-graph.json"))
	if err != nil {
		t.Fatalf("讀星域圖：%v", err)
	}
	owner := make([]int, 32)
	for i := range owner {
		owner[i] = RegionOwnerNone
	}
	for _, p := range g.Postures(owner, nil) {
		if p.CapitalCost != -1 || p.Far || p.Frontier || p.Contested {
			t.Fatalf("整張無主時應該什麼都不設，得到 %+v", p)
		}
	}
}

// Contested 不計 `Flags` bit0 有設的那些（原版 `297B:14F8` 的 `test 1`）。
func TestPostureContestedSkipsSplitFleets(t *testing.T) {
	dir := filepath.Join("..", "..", "docs", "internals", "data")
	g, _ := LoadStarGraph(filepath.Join(dir, "star-graph.json"))
	owner := make([]int, 32)
	both := []Fleet{{Region: 5, Faction: 0}, {Region: 5, Faction: 1}}
	if !g.Postures(owner, both)[5].Contested {
		t.Error("雙方都有艦隊，應該是 Contested")
	}
	both[1].Flags = 1
	if g.Postures(owner, both)[5].Contested {
		t.Error("同盟那支的 bit0 有設，不該計入 Contested")
	}
}
