package game

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

func reachLoadFaction(t *testing.T) *Game {
	t.Helper()
	r := load(t)
	var err error
	r.Graph, err = rules.LoadStarGraph(filepath.Join(root, "docs/internals/data/star-graph.json"))
	if err != nil {
		t.Fatal(err)
	}
	g, err := New(r)
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, menuRowY(t, r, menuRowLoad))
	for _, c := range []byte("138\r") {
		g.Key(c)
	}
	wantState(t, g, StateLoadGame, "正常主選單取進度")
	before := bytes.Clone(g.res.Stellar)
	g.Click(300, r.SlotList.Top)
	wantState(t, g, StateLoadGame, "空存檔不開選單")
	g.Click(300, r.SlotList.Top+10*r.SlotList.Height)
	wantState(t, g, StateLoadFaction, "S3尚未讀取，先選陣營")
	if g.slot != nil || !bytes.Equal(g.res.Stellar, before) {
		t.Fatal("選陣營前已換局面")
	}
	return g
}

func TestLoadFactionOriginalMenuAndEdges(t *testing.T) {
	for sel := -1; sel < 3; sel++ {
		g := reachLoadFaction(t)
		s := g.res.LoadFactionMenu
		x, y := loadFactionLeft+s.Layout.HighlightDX, loadFactionTop+s.Layout.HighlightDY
		batch := "menu"
		if sel < 0 {
			g.MoveMouse(300, 322)
		} else {
			g.RightClick(x+14, y+sel*s.Layout.RowPitch+8)
			batch = "hover" + string(rune('0'+sel))
		}
		wantState(t, g, StateLoadFaction, "右鍵不選擇")
		p := filepath.Join(root, "workplace/oracle/normal89-r46-"+batch+"-a/screen.bin")
		orig, err := os.ReadFile(p)
		if err != nil {
			t.Skipf("缺原版選單收據：%v", err)
		}
		got := g.Frame().Pix
		w, h := s.Size(screen.CommandMenuLoadFaction, 3)
		for row := loadFactionTop; row < loadFactionTop+h; row++ {
			start := row*640 + loadFactionLeft
			if !bytes.Equal(orig[start:start+w], got[start:start+w]) {
				t.Fatalf("反白%d選單第%d行與原版不同", sel, row)
			}
		}
		for _, p := range [][2]int{{x - 1, y}, {x + 2*s.Layout.CellW, y}, {x, y - 1}, {x, y + 3*s.Layout.RowPitch}} {
			g.Click(p[0], p[1])
			wantState(t, g, StateLoadFaction, "邊界外不選擇")
		}
	}
}

func TestLoadFactionNormalSelectionAndSave(t *testing.T) {
	for mode, name := range []string{"empire", "alliance", "versus"} {
		t.Run(name, func(t *testing.T) {
			g := reachLoadFaction(t)
			// 從原版雙跑完整RAM取驗收值，未提供原版就skip。
			m, err := os.ReadFile(filepath.Join(root, "workplace/oracle/normal89-r46-"+name+"-a/loaded-mem.bin"))
			if err != nil {
				t.Skipf("缺原版座位收據：%v", err)
			}
			g.versus = true // 從舊雙人局讀成單人也必須重設。
			g.Click(loadFactionLeft+g.res.LoadFactionMenu.Layout.HighlightDX, loadFactionTop+g.res.LoadFactionMenu.Layout.HighlightDY+mode*g.res.LoadFactionMenu.Layout.RowPitch)
			want := StateCommandPhase
			if mode == 2 {
				want = StateVersusInput
			}
			wantState(t, g, want, "正常選擇陣營")
			f := mode
			if mode == 2 {
				f = 0
			}
			if g.faction != f || g.versus != (mode == 2) || g.turn != 380 {
				t.Fatal("選擇陣營、模式或日期錯誤")
			}
			if !bytes.Equal(g.workRecord()[0xC:0x15], m[0x5da50:0x5da59]) {
				t.Fatal("九個控制bytes不符合原版")
			}
			if !bytes.Equal(g.workRecord()[0x09:0x0B], m[0x5DA4D:0x5DA4F]) {
				t.Fatal("活動陣營／座位不符合原版")
			}
			if g.pct != g.slot.BuildPct[f] || g.treasury.Budget != g.slot.BudgetStart[f] || g.treasury.ShipPaid != g.slot.ShipPaid[f] {
				t.Fatal("財政沿用了另一陣營")
			}
			if g.Military() != int(binary.LittleEndian.Uint32(m[0x5da82+4*f:])) {
				t.Fatal("軍事費與原版不同")
			}
			dir := t.TempDir()
			if err := g.SaveSlot(dir, 0); err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile(filepath.Join(dir, "WORK.DT0"))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(b[0xC:0x15], m[0x5da50:0x5da59]) {
				t.Fatal("保存遺失座位控制")
			}
			if !bytes.Equal(b[0x09:0x0B], m[0x5DA4D:0x5DA4F]) {
				t.Fatal("保存遺失活動控制")
			}
			fleet := bytes.Clone(g.slot.Fleet)
			ship := bytes.Clone(g.slot.Ship)
			g.res.SaveDir = dir
			if mode == 2 {
				g.Click(432, 270)
			}
			wantState(t, g, StateCommandPhase, "讀檔確認後正常命令畫面")
			g.Click(544, 125)
			g.Click(560, 145)
			g.Click(300, g.res.SlotList.Top)
			wantState(t, g, StateLoadFaction, "正常保存後取進度")
			g.Click(loadFactionLeft+g.res.LoadFactionMenu.Layout.HighlightDX, loadFactionTop+g.res.LoadFactionMenu.Layout.HighlightDY+mode*g.res.LoadFactionMenu.Layout.RowPitch)
			if !bytes.Equal(g.slot.Fleet, fleet) || !bytes.Equal(g.slot.Ship, ship) {
				t.Fatal("完整艦隊／艦艇保存載入改變")
			}
		})
	}
}

func TestLoadFactionVersusFirstRoundKeepsGlobalFleetPool(t *testing.T) {
	g := reachLoadFaction(t)
	g.Click(318, 248)
	wantState(t, g, StateVersusInput, "雙人帝國交接")
	before := append([]rules.Fleet(nil), g.fleets...)
	g.Click(432, 270)
	wantState(t, g, StateCommandPhase, "載入雙人跳過新局月報")
	g.Click(544, 125)
	g.Click(560, 125)
	wantState(t, g, StateVersusInput, "帝國實行交接同盟")
	if g.faction != 1 || g.turn != 380 || !reflect.DeepEqual(g.fleets, before) {
		t.Fatal("交接改日期或丟失全域艦隊")
	}
	g.Click(432, 270)
	wantState(t, g, StateCommandPhase, "同盟確認直接命令畫面")
	if g.pct != g.slot.BuildPct[1] || g.treasury.Budget != g.slot.BudgetStart[1] {
		t.Fatal("同盟財政錯誤")
	}
	// 追加與刪除都可能重置slice標頭，交接不能復原過期的艦隊快照。
	g.fleets = append(g.fleets, rules.Fleet{Faction: 1, Region: 31})
	g.fleets = g.fleets[1:]
	after := append([]rules.Fleet(nil), g.fleets...)
	g.enterSide(0)
	g.enterSide(1)
	if !reflect.DeepEqual(g.fleets, after) {
		t.Fatal("兩次交接復原過期艦隊slice")
	}
}

func TestLoadFactionCorruptSlotPreservesWorld(t *testing.T) {
	g := reachLoadFaction(t)
	stellar := bytes.Clone(g.res.Stellar)
	planets := bytes.Clone(g.res.Planets.Data)
	f, v, turn := g.faction, g.versus, g.turn
	g.res.OrigDir = t.TempDir()
	g.Click(318, 228)
	wantState(t, g, StateLoadFaction, "缺檔保留選單")
	if g.faction != f || g.versus != v || g.turn != turn || g.slot != nil || !bytes.Equal(stellar, g.res.Stellar) || !bytes.Equal(planets, g.res.Planets.Data) {
		t.Fatal("讀檔失敗破壞原局面")
	}
}
