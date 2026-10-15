package rules

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestFlagshipTransferChoiceAndChance(t *testing.T) {
	units := []BattleUnit{
		{Fleet: 2, Faction: 0, Squadron: 0, Ships: 0},
		{Fleet: 2, Faction: 0, Squadron: 3, Kind: 0, Ships: 100},
		{Fleet: 2, Faction: 0, Squadron: 1, Kind: 1, Ships: 100},
		{Fleet: 2, Faction: 0, Squadron: 2, Kind: 2, Ships: 999},
		{Fleet: 2, Faction: 0, Squadron: -1, Kind: 7, Ships: 999},
		{Fleet: 3, Faction: 0, Squadron: 1, Kind: 0, Ships: 999},
		{Fleet: 2, Faction: 0, Squadron: 4, Kind: 0, Ships: 999, Out: true},
	}
	if FlagshipTransferTarget(units, 0) != 2 {
		t.Fatal("艦種、艦隊、離場或同艘數順序錯誤")
	}
	units[1].Ships = 101
	if FlagshipTransferTarget(units, 0) != 1 {
		t.Fatal("未選最多艘數")
	}
	units[1].Ships, units[2].Ships = 0, 0
	if FlagshipTransferTarget(units, 0) != -1 {
		t.Fatal("沒有有效戰艦卻轉移")
	}
	for roll := 0; roll < 100; roll++ {
		if FlagshipTransferSucceeds(5, roll) != true || FlagshipTransferSucceeds(0, roll) != (roll < 25) || FlagshipTransferSucceeds(52, roll) != (roll < 25) {
			t.Fatalf("機率邊界 %d", roll)
		}
	}
}

func transferSlotFixture() *Slot {
	s := &Slot{Fleet: make([]byte, FleetCount*FleetRecord), Ship: make([]byte, ShipCount*ShipRecord)}
	r := s.Fleet[:FleetRecord]
	r[fleetUnused], r[fleetAdjutant], r[fleetGroups] = 5, 8, 3
	for i, ships := range []int{0, 70, 90} {
		o, pool := fleetSlots+4*i, 10+i
		binary.LittleEndian.PutUint16(r[o:], uint16(pool))
		r[o+2], r[o+3] = byte(i%2), byte(20+i)
		sh := s.ShipAt(pool)
		sh[shipOwner], sh[shipKind], sh[3], sh[5] = 0, byte(i%2), byte(i), byte(20+i)
		binary.LittleEndian.PutUint16(sh[shipAlive:], uint16(ships))
		binary.LittleEndian.PutUint16(sh[shipPosX:], uint16(100+i*40))
		binary.LittleEndian.PutUint16(sh[shipPosX+2:], uint16(200+i*30))
	}
	return s
}

func TestTransferFlagshipPoolTransaction(t *testing.T) {
	s := transferSlotFixture()
	before := bytes.Clone(s.Ship)
	if err := s.TransferFlagship(0, 2); err != nil {
		t.Fatal(err)
	}
	if got := s.FleetGroupShips(0); got[0] != 12 || got[2] != 10 {
		t.Fatal("池引用未交換", got)
	}
	if s.Fleet[4] != 5 || s.Fleet[5] != 8 || s.Fleet[15] != 3 {
		t.Fatal("提督／副官／組數變動")
	}
	for i := range before {
		if i == 10*ShipRecord+3 || i == 10*ShipRecord+5 || i == 12*ShipRecord+3 || i == 12*ShipRecord+5 {
			continue
		}
		if before[i] != s.Ship[i] {
			t.Fatalf("轉移改動艦艇資源或位置：%x", i)
		}
	}
	if s.ShipAt(12)[5] != FlagshipFormationSlot || s.ShipAt(10)[5] != 22 {
		t.Fatal("旗艦隊形槽不同步")
	}
	if int(binary.LittleEndian.Uint16(s.Fleet[fleetPosX:])) != 180 {
		t.Fatal("艦隊位置未跟隨新旗艦")
	}
	for _, kind := range []string{"alive", "nonbattleship", "owner", "duplicate", "out"} {
		t.Run(kind, func(t *testing.T) {
			s := transferSlotFixture()
			switch kind {
			case "alive":
				s.ShipAt(10)[shipAlive] = 1
			case "nonbattleship":
				s.ShipAt(12)[shipKind] = 2
			case "owner":
				s.ShipAt(12)[shipOwner] = 7
			case "duplicate":
				binary.LittleEndian.PutUint16(s.Fleet[fleetSlots+4:], 12)
			case "out":
				s.ShipAt(12)[0x1e] = 0x80
			}
			fleet, ship := bytes.Clone(s.Fleet), bytes.Clone(s.Ship)
			if s.TransferFlagship(0, 2) == nil || !bytes.Equal(fleet, s.Fleet) || !bytes.Equal(ship, s.Ship) {
				t.Fatal("無效輸入未完整回滾")
			}
		})
	}
}
