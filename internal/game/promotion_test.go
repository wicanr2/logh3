package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

type promotionGameRow struct {
	Name          string
	Case          []int
	Before, After map[string]string
	SeedAfter     string `json:"seed_after_hex"`
	Calls         []struct {
		Location string
		Args     []int
		Text     string `json:"text_hex"`
	}
	Frames []struct {
		Screen  []byte
		Palette []byte `json:"palette_rgb"`
		PNG     []byte `json:"png"`
	}
}

func promotionGameRows(t *testing.T, path, hash string) []promotionGameRow {
	t.Helper()
	b, e := os.ReadFile(filepath.Join(root, path))
	if os.IsNotExist(e) {
		t.Skip("缺原版升級收據；沿tools/battle_promotion_global_oracle.sh重生")
	}
	if e != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != hash {
		t.Fatal("原版升級收據雜湊不符", e)
	}
	var raw struct{ Rows []promotionGameRow }
	if e = json.Unmarshal(b, &raw); e != nil {
		t.Fatal(e)
	}
	return raw.Rows
}

func promotionGame(t *testing.T, base *Resources, row promotionGameRow) *Game {
	t.Helper()
	data := func(key string) []byte { return ownershipHex(t, row.Before[key]) }
	p := data("planet")
	s, e := rules.LoadSlot(rules.SlotFiles{Admiral: data("admiral"), Fleet: data("fleet"), Ship: data("ship"), Stellar: data("stellar"), Work: data("work"), PlanetHeader: p[:64], Planet: p[64:]})
	if e != nil {
		t.Fatal(e)
	}
	r := *base
	r.Stellar, r.Planets, r.Admiral = s.Stellar, s.Planets, s.Admiral
	g := &Game{res: &r, slot: s, state: StateBattle, battleRegion: 26, turn: s.Turn, military: s.Military, taxTotal: s.TaxTotal, blend: s.Blend, rate: s.Rate,
		shipPct: s.ShipPct, building: s.Building, battleRand: rules.NewRand(0x12345678), display: DisplaySettings{Anim: s.AnimEffect, AnimSpeed: s.AnimSpeed, MsgSpeed: s.MsgSpeed}}
	g.fleets = s.Fleets(r.Weights)
	g.battleFleets = s.BattleFleetsFor([]int{row.Case[0]})
	if f := g.fleetByIndex(row.Case[0]); f != nil {
		g.battleUnits = s.BattleUnitsFor(r.Weights, *f)
	}
	switch row.Case[3] {
	case 0:
		g.faction = -1
	case 1:
		g.faction = 0
	case 2:
		g.faction = 1
	case 3:
		g.faction = 0
		g.versus = true
	}
	g.paidUpkeep, g.paidMove, g.campaignScore = s.PaidUpkeep, s.PaidMove, s.StrategicScore
	g.music = MusicSettings{Device: s.MusicDevice, BGM: s.MusicBGM, Battle: s.MusicBattle}
	if g.faction >= 0 {
		g.pct, g.foePct = s.BuildPct[g.faction], s.BuildPct[1-g.faction]
		g.treasury = rules.Treasury{Budget: s.BudgetStart[g.faction], ShipPaid: s.ShipPaid[g.faction]}
	}
	g.cacheBattlePromotionPaths()
	return g
}

func promotionGameTables(t *testing.T, g *Game, want map[string]string) {
	t.Helper()
	for name, b := range map[string][]byte{"admiral": g.slot.Admiral, "fleet": g.slot.Fleet, "ship": g.slot.Ship, "stellar": g.res.Stellar, "work": g.workRecord()} {
		w := ownershipHex(t, want[name])
		if !bytes.Equal(b, w) {
			for i, v := range b {
				if v != w[i] {
					t.Fatalf("%s+%X：%02X／原版%02X", name, i, v, w[i])
				}
			}
		}
	}
}

func TestBattlePromotionGameOriginal(t *testing.T) {
	rows := promotionGameRows(t, "workplace/promotion79-r86-global-a.json", "2bc8fa7bf8dc4081fad414d23e7bfeabb76cd8df510d53f3442136283cb1a435")
	base := loadWithRules(t)
	reports := 0
	for _, row := range rows {
		t.Run(row.Name, func(t *testing.T) {
			g := promotionGame(t, base, row)
			done := 0
			g.battlePromotionPhase(0, false, func() { done++ })
			frame := 0
			face := -1
			for _, call := range row.Calls {
				if call.Location == "185C:06E8" {
					face = call.Args[0] & 255
				}
				if call.Location != "3826:0405" {
					continue
				}
				if done != 0 || g.state != StateTurnReport || len(g.turnRep.pending) != 1 {
					t.Fatal("報告等待與逐人重試不符", g.saveErr)
				}
				got := g.turnRep.pending[0]
				if got.face != face || !bytes.Equal(got.text, ownershipHex(t, call.Text)) {
					t.Fatalf("文字／頭像不符：%d/%d，%x／%s", got.face, face, got.text, call.Text)
				}
				beforeSeed := g.battleRand.Seed()
				g.turnReportTick(10)
				if g.state != StateTurnReport || beforeSeed != g.battleRand.Seed() {
					t.Fatal("零速度未等待實際按下")
				}
				if frame >= len(row.Frames) {
					t.Fatal("原版報告畫面缺失")
				}
				promotionWindowCompare(t, g, row.Name, frame, row.Frames[frame].Screen, row.Frames[frame].Palette, row.Frames[frame].PNG)
				if reports%2 == 0 {
					g.Click(240, 240)
					g.LeftRelease()
				} else {
					g.RightClick(240, 240)
				}
				frame++
				reports++
			}
			if done != 1 || g.state != StateBattle || g.battleRand.Seed() != binary.LittleEndian.Uint32(ownershipHex(t, row.SeedAfter)) {
				t.Fatal("末端續接或seed不同", g.saveErr)
			}
			promotionGameTables(t, g, row.After)
			for _, f := range g.fleets {
				r := g.slot.Fleet[f.Index*122 : (f.Index+1)*122]
				if int(r[0]) != f.Flags || int(r[14]) != f.Expected || int(r[0x3c]) != f.Morale || int(r[0x47]) != f.Operations {
					t.Fatal("保存欄位未同步")
				}
			}
		})
	}
	if reports != 11 {
		t.Fatal("報告數不同", reports)
	}
	t.Log("66案正式Game、11則文字／頭像／零速度等待及完整報告矩形符合原版")
}

func promotionWindowCompare(t *testing.T, g *Game, name string, index int, want, palette, originalPNG []byte) {
	t.Helper()
	if len(palette) != 48 {
		t.Fatal("原版16色色盤不完整")
	}
	for i := 0; i < 16; i++ {
		r, gc, b := g.Palette().RGB8(i)
		if !bytes.Equal([]byte{r, gc, b}, palette[i*3:i*3+3]) {
			t.Fatal("報告有效色盤不同", i)
		}
	}
	fb := gfx.NewFramebuffer(0)
	g.drawTurnReport(fb)
	g.res.Cursor.Draw(fb, 240, 200)
	r := g.res.TurnWindow.BattleReport
	x, y := r.CellX*ui.FrameTileW, ui.FrameCellOriginY+r.CellY*ui.FrameTileH
	w, h := (r.Cols+2)*ui.FrameTileW, (r.Rows+2)*ui.FrameTileH
	stats := func(x, y, w, h int) map[string]any {
		bad, maxDiff := 0, 0
		box := [4]int{640, 480, -1, -1}
		for yy := y; yy < y+h; yy++ {
			for xx := x; xx < x+w; xx++ {
				i := yy*640 + xx
				if fb.Pix[i] == want[i] {
					continue
				}
				bad++
				box = [4]int{min(box[0], xx), min(box[1], yy), max(box[2], xx), max(box[3], yy)}
				for c := 0; c < 3; c++ {
					d := int(palette[int(fb.Pix[i])*3+c]) - int(palette[int(want[i])*3+c])
					if d < 0 {
						d = -d
					}
					maxDiff = max(maxDiff, d)
				}
			}
		}
		var bbox any
		if bad > 0 {
			bbox = box
		}
		return map[string]any{"rect": [4]int{x, y, w, h}, "pixels": w * h, "same": w*h - bad, "different": bad, "maximum_channel_difference": maxDiff, "bbox": bbox}
	}
	window, full := stats(x, y, w, h), stats(0, 0, 640, 480)
	if out := os.Getenv("LOGH3_PROMOTION_OUTPUT"); out != "" {
		if e := os.MkdirAll(out, 0755); e != nil {
			t.Fatal(e)
		}
		stem := filepath.Join(out, fmt.Sprintf("%s-frame%d", name, index))
		diff := gfx.NewFramebuffer(0)
		for i := range want {
			if fb.Pix[i] != want[i] {
				diff.Pix[i] = 15
			}
		}
		for suffix, pic := range map[string]*gfx.Framebuffer{"remake": fb, "diff": diff} {
			if e := pic.WritePNG(stem+"-"+suffix+".png", g.Palette()); e != nil {
				t.Fatal(e)
			}
		}
		b, _ := json.MarshalIndent(map[string]any{"scope": "same-state controlled report window only", "window": window, "full": full, "mask": nil, "background": "blank diagnostic; full-frame differences retained"}, "", "  ")
		if e := os.WriteFile(stem+"-comparison.json", append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(stem+"-original.png", originalPNG, 0644); e != nil {
			t.Fatal(e)
		}
	}
	if window["different"].(int) != 0 {
		t.Fatalf("報告矩形差異%v", window)
	}
}

func TestBattlePromotionWaitAndRanks(t *testing.T) {
	base := loadWithRules(t)
	rows := promotionGameRows(t, "workplace/promotion79-r86-global-a.json", "2bc8fa7bf8dc4081fad414d23e7bfeabb76cd8df510d53f3442136283cb1a435")
	g := promotionGame(t, base, rows[50])
	g.display.MsgSpeed = 32770
	original := g.display
	g.battlePromotionPhase(0, false, func() {})
	if g.state != StateTurnReport || g.turnRep.pending[0].wait != 4 || g.display != original {
		t.Fatal("word乘二回繞或永久設定不符", g.saveErr)
	}
	g.turnReportTick(3)
	if g.state != StateTurnReport {
		t.Fatal("提早返回")
	}
	g.turnReportTick(1)
	if g.state != StateBattle {
		t.Fatal("計時未返回")
	}
	for _, f := range []int{0, 1} {
		for _, level := range []int{28, 29, 33, 34, 38, 39, 99, 255} {
			r := make([]byte, 28)
			r[1], r[6] = byte(f), byte(level)
			if !bytes.Equal(codesBig5(g.admiralRankCodes(r)), codesBig5(base.Promotion.RankCodes(f, rules.RankOf(level)))) {
				t.Fatal("軍階未取記錄陣營")
			}
		}
	}
	if hex.EncodeToString(codesBig5(base.Promotion.RankCodes(0, 10))) != "a440afc5a457b14e" || hex.EncodeToString(codesBig5(base.Promotion.RankCodes(1, 10))) != "a457b14e" {
		t.Fatal("兩側一級上將名稱不符原版")
	}
}

func TestBattlePromotionSaveReload(t *testing.T) {
	base := loadWithRules(t)
	rows := promotionGameRows(t, "workplace/promotion79-r86-global-a.json", "2bc8fa7bf8dc4081fad414d23e7bfeabb76cd8df510d53f3442136283cb1a435")
	for _, row := range rows {
		if row.Case[3] == 0 {
			continue
		}
		t.Run(row.Name, func(t *testing.T) {
			g := promotionGame(t, base, row)
			done := 0
			g.battlePromotionPhase(0, false, func() { done++ })
			for g.state == StateTurnReport {
				g.Click(240, 240)
				g.LeftRelease()
			}
			if done != 1 {
				t.Fatal(g.saveErr)
			}
			before := g.battleRand.Seed()
			dir := t.TempDir()
			if out := os.Getenv("LOGH3_PROMOTION_OUTPUT"); out != "" {
				dir = filepath.Join(out, row.Name+"-save")
				if e := os.MkdirAll(dir, 0755); e != nil {
					t.Fatal(e)
				}
			}
			g.state = StateSaveGame // 明示保存診斷，不宣稱經正常戰場選單
			if e := g.SaveSlot(dir, 0); e != nil {
				t.Fatal(e)
			}
			read := func(name string) []byte {
				b, e := os.ReadFile(filepath.Join(dir, name+".DT0"))
				if e != nil {
					t.Fatal(e)
				}
				return b
			}
			for name, key := range map[string]string{"ADMIRAL": "admiral", "FLEET": "fleet", "SHIP": "ship", "STELLAR": "stellar", "WORK": "work"} {
				b := read(name)
				if !bytes.Equal(b, ownershipHex(t, row.After[key])) {
					for i, v := range b {
						w := ownershipHex(t, row.After[key])
						if v != w[i] {
							t.Fatalf("保存%s+%X不同：%02X／原版%02X", name, i, v, w[i])
						}
					}
				}
			}
			p := ownershipHex(t, row.After["planet"])
			if !bytes.Equal(read("PLANET"), p[64:]) {
				t.Fatal("行星保存不同")
			}
			s, e := rules.LoadSlot(rules.SlotFiles{Admiral: read("ADMIRAL"), Fleet: read("FLEET"), Ship: read("SHIP"), Stellar: read("STELLAR"), Work: read("WORK"), Planet: read("PLANET"), PlanetHeader: p[:64]})
			if e != nil {
				t.Fatal(e)
			}
			if s.ShipBudget != g.slot.ShipBudget || s.TaxTotal != g.taxTotal || !bytes.Equal(s.Admiral, g.res.Admiral) || g.battleRand.Seed() != before {
				t.Fatal("回讀不同或保存消耗RND")
			}
			g.LoadSlot(s)
			if g.battlePromotionPaths != nil || !bytes.Equal(g.res.Admiral, s.Admiral) {
				t.Fatal("回讀未清快取或提督表分家")
			}
		})
	}
}

func TestBattlePromotionCachedPaths(t *testing.T) {
	rows := promotionGameRows(t, "workplace/promotion79-r87-cache-a.json", "c724ea9bf26aaf922eda7bc322627cd00383139894fb343b379a2ff7d136f861")
	base := loadWithRules(t)
	for _, row := range rows {
		t.Run(row.Name, func(t *testing.T) {
			g := promotionGame(t, base, row)
			// pattern8在原版8-4入口注入13歸屬1。路徑仍沿注入前歸屬；pattern9為同值控制。
			if row.Case[5] == 8 {
				g.res.Stellar[13*126+0x12] = 0
				g.cacheBattlePromotionPaths()
				g.res.Stellar[13*126+0x12] = 1
			}
			cache := g.battlePromotionPaths
			done := 0
			g.battlePromotionPhase(0, false, func() { done++ })
			if done != 1 || cache != g.battlePromotionPaths {
				t.Fatal("提早重建路徑", g.saveErr)
			}
			promotionGameTables(t, g, row.After)
		})
	}
}

// 戰略處理已扣軍事費，WORK底稿要到保存才同步；升級不能沿舊底稿算預算。
func TestBattlePromotionCurrentMilitary(t *testing.T) {
	base := loadWithRules(t)
	rows := promotionGameRows(t, "workplace/promotion79-r86-global-a.json", "2bc8fa7bf8dc4081fad414d23e7bfeabb76cd8df510d53f3442136283cb1a435")
	g := promotionGame(t, base, rows[1])
	stale := bytes.Clone(g.workRecord()[0x3e:0x46])
	g.military = [2]int{1234567, 7654321}
	flags, e := g.promoteBattleAdmiral(rows[1].Case[0], 12)
	if e != nil || flags == 0 {
		t.Fatal("未完成測試升級", e)
	}
	for f := 0; f < 2; f++ {
		budget := rules.Budget{Treasury: g.military[f], Tax: g.slot.TaxTotal[f], ShipPct: g.slot.ShipPct[f], Upkeep: g.slot.Upkeep[f], MoveCost: g.slot.MoveCost[f]}
		if g.slot.ShipBudget[f] != budget.ShipCost() {
			t.Fatal("造艦預估未用當下經費", f, g.slot.ShipBudget[f], budget.ShipCost())
		}
	}
	if !bytes.Equal(stale, g.workRecord()[0x3e:0x46]) {
		t.Fatal("8-4提前覆寫軍事費底稿")
	}
}
