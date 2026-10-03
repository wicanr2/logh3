package rules

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func visibleFleetMemory(t *testing.T, name, file, hash string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("../../workplace/oracle", "intel59-r60-"+name+"-b", file+".bin"))
	if os.IsNotExist(err) {
		t.Skip("缺第60輪原版收據；tools/region_visible_fleets_oracle.sh重播")
	}
	if err != nil || len(b) != 1048576 || fmt.Sprintf("%x", sha256.Sum256(b)) != hash {
		t.Fatalf("原版RAM漂移或無效：%v", err)
	}
	return b
}

// 原版輸入RAM→正式Slot.Fleets→導出規則，比全部32星域的64格。
// 晚注入案例只裁決受控規則；normal完全不注入。
func TestRegionVisibleFleetsMatchOracle(t *testing.T) {
	for _, c := range []struct{ name, input, command string }{
		{"normal", "aff5a1ba00c35265c9307b7a9104bc4c272514124a2e1ed87e5d9d2dbe8fa37d", "688ff1c526efee10bc98d0e16332de511516b3f76302c643d5f2ac43f734ddf1"},
		{"late1", "2ed0cc65c7f6e43a878c1acca37e7a90e2173b8f0173508642b73d70669f3e16", "e5d57a270fbeaba1685c3de11989cbd3bb0ed59466a19a14b05577398927be16"},
		{"late255", "f0c5c84f790f9f9aaf2413b722d097b40550ec093ef6357c5b59f96e11a35f5d", "c451b44d88f23fa02437e57fc16f193ae716860a052ff2b0b0ab7bf9193b91c3"},
	} {
		t.Run(c.name, func(t *testing.T) {
			input := visibleFleetMemory(t, c.name, "input", c.input)
			command := visibleFleetMemory(t, c.name, "command", c.command)
			pool := int(binary.LittleEndian.Uint16(input[0x61AE4:])) + 16*int(binary.LittleEndian.Uint16(input[0x61AE6:]))
			s := &Slot{Fleet: bytes.Clone(input[0x60BA4 : 0x60BA4+FleetCount*FleetRecord]), Ship: bytes.Clone(input[pool : pool+ShipCount*ShipRecord])}
			records := bytes.Clone(input[0x5F2CC : 0x5F2CC+RegionCount*RegionRecord])
			before := bytes.Clone(records)
			RecountRegionVisibleFleets(records, s.Fleets(nil))
			for r := 0; r < RegionCount; r++ {
				o := r*RegionRecord + 0x2A
				if !bytes.Equal(records[o:o+2], command[0x5F2CC+o:0x5F2CC+o+2]) {
					t.Fatalf("星域%d可見數不符原版：%x", r, records[o:o+2])
				}
				copy(before[o:o+2], records[o:o+2])
			}
			if !bytes.Equal(records, before) {
				t.Fatal("重算改寫其他星域欄位")
			}
		})
	}
}

func TestRegionVisibleFleetsGroupAndBounds(t *testing.T) {
	records := bytes.Repeat([]byte{0xA5}, 2*RegionRecord+3)
	visible := []FleetGroup{{Detail: 1}, {Detail: 3}}
	fleets := []Fleet{
		{Faction: 0, Region: 0, Composition: visible, Flags: 1}, // 兩組也只算一支
		{Faction: 1, Region: 0, Composition: []FleetGroup{{Detail: 255, Ships: 0}}},
		{Faction: 1, Region: 0, Composition: visible},
		{Faction: 0, Region: 1, Composition: []FleetGroup{{Detail: 0}}},
		{Faction: 1, Region: 1, Scout: 0}, // 不讀獨立偵察隊
		{Faction: 2, Region: 0, Composition: visible},
		{Faction: 0, Region: -1, Composition: visible},
		{Faction: 0, Region: 2, Composition: visible},
	}
	before := bytes.Clone(records)
	RecountRegionVisibleFleets(records, fleets)
	copy(before[0x2A:0x2C], []byte{1, 2})
	copy(before[RegionRecord+0x2A:RegionRecord+0x2C], []byte{0, 0})
	if !bytes.Equal(records, before) {
		t.Fatal("多組、陣營、艘數、旗標、偵察隊、範圍或尾端處理錯誤")
	}
	RecountRegionVisibleFleets(nil, fleets)
}
