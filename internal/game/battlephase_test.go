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

// 僅入口注入原版局面，35邊界全部由同一正式規則向前演算。
// 比較等級在第三十四輪執行原版前宣告；正常玩家抽樣另列。
func TestMovementAllBoundariesMatchOriginal(t *testing.T) {
	base := filepath.Join(root, "workplace/oracle/battle79-phase-r34-a")
	mem, err := os.ReadFile(filepath.Join(base, "entry-mem.bin"))
	if os.IsNotExist(err) {
		t.Skip("缺原版逐圈RAM")
	}
	if err != nil {
		t.Fatal(err)
	}
	if fmtHash(sha256.Sum256(mem)) != "9c4a4c28a1f53056962f613b1b643e5a5ba199d9231df291918d9499111be38d" {
		t.Fatal("原版入口RAM身份不同")
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
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/battle-phase-frames.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Frames) != 35 {
		t.Fatal("邊界數不同")
	}
	g := preparationGame(t, mem)
	g.originalBattleAI = true // 原版逐圈收據必須啟用兩方步進。
	loadMovementTiles(t, g)
	for i := range g.battleUnits {
		u := &g.battleUnits[i]
		r := g.slot.ShipAt(g.battleUnitPool(*u))
		u.Detail, u.Seen, u.DetailHold = int(r[0x17]), r[0x16]&1 != 0, int(r[0x18])
	}
	g.battleFleets = nil
	for _, r := range data.Frames[0].Fleets {
		raw, err := hex.DecodeString(r.Hex)
		if err != nil {
			t.Fatal(err)
		}
		g.battleFleets = append(g.battleFleets, rules.NewBattleFleet(r.Index, raw, [rules.FleetAbilityCount]byte{}))
	}
	g.battleView = screen.BattleView{Col: int(binary.LittleEndian.Uint16(mem[0x5E724:])) / g.res.Battle.Field.TileW, Row: int(binary.LittleEndian.Uint16(mem[0x5E726:])) / g.res.Battle.Field.TileH}
	g.MoveMouse(240, 240)
	g.battleMoving = true
	g.battlePhase = screen.BattlePhaseRepair
	check := func(i int) {
		t.Helper()
		f := data.Frames[i]
		if g.battleRand.Seed() != 0x12345678 || g.battleView != (screen.BattleView{Col: 66, Row: 30}) {
			t.Fatal("移動改變固定seed或鏡頭")
		}
		// 以正式寫回adapter重建全紀錄，未知byte從唯一入口保留；不從原版每幀還原。
		s := &rules.Slot{Fleet: bytes.Clone(g.slot.Fleet), Ship: bytes.Clone(g.slot.Ship)}
		for _, group := range g.battleMovementGroups {
			s.ApplyBattleMovementFleet(*group.Fleet)
			for _, u := range group.Units {
				s.ApplyBattleUnitMovement(u.Pool, *u.Unit)
			}
		}
		if len(f.Fleets) != 7 || len(f.Units) != 55 {
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
			t.Fatal("畫布大小不同")
		}
		movementFrame(t, g, "boundary-"+f.Name)
		diff := 0
		for j := range want {
			if got[j] != want[j] {
				diff++
			}
		}
		if diff != 0 {
			t.Errorf("%s全畫面不同%d／307200", f.Name, diff)
		}
		t.Logf("%s：7艦隊／55分隊全紀錄相同，色號差異%d", f.Name, diff)
	}
	check(0)
	g.battlePhase = screen.BattlePhaseMovement
	m := g.newBattleMovement()
	check(1)
	for i := 2; i < 34; i++ {
		if !m.Next() {
			t.Fatalf("%s提前結束", data.Frames[i].Name)
		}
		check(i)
	}
	if m.Next() {
		t.Fatal("第33圈未結束")
	}
	g.saveBattleMovement()
	check(34)
}
