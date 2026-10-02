package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 第36輪事前宣告same-state；S1唯一入口注入，不以原版每圈資料餵引擎。
// 正常玩家路徑另由TestBattleShippedSavesNormalExecuteAndSave抽驗。
func TestMovementSecondBattleRecordsMatchOriginal(t *testing.T) {
	base := filepath.Join(root, "workplace/oracle/battle79-second-r36-a")
	mem, err := os.ReadFile(filepath.Join(base, "entry-mem.bin"))
	if os.IsNotExist(err) {
		t.Skip("缺S1原版收據；入口tools/battle_second_oracle.sh")
	}
	if err != nil {
		t.Fatal(err)
	}
	if fmtHash(sha256.Sum256(mem)) != "9095dfa08ab36201784e6d4555b5e42312ac76bb20049614cf53bf93bc09e7ec" {
		t.Fatal("原版S1入口身份不同")
	}
	var data struct {
		Frames []struct {
			Name   string
			Hashes map[string]string
			Fleets []struct {
				Index int
				Hex   string
			}
			Units []struct {
				Index int
				Hex   string
			}
		}
	}
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/battle-second-frames.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Frames) != 43 {
		t.Fatal("S1邊界數不同")
	}
	g := newTestGame(t)
	g.loadSlotFromDisk(8)
	if g.turn != 93 {
		t.Fatal("未載入S1")
	}
	loadPreparationPresentation(t, g)
	loadMovementTiles(t, g)
	// 明示診斷：參戰清單由原版入口取得。包括尚未改戰略星域的進攻艦隊。
	g.slot.Fleet = bytes.Clone(mem[0x60BA4 : 0x60BA4+32*122])
	g.slot.Ship = bytes.Clone(mem[0x668B2 : 0x668B2+352*78])
	copy(g.slot.Planets.Data[rules.PlanetHeader:], mem[0x53132:0x53132+81*82])
	g.LoadSlot(g.slot)
	g.state, g.battleRegion, g.battleTurn = StateBattle, 8, 1
	g.battleUnits, g.battleFleets = nil, nil
	for _, f := range data.Frames[0].Fleets {
		raw, err := hex.DecodeString(f.Hex)
		if err != nil {
			t.Fatal(err)
		}
		bf := rules.NewBattleFleet(f.Index, raw, [rules.FleetAbilityCount]byte{})
		g.battleFleets = append(g.battleFleets, bf)
		fleet := g.fleetByIndex(f.Index)
		if fleet == nil {
			t.Fatalf("缺艦隊%d", f.Index)
		}
		g.battleUnits = append(g.battleUnits, g.slot.Deploy(g.res.Weights, g.res.Formation, *fleet, bf.X, bf.Y, bf.Facing)...)
	}
	for i := range g.battleUnits {
		u := &g.battleUnits[i]
		r := g.slot.ShipAt(g.battleUnitPool(*u))
		u.X, u.Y = int(binary.LittleEndian.Uint16(r[0x34:])), int(binary.LittleEndian.Uint16(r[0x36:]))
		u.TargetX, u.TargetY = int(binary.LittleEndian.Uint16(r[0x28:])), int(binary.LittleEndian.Uint16(r[0x2A:]))
		u.Detail, u.Seen, u.DetailHold = int(r[0x17]), r[0x16]&1 != 0, int(r[0x18])
	}
	g.battlePlanets, g.battleRally = g.battleRegionPlanets(8), g.battleRegionRally(8)
	for i := range g.battlePlanets {
		p := &g.battlePlanets[i]
		r := mem[0x53132+p.Index*82:]
		p.Owner, p.X, p.Y = int(r[17]), int(binary.LittleEndian.Uint16(r[70:])), int(binary.LittleEndian.Uint16(r[72:]))
		p.Defence, p.DefenceMax, p.Recover = int(binary.LittleEndian.Uint16(r[20:])), int(binary.LittleEndian.Uint16(r[22:])), int(binary.LittleEndian.Uint16(r[24:]))
		p.Stock, _ = rules.ReadPlanetStock(r)
	}
	g.battleView = screen.BattleView{Col: 58, Row: 1}
	g.battleRand = rules.NewRand(0xECC6CD87)
	g.MoveMouse(100, 150)
	g.battleMoving, g.battlePhase = true, screen.BattlePhaseRepair
	check := func(i int) {
		t.Helper()
		f := data.Frames[i]
		if g.battleRand.Seed() != 0xECC6CD87 || g.battleView != (screen.BattleView{Col: 58, Row: 1}) {
			t.Fatal("seed或鏡頭不同")
		}
		s := &rules.Slot{Fleet: bytes.Clone(g.slot.Fleet), Ship: bytes.Clone(g.slot.Ship)}
		for _, group := range g.battleMovementGroups {
			s.ApplyBattleMovementFleet(*group.Fleet)
			for _, u := range group.Units {
				s.ApplyBattleUnitMovement(u.Pool, *u.Unit)
			}
		}
		if len(f.Fleets) != 18 || len(f.Units) != 135 {
			t.Fatal("原版清單大小不同")
		}
		for _, r := range f.Fleets {
			want, err := hex.DecodeString(r.Hex)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(s.Fleet[r.Index*122:(r.Index+1)*122], want) {
				t.Fatalf("%s艦隊%d完整紀錄不同", f.Name, r.Index)
			}
		}
		for _, r := range f.Units {
			want, err := hex.DecodeString(r.Hex)
			if err != nil {
				t.Fatal(err)
			}
			got := s.Ship[r.Index*78 : (r.Index+1)*78]
			if !bytes.Equal(got, want) {
				t.Fatalf("%s池%d完整紀錄不同：%x / %x", f.Name, r.Index, got, want)
			}
		}
		want, err := os.ReadFile(filepath.Join(base, f.Name+".bin"))
		if err != nil {
			t.Fatal(err)
		}
		if fmtHash(sha256.Sum256(want)) != f.Hashes[".bin"] {
			t.Fatal("原版畫面雜湊不同")
		}
		got := g.Frame().Pix
		if len(got) != len(want) || len(want) != 640*480 {
			t.Fatal("畫布不同")
		}
		movementFrame(t, g, "second-"+f.Name)
		// 只用每幀原版RAM作斷言，不更新Game。方向位置直接讀原版物件，
		// 不在測試複製renderer的八向座標公式。
		frameMem, err := os.ReadFile(filepath.Join(base, f.Name+"-mem.bin"))
		if err != nil {
			t.Fatal(err)
		}
		if fmtHash(sha256.Sum256(frameMem)) != f.Hashes["-mem.bin"] {
			t.Fatal("原版每幀RAM身份不同")
		}
		directionPixels := 0
		field := &g.res.Battle.Field
		for _, r := range f.Fleets {
			fleet := frameMem[0x60BA4+r.Index*122:]
			obj := int(binary.LittleEndian.Uint16(fleet[0x6E:]))*16 + int(binary.LittleEndian.Uint16(fleet[0x6C:]))
			if obj == 0 {
				continue
			}
			dir := int(frameMem[obj+12])
			if dir >= 8 {
				continue
			}
			sx := field.X + int(int16(binary.LittleEndian.Uint16(frameMem[obj+6:])))
			sy := field.Y + int(int16(binary.LittleEndian.Uint16(frameMem[obj+8:])))
			tile := g.res.Sprites[dir]
			for y := 0; y < tile.H; y++ {
				for x := 0; x < tile.W; x++ {
					px, py := sx+x, sy+y
					if tile.Pix[y*tile.W+x] == 0 || px < field.X || py < field.Y || px >= field.X+field.Cols*field.TileW || py >= field.Y+field.Rows*field.TileH {
						continue
					}
					index := py*640 + px
					if got[index] != want[index] {
						t.Errorf("%s艦隊%d方向像素%d,%d不同", f.Name, r.Index, px, py)
					}
					directionPixels++
				}
			}
		}
		if directionPixels != 440 {
			t.Fatalf("%s方向樣本%d，不是原版440", f.Name, directionPixels)
		}
		diff := 0
		fieldDiff := 0
		for j := range want {
			if got[j] != want[j] {
				diff++
				x, y := j%640, j/640
				if x >= field.X && x < field.X+field.Cols*field.TileW && y >= field.Y && y < field.Y+field.Rows*field.TileH {
					fieldDiff++
				}
			}
		}
		if fieldDiff != 0 {
			t.Errorf("%s正式Field有%d個差異像素，分隊繪製次序尚未符合085", f.Name, fieldDiff)
		}
		if diff != 0 {
			t.Errorf("%s正式全畫面差%d／307200，尚未符合086", f.Name, diff)
		}
		t.Logf("%s：18艦隊／135分隊完整紀錄與440個方向像素相同；Field差%d，全畫面差%d／307200", f.Name, fieldDiff, diff)
	}
	check(0)
	g.battlePhase = screen.BattlePhaseMovement
	m := g.newBattleMovement()
	refs := g.preparationUnits()
	if len(refs) != 135 {
		t.Fatalf("正式清單%d，不是135", len(refs))
	}
	for i, r := range refs {
		if r.Pool != data.Frames[0].Units[i].Index {
			t.Fatalf("清單%d原池不同", i)
		}
	}
	check(1)
	for i := 2; i < 42; i++ {
		if !m.Next() {
			t.Fatalf("%s提前結束", data.Frames[i].Name)
		}
		check(i)
	}
	if m.Next() {
		t.Fatal("第41圈未結束")
	}
	g.saveBattleMovement()
	check(42)
}
