package rules

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func upkeepPlanets(count int) *Planets {
	data := make([]byte, PlanetHeader+count*PlanetRecord)
	binary.LittleEndian.PutUint16(data, uint16(count))
	p, _ := LoadPlanets(data)
	return p
}

func TestDepartNeutralPlanet(t *testing.T) {
	p := upkeepPlanets(3)
	for i, owner := range []byte{0, RegionOwnerNone, 1} {
		p.Record(i)[planetOwner] = owner
		binary.LittleEndian.PutUint32(p.Record(i)[planetEconomy:], 300)
		binary.LittleEndian.PutUint16(p.Record(i)[planetDefence:], 90)
		binary.LittleEndian.PutUint16(p.Record(i)[planetGrowthDef:], 120)
	}
	record := make([]byte, RegionRecord)
	record[regionOwner], record[regionCount+2] = RegionOwnerNone, 1
	rnd := NewRand(1)
	before := rnd.Seed()
	picked := DepartNeutralPlanet(record, p, 0, rnd)
	if picked != 0 && picked != 2 {
		t.Fatalf("選到不合資格的行星 %d", picked)
	}
	if rnd.Seed() == before {
		t.Fatal("有候選行星時應固定消耗一次亂數")
	}
	if p.Owner(picked) != RegionOwnerNone || p.PlanetEconomy(picked) != 200 {
		t.Errorf("行星 %d 沒完整套用脫離效果", picked)
	}

	seed := rnd.Seed()
	record[regionCount+2] = 0
	if got := DepartNeutralPlanet(record, p, 0, rnd); got != -1 || rnd.Seed() != seed {
		t.Errorf("+0x18 為 0 時仍處理 %d 或消耗亂數", got)
	}
}

func TestRecoverRegionDefences(t *testing.T) {
	p := upkeepPlanets(3)
	for i, v := range []struct{ defence, limit, rate int }{
		{1000, 7000, 2}, // 收據 rtu22：1000 + 9×140 = 2260
		{95, 100, 10},   // 第一輪即夾到上限
		{10, 99, 0},     // step 為 0，九輪後仍不變
	} {
		r := p.Record(i)
		binary.LittleEndian.PutUint16(r[planetDefence:], uint16(v.defence))
		binary.LittleEndian.PutUint16(r[planetGrowthDef:], uint16(v.limit))
		binary.LittleEndian.PutUint16(r[planetRecoveryRate:], uint16(v.rate))
	}
	RecoverRegionDefences(p, 0)
	for i, want := range []int{2260, 100, 10} {
		if got := p.Defence(i); got != want {
			t.Errorf("行星 %d 防衛力 %d，應為 %d", i, got, want)
		}
	}
}

func TestResetRegionFortressGuns(t *testing.T) {
	p := upkeepPlanets(2)
	for i := 0; i < 2; i++ {
		p.Record(i)[planetGunState] = 3
		binary.LittleEndian.PutUint16(p.Record(i)[planetGunAim:], 0x40)
	}
	p.Record(0)[planetFortress] = 1
	ResetRegionFortressGuns(p, 0)
	if p.Record(0)[planetGunState] != FortressGunIdle ||
		binary.LittleEndian.Uint16(p.Record(0)[planetGunAim:]) != FortressGunNoAim {
		t.Error("要塞砲沒有重置")
	}
	if p.Record(1)[planetGunState] != 3 || binary.LittleEndian.Uint16(p.Record(1)[planetGunAim:]) != 0x40 {
		t.Error("沒有要塞的行星不應被修改")
	}
}

func readUpkeepReceipt(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "workplace", "oracle", "rtu22b", name))
	if err != nil {
		t.Skipf("缺 #22 收據 %s，跳過（不用代用品）", name)
	}
	return b
}

func upkeepReceiptPlanets(t *testing.T, raw []byte) *Planets {
	t.Helper()
	counts := []int{4, 3, 2, 3, 1, 2, 3, 3, 4, 3, 2, 2, 1, 1, 1, 2,
		3, 2, 3, 3, 3, 4, 2, 2, 1, 3, 3, 2, 3, 2, 3, 5}
	data := make([]byte, PlanetHeader+len(raw))
	for i, n := range counts {
		binary.LittleEndian.PutUint16(data[i*2:], uint16(n))
	}
	copy(data[PlanetHeader:], raw)
	p, err := LoadPlanets(data)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// TestRegionUpkeepMatchesReceipt 對 dosgolem `rtu22b` 的兩個受控局面：
// 星域 3 無主但仍含一顆帝國行星；星域 5 的要塞防衛力未滿且砲仍在瞄準。
func TestRegionUpkeepMatchesReceipt(t *testing.T) {
	pre := upkeepReceiptPlanets(t, readUpkeepReceipt(t, "planets-pre.bin"))
	post := upkeepReceiptPlanets(t, readUpkeepReceipt(t, "planets-post.bin"))
	regions := readUpkeepReceipt(t, "regions-pre.bin")

	// 星域 3 只有行星 9 是非無主，所以亂數結果無論為何都必須選它；脫離
	// 後同一回合再由 3867 依 3% 回復九輪到 5433。
	record3 := regions[3*RegionRecord : 4*RegionRecord]
	rnd := NewRand(1)
	before := rnd.Seed()
	if got := DepartNeutralPlanet(record3, pre, 3, rnd); got != 9 {
		t.Fatalf("無主星域選到行星 %d，收據是 9", got)
	}
	RecoverRegionDefences(pre, 3)
	ResetRegionFortressGuns(pre, 3)
	if rnd.Seed() == before {
		t.Fatal("無主星域沒有消耗一次亂數")
	}
	if got, want := pre.Record(9), post.Record(9); got[planetOwner] != want[planetOwner] ||
		pre.PlanetEconomy(9) != post.PlanetEconomy(9) || pre.Defence(9) != post.Defence(9) {
		t.Errorf("行星 9：擁有者／經濟力／防衛力 = %d/%d/%d，收據 %d/%d/%d",
			got[planetOwner], pre.PlanetEconomy(9), pre.Defence(9),
			want[planetOwner], post.PlanetEconomy(9), post.Defence(9))
	}

	// 星域 5 的行星 13：1000 + 9×(7000×2/100) = 2260；要塞砲
	// 由狀態 3、方 0x40 重置為 0、0xFFFF。
	RecoverRegionDefences(pre, 5)
	ResetRegionFortressGuns(pre, 5)
	got, want := pre.Record(13), post.Record(13)
	if pre.Defence(13) != post.Defence(13) || got[planetGunState] != want[planetGunState] ||
		binary.LittleEndian.Uint16(got[planetGunAim:]) != binary.LittleEndian.Uint16(want[planetGunAim:]) {
		t.Errorf("行星 13：防衛力／砲狀態／瞄準 = %d/%d/%04X，收據 %d/%d/%04X",
			pre.Defence(13), got[planetGunState], binary.LittleEndian.Uint16(got[planetGunAim:]),
			post.Defence(13), want[planetGunState], binary.LittleEndian.Uint16(want[planetGunAim:]))
	}
}
