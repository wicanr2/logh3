package game

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/screen"
	"github.com/wicanr2/logh3/internal/ui"
)

// 選了「移動」之後點一個過不去的地方：原版跳「這是無法行動的地方。」，
// 而且那個框是**模態**的——點別處沒有反應，只有點到「確認」才關
// （`docs/re/121`、`138`）。
func TestMoveBlockedShowsMessage(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	loadMessageBoxSpec(t, g)
	u := findUnit(g, g.faction, false)
	if u < 0 {
		t.Skip("這個局面畫面上沒有自己的分隊，跳過")
	}
	ux, uy := unitScreenPos(g, u)
	g.MoveMouse(ux+8, uy+8)
	g.Click(ux+8, uy+8)
	if !g.CommandMenuOpen() {
		t.Skip("點分隊沒開出命令選單，跳過")
	}
	s := g.res.CommandMenus
	g.Click(g.battleCmdX+s.Layout.HighlightDX+8, g.battleCmdY+s.Layout.HighlightDY+8)
	if !g.battleMove {
		t.Fatal("選了「移動」之後該在等玩家點目的地")
	}

	// 戰場那一塊外面一定不合法（`battleWorldAt` 的第一個判斷）。
	g.Click(0, 0)
	if !g.msgOn {
		t.Fatal("點在戰場外沒有跳訊息框")
	}
	if g.battleMove {
		t.Error("跳訊息框的同時命令就該作廢（原版回 0xFFFF）")
	}

	// 模態：點框外面沒有反應。
	g.Click(0, 0)
	if !g.msgOn {
		t.Error("點框外面把框關掉了——原版那個迴圈不理會別處的點擊")
	}

	// 點「確認」才關。
	tp := messageBoxButtonPoint(t, g.res.MessageBox, g.msgKind)
	g.Click(tp[0], tp[1])
	if g.msgOn {
		t.Error("點了「確認」框還在")
	}
}

// 半徑內沒有候選時，原版跳「沒有需要補給的船艦。」／「沒有需要修理的
// 船艦在附近。」（`docs/re/131`）。
func TestSupplyWithNoCandidateShowsMessage(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	loadMessageBoxSpec(t, g)
	// 直接走內部入口：這裡驗的是「沒有候選時會跳哪一句」，不是玩家路徑。
	u := findUnit(g, g.faction, false)
	if u < 0 {
		t.Skip("這個局面畫面上沒有自己的分隊，跳過")
	}
	for _, c := range []struct {
		repair bool
		name   string
	}{{false, "supply_none"}, {true, "repair_none"}} {
		if len(g.supplyCandidates(u, c.repair)) != 0 {
			continue // 這個分隊有候選，換下一種
		}
		g.msgOn = false
		g.beginSupplyPick(u, c.repair)
		if !g.msgOn {
			t.Errorf("%s：沒有候選卻沒跳訊息框", c.name)
			continue
		}
		_, want, _ := g.res.MessageBox.Message(c.name)
		if string(g.msgRaw) != string(want) {
			t.Errorf("%s：跳的不是那一句", c.name)
		}
	}
}

// messageBoxButtonPoint 回第 kind 種訊息框第一顆按鈕中間那一點。
func messageBoxButtonPoint(t *testing.T, s *screen.MessageBoxSpec, kind int) [2]int {
	t.Helper()
	for y := 0; y < 480; y += 4 {
		for x := 0; x < 640; x += 4 {
			if _, ok := s.MessageBoxButtonAt(kind, x, y); ok {
				return [2]int{x + ui.FrameTileW/2, y + ui.FrameTileH/2}
			}
		}
	}
	t.Fatal("找不到訊息框的按鈕")
	return [2]int{}
}

// loadMessageBoxSpec 補上訊息框版面（測試用的 Game 不走 bootstrap）。
func loadMessageBoxSpec(t *testing.T, g *Game) {
	t.Helper()
	if g.res.MessageBox != nil {
		return
	}
	s, err := screen.LoadMessageBoxSpec(filepath.Join(root,
		"docs/internals/data/message-box.json"))
	if err != nil {
		t.Skipf("缺訊息框的版面，跳過：%v", err)
	}
	g.res.MessageBox = s
}
