package rules

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 整階段由原版entry計算；預期end只在終點比較，沒有選骰、排除byte或逐欄遮罩。
func TestBattleMovementPhaseMatchesOriginal(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("../..", "docs/internals/data/battle-movement-phase.json"))
	if err != nil {
		t.Fatal(err)
	}
	type record struct {
		Fleet int `json:"fleet_index"`
		Pool  int `json:"ship_index"`
		Scout bool
		Entry string `json:"entry_hex"`
		Init  string `json:"init_hex"`
		End   string `json:"end_hex"`
	}
	var d struct {
		Fleets, Units []record
		Seed          string
	}
	if err = json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	if len(d.Fleets) != 7 || len(d.Units) != 55 || d.Seed != "0x12345678" {
		t.Fatal("原版完整局面不符")
	}
	formation, err := LoadFormationSlots(filepath.Join("../..", "docs/internals/data/battle-formation-slots.json"))
	if err != nil {
		t.Fatal(err)
	}
	terrain, err := LoadBattleTerrain(filepath.Join("../..", "docs/internals/data/battle-terrain.json"))
	if err != nil {
		t.Fatal(err)
	}
	// 同態地圖只從原版RAM取得，沒有原版時明確skip，不能填一張全通行地圖。
	mem, err := os.ReadFile(filepath.Join("../..", "workplace/oracle/battle79-move-entry-a/mem-in.bin"))
	if os.IsNotExist(err) {
		t.Skip("缺原版移動入口RAM")
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(mem) != 1048576 {
		t.Fatal("入口RAM長度不符")
	}
	if got := sha256.Sum256(mem); got != [32]byte{0x9c, 0x4a, 0x4c, 0x28, 0xa1, 0xf5, 0x30, 0x56, 0x96, 0x2f, 0x61, 0x3b, 0x1b, 0x64, 0x3e, 0x5a, 0x5b, 0xa1, 0x99, 0xd9, 0x23, 0x1d, 0xf2, 0x91, 0x91, 0x8d, 0x94, 0x99, 0x11, 0x1b, 0xe3, 0x8d} {
		t.Fatal("原版入口雜湊不符")
	}
	slot := &Slot{Fleet: make([]byte, 32*FleetRecord), Ship: make([]byte, ShipCount*ShipRecord)}
	fleets := make([]BattleFleet, 7)
	units := make([]BattleUnit, 55)
	groups := make([]MovementFleet, 7)
	for i, r := range d.Fleets {
		raw := stepBytes(t, r.Entry)
		copy(slot.Fleet[r.Fleet*FleetRecord:], raw)
		fleets[i] = NewBattleFleet(r.Fleet, raw, [FleetAbilityCount]byte{})
		groups[i].Fleet = &fleets[i]
	}
	for i, r := range d.Units {
		raw := stepBytes(t, r.Entry)
		copy(slot.Ship[r.Pool*ShipRecord:], raw)
		units[i] = *stepUnit(t, r.Entry)
		units[i].Fleet = r.Fleet
		for j := range groups {
			if groups[j].Fleet.Index != r.Fleet {
				continue
			}
			units[i].Squadron = len(groups[j].Units)
			if r.Scout {
				units[i].Squadron = -1
			}
			groups[j].Units = append(groups[j].Units, MovementOccupant{Pool: r.Pool, Unit: &units[i]})
		}
	}
	m := NewBattleMovement(groups, formation, terrain, mem[0x6D3F4:0x6D3F4+128*102])
	save := func() {
		for _, f := range fleets {
			slot.ApplyBattleMovementFleet(f)
		}
		for i, u := range units {
			slot.ApplyBattleUnitMovement(d.Units[i].Pool, u)
		}
	}
	save()
	for _, r := range d.Fleets {
		if !bytes.Equal(slot.Fleet[r.Fleet*FleetRecord:(r.Fleet+1)*FleetRecord], stepBytes(t, r.Init)) {
			t.Fatalf("艦隊%d初始化不同", r.Fleet)
		}
	}
	for _, r := range d.Units {
		if !bytes.Equal(slot.Ship[r.Pool*ShipRecord:(r.Pool+1)*ShipRecord], stepBytes(t, r.Init)) {
			t.Fatalf("分隊%d初始化不同", r.Pool)
		}
	}
	rounds := 1
	for m.Next() {
		rounds++
		if rounds > 1000 {
			t.Fatal("交錯未停止")
		}
	}
	save()
	for _, r := range d.Fleets {
		got := slot.Fleet[r.Fleet*FleetRecord : (r.Fleet+1)*FleetRecord]
		if !bytes.Equal(got, stepBytes(t, r.End)) {
			t.Errorf("艦隊%d完整122 bytes不同：\n%x\n%x", r.Fleet, got, stepBytes(t, r.End))
		}
	}
	for _, r := range d.Units {
		got := slot.Ship[r.Pool*ShipRecord : (r.Pool+1)*ShipRecord]
		if !bytes.Equal(got, stepBytes(t, r.End)) {
			t.Errorf("分隊%d完整78 bytes不同：\n%x\n%x", r.Pool, got, stepBytes(t, r.End))
		}
	}
	if rounds != 33 {
		t.Errorf("交錯圈數%d，模型33", rounds)
	}
	t.Log("原版7艦隊／55分隊初始化與完整終態零差異，33圈；兩側seed=0x12345678；規則無RNG輸入")
}

func TestBattleUnitStepsMatchOriginal(t *testing.T) {
	fixture := loadFleetStepReceipt(t)
	formation, err := LoadFormationSlots(filepath.Join("../..", "docs/internals/data/battle-formation-slots.json"))
	if err != nil {
		t.Fatal(err)
	}
	terrain, err := LoadBattleTerrain(filepath.Join("../..", "docs/internals/data/battle-terrain.json"))
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for i, row := range fixture.Receipt.Rows {
		if strings.HasPrefix(row.Name, "fleet-") || strings.HasPrefix(row.Name, "collision-") {
			continue
		}
		count++
		t.Run(row.Name, func(t *testing.T) {
			call := fixture.Receipt.Plan.Cases[i].Calls[0]
			label, pool := "unit", 104
			if strings.Contains(row.Name, "scout") {
				label, pool = "scout", 103
			}
			if strings.Contains(row.Name, "flagship") {
				label, pool = "flagship", 102
			}
			if call.Offset == 0x3AA && call.Args[0] == 2+103*78 {
				label, pool = "scout", 103
			}
			u := stepUnit(t, row.Before[label])
			u.Squadron = 1
			if label == "scout" {
				u.Squadron = -1
			}
			if label == "flagship" {
				u.Squadron = 0
			}
			raw := stepBytes(t, row.Before["fleet"])
			f := NewBattleFleet(16, raw, [FleetAbilityCount]byte{})
			cells, enemyCount := stepMap(t, fixture, i)
			var occupants []MovementOccupant
			if enemyCount == 1 {
				occupants = []MovementOccupant{{Pool: 142, Unit: stepUnit(t, row.Before["enemy"])}}
			}
			check := func(x, y int) int { return BattleMovementCheck(terrain, cells, occupants, u.Faction, x, y) }
			processed := false
			switch call.Offset {
			case 0x444:
				MoveBattleUnitStep(&f, u, call.Args[2], check)
			case 0x229: // 原版直接呼叫同步，等價於旗艦分派。
				saved := u.Squadron
				u.Squadron = 0
				processed = AdvanceBattleUnitStep(&f, u, formation, check)
				u.Squadron = saved
			case 0x3AA:
				processed = AdvanceBattleUnitOwnStep(&f, u, check)
			case 0x29C, 0x153:
				processed = AdvanceBattleUnitStep(&f, u, formation, check)
			default:
				t.Fatalf("未處理原版入口%X", call.Offset)
			}
			if call.Offset != 0x444 && processed != (row.Returns[0]&255 != 0) {
				t.Errorf("AL%d，處理%v", row.Returns[0]&255, processed)
			}
			slot := &Slot{Fleet: make([]byte, 17*FleetRecord), Ship: make([]byte, 105*ShipRecord)}
			copy(slot.Fleet[16*FleetRecord:], raw)
			copy(slot.Ship[pool*ShipRecord:], stepBytes(t, row.Before[label]))
			slot.ApplyBattleMovementFleet(f)
			slot.ApplyBattleUnitMovement(pool, *u)
			if !bytes.Equal(slot.Fleet[16*FleetRecord:], stepBytes(t, row.After["fleet"])) {
				t.Errorf("完整艦隊不同")
			}
			if !bytes.Equal(slot.Ship[pool*ShipRecord:(pool+1)*ShipRecord], stepBytes(t, row.After[label])) {
				t.Errorf("完整分隊不同：\n%x\n%x", slot.Ship[pool*ShipRecord:(pool+1)*ShipRecord], stepBytes(t, row.After[label]))
			}
			if row.Seed != "78563412" {
				t.Fatal("原版seed不同")
			}
		})
	}
	if count != 55 {
		t.Fatalf("分隊案例%d，要55", count)
	}
}
