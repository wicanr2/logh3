package rules

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func saveFleetTables(t *testing.T) (*ShipKindTable, []int) {
	t.Helper()
	k, err := LoadShipKindTable("../../docs/internals/data/ship-kind-table.json")
	if err != nil {
		t.Fatal(err)
	}
	e, err := LoadEditorTables("../../docs/internals/data/fleet-editor.json")
	if err != nil {
		t.Fatal(err)
	}
	return k, e.Mobility
}

func blankFleetSlot() *Slot {
	return &Slot{Fleet: bytes.Repeat([]byte{255}, FleetCount*FleetRecord), Ship: bytes.Repeat([]byte{255}, ShipCount*ShipRecord)}
}

func TestSyncFleetsPreservesOriginalTables(t *testing.T) {
	k, mobility := saveFleetTables(t)
	for _, ext := range []string{"8", "9", "A"} {
		t.Run(ext, func(t *testing.T) {
			read := func(name string) []byte {
				b, err := os.ReadFile(filepath.Join("../../org_game/extracted", name+".DT"+ext))
				if err != nil {
					t.Skip(err)
				}
				return b
			}
			s := &Slot{Fleet: read("FLEET"), Ship: read("SHIP"), Admiral: read("ADMIRAL")}
			fb, sb := bytes.Clone(s.Fleet), bytes.Clone(s.Ship)
			f := s.Fleets(nil)
			out, err := s.SyncFleets(f, k, mobility)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(out, f) {
				t.Fatal("未操作的編制被改動")
			}
			if !bytes.Equal(s.Fleet, fb) || !bytes.Equal(s.Ship, sb) {
				t.Fatal("未操作的原版艦隊／艦艇整表被改動")
			}
			// 改一個已建模欄位，既有池索引與未知 bytes 仍必須保留。
			out[0].Composition[0].Ships--
			pi := out[0].Composition[0].Pool
			if _, err = s.SyncFleets(out, k, mobility); err != nil {
				t.Fatal(err)
			}
			binary.LittleEndian.PutUint16(sb[pi*ShipRecord+shipAlive:], uint16(out[0].Composition[0].Ships))
			if !bytes.Equal(s.Ship, sb) {
				t.Fatal("更新艘數覆寫了其他欄位")
			}
		})
	}
}

func TestSyncFleetsSpeedsMatchOriginal(t *testing.T) {
	m, err := os.ReadFile("../../workplace/oracle/runtime85-normal/command.bin")
	if err != nil {
		t.Skip(err)
	}
	if len(m) != 1048576 {
		t.Fatal("原版記憶體收據長度錯誤")
	}
	po := int(binary.LittleEndian.Uint16(m[0x61AE4:])) + 16*int(binary.LittleEndian.Uint16(m[0x61AE6:]))
	orig := &Slot{Fleet: bytes.Clone(m[0x60BA4 : 0x60BA4+FleetCount*FleetRecord]), Ship: bytes.Clone(m[po : po+ShipCount*ShipRecord]), Admiral: bytes.Clone(m[0x602AC:0x60BA4])}
	in := orig.Fleets(nil)
	for i := range in {
		in[i].Scout = -1
		for j := range in[i].Composition {
			in[i].Composition[j].Pool = -1
		}
	}
	s := blankFleetSlot()
	s.Admiral = orig.Admiral
	k, mobility := saveFleetTables(t)
	out, err := s.SyncFleets(in, k, mobility)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(s.Fleets(nil), out) {
		t.Fatal("原版樣本的位置、命令或編制欄位存讀後不同")
	}
	n := 0
	for _, f := range out {
		old := orig.FleetComposition(f.Index)
		for j, g := range f.Composition {
			for _, o := range []int{0x1F, 0x21, 0x22} {
				if s.Ship[g.Pool*ShipRecord+o] != orig.Ship[old[j].Pool*ShipRecord+o] {
					t.Fatalf("艦隊 %d 組 %d 速度 +%X 不符原版", f.Index, j, o)
				}
			}
			n++
		}
		oldScout := int(binary.LittleEndian.Uint16(orig.fleet(f.Index)[fleetScout:]))
		for _, o := range []int{0x1F, 0x21, 0x22} {
			if s.Ship[f.Scout*ShipRecord+o] != orig.Ship[oldScout*ShipRecord+o] {
				t.Fatal("索敵隊速度不符原版")
			}
		}
		n++
	}
	if n != 19 {
		t.Fatalf("收據局面改變：%d 筆", n)
	}
	t.Log("原版 15 組編制＋4 組索敵隊，全部三項速度相同")
}

func TestSyncFleetsRejectsInvalidPoolWithoutMutation(t *testing.T) {
	k, mobility := saveFleetTables(t)
	for _, mode := range []string{"duplicate", "full", "out-of-range"} {
		t.Run(mode, func(t *testing.T) {
			s := blankFleetSlot()
			f := Fleet{Index: 0, Number: 1, Faction: 0, Admiral: 0, Adjutant: -1, Scout: -1, Composition: []FleetGroup{{Kind: 0, Pool: -1, Ships: 3000, Full: 3000, Slot2: 0, Slot3: 20}}}
			switch mode {
			case "full":
				for i := 0; i < ShipCount; i++ {
					s.Ship[i*ShipRecord+shipOwner] = 31
				}
			case "out-of-range":
				f.Composition[0].Pool = ShipCount
			case "duplicate":
				s.Ship[shipOwner], s.Ship[shipKind] = 0, 0
				f.Composition[0].Pool = 0
				f.Composition = append(f.Composition, f.Composition[0])
			}
			fb, sb := bytes.Clone(s.Fleet), bytes.Clone(s.Ship)
			if _, err := s.SyncFleets([]Fleet{f}, k, mobility); err == nil {
				t.Fatal("錯誤輸入未拒絕")
			}
			if !bytes.Equal(s.Fleet, fb) || !bytes.Equal(s.Ship, sb) {
				t.Fatal("失敗改動原始表")
			}
		})
	}
}

func TestFleetPoolIndex255IsValid(t *testing.T) {
	s := blankFleetSlot()
	r := s.Fleet[:FleetRecord]
	r[fleetUnused], r[fleetGroups] = 0, 1
	binary.LittleEndian.PutUint16(r[fleetSlots:], 255)
	r[fleetSlots+2], r[fleetSlots+3] = 0, 20
	sh := s.Ship[255*ShipRecord : 256*ShipRecord]
	sh[shipOwner], sh[shipKind] = 0, 0
	binary.LittleEndian.PutUint16(sh[shipAlive:], 3000)
	if f := s.FleetComposition(0); len(f) != 1 || f[0].Pool != 255 || f[0].Ships != 3000 {
		t.Fatal("00FF 池索引被誤判為 FFFF")
	}
	if n := len(s.FleetShips(0)); n != 1 {
		t.Fatalf("部署索引數 %d", n)
	}
	if n := len(s.FleetGroupShips(0)); n != 1 {
		t.Fatalf("計分索引數 %d", n)
	}
}

func TestSaveShrinkDoesNotDeployStaleSlots(t *testing.T) {
	k, mobility := saveFleetTables(t)
	s := blankFleetSlot()
	f := Fleet{Index: 0, Number: 1, Faction: 0, Admiral: 0, Adjutant: -1, Scout: -1, Composition: []FleetGroup{
		{Kind: 0, Pool: -1, Ships: 3000, Full: 3000, Slot2: 0, Slot3: 20},
		{Kind: 2, Pool: -1, Ships: 4000, Full: 4000, Slot2: 2, Slot3: 10},
	}}
	out, err := s.SyncFleets([]Fleet{f}, k, mobility)
	if err != nil {
		t.Fatal(err)
	}
	removed := out[0].Composition[1].Pool
	out[0].Composition = out[0].Composition[:1]
	out, err = s.SyncFleets(out, k, mobility)
	if err != nil {
		t.Fatal(err)
	}
	if s.Ship[removed*ShipRecord+shipOwner] != 255 {
		t.Fatal("縮編後的池槽未作廢")
	}
	if units := s.Deploy(nil, nil, out[0], 100, 100, 64); len(units) != 1 {
		t.Fatalf("縮編後部署了 %d 組，讀到舊槽", len(units))
	}
	if got := s.Fleets(nil)[0].Groups; got != 1 {
		t.Fatalf("存檔組數仍是 %d", got)
	}
}
