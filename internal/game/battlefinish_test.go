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

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/rules"
)

type finishGameRow struct {
	promotionGameRow
	Mask, Human, Round, AX int
}

func finishGameRows(t *testing.T, name, hash string) []finishGameRow {
	t.Helper()
	b, e := os.ReadFile(filepath.Join(root, "workplace", name))
	if os.IsNotExist(e) {
		t.Skip("缺固定dosgolem收據", name)
	}
	if e != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != hash {
		t.Fatal("戰場結束收據身分不同", e)
	}
	var d struct{ Rows []finishGameRow }
	if e = json.Unmarshal(b, &d); e != nil {
		t.Fatal(e)
	}
	return d.Rows
}

func finishGame(t *testing.T, base *Resources, row finishGameRow) *Game {
	t.Helper()
	list := ownershipHex(t, row.Before["lists"])
	row.Case = []int{int(list[4]), 0, 0, row.Human}
	g := promotionGame(t, base, row.promotionGameRow)
	var err error
	g.res.Terrain, err = rules.LoadBattleTerrain(filepath.Join(root, "docs/internals/data/battle-terrain.json"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.Formation, err = rules.LoadFormationSlots(filepath.Join(root, "docs/internals/data/battle-formation-slots.json"))
	if err != nil {
		t.Fatal(err)
	}
	g.res.Reporter = [2]int{int(g.workRecord()[0x39]), int(g.workRecord()[0x3a])}
	g.battleRegion = int(g.workRecord()[3])
	input, err := os.ReadFile(filepath.Join(root, fmt.Sprintf("org_game/extracted/MAP%02d.CMP", g.battleRegion)))
	if err != nil {
		t.Fatal(err)
	}
	cells, err := asset.DecodeCMP(input, asset.BattleMapCols, asset.BattleMapRows)
	if err != nil {
		t.Fatal(err)
	}
	g.res.BattleMaps = make([][]byte, rules.RegionCount)
	g.res.BattleMaps[g.battleRegion] = cells
	g.battleTurn = row.Round + 1
	g.battleScore = [2]int{int(binary.LittleEndian.Uint32(g.workRecord()[0xd0:])), int(binary.LittleEndian.Uint32(g.workRecord()[0xd4:]))}
	var ids []int
	for _, b := range list[4 : 4+int(list[0])] {
		ids = append(ids, int(b))
	}
	g.battleFleets = g.slot.BattleFleetsFor(ids)
	byPool := map[int]rules.BattleUnit{}
	for _, id := range ids {
		f := g.fleetByIndex(id)
		for i, u := range g.slot.BattleUnitsFor(g.res.Weights, *f) {
			byPool[f.Composition[i].Pool] = u
		}
		if u, ok := g.slot.ScoutUnit(g.res.Weights, *f, 0, 0); ok {
			r := g.slot.Ship[f.Scout*rules.ShipRecord:]
			u.X, u.Y = int(binary.LittleEndian.Uint16(r[0x34:])), int(binary.LittleEndian.Uint16(r[0x36:]))
			u.TargetX, u.TargetY = int(binary.LittleEndian.Uint16(r[0x28:])), int(binary.LittleEndian.Uint16(r[0x2a:]))
			byPool[f.Scout] = u
		}
	}
	g.battleUnits = nil
	for i := 0; i < int(binary.LittleEndian.Uint16(list[2:])); i++ {
		pool := int(binary.LittleEndian.Uint16(list[38+2*i:]))
		u, ok := byPool[pool]
		if !ok {
			t.Fatal("固定清單缺正式單位", pool)
		}
		g.battleUnits = append(g.battleUnits, u)
	}
	start, n := g.res.Planets.Range(g.battleRegion)
	var sites []rules.PlanetSite
	for i := start; i < start+n; i++ {
		r := g.res.Planets.Record(i)
		sites = append(sites, rules.PlanetSite{X: int(binary.LittleEndian.Uint16(r[0x46:])), Y: int(binary.LittleEndian.Uint16(r[0x48:]))})
	}
	g.battlePlanets = g.res.Planets.BattlePlanetsFor(g.battleRegion, sites)
	return g
}

func finishGameTables(t *testing.T, g *Game, want map[string]string) {
	t.Helper()
	promotionGameTables(t, g, want)
	if !bytes.Equal(g.res.Planets.Data, ownershipHex(t, want["planet"])) {
		t.Fatal("完整行星表不同")
	}
	list := ownershipHex(t, want["lists"])
	if len(g.battleFleets) != int(list[0]) || len(g.battleUnits) != int(binary.LittleEndian.Uint16(list[2:])) {
		t.Fatal("清單筆數不同")
	}
	for i, f := range g.battleFleets {
		if f.Index != int(list[4+i]) {
			t.Fatal("艦隊清單順序不同")
		}
	}
	for i, u := range g.battleUnits {
		if g.battleUnitPool(u) != int(binary.LittleEndian.Uint16(list[38+2*i:])) {
			t.Fatal("單位清單池序不同")
		}
	}
}

func TestBattleFinishGameOriginal(t *testing.T) {
	base := loadWithRules(t)
	rows := finishGameRows(t, "finish79-r89-a.json", "d26f893d7a617ed566f6ef7852abfcd1cf8476e746f123f3bad7551d3cc6e477")
	guard := finishGameRows(t, "finish79-r90-guard-a.json", "85fbd698622c307051183167d2315e9977abe5cac8fc8068611b31f4287ee335")
	if len(rows) != 60 || len(guard) != 7 {
		t.Fatal("案例數不同")
	}
	reports := 0
	for _, row := range append(rows, guard...) {
		t.Run(row.Name, func(t *testing.T) {
			g := finishGame(t, base, row)
			done := 0
			g.battleFinishPhase(func() { done++ })
			frame := 0
			for _, call := range row.Calls {
				if call.Location != "3826:0405" {
					continue
				}
				if done != 0 || g.state != StateTurnReport || !g.battleFinishing || len(g.turnRep.pending) != 1 {
					t.Fatal("報告未等待", g.saveErr)
				}
				r := g.turnRep.pending[0]
				face := call.Args[3] & 255
				if call.Args[2] != 65535 {
					face = g.res.Reporter[call.Args[2]]
				}
				if r.face != face || !bytes.Equal(r.text, ownershipHex(t, call.Text)) {
					t.Fatalf("報告文字／頭像不同 %d/%d %x/%s", r.face, call.Args[3], r.text, call.Text)
				}
				seed := g.battleRand.Seed()
				before := map[string]string{}
				for k, v := range map[string][]byte{"fleet": g.slot.Fleet, "ship": g.slot.Ship, "work": g.workRecord(), "admiral": g.slot.Admiral, "stellar": g.res.Stellar, "planet": g.res.Planets.Data, "lists": ownershipHex(t, row.Before["lists"])} {
					before[k] = fmt.Sprintf("%x", v)
				}
				if g.SaveSlot(t.TempDir(), 0) == nil {
					t.Fatal("占領等待可保存")
				}
				g.AdvanceBattleTurn()
				finishGameTables(t, g, before)
				if g.battleRand.Seed() != seed {
					t.Fatal("等待消費亂數")
				}
				if len(row.Frames) > 0 {
					if frame >= len(row.Frames) {
						t.Fatal("缺原版報告圖")
					}
					f := row.Frames[frame]
					promotionWindowCompare(t, g, row.Name, frame, f.Screen, f.Palette, f.PNG)
				}
				if reports%2 == 0 {
					g.Click(240, 240)
					g.LeftRelease()
				} else {
					g.RightClick(240, 240)
				}
				frame++
				reports++
			}
			if done != 1 || g.state != StateBattle || g.battleFinishing || g.battleOccupation != nil {
				t.Fatal("末端未續接一次", g.saveErr, done)
			}
			if g.battleFinishEnded != (row.AX == 1) {
				t.Fatal("勝負決定不同")
			}
			finishGameTables(t, g, row.After)
			if g.battleRand.Seed() != binary.LittleEndian.Uint32(ownershipHex(t, row.SeedAfter)) {
				t.Fatal("末端seed不同")
			}
			if err := g.syncFleetTables(); err != nil {
				t.Fatal(err)
			}
			g.slot.ApplyBattleResult(g.battleUnits)
			g.res.Planets.ApplyBattlePlanets(g.battlePlanets)
			finishGameTables(t, g, row.After)
		})
	}
	t.Log("67案完整六表／清單／seed及報告等待；報告", reports)
}

func TestBattleFinishSaveReload(t *testing.T) {
	base := loadWithRules(t)
	rows := finishGameRows(t, "finish79-r89-a.json", "d26f893d7a617ed566f6ef7852abfcd1cf8476e746f123f3bad7551d3cc6e477")
	checked := 0
	for _, row := range rows {
		if (row.Mask != 1 && row.Mask != 126) || row.Human == 0 || row.Round != 5 {
			continue
		}
		t.Run(row.Name, func(t *testing.T) {
			g := finishGame(t, base, row)
			done := 0
			g.battleFinishPhase(func() { done++ })
			for g.state == StateTurnReport {
				g.Click(240, 240)
				g.LeftRelease()
			}
			if done != 1 || g.saveErr != nil {
				t.Fatal("未完成", g.saveErr)
			}
			seed := g.battleRand.Seed()
			dir := t.TempDir()
			if out := os.Getenv("LOGH3_FINISH_OUTPUT"); out != "" {
				dir = filepath.Join(out, row.Name+"-save")
				if e := os.MkdirAll(dir, 0755); e != nil {
					t.Fatal(e)
				}
			}
			g.state = StateSaveGame // 第8-5步保存診斷；不宣稱正常玩家選單。
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
				if !bytes.Equal(read(name), ownershipHex(t, row.After[key])) {
					t.Fatal("保存不同", name)
				}
			}
			p := ownershipHex(t, row.After["planet"])
			if !bytes.Equal(read("PLANET"), p[64:]) {
				t.Fatal("保存行星不同")
			}
			s, e := rules.LoadSlot(rules.SlotFiles{Admiral: read("ADMIRAL"), Fleet: read("FLEET"), Ship: read("SHIP"), Stellar: read("STELLAR"), Work: read("WORK"), Planet: read("PLANET"), PlanetHeader: p[:64]})
			if e != nil {
				t.Fatal(e)
			}
			if g.battleRand.Seed() != seed || s.Upkeep != g.slot.Upkeep || s.StrategicScore != g.campaignScore {
				t.Fatal("保存消費亂數或回讀不同")
			}
			g.LoadSlot(s)
			if g.battleFinishing || g.battleOccupation != nil || g.battleFinishEnded {
				t.Fatal("載入殘留占領狀態")
			}
			if !bytes.Equal(g.slot.Ship, ownershipHex(t, row.After["ship"])) {
				t.Fatal("載入覆寫庫存")
			}
		})
		checked++
	}
	if checked != 6 {
		t.Fatal("保存案例數不同", checked)
	}
}

func TestBattleFinishGameCaptureWait(t *testing.T) {
	base := loadWithRules(t)
	rows := finishGameRows(t, "finish79-r91-cursor-a.json", "7ae71c1bea82bf045d5083b9326e41c1defdde899355164d1df2d699a82978f8")
	// 此收據在真正callback入口附完整六表。
	b, e := os.ReadFile(filepath.Join(root, "workplace/finish79-r91-cursor-a.json"))
	if e != nil {
		t.Fatal(e)
	}
	var raw struct {
		Rows []struct {
			Name  string
			Calls []struct {
				Location string
				Tables   map[string]string
				Seed     string `json:"seed_hex"`
			}
		}
	}
	if e = json.Unmarshal(b, &raw); e != nil {
		t.Fatal(e)
	}
	checked := 0
	for i, row := range rows {
		g := finishGame(t, base, row)
		done := 0
		g.battleFinishPhase(func() { done++ })
		for _, call := range row.Calls {
			if call.Location != "3826:0405" {
				continue
			}
			if g.battleOccupation != nil {
				for _, c := range raw.Rows[i].Calls {
					if c.Location == "185C:0B18" {
						finishGameTables(t, g, c.Tables)
						if g.battleRand.Seed() != binary.LittleEndian.Uint32(ownershipHex(t, c.Seed)) {
							t.Fatal("callbackseed不同")
						}
						checked++
					}
				}
			}
			g.Click(240, 240)
			g.LeftRelease()
		}
		if done != 1 {
			t.Fatal("未返回", g.saveErr)
		}
		finishGameTables(t, g, row.After)
	}
	if checked != 1 {
		t.Fatal("沒有驗真正占領等待點", checked)
	}
}
