package rules

import (
	"encoding/binary"
	"path/filepath"
	"testing"
)

// TestAllianceDeploymentMatchesOracle 釘住 issue #11：同盟分隊不是另一套
// 陣營專用公式。原版 `297B:091D` → `2BE1:0000` 逐支讀艦艇 `+0x05`，
// 再由 `2BE1:0059` 用同一張 `ds:59BA` 槽位表旋轉定位。
//
// 收據 s3ai6 的同盟六支艦隊都在星域 26；這裡從 1 MB 傾印重建艦隊表與
// 艦艇池，再以每支艦隊的旗艦位置／朝向重算所有分隊位置。
func TestAllianceDeploymentMatchesOracle(t *testing.T) {
	mem := readReceipt(t, "s3ai6.bin")
	if len(mem) != 1<<20 {
		t.Fatalf("s3ai6.bin 是 %d bytes，要 1 MB", len(mem))
	}

	const poolPointer = dumpDS + 0xCFB4
	poolOff := int(binary.LittleEndian.Uint16(mem[poolPointer:]))
	poolSeg := int(binary.LittleEndian.Uint16(mem[poolPointer+2:]))
	pool := poolSeg*16 + poolOff
	if pool < 0 || pool+ShipCount*ShipRecord > len(mem) {
		t.Fatalf("艦艇池線性位址 %#x 超出傾印", pool)
	}

	formation, err := LoadFormationSlots(filepath.Join("..", "..",
		"docs", "internals", "data", "battle-formation-slots.json"))
	if err != nil {
		t.Fatal(err)
	}
	s := &Slot{
		Fleet: mem[dumpFleets : dumpFleets+FleetCount*FleetRecord],
		Ship:  mem[pool : pool+ShipCount*ShipRecord],
	}

	fleets := 0
	units := 0
	for _, f := range s.Fleets(nil) {
		if f.Region != 26 || f.Faction != 1 {
			continue
		}
		fleets++
		r := s.fleet(f.Index)
		x := int(binary.LittleEndian.Uint16(r[0x5C:]))
		y := int(binary.LittleEndian.Uint16(r[0x5E:]))
		got := s.Deploy(nil, formation, f, x, y, int(r[0x4A]))
		ships := s.FleetShips(f.Index)
		if len(got) != len(ships) {
			t.Fatalf("艦隊 %d 分隊數 %d，要 %d", f.Number, len(got), len(ships))
		}
		for i, u := range got {
			a := pool + ships[i]*ShipRecord
			wantX := int(binary.LittleEndian.Uint16(mem[a+0x28:]))
			wantY := int(binary.LittleEndian.Uint16(mem[a+0x2A:]))
			if u.X != wantX || u.Y != wantY {
				t.Errorf("同盟第 %d 艦隊分隊 %d（槽 %d）：算出 (%d,%d)，收據 (%d,%d)",
					f.Number, i, u.Slot, u.X, u.Y, wantX, wantY)
			}
			units++
		}
	}
	if fleets != 6 || units == 0 {
		t.Fatalf("比對到 %d 支同盟艦隊、%d 個分隊，要 6 支且分隊非零", fleets, units)
	}
}
