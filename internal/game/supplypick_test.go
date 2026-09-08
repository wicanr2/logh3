package game

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// loadKinds 把艦種表補進測試局面——那份資料在版控裡，不是原版素材。
func loadKinds(t *testing.T, g *Game) {
	t.Helper()
	if g.res.ShipKinds != nil {
		return
	}
	k, err := rules.LoadShipKindTable(filepath.Join(root,
		"docs/internals/data/ship-kind-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.ShipKinds = k
}

// 從玩家路徑走一次「補給」：點運補艦 → 選單第三項 → 點一個候選，
// 對象的飛彈與艦載機補上去、運補艦身上少掉同樣多（`docs/re/131`）。
func TestSupplyOrderMovesAmmo(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	loadKinds(t, g)
	// 這個局面沒有自己的運補艦，所以**造一個**：把一個點得到的分隊
	// 改成艦種 5。這是 fixture 不是收據——驗的是玩家路徑接得起來，
	// 數字本身另外對原版（`internal/rules` 的 `TestSupplyCapacityMatchesPool`）。
	src, dst := -1, -1
	for i := range g.battleUnits {
		u := g.battleUnits[i]
		if u.Faction != g.faction || u.Squadron == 0 {
			continue
		}
		x, y := unitScreenPos(g, i)
		if !inField(g, x, y) {
			continue
		}
		if src < 0 {
			src = i
			continue
		}
		if dst < 0 {
			dst = i
		}
	}
	if src < 0 || dst < 0 {
		t.Skip("這個局面畫面上不足兩個自己的分隊，跳過")
	}
	g.battleUnits[src].Kind = shipKindSupply
	// 兩者必須落在收集半徑內，否則這個 fixture 沒有意義。
	me := g.battleUnits[src]
	r := rules.SupplyListRange(g.res.ShipKinds, me.Faction, rules.SupplyKind)
	if !rules.InSupplyRange(me.X, me.Y, g.battleUnits[dst].X, g.battleUnits[dst].Y, r) {
		t.Skip("兩個分隊不在補給半徑內，跳過")
	}

	// 把對象的彈藥打掉一半，運補艦裝滿。
	g.battleUnits[dst].Missiles /= 2
	g.battleUnits[dst].Fighters /= 2
	g.battleUnits[src].Missiles = 1 << 20
	g.battleUnits[src].Fighters = 1 << 20
	before := g.battleUnits[dst]
	k, ok := g.res.ShipKinds.Kind(before.Faction, before.Kind)
	if !ok {
		t.Skip("對象的艦種不在表裡，跳過")
	}

	ux, uy := unitScreenPos(g, src)
	g.MoveMouse(ux+8, uy+8)
	g.Click(ux+8, uy+8)
	if !g.CommandMenuOpen() || g.battleCmdKind != screen.CommandMenuSupply {
		t.Fatalf("點運補艦開出來的是 %v，該是運補那張", g.battleCmdKind)
	}
	s := g.res.CommandMenus
	g.Click(g.battleCmdX+s.Layout.HighlightDX+8,
		g.battleCmdY+s.Layout.HighlightDY+2*s.Layout.RowPitch+8)
	if !g.SupplyPickPending() {
		t.Fatal("選了「補給」之後應該在等玩家點對象")
	}

	tx, ty := unitScreenPos(g, dst)
	g.MoveMouse(tx+8, ty+8)
	g.Click(tx+8, ty+8)
	if g.SupplyPickPending() {
		t.Error("點完對象之後不該還在等")
	}
	got := g.battleUnits[dst]
	wantM := k.Byte(0x10) * got.Ships
	wantF := k.Byte(0x14) * got.Ships
	if got.Missiles != wantM || got.Fighters != wantF {
		t.Errorf("補完是 %d／%d，公式算出 %d／%d",
			got.Missiles, got.Fighters, wantM, wantF)
	}
	gaveM := wantM - before.Missiles
	gaveF := wantF - before.Fighters
	if g.battleUnits[src].Missiles != 1<<20-gaveM ||
		g.battleUnits[src].Fighters != 1<<20-gaveF {
		t.Errorf("運補艦少了 %d／%d，該少 %d／%d",
			1<<20-g.battleUnits[src].Missiles, 1<<20-g.battleUnits[src].Fighters,
			gaveM, gaveF)
	}
}

// 候選的條件：同陣營、沒出局、在半徑內；補給排除艦種 5／6／7，
// 修理只排除 6／7——**修理修得了運補艦**（`1930:0485` 對 `1AAB:043D`）。
func TestSupplyCandidatesExcludeKinds(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	loadKinds(t, g)
	src := -1
	for i, u := range g.battleUnits {
		if u.Faction == g.faction && u.Squadron != 0 {
			src = i
			break
		}
	}
	if src < 0 {
		t.Skip("這個局面沒有自己的分隊，跳過")
	}
	me := g.battleUnits[src]
	for _, repair := range []bool{false, true} {
		kind := rules.SupplyKind
		if repair {
			kind = rules.RepairKind
		}
		r := rules.SupplyListRange(g.res.ShipKinds, me.Faction, kind)
		got := g.supplyCandidates(src, repair)
		in := map[int]bool{}
		for _, c := range got {
			in[c] = true
		}
		for i, u := range g.battleUnits {
			want := u.Faction == me.Faction && !u.Out &&
				u.Kind != shipKindRepair && u.Kind != shipKindScout &&
				(repair || u.Kind != shipKindSupply) &&
				rules.InSupplyRange(me.X, me.Y, u.X, u.Y, r)
			if in[i] != want {
				t.Errorf("修理=%v：第 %d 個（艦種 %d、陣營 %d）收了 %v，該是 %v",
					repair, i, u.Kind, u.Faction, in[i], want)
			}
		}
	}
}

func inField(g *Game, x, y int) bool {
	f := g.res.Battle.Field
	return x >= f.X && y >= f.Y &&
		x+battleUnitTile <= f.X+f.Cols*f.TileW &&
		y+battleUnitTile <= f.Y+f.Rows*f.TileH
}
