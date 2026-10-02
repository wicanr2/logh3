package game

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 主選單→S3→正常戰場→兩段移動命令→實行→正式存讀；這是remake
// 玩家路徑驗證，原版同狀態證據另在38案單步測試，不混稱整戰場對拍。
func TestFleetMoveNormalExecuteAndSave(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	loadPreparationPresentation(t, g)
	var err error
	g.res.BattleMenu, err = screen.LoadBattleMenuSpec(filepath.Join(root, "docs/internals/data/battle-menu.json"))
	if err != nil {
		t.Fatal(err)
	}
	flag := findUnit(g, g.faction, true)
	if flag < 0 {
		t.Fatal("正常局面沒有可見己方旗艦")
	}
	fleetIndex := g.battleUnits[flag].Fleet
	bf := g.battleFleet(fleetIndex)
	if bf == nil || bf.Flag0 {
		t.Fatal("正常己方艦隊不能下移動命令")
	}
	ux, uy := unitScreenPos(g, flag)
	x0, y0 := g.battleUnits[flag].X, g.battleUnits[flag].Y
	g.MoveMouse(ux+8, uy+8)
	g.Click(ux+8, uy+8)
	if !g.CommandMenuOpen() {
		t.Fatal("正常點旗艦未開命令選單")
	}
	l := g.res.CommandMenus.Layout
	g.Click(g.battleCmdX+l.HighlightDX+8, g.battleCmdY+l.HighlightDY+8)
	if g.FleetOrderStage() != FleetOrderPoint {
		t.Fatal("正常移動命令未進選點")
	}
	tx, ty := -1, -1
	for _, dx := range []int{48, -48, 96, -96} {
		for _, dy := range []int{0, 40, -40} {
			x, y := ux+8+dx, uy+8+dy
			if _, _, ok := g.battleWorldAt(x, y); ok && tx < 0 {
				tx, ty = x, y
			}
		}
	}
	if tx < 0 {
		t.Fatal("正常視野無可走目的地")
	}
	g.MoveMouse(tx, ty)
	g.Click(tx, ty)
	if g.FleetOrderStage() != FleetOrderFacing {
		t.Fatal("未進抵達朝向選擇")
	}
	g.MoveMouse(tx+32, ty)
	g.Click(tx+32, ty)
	if !bf.Ordered || g.FleetOrderStage() != FleetOrderNone {
		t.Fatal("兩段命令未提交")
	}
	g.battleRand = rules.NewRand(0x12345678)
	turn := g.battleTurn
	g.MoveMouse(600, 240)
	g.Click(600, 240)
	if !g.battleMenu {
		t.Fatal("正常戰術選單未開")
	}
	b := g.res.BattleMenu.Layout
	g.Click(g.battleMenuX+b.Highlight.X-b.Window.X+2, g.battleMenuY+b.Highlight.Y-b.Window.Y+2)
	for i := 0; i < 4000 && (g.battlePreparing || g.state == StateTurnReport || g.state == StateEngagement); i++ {
		g.Frame()
		g.Tick(time.Second / 60)
	}
	if g.state != StateBattle || g.battleTurn != turn+1 {
		t.Fatalf("實行未回戰場下一回合：%v/%d", g.state, g.battleTurn)
	}
	if bf.X == x0 && bf.Y == y0 {
		t.Fatal("正常艦隊移動命令未改變位置")
	}
	if err := os.WriteFile(filepath.Join(root, "workplace/battle79-round30-normal.bin"), g.Frame().Pix, 0644); err != nil {
		t.Fatal(err)
	}
	var palette []byte
	for _, c := range g.Palette() {
		palette = append(palette, c[:]...)
	}
	if err := os.WriteFile(filepath.Join(root, "workplace/battle79-round30-normal.pal"), palette, 0644); err != nil {
		t.Fatal(err)
	}
	off := fleetIndex * rules.FleetRecord
	raw := g.slot.Fleet[off : off+rules.FleetRecord]
	if int(binary.LittleEndian.Uint16(raw[0x5C:])) != bf.X || int(binary.LittleEndian.Uint16(raw[0x5E:])) != bf.Y || int(raw[0x4A]) != bf.Facing || int(raw[0x49]) != bf.MoveFlags || int(int32(binary.LittleEndian.Uint32(raw[0x56:]))) != bf.Budget {
		t.Fatal("正式移動結果未寫回FLEET")
	}
	want := bytes.Clone(raw)
	dir := t.TempDir()
	if err := g.SaveSlot(dir, 0); err != nil {
		t.Fatal(err)
	}
	g.res.SaveDir = dir
	g.loadSlotFromDisk(0)
	if g.SaveErr() != nil {
		t.Fatal(g.SaveErr())
	}
	if !bytes.Equal(want, g.slot.Fleet[off:off+rules.FleetRecord]) {
		for i, v := range want {
			if got := g.slot.Fleet[off+i]; v != got {
				t.Logf("FLEET+%02X：存前%02X／讀後%02X", i, v, got)
			}
		}
		t.Fatal("完整艦隊記錄經正式存讀變動")
	}
	t.Logf("正常移動 (%d,%d)→(%d,%d)，回合%d；完整122-byte FLEET正式存讀相同", x0, y0, bf.X, bf.Y, turn+1)
}
