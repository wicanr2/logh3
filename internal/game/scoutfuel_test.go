package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
	"github.com/wicanr2/logh3/internal/ui"
)

type scoutFuelCall struct {
	Entry string
	Args  []int
	Fleet string `json:"fleet_hex"`
	Ship  string `json:"ship_hex"`
	Text  string `json:"text_hex"`
}
type scoutFuelCase struct {
	Name                    string
	Fleet, Pool, Mode, Fuel int
	Humans                  int    `json:"human_mask"`
	Count                   int    `json:"fleet_count"`
	BeforeFleet             string `json:"before_fleet_hex"`
	AfterFleet              string `json:"after_fleet_hex"`
	BeforeShip              string `json:"before_ship_hex"`
	AfterShip               string `json:"after_ship_hex"`
	Calls                   []scoutFuelCall
}
type scoutFuelManifest struct {
	Cases   []scoutFuelCase
	Sources map[string]struct {
		SHA256 string
		Size   int
	}
}

func scoutFuelSource(t *testing.T) scoutFuelManifest {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/scout-fuel.json"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != "9934ca4f7832739eaea663e851c82c4684d7d53b20432267ff2f76f329d35ef0" {
		t.Fatal("第83輪原版凍結來源不符", err)
	}
	var m scoutFuelManifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func (m scoutFuelManifest) read(t *testing.T, name string) []byte {
	t.Helper()
	s, ok := m.Sources[name]
	if !ok {
		t.Fatal("未登錄的原版來源", name)
	}
	b, err := os.ReadFile(filepath.Join(root, name))
	if os.IsNotExist(err) {
		t.Skip("缺原版收據；重生入口tools/scout_fuel_oracle.sh", name)
	}
	if err != nil || len(b) != s.Size || fmt.Sprintf("%x", sha256.Sum256(b)) != s.SHA256 {
		t.Fatal("原版來源不符", name, err)
	}
	return b
}

// 窄幅原版控制狀態；完整兩張原池底稿及typed資料，不冒充正常玩家流程。
func scoutFuelFixture(t *testing.T, base *Game, c scoutFuelCase, visible bool) *Game {
	t.Helper()
	slot := *base.slot
	slot.Fleet, slot.Ship = bytes.Clone(slot.Fleet), bytes.Clone(slot.Ship)
	copy(slot.Fleet[c.Fleet*rules.FleetRecord:], ownershipHex(t, c.BeforeFleet))
	copy(slot.Ship[c.Pool*rules.ShipRecord:], ownershipHex(t, c.BeforeShip))
	res := *base.res
	if res.FleetWindow == nil {
		var err error
		res.FleetWindow, err = screen.LoadFleetWindowSpec(filepath.Join(root, "docs/internals/data/fleet-window.json"))
		if err != nil {
			t.Fatal(err)
		}
	}
	g := &Game{res: &res, slot: &slot, state: StateBattle, faction: -1, battleRand: rules.NewRand(0x12345678)}
	g.fleets = slot.Fleets(res.Weights)
	switch c.Humans {
	case 1:
		g.faction = 0
	case 2:
		g.faction = 1
	case 3:
		g.faction, g.versus = 0, true
	}
	if c.Count != 0 {
		r := slot.Fleet[c.Fleet*rules.FleetRecord : (c.Fleet+1)*rules.FleetRecord]
		g.battleFleets = []rules.BattleFleet{rules.NewBattleFleet(c.Fleet, r, [rules.FleetAbilityCount]byte{})}
	}
	if visible {
		f := g.fleetByIndex(c.Fleet)
		r := slot.ShipAt(c.Pool)
		u, ok := slot.ScoutUnit(res.Weights, *f, int(binary.LittleEndian.Uint16(r[0x34:])), int(binary.LittleEndian.Uint16(r[0x36:])))
		if !ok {
			t.Fatal("原版索敵池無法解碼")
		}
		u.TargetX, u.TargetY = int(binary.LittleEndian.Uint16(r[0x28:])), int(binary.LittleEndian.Uint16(r[0x2A:]))
		g.battleUnits = []rules.BattleUnit{u}
	}
	return g
}

func TestScoutFuelOriginalCases(t *testing.T) {
	m := scoutFuelSource(t)
	base := newTestGame(t)
	reports := 0
	for _, c := range m.Cases {
		// 原池存在但沒有可見typed索敵隊也必須同步；兩種合法保存形態均核完整byte。
		for _, visible := range []bool{false, true} {
			g := scoutFuelFixture(t, base, c, visible)
			beforeF, beforeS := bytes.Clone(g.slot.Fleet), bytes.Clone(g.slot.Ship)
			done := 0
			g.battleScoutFuelPhase(0, func() { done++ })
			for _, call := range c.Calls {
				if call.Entry != "3826:0405" {
					continue
				}
				if done != 0 || g.state != StateTurnReport || !g.turnRep.battle || len(g.turnRep.pending) != 1 {
					t.Fatal(c.Name, "報告沒有等待")
				}
				r := g.turnRep.pending[0]
				if r.face != call.Args[3]&255 || !bytes.Equal(r.text, ownershipHex(t, call.Text)) ||
					!bytes.Equal(g.slot.Fleet[c.Fleet*122:(c.Fleet+1)*122], ownershipHex(t, call.Fleet)) ||
					!bytes.Equal(g.slot.ShipAt(c.Pool), ownershipHex(t, call.Ship)) {
					t.Fatalf("%s 報告文字／頭像／回收前記錄不符：face=%d/%d text=%x/%s fleet=%x/%s ship=%x/%s", c.Name, r.face, call.Args[3]&255, r.text, call.Text, g.slot.Fleet[c.Fleet*122:(c.Fleet+1)*122], call.Fleet, g.slot.ShipAt(c.Pool), call.Ship)
				}
				g.Frame()
				g.Click(240, 240)
				if visible {
					reports++
				}
			}
			copy(beforeF[c.Fleet*122:], ownershipHex(t, c.AfterFleet))
			copy(beforeS[c.Pool*78:], ownershipHex(t, c.AfterShip))
			if done != 1 || g.state != StateBattle || !bytes.Equal(beforeF, g.slot.Fleet) || !bytes.Equal(beforeS, g.slot.Ship) || g.battleRand.Seed() != 0x12345678 {
				t.Fatal(c.Name, "完整兩池／續接／seed不符")
			}
			if c.Count != 0 {
				r := ownershipHex(t, c.AfterFleet)
				bf := g.battleFleets[0]
				if bf.ScoutMode != int(r[0x3A]) || bf.ScoutState != int(r[0x3B]) || bf.ScoutOut != (r[0x3A] != 255 || r[0x3B] != 0) {
					t.Fatal(c.Name, "typed燃料不同步")
				}
				if visible && c.Mode != 255 && byte(c.Fuel+1) >= 4 {
					u := g.battleUnits[0]
					if !u.Out || u.Marked || u.Action != 255 || u.MarkerFlags != int(beforeS[c.Pool*78+0x19]) ||
						u.X != bf.X || u.Y != bf.Y || u.TargetX != bf.X || u.TargetY != bf.Y {
						t.Fatal(c.Name, "typed回收不同步")
					}
				}
			}
		}
	}
	if len(m.Cases) != 281 || reports != 48 {
		t.Fatal("案例／報告數", len(m.Cases), reports)
	}
	t.Log("281原版案×兩種typed形態，完整兩池、48報告及seed通過")
}

func TestScoutFuelWindowOriginal(t *testing.T) {
	m := scoutFuelSource(t)
	base := newTestGame(t)
	for _, fleet := range []int{12, 16} {
		t.Run(fmt.Sprint(fleet), func(t *testing.T) {
			var c scoutFuelCase
			for _, r := range m.Cases {
				if r.Fleet == fleet && r.Mode == 0 && r.Fuel == 3 && r.Humans == 3 {
					c = r
					break
				}
			}
			g := scoutFuelFixture(t, base, c, true)
			g.battleScoutFuelPhase(0, func() {})
			fb := gfx.NewFramebuffer(0)
			g.drawTurnReport(fb)
			g.res.Cursor.Draw(fb, 240, 200)
			prefix := fmt.Sprintf("workplace/scout79-r83-fleet%d-original", fleet)
			want := m.read(t, prefix+".screen.bin")
			rgb := m.read(t, prefix+".palette.bin")
			pal := g.Palette()
			for i := 0; i < 16; i++ {
				r, g, b := pal.RGB8(i)
				if !bytes.Equal([]byte{r, g, b}, rgb[i*3:i*3+3]) {
					t.Fatal("有效16色色盤不同", i)
				}
			}
			r := g.res.TurnWindow.BattleReport
			x, y, w, h := r.CellX*ui.FrameTileW, ui.FrameCellOriginY+r.CellY*ui.FrameTileH, (r.Cols+2)*ui.FrameTileW, (r.Rows+2)*ui.FrameTileH
			stats := func(rect [4]int) map[string]any {
				bad, maxdiff := 0, 0
				box := [4]int{640, 480, -1, -1}
				var bbox any
				for yy := rect[1]; yy < rect[1]+rect[3]; yy++ {
					for xx := rect[0]; xx < rect[0]+rect[2]; xx++ {
						i := yy*640 + xx
						if fb.Pix[i] == want[i] {
							continue
						}
						bad++
						box = [4]int{min(box[0], xx), min(box[1], yy), max(box[2], xx), max(box[3], yy)}
						for ch := 0; ch < 3; ch++ {
							d := int(rgb[int(fb.Pix[i])*3+ch]) - int(rgb[int(want[i])*3+ch])
							if d < 0 {
								d = -d
							}
							maxdiff = max(maxdiff, d)
						}
					}
				}
				if bad > 0 {
					bbox = box
				}
				return map[string]any{"rect": rect, "pixels": rect[2] * rect[3], "same": rect[2]*rect[3] - bad, "different": bad, "maximum_channel_difference": maxdiff, "bbox": bbox}
			}
			window, full := stats([4]int{x, y, w, h}), stats([4]int{0, 0, 640, 480})
			if out := os.Getenv("LOGH3_SCOUT_OUTPUT"); out != "" {
				if err := os.MkdirAll(out, 0755); err != nil {
					t.Fatal(err)
				}
				diff := gfx.NewFramebuffer(0)
				for i := range want {
					if want[i] != fb.Pix[i] {
						diff.Pix[i] = 15
					}
				}
				for label, pic := range map[string]*gfx.Framebuffer{"remake": fb, "diff": diff} {
					if err := pic.WritePNG(filepath.Join(out, fmt.Sprintf("fleet%d-%s.png", fleet, label)), pal); err != nil {
						t.Fatal(err)
					}
				}
				b, _ := json.MarshalIndent(map[string]any{"fleet": fleet, "window": window, "full": full, "mask": nil, "comparison_level": "same-state controlled report window only", "background": "blank remake diagnostic; full-frame differences retained"}, "", "  ")
				if err := os.WriteFile(filepath.Join(out, fmt.Sprintf("fleet%d-comparison.json", fleet)), append(b, '\n'), 0644); err != nil {
					t.Fatal(err)
				}
			}
			if window["different"].(int) != 0 {
				t.Fatal("完整報告矩形不符", window)
			}
			t.Logf("報告%d像素零差，完整診斷底圖差%v像素", w*h, full["different"])
		})
	}
}

// 靜態原caller順序另已由IDA核；驗產品兩支報告不提前回收或跳過。
func TestScoutFuelSequentialReports(t *testing.T) {
	m := scoutFuelSource(t)
	base := newTestGame(t)
	var cases []scoutFuelCase
	for _, f := range []int{16, 12} {
		for _, c := range m.Cases {
			if c.Fleet == f && c.Mode == 0 && c.Fuel == 3 && c.Humans == 3 {
				cases = append(cases, c)
				break
			}
		}
	}
	g := scoutFuelFixture(t, base, cases[0], true)
	h := scoutFuelFixture(t, base, cases[1], true)
	copy(g.slot.Fleet[12*122:], ownershipHex(t, cases[1].BeforeFleet))
	copy(g.slot.ShipAt(cases[1].Pool), ownershipHex(t, cases[1].BeforeShip))
	g.battleFleets = append(g.battleFleets, h.battleFleets...)
	g.battleUnits = append(g.battleUnits, h.battleUnits...)
	done := 0
	g.battleScoutFuelPhase(0, func() { done++ })
	if done != 0 || g.turnRep.pending[0].face != 40 || g.battleFleets[1].ScoutState != 3 {
		t.Fatal("第二艦隊提前加燃料")
	}
	g.Click(240, 240)
	if done != 0 || g.turnRep.pending[0].face != 12 || g.battleFleets[0].ScoutMode != 255 || g.battleFleets[1].ScoutMode == 255 {
		t.Fatal("一次Click跳過第二報告或回收順序錯")
	}
	g.Click(240, 240)
	if done != 1 || g.battleFleets[1].ScoutMode != 255 {
		t.Fatal("最後報告未續接")
	}
}

// 正常主選單／S3首戰與公開索敵實行一次；後續只跑第8步作回收保存診斷。
// 本視野的合法近距離目的地會先正常歸隊，不能宣稱正常四回合燃料到期。
func TestScoutFuelPlayerDeployTurnSave(t *testing.T) {
	g := commandTestGame(t)
	if g == nil {
		return
	}
	var err error
	g.res.FleetWindow, err = screen.LoadFleetWindowSpec(filepath.Join(root, "docs/internals/data/fleet-window.json"))
	if err != nil {
		t.Fatal(err)
	}
	flag := findUnit(g, g.faction, true)
	if flag < 0 {
		t.Fatal("首戰看不到己方旗艦")
	}
	fleet := g.battleUnits[flag].Fleet
	bf := g.battleFleet(fleet)
	g.battleRand = rules.NewRand(0x12345678)
	// 選目前視野內最遠的直線可通行目的地。避免近處到達後先正常歸隊，
	// 把另一個分支誤當燃料回收。只送公開點擊，不改任何局面資料。
	tx, ty, best := -1, -1, -1
	for yy := g.res.Battle.Field.Y + 8; yy < g.res.Battle.Field.Y+g.res.Battle.Field.Rows*g.res.Battle.Field.TileH; yy += 16 {
		for xx := g.res.Battle.Field.X + 8; xx < g.res.Battle.Field.X+g.res.Battle.Field.Cols*g.res.Battle.Field.TileW; xx += 16 {
			wx, wy, ok := g.battleWorldAt(xx, yy)
			if !ok {
				continue
			}
			clear := true
			dx, dy := wx-bf.X, wy-bf.Y
			for k := 1; k <= 128; k++ {
				if !g.res.Terrain.Passable(g.battleCells(g.battleRegion), bf.X+dx*k/128, bf.Y+dy*k/128) {
					clear = false
					break
				}
			}
			if clear && dx*dx+dy*dy > best {
				tx, ty, best = xx, yy, dx*dx+dy*dy
			}
		}
	}
	if tx < 0 {
		t.Fatal("首戰缺可通行索敵目的地")
	}
	t.Logf("固定正常目的地screen=(%d,%d)，距離平方%d，seed12345678", tx, ty, best)
	deploy := func() {
		t.Helper()
		x, y := unitScreenPos(g, flag)
		g.MoveMouse(x+8, y+8)
		g.Click(x+8, y+8)
		if !g.CommandMenuOpen() {
			t.Fatal("正常旗艦選單未開")
		}
		m := g.res.CommandMenus
		g.Click(g.battleCmdX+m.Layout.HighlightDX+8, g.battleCmdY+m.Layout.HighlightDY+2*m.Layout.RowPitch+8)
		if !g.MoveTargetPending() {
			t.Fatal("正常索敵不能派出", *bf)
		}
		g.Click(tx, ty)
		if g.MoveTargetPending() {
			t.Fatal("索敵目的地沒有提交")
		}
	}
	advance := func() {
		t.Helper()
		turn := g.battleTurn
		x, y := -1, -1
		for yy := 280; yy < 420 && x < 0; yy += 16 {
			for xx := 160; xx < 420; xx += 16 {
				if g.battleUnitAt(xx, yy) < 0 && g.battlePlanetAt(xx, yy) < 0 {
					x, y = xx, yy
					break
				}
			}
		}
		if x < 0 {
			t.Fatal("找不到正常戰場選單落點")
		}
		g.MoveMouse(x, y)
		g.Click(x, y)
		if !g.battleMenu {
			t.Fatal("正常戰場選單未開")
		}
		l := g.res.BattleMenu.Layout
		g.Click(g.battleMenuX+l.Highlight.X-l.Window.X+8, g.battleMenuY+l.Highlight.Y-l.Window.Y+8)
		for i := 0; i < 20000 && g.battleTurn == turn; i++ {
			g.Tick(time.Second / 60)
			g.Frame()
		}
		if g.state != StateBattle || g.battleTurn != turn+1 || g.battlePhase != screen.BattlePhaseInput {
			t.Fatal("正常實行未返回戰場", g.state, g.battleTurn, turn)
		}
	}
	before := len(g.battleUnits)
	deploy()
	if len(g.battleUnits) != before+1 {
		t.Fatal("初次索敵池未加入")
	}
	for fuel := 1; fuel <= 4; fuel++ {
		if fuel == 1 {
			advance()
		} else {
			g.finishBattleTurn()
			for g.state == StateTurnReport {
				g.Frame()
				g.Click(240, 240)
			}
		}
		if fuel < 4 {
			if bf.ScoutState != fuel || bf.ScoutMode == 255 {
				t.Fatal("正常燃料遞增不符", fuel, *bf)
			}
		} else if bf.ScoutState != 3 || bf.ScoutMode != 255 {
			t.Fatal("第四回合未回收冷卻", *bf)
		}
	}
	f := g.fleetByIndex(fleet)
	u := g.battleUnits[len(g.battleUnits)-1]
	if !u.Out || u.Action != 255 || u.X != bf.X || u.Y != bf.Y {
		t.Fatal("正常回收typed位置不符", u, *bf)
	}
	dir := t.TempDir()
	if out := os.Getenv("LOGH3_SCOUT_OUTPUT"); out != "" {
		dir = filepath.Join(out, "normal-save")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.SaveSlot(dir, 0); err != nil {
		t.Fatal(err)
	}
	read := func(name string) []byte {
		b, e := os.ReadFile(filepath.Join(dir, name+".DT0"))
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	header, e := os.ReadFile(filepath.Join(root, "org_game/extracted/PLADAT.DAT"))
	if e != nil {
		t.Fatal(e)
	}
	saved, e := rules.LoadSlot(rules.SlotFiles{Stellar: read("STELLAR"), Planet: read("PLANET"), Admiral: read("ADMIRAL"), Fleet: read("FLEET"), Ship: read("SHIP"), Work: read("WORK"), PlanetHeader: header})
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(saved.Fleet, g.slot.Fleet) || !bytes.Equal(saved.Ship, g.slot.Ship) || saved.Fleet[fleet*122+0x3A] != 255 || saved.Fleet[fleet*122+0x3B] != 3 || saved.ShipAt(f.Scout)[0x1E] != 255 {
		t.Fatal("完整原生FLEET／SHIP沒有保存回收結果")
	}
	other := newTestGame(t)
	other.LoadSlot(saved)
	if other.state != StateCommandPhase || !bytes.Equal(other.slot.Fleet, saved.Fleet) || !bytes.Equal(other.slot.Ship, saved.Ship) {
		t.Fatal("原生保存重讀不符")
	}
	for remaining := 2; remaining >= 0; remaining-- {
		g.finishBattleTurn()
		for g.state == StateTurnReport {
			g.Frame()
			g.Click(240, 240)
		}
		if bf.ScoutState != remaining || bf.ScoutOut != (remaining != 0) {
			t.Fatal("正常冷卻不符", remaining, *bf)
		}
	}
	deploy()
	if len(g.battleUnits) != before+1 || bf.ScoutState != 0 || bf.ScoutMode != 0 {
		t.Fatal("再次派出沒有重用池", len(g.battleUnits), *bf)
	}
	t.Logf("正常旗艦%d：公開索敵／一次實行；後續第8步診斷回收／冷卻／再派出，六檔保存與完整FLEET／SHIP重讀通過；不宣稱完整正常到期或讀檔續戰", fleet)
}
