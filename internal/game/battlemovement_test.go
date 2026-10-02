package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// 正常入口後明示注入原版局面，只驗Game轉接與完整紀錄，不冒充正常同狀態GUI。
func TestGameMovementPhaseMatchesOriginal(t *testing.T) {
	mem, err := os.ReadFile(filepath.Join(root, "workplace/oracle/battle79-move-entry-a/mem-in.bin"))
	if os.IsNotExist(err) {
		t.Skip("缺原版移動入口RAM")
	}
	if err != nil {
		t.Fatal(err)
	}
	if got := sha256.Sum256(mem); fmtHash(got) != "9c4a4c28a1f53056962f613b1b643e5a5ba199d9231df291918d9499111be38d" {
		t.Fatal("原版RAM雜湊不同")
	}
	g := preparationGame(t, mem)
	// 資料測試的StartBattle重設索敵層級；畫面診斷須還原原版完整顯示狀態。
	for i := range g.battleUnits {
		u := &g.battleUnits[i]
		r := g.slot.ShipAt(g.battleUnitPool(*u))
		u.Detail, u.Seen, u.DetailHold = int(r[0x17]), r[0x16]&1 != 0, int(r[0x18])
	}
	loadMovementTiles(t, g)
	var d struct {
		Fleets []struct {
			Fleet int    `json:"fleet_index"`
			Entry string `json:"entry_hex"`
			End   string `json:"end_hex"`
		}
		Units []struct {
			Pool int    `json:"ship_index"`
			End  string `json:"end_hex"`
		}
	}
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/battle-movement-phase.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	g.battleFleets = nil
	for _, r := range d.Fleets {
		raw, err := hex.DecodeString(r.Entry)
		if err != nil {
			t.Fatal(err)
		}
		g.battleFleets = append(g.battleFleets, rules.NewBattleFleet(r.Fleet, raw, [rules.FleetAbilityCount]byte{}))
	}
	// 已證實鏡頭來源ds:9BF4/9BF6，線性DS=54B30，研究326。
	// 診斷明示注入原版鏡頭；不從截圖反推位置、不改正式玩家鏡頭。
	g.battleView = screen.BattleView{Col: int(binary.LittleEndian.Uint16(mem[0x5E724:])) / g.res.Battle.Field.TileW,
		Row: int(binary.LittleEndian.Uint16(mem[0x5E726:])) / g.res.Battle.Field.TileH}
	g.MoveMouse(240, 240) // 原版最後點擊240,200，產品畫布加40；游標相位仍分列。
	g.battleMoving = true // 明示診斷移動邊界；正常玩家測試則由實行自行設定。
	g.battlePhase = screen.BattlePhaseRepair
	movementFrame(t, g, "phase-entry")
	g.battlePhase = screen.BattlePhaseMovement
	m := g.newBattleMovement()
	if len(g.battleMovementGroups) != 7 {
		t.Fatal("正式Game艦隊清單不同")
	}
	n := 0
	for _, group := range g.battleMovementGroups {
		n += len(group.Units)
	}
	if n != 55 {
		t.Fatal("正式Game分隊清單不同")
	}
	count := 1
	for m.Next() {
		count++
		if count > 1000 {
			t.Fatal("交錯未停止")
		}
	}
	g.saveBattleMovement()
	movementFrame(t, g, "phase-end")
	g.battleMoving = false
	for _, r := range d.Fleets {
		want, err := hex.DecodeString(r.End)
		if err != nil {
			t.Fatal(err)
		}
		got := g.slot.Fleet[r.Fleet*rules.FleetRecord : (r.Fleet+1)*rules.FleetRecord]
		if !bytes.Equal(got, want) {
			t.Errorf("正式Game艦隊%d完整紀錄不同", r.Fleet)
		}
	}
	for _, r := range d.Units {
		want, err := hex.DecodeString(r.End)
		if err != nil {
			t.Fatal(err)
		}
		got := g.slot.Ship[r.Pool*rules.ShipRecord : (r.Pool+1)*rules.ShipRecord]
		if !bytes.Equal(got, want) {
			t.Errorf("正式Game池%d完整紀錄不同", r.Pool)
		}
	}
	if count != 33 {
		t.Errorf("正式Game交錯%d圈，要33", count)
	}
	t.Log("正式Game清單、隱藏索敵隊及保存轉接：7艦隊／55分隊完整紀錄零差異")
}

func fmtHash(v [32]byte) string { return hex.EncodeToString(v[:]) }

// 主選單→S3→戰術，所有命令走正常滑鼠，固定測試seed不改正式遊戲RNG。
func TestMovementCommandsExecuteAndSave(t *testing.T) {
	for _, mode := range []string{"long", "formation", "scout", "unit"} {
		t.Run(mode, func(t *testing.T) {
			g := commandTestGame(t)
			if g == nil {
				return
			}
			loadPreparationPresentation(t, g)
			var err error
			g.res.EditorTables, err = rules.LoadEditorTables(filepath.Join(root, "docs/internals/data/fleet-editor.json"))
			if err != nil {
				t.Fatal(err)
			}
			index := findUnit(g, g.faction, true)
			if mode == "formation" || mode == "unit" {
				index = -1
				for i, u := range g.battleUnits {
					if u.Faction == g.faction && u.Squadron > 0 && !u.Out {
						x, y := unitScreenPos(g, i)
						if _, _, ok := g.battleFieldPoint(x+8, y+8); ok {
							index = i
							break
						}
					}
				}
			}
			if index < 0 {
				t.Fatal("正常視野缺所需己方分隊")
			}
			start := g.battleUnits[index]
			fleet := g.battleFleet(start.Fleet)
			if fleet == nil || fleet.Flag0 {
				t.Fatal("正常命令不可用")
			}
			ux, uy := unitScreenPos(g, index)
			g.MoveMouse(ux+8, uy+8)
			g.Click(ux+8, uy+8)
			if !g.CommandMenuOpen() {
				t.Fatal("點選未開命令選單")
			}
			row := 0
			if mode == "formation" {
				row = 1
			}
			if mode == "scout" {
				row = 2
			}
			l := g.res.CommandMenus.Layout
			g.Click(g.battleCmdX+l.HighlightDX+8, g.battleCmdY+l.HighlightDY+row*l.RowPitch+8)
			selected := index
			if mode == "formation" {
				if !g.FormationPickPending() {
					t.Fatal("未開隊形選格")
				}
				picked := false
				for _, s := range g.FormationSlots() {
					if s.Index == start.Slot {
						continue
					}
					x := s.X - g.battleView.Col*g.res.Battle.Field.TileW + g.res.Battle.Field.X + 8
					y := s.Y - g.battleView.Row*g.res.Battle.Field.TileH + g.res.Battle.Field.Y + 8
					if _, _, ok := g.battleFieldPoint(x, y); ok {
						g.Click(x, y)
						picked = true
						break
					}
				}
				if !picked || g.FormationPickPending() || g.battleUnits[index].Slot == start.Slot {
					t.Fatal("正常選格未提交")
				}
			} else {
				if mode == "scout" {
					if !g.MoveTargetPending() {
						t.Fatal("索敵未進選點")
					}
					selected = len(g.battleUnits) - 1
				}
				picked := false
				for _, dx := range []int{160, -160, 120, -120, 64, -64} {
					x, y := ux+8+dx, uy+8
					if _, _, ok := g.battleWorldAt(x, y); !ok {
						continue
					}
					g.MoveMouse(x, y)
					g.Click(x, y)
					if mode == "long" {
						if g.FleetOrderStage() != FleetOrderFacing {
							continue
						}
						g.MoveMouse(x+24, y)
						g.Click(x+24, y)
						if fleet.ShortMove {
							t.Fatal("長程抽樣意外成短程")
						}
					} else if g.MoveTargetPending() {
						continue
					}
					picked = true
					break
				}
				if !picked {
					t.Fatal("正常視野找不到可走目的地")
				}
			}
			before := g.battleUnits[selected]
			g.battleRand = rules.NewRand(0x12345678)
			g.MoveMouse(600, 240)
			g.Click(600, 240)
			if !g.battleMenu {
				t.Fatal("正常實行未開戰術選單")
			}
			b := g.res.BattleMenu.Layout
			g.Click(g.battleMenuX+b.Highlight.X-b.Window.X+2, g.battleMenuY+b.Highlight.Y-b.Window.Y+2)
			frames := 0
			for i := 0; i < 12000 && (g.battlePreparing || g.battleMoving || g.battleSpotting.active || g.state == StateTurnReport || g.state == StateEngagement); i++ {
				if g.battleMoving {
					if g.battlePhase != screen.BattlePhaseMovement {
						t.Fatal("正常移動未切換階段文字")
					}
					frames++
					if frames == 2 {
						movementFrame(t, g, mode+"-moving")
						turn := g.battleTurn
						g.Click(600, 240)
						g.RightClick(600, 240)
						g.AdvanceBattleTurn()
						if g.battleMenu || g.battleTurn != turn {
							t.Fatal("移動中可重入命令")
						}
					}
				}
				g.Frame()
				g.Tick(time.Second / 60)
			}
			if g.battleMoving || g.battlePreparing || g.battleSpotting.active || g.state != StateBattle || g.battleTurn != 2 || frames < 2 {
				t.Fatalf("正常實行未完成／無中間幀：%v、回合%d、%d幀", g.state, g.battleTurn, frames)
			}
			after := g.battleUnits[selected]
			if g.battlePhase != screen.BattlePhaseInput {
				t.Fatal("正常回合結束未還原輸入面板")
			}
			if after.X == before.X && after.Y == before.Y {
				t.Fatal("所選單位未移動")
			}
			movementFrame(t, g, mode+"-after")
			// SyncFleets不得用戰略編制舊座標覆蓋正式移動。驗每支現有分隊的移動欄位。
			wants := map[int][]byte{}
			for _, u := range g.battleUnits {
				pool := g.battleUnitPool(u)
				if pool < 0 {
					t.Fatal("丟失池身份")
				}
				wants[pool] = bytes.Clone(g.slot.Ship[pool*rules.ShipRecord : (pool+1)*rules.ShipRecord])
			}
			dir := t.TempDir()
			if err := g.SaveSlot(dir, 0); err != nil {
				t.Fatal(err)
			}
			for pool, want := range wants {
				r := g.slot.Ship[pool*rules.ShipRecord : (pool+1)*rules.ShipRecord]
				for _, off := range []int{5, 0x16, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2A, 0x2B, 0x34, 0x35, 0x36, 0x37} {
					if r[off] != want[off] {
						t.Fatalf("池%d+%02X被存檔覆蓋：%02X／%02X", pool, off, want[off], r[off])
					}
				}
			}
			ship, fb := bytes.Clone(g.slot.Ship), bytes.Clone(g.slot.Fleet)
			g.res.SaveDir = dir
			g.loadSlotFromDisk(0)
			if g.SaveErr() != nil {
				t.Fatal(g.SaveErr())
			}
			if !bytes.Equal(ship, g.slot.Ship) || !bytes.Equal(fb, g.slot.Fleet) {
				t.Fatal("完整SHIP／FLEET正式存讀不同")
			}
			t.Logf("正常%s：(%d,%d)→(%d,%d)，%d中間幀，%d池移動欄位與完整SHIP／FLEET存讀相同", mode, before.X, before.Y, after.X, after.Y, frames, len(wants))
		})
	}
}

func movementFrame(t *testing.T, g *Game, label string) {
	t.Helper()
	prefix := os.Getenv("LOGH3_MOVEMENT_OUTPUT")
	if prefix == "" {
		return
	}
	p := g.Palette()
	var pal []byte
	for _, rgb := range p {
		pal = append(pal, rgb[:]...)
	}
	for _, f := range []struct {
		ext string
		b   []byte
	}{{"bin", g.Frame().Pix}, {"pal", pal}} {
		path := filepath.Join(root, "workplace", prefix+"-"+label+"."+f.ext)
		if err := os.WriteFile(path, f.b, 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func loadMovementTiles(t *testing.T, g *Game) {
	t.Helper()
	if g.res.BattleTiles != nil {
		return
	}
	b, err := os.ReadFile(filepath.Join(root, "org_game/extracted", g.res.Battle.Field.Tiles))
	if err != nil {
		t.Fatal(err)
	}
	g.res.BattleTiles, err = asset.DecodeTiles(b, g.res.Battle.Field.TileW, g.res.Battle.Field.TileH)
	if err != nil {
		t.Fatal(err)
	}
}

// 正常索敵命令與實行，驗回收後原池重用及不可點選；不注入回收旗標。
func TestScoutRecoverAndRelaunchNormal(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	loadPreparationPresentation(t, g)
	loadMovementTiles(t, g)
	var err error
	g.res.EditorTables, err = rules.LoadEditorTables(filepath.Join(root, "docs/internals/data/fleet-editor.json"))
	if err != nil {
		t.Fatal(err)
	}
	flag := findUnit(g, g.faction, true)
	if flag < 0 {
		t.Fatal("正常視野缺己方旗艦")
	}
	fleet := g.battleUnits[flag].Fleet
	bf := g.battleFleet(fleet)
	if bf == nil || bf.Flag0 || bf.ScoutOut {
		t.Fatal("正常旗艦不能放出索敵隊")
	}
	f, ok := g.saveFleet(fleet)
	if !ok {
		t.Fatal("缺原始艦隊")
	}
	pool := f.Scout
	base := len(g.battleUnits)
	clickRow := func(index, row int) {
		t.Helper()
		x, y := unitScreenPos(g, index)
		g.MoveMouse(x+8, y+8)
		g.Click(x+8, y+8)
		if !g.CommandMenuOpen() || g.battleCmdUnit != index {
			t.Fatalf("正常點選%d命中%d，選單%v", index, g.battleCmdUnit, g.CommandMenuOpen())
		}
		l := g.res.CommandMenus.Layout
		g.Click(g.battleCmdX+l.HighlightDX+8, g.battleCmdY+l.HighlightDY+row*l.RowPitch+8)
	}
	deploy := func() int {
		t.Helper()
		clickRow(flag, 2)
		if !g.MoveTargetPending() {
			t.Fatal("索敵未進選點")
		}
		index := g.battleMoveUnit
		if index < 0 || g.battleUnitPool(g.battleUnits[index]) != pool {
			t.Fatal("索敵原池身份不同")
		}
		x, y := unitScreenPos(g, flag)
		picked := false
		for _, dx := range []int{64, -64, 96, -96} {
			if _, _, ok := g.battleWorldAt(x+8+dx, y+8); ok {
				g.MoveMouse(x+8+dx, y+8)
				g.Click(x+8+dx, y+8)
				picked = true
				break
			}
		}
		if !picked || g.MoveTargetPending() || g.battleUnits[index].Out {
			t.Fatal("正常放出選點失敗")
		}
		return index
	}
	// 測試seed在第一次實行前固定一次，不重擲。
	g.battleRand = rules.NewRand(0x12345678)
	execute := func() {
		t.Helper()
		g.MoveMouse(600, 240)
		g.Click(600, 240)
		if !g.battleMenu {
			t.Fatal("未開戰術選單")
		}
		b := g.res.BattleMenu.Layout
		g.Click(g.battleMenuX+b.Highlight.X-b.Window.X+2, g.battleMenuY+b.Highlight.Y-b.Window.Y+2)
		settleBattleExecution(t, g)
		if g.state != StateBattle {
			t.Fatal("索敵流程意外離開戰場")
		}
	}
	scout := deploy()
	if len(g.battleUnits) != base+1 {
		t.Fatal("首次放出數量不同")
	}
	movementFrame(t, g, "scout-deployed")
	execute()
	if g.battleUnits[scout].Out {
		t.Fatal("首次正常移動意外已回收")
	}
	movementFrame(t, g, "scout-away")
	clickRow(scout, 1)
	if g.battleUnits[scout].Action != squadronRejoin {
		t.Fatal("歸隊命令未設3")
	}
	turns := 0
	for !g.battleUnits[scout].Out && turns < 6 {
		execute()
		turns++
	}
	u := g.battleUnits[scout]
	// 原版26E5刷新仍處理已回收隊；Detail1時會重新設C5，不能用Out替代標記（381／082）。
	oracleRaw, err := os.ReadFile(filepath.Join(root, "docs/internals/data/battle-spotting-display.json"))
	if err != nil {
		t.Fatal(err)
	}
	var markerOracle struct {
		RecoveredScout struct {
			Rows []struct{ Before, After map[string]string }
		} `json:"recovered_scout"`
	}
	if err := json.Unmarshal(oracleRaw, &markerOracle); err != nil {
		t.Fatal(err)
	}
	wantMarker := -1
	for _, row := range markerOracle.RecoveredScout.Rows {
		before, err := hex.DecodeString(row.Before["scout"])
		if err != nil || len(before) != rules.ShipRecord {
			t.Fatal("原版回收紀錄格式不同")
		}
		if int(before[0x17]) == u.Detail {
			after, err := hex.DecodeString(row.After["scout"])
			if err != nil || len(after) != rules.ShipRecord {
				t.Fatal("原版刷新後紀錄格式不同")
			}
			wantMarker = int(after[0x19])
		}
	}
	if wantMarker < 0 || !u.Out || u.Action != 0xFF || u.MarkerFlags != wantMarker || u.Marked != (wantMarker&0x80 != 0) || bf.ScoutOut || bf.ScoutMode != 0xFF || bf.ScoutState != 0 {
		t.Fatalf("正常回收未完成：%+v／%+v", u, *bf)
	}
	x, y := unitScreenPos(g, scout)
	if g.battleUnitAt(x+8, y+8) == scout {
		t.Fatal("已回收索敵隊仍攔截正常點選")
	}
	movementFrame(t, g, "scout-recovered")
	dir := t.TempDir()
	if err := g.SaveSlot(dir, 0); err != nil {
		t.Fatal(err)
	}
	if g.slot.Ship[pool*rules.ShipRecord+0x1E] != 0xFF || int(g.slot.Ship[pool*rules.ShipRecord+0x19]) != wantMarker || g.slot.Fleet[fleet*rules.FleetRecord+0x3A] != 0xFF {
		t.Fatal("正式存檔遺失回收狀態")
	}
	ship, fleets := bytes.Clone(g.slot.Ship), bytes.Clone(g.slot.Fleet)
	loaded := newTestGame(t)
	loaded.res.SaveDir = dir
	loaded.loadSlotFromDisk(0)
	if loaded.SaveErr() != nil || !bytes.Equal(ship, loaded.slot.Ship) || !bytes.Equal(fleets, loaded.slot.Fleet) {
		t.Fatal("回收後完整SHIP／FLEET正式存讀不同")
	}
	if next := deploy(); next != scout || len(g.battleUnits) != base+1 {
		t.Fatal("再放出未重用原池與既有單位")
	}
	if !bf.ScoutOut || bf.ScoutMode != 0 || bf.ScoutState != 0 || g.battleUnits[scout].Action != 0 || g.battleUnits[scout].MarkerFlags&0x80 == 0 {
		t.Fatal("再放出命令旗標不符")
	}
	movementFrame(t, g, "scout-relaunched")
	execute()
	if g.battleUnitPool(g.battleUnits[scout]) != pool {
		t.Fatal("再放出實行丟失原池")
	}
	t.Logf("正常索敵放出／移動／歸隊%d回合／回收／再放出，原池%d、單位%d保持", turns, pool, scout)
}
