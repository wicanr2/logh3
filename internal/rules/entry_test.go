package rules

import (
	"path/filepath"
	"testing"
)

func loadEntries(t *testing.T) *BattleEntries {
	t.Helper()
	e, err := LoadBattleEntries(filepath.Join("../..", "docs/internals/data/battle-entry-points.json"))
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// 入口的鍵應該是相鄰星域（或星域自己）——這是拿星域相鄰圖來交叉驗證
// 入口表的判讀，兩張表的來源完全不同。
func TestBattleEntryKeysAreNeighbours(t *testing.T) {
	e := loadEntries(t)
	g, err := LoadStarGraph(filepath.Join("../..", "docs/internals/data/star-graph.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range e.Regions {
		near := map[int]bool{r.Region: true}
		for _, n := range g.Neighbours(r.Region) {
			near[n] = true
		}
		for _, x := range r.Entries {
			if !near[x.From] {
				t.Errorf("星域 %d 的入口鍵 %d 不是它的鄰域（鄰域 %v）",
					r.Region, x.From, g.Neighbours(r.Region))
			}
		}
	}
}

// 座標就是格乘上圖塊尺寸（`297B:0AFF`／`0B0D`）。
func TestBattleEntryCoordsAreTileMultiples(t *testing.T) {
	e := loadEntries(t)
	for _, r := range e.Regions {
		for _, x := range r.Entries {
			if x.X != x.Col*e.TileW || x.Y != x.Row*e.TileH {
				t.Errorf("星域 %d 來自 %d：格 (%d,%d) 對不上座標 (%d,%d)",
					r.Region, x.From, x.Col, x.Row, x.X, x.Y)
			}
			if x.Col >= 128 || x.Row >= 102 {
				t.Errorf("星域 %d 來自 %d：格 (%d,%d) 超出 128×102",
					r.Region, x.From, x.Col, x.Row)
			}
		}
	}
}

// 收據 s3ai6 那一場：在星域 26 打，帝國第 13 艦隊從星域 28 進來，
// 佈在格 (26,14) ＝ (416,280)（`docs/re/84`）。
func TestBattleEntryMatchesReceipt(t *testing.T) {
	e := loadEntries(t)
	x, ok := e.EntryFor(26, 28)
	if !ok {
		t.Fatal("星域 26 找不到來自 28 的入口")
	}
	if x.Col != 26 || x.Row != 14 || x.X != 416 || x.Y != 280 {
		t.Fatalf("算出格 (%d,%d) ＝ (%d,%d)，原版是 (26,14) ＝ (416,280)",
			x.Col, x.Row, x.X, x.Y)
	}
	// 同一場的另一邊：同盟第 1 艦隊從星域 27 進來，量到 (1536,1840)。
	y, ok := e.EntryFor(26, 27)
	if !ok {
		t.Fatal("星域 26 找不到來自 27 的入口")
	}
	if y.X != 1536 || y.Y != 1840 {
		t.Fatalf("算出 (%d,%d)，原版是 (1536,1840)", y.X, y.Y)
	}
}
