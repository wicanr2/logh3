package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

type ownershipReportReceipt struct {
	Seat                 int
	Text                 string `json:"text_hex"`
	PendingBeforeDisplay int    `json:"pending_before_display"`
	Record               string `json:"region_record_hex"`
	Face                 int
}

type ownershipCaseReceipt struct {
	Name         string
	Region       int
	HumanMask    int `json:"human_mask"`
	Occurred     int
	BeforeRecord string `json:"before_record_hex"`
	AfterRecord  string `json:"after_record_hex"`
	PendingAfter int    `json:"pending_after"`
	Owners       []int  `json:"planet_owners"`
	Reports      []ownershipReportReceipt
}

type ownershipReceipt struct {
	Cases     []ownershipCaseReceipt
	Reporters [2]int
	Sources   map[string]struct {
		SHA256 string
		Size   int
	}
}

func ownershipManifest(t *testing.T) ownershipReceipt {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/battle-ownership-v2.json"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != "306f7612cb536cdb441bba2e245baecb9786fbc84a9210b066b06a09934ac262" {
		t.Fatal("第82輪原版凍結來源不符", err)
	}
	var m ownershipReceipt
	if err = json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func (m ownershipReceipt) read(t *testing.T, name string) []byte {
	t.Helper()
	s, ok := m.Sources[name]
	if !ok {
		t.Fatal("未登錄的第82輪來源", name)
	}
	b, err := os.ReadFile(filepath.Join(root, name))
	if os.IsNotExist(err) {
		t.Skip("缺原版收據", name, "；重生入口tools/battle_ownership_oracle.sh")
	}
	if err != nil || len(b) != s.Size || fmt.Sprintf("%x", sha256.Sum256(b)) != s.SHA256 {
		t.Fatal("來源大小或雜湊不符", name, err)
	}
	return b
}

func ownershipHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// 原版控制常式診斷：還原完整星域及typed行星，不冒充正常玩家流程。
func ownershipFixture(t *testing.T, base *Game, c ownershipCaseReceipt) *Game {
	t.Helper()
	res := *base.res
	res.Stellar = make([]byte, rules.RegionCount*rules.RegionRecord)
	g := &Game{res: &res, state: StateBattle, battleRegion: c.Region,
		faction: -1, work: make([]byte, rules.WorkRecord), battleRand: rules.NewRand(0x12345678)}
	copy(g.regionRecord(c.Region), ownershipHex(t, c.BeforeRecord))
	if c.Occurred != 0 {
		rules.SetPhezzanOccurred(g.work)
	}
	switch c.HumanMask {
	case 1:
		g.faction = 0
	case 2:
		g.faction = 1
	case 3:
		g.faction, g.versus = 0, true
	}
	for _, owner := range c.Owners {
		g.battlePlanets = append(g.battlePlanets, rules.BattlePlanet{Owner: owner})
	}
	return g
}

func TestBattleOwnershipOriginalCases(t *testing.T) {
	m := ownershipManifest(t)
	base := newTestGame(t)
	base.res.Reporter = m.Reporters
	reports := 0
	for _, c := range m.Cases {
		g := ownershipFixture(t, base, c)
		work := bytes.Clone(g.work)
		done := 0
		g.battleOwnershipPhase(func() { done++ })
		for i, want := range c.Reports {
			if done != 0 || g.state != StateTurnReport || !g.turnRep.battle || len(g.turnRep.pending) != 1 {
				t.Fatalf("%s報告%d等待或續接不符", c.Name, i)
			}
			got := g.turnRep.pending[0]
			if !bytes.Equal(got.text, ownershipHex(t, want.Text)) || got.face != want.Face ||
				!bytes.Equal(g.regionRecord(c.Region), ownershipHex(t, want.Record)) ||
				g.phezzanPending != (want.PendingBeforeDisplay != 0) {
				t.Fatalf("%s座位%d文字／頭像／完整星域／費沙時點不符", c.Name, want.Seat)
			}
			// 正式一次Click只關閉目前這則，不跳過下一座位。
			g.Click(240, 240)
			reports++
		}
		if done != 1 || g.state != StateBattle ||
			!bytes.Equal(g.regionRecord(c.Region), ownershipHex(t, c.AfterRecord)) ||
			g.phezzanPending != (c.PendingAfter != 0) || g.battleRand.Seed() != 0x12345678 ||
			!bytes.Equal(work, g.work) {
			t.Fatalf("%s完成記錄／續接／費沙／seed／WORK不符", c.Name)
		}
	}
	if len(m.Cases) != 156 || reports != 104 {
		t.Fatalf("原版案例%d／報告%d，預期156／104", len(m.Cases), reports)
	}
	t.Log("156案完整126-byte星域、104則文字與等待、費沙時點、WORK及seed符合原版")
}

// 只比較原版控制報告矩形。底圖由空白重畫，不複製原版像素。
func TestBattleOwnershipWindowOriginal(t *testing.T) {
	m := ownershipManifest(t)
	base := newTestGame(t)
	base.res.Reporter = m.Reporters
	for _, name := range []string{"capture0", "capture1", "partial0", "partial1", "capture0-second", "capture1-second"} {
		t.Run(name, func(t *testing.T) {
			original := name
			if name == "capture0" || name == "capture1" || name == "partial0" || name == "partial1" {
				original += "-complete"
			}
			prefix := "workplace/oracle/ownership79-r82-" + original + "-a/"
			want := m.read(t, prefix+"screen.bin")
			ram := m.read(t, "workplace/oracle/ownership79-r82-entry-a/mem-in.bin")
			owners := []int{0, 1, 1}
			old := 2
			if name == "partial0" {
				old = 0
			} else if name == "partial1" {
				old = 1
			} else {
				o := 0
				if name == "capture1" || name == "capture1-second" {
					o = 1
				}
				owners = []int{o, o, o}
			}
			// 原版entry仍在8-2前：重建該局面，再走產品規則及報告續接。
			rec := bytes.Clone(ram[0x5f2cc+26*126 : 0x5f2cc+27*126])
			rec[regionOwnerOff] = byte(old)
			g := ownershipFixture(t, base, ownershipCaseReceipt{Region: 26, HumanMask: 3,
				BeforeRecord: hex.EncodeToString(rec), Owners: owners})
			g.battleOwnershipPhase(func() {})
			if name == "capture0-second" || name == "capture1-second" {
				g.Click(240, 240)
			}
			viewer := 0
			if name == "capture0-second" || name == "capture1-second" {
				viewer = 1
			}
			if g.state != StateTurnReport || g.turnRep.pending[0].face != m.Reporters[viewer] {
				t.Fatal("沒有正式占領報告或情報參謀不符")
			}
			fb := gfx.NewFramebuffer(0)
			g.drawTurnReport(fb)
			// 原版完整滑鼠位置保留，視窗內也不遮罩游標。
			g.res.Cursor.Draw(fb, 240, 200)
			r := g.res.TurnWindow.BattleReport
			x, y := r.CellX*ui.FrameTileW, ui.FrameCellOriginY+r.CellY*ui.FrameTileH
			w, h := (r.Cols+2)*ui.FrameTileW, (r.Rows+2)*ui.FrameTileH
			pal := g.Palette()
			originalPNG, err := png.Decode(bytes.NewReader(m.read(t, prefix+"screen.png")))
			if err != nil {
				t.Fatal(err)
			}
			cp, ok := originalPNG.ColorModel().(color.Palette)
			if !ok || len(cp) < gfx.Colors {
				t.Fatal("原版有效色盤缺失")
			}
			for i := 0; i < gfx.Colors; i++ {
				ra, ga, ba, _ := cp[i].RGBA()
				rb, gb, bb := pal.RGB8(i)
				if uint8(ra>>8) != rb || uint8(ga>>8) != gb || uint8(ba>>8) != bb {
					t.Fatal("原版與正式renderer有效色盤不同", i)
				}
			}
			stats := func(rect [4]int) map[string]any {
				bad, maximum := 0, 0
				var bbox any
				box := [4]int{gfx.Width, gfx.Height, -1, -1}
				for yy := rect[1]; yy < rect[1]+rect[3]; yy++ {
					for xx := rect[0]; xx < rect[0]+rect[2]; xx++ {
						i := yy*gfx.Width + xx
						if fb.Pix[i] == want[i] {
							continue
						}
						bad++
						box = [4]int{min(box[0], xx), min(box[1], yy), max(box[2], xx), max(box[3], yy)}
						ra, ga, ba := pal.RGB8(int(fb.Pix[i]))
						rb, gb, bb := pal.RGB8(int(want[i]))
						for _, delta := range []int{int(ra) - int(rb), int(ga) - int(gb), int(ba) - int(bb)} {
							if delta < 0 {
								delta = -delta
							}
							maximum = max(maximum, delta)
						}
					}
				}
				if bad != 0 {
					bbox = box
				}
				return map[string]any{"rect": rect, "pixels": rect[2] * rect[3], "same": rect[2]*rect[3] - bad,
					"different": bad, "maximum_channel_difference": maximum, "bbox": bbox}
			}
			window := stats([4]int{x, y, w, h})
			full := stats([4]int{0, 0, gfx.Width, gfx.Height})
			if out := os.Getenv("LOGH3_OWNERSHIP_OUTPUT"); out != "" {
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
					if err := pic.WritePNG(filepath.Join(out, name+"-"+label+".png"), pal); err != nil {
						t.Fatal(err)
					}
				}
				b, _ := json.MarshalIndent(map[string]any{"case": name, "window": window, "full": full,
					"comparison_level": "same-state controlled report window only", "mask": nil,
					"background": "blank remake diagnostic; full-frame differences retained, not parity"}, "", "  ")
				if err := os.WriteFile(filepath.Join(out, name+"-comparison.json"), append(b, '\n'), 0644); err != nil {
					t.Fatal(err)
				}
			}
			if window["different"].(int) != 0 {
				t.Fatalf("%s完整報告矩形差異%v", name, window)
			}
			t.Logf("%s報告%d像素零差；完整底圖差%d像素，未宣稱全圖同態", name, w*h, full["different"])
		})
	}
}

// 正常公開輸入到首戰後，明示注入戰場行星歸屬作8-2保存診斷。
// 從正式8-1續到8-2，不宣稱玩家已完成占領這些行星。
func TestBattleOwnershipTurnSave(t *testing.T) {
	for _, owner := range []int{0, 1} {
		t.Run(fmt.Sprint(owner), func(t *testing.T) {
			g := battleTestGame(t)
			if g == nil {
				return
			}
			region := g.battleRegion
			before := bytes.Clone(g.regionRecord(region))
			if before[regionOwnerOff] != 2 || len(g.battlePlanets) != 3 {
				t.Fatal("首戰診斷父狀態不符")
			}
			for i := range g.battlePlanets {
				g.battlePlanets[i].Owner = owner
			}
			turn := g.battleTurn
			g.battleRand = rules.NewRand(0x12345678)
			g.finishBattleTurn()
			if g.state != StateTurnReport || g.battleTurn != turn || g.regionOwner(region) != 2 {
				t.Fatal("正式8-2沒有等待或提前更新Owner／回合")
			}
			g.Frame()
			g.Click(240, 240)
			if g.state != StateBattle || g.battleTurn != turn+1 || g.battleRand.Seed() != 0x12345678 {
				t.Fatal("報告返回未續到原回合尾巴或多消耗RND")
			}
			want := bytes.Clone(before)
			want[regionOwnerOff] = byte(owner)
			want[0x16], want[0x17], want[0x18] = 0, 0, 0
			want[0x16+owner] = 3
			if !bytes.Equal(g.regionRecord(region), want) {
				t.Fatal("正式續接的完整星域記錄不符")
			}
			dir := t.TempDir()
			if out := os.Getenv("LOGH3_OWNERSHIP_OUTPUT"); out != "" {
				dir = filepath.Join(out, fmt.Sprintf("save-owner%d", owner))
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := g.Frame().WritePNG(filepath.Join(dir, "screen.png"), g.Palette()); err != nil {
					t.Fatal(err)
				}
			}
			g.res.SaveDir = dir
			// 玩家先開戰場選單，再進「存進度」；保存時已是StateSaveGame。
			mx, my := -1, -1
			for y := 240; y < 340 && mx < 0; y += 16 {
				for x := 240; x < 480; x += 16 {
					if g.battleUnitAt(x, y) < 0 && g.battlePlanetAt(x, y) < 0 {
						mx, my = x, y
						break
					}
				}
			}
			if mx < 0 {
				t.Fatal("找不到戰場選單空白點")
			}
			g.Click(mx, my)
			g.Click(mx, my+4*g.res.BattleMenu.Layout.Text.RowPitch)
			if g.state != StateSaveGame {
				t.Fatal("正式戰場選單未到存進度", g.state)
			}
			sy := -1
			for y := 0; y < gfx.Height; y++ {
				if g.res.SlotList.SaveRowAt(y) == 0 {
					sy = y
					break
				}
			}
			if sy < 0 {
				t.Fatal("存檔第一列缺失")
			}
			g.Click(300, sy)
			if err := g.SaveErr(); err != nil {
				t.Fatal(err)
			}
			read := func(name string) []byte {
				b, err := os.ReadFile(filepath.Join(dir, name+".DT"+rules.SlotExt(0)))
				if err != nil {
					t.Fatal(err)
				}
				return b
			}
			header, err := os.ReadFile(filepath.Join(root, "org_game/extracted/PLADAT.DAT"))
			if err != nil {
				t.Fatal(err)
			}
			slot, err := rules.LoadSlot(rules.SlotFiles{Stellar: read("STELLAR"), Planet: read("PLANET"),
				Admiral: read("ADMIRAL"), Fleet: read("FLEET"), Ship: read("SHIP"), Work: read("WORK"), PlanetHeader: header})
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(slot.Stellar, g.res.Stellar) {
				t.Fatal("整張星域表沒有原樣保存")
			}
			for _, p := range g.battlePlanets {
				if int(slot.Planets.Record(p.Index)[0x11]) != owner {
					t.Fatal("保存的星域與行星歸屬分家", p.Index)
				}
			}
			other := newTestGame(t)
			other.LoadSlot(slot)
			// LoadSlot目前返回命令階段；088明列戰場讀檔續戰尚未完成。
			if other.state != StateCommandPhase || other.regionOwner(region) != owner ||
				!bytes.Equal(other.regionRecord(region), want) {
				t.Fatal("原生保存重讀沒有保留完整星域")
			}
			for _, p := range g.battlePlanets {
				if int(other.res.Planets.Record(p.Index)[0x11]) != owner {
					t.Fatal("重讀行星不是當前擁有者")
				}
			}
			t.Logf("Owner%d：正常入口＋明示行星注入，正式8-1／8-2／等待／回合續接及六檔存讀通過", owner)
		})
	}
}
