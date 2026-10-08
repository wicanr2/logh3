package rules

import "testing"

// TestStrategicAttackAgainstOriginalVectors 重放原版 203A:0424 的行為紀錄（decomp/），驗證 strategicAttack。
//
//	LOGH3_DECOMP_FLEET_ATTACK=/src/workplace/decomp/diff/203A_0424/<批次>/vec/vectors.txt \
//	  tools/go.sh test ./internal/rules -run TestStrategicAttackAgainstOriginalVectors -v
//
// 向量裡的記憶體是被突變過的任意狀態：槽數可以大到 255、艦艇池索引可以指到池外。
// adapter 照原版的方式逐槽讀記憶體（位移 16 位元回繞）建出 remake 的型別，所以 remake 看到的
// 輸入與原版完全相同。
//
// 原版有 16 位元回繞（艘數合計）與截斷（第一次除法的商只留低 16 位元），remake 用 int，
// 兩者只可能在溢位區不同：艘數合計 >= 65536，或 艘數合計 × 能力 ÷ 100 >= 65536。
// 測試要求溢位區以外一筆都不能不符；溢位區內的差異數量記進 log。
// 溢位區是否可達另案處理，見 docs/re/443。
func TestStrategicAttackAgainstOriginalVectors(t *testing.T) {
	v := loadDecompVectors(t, "LOGH3_DECOMP_FLEET_ATTACK")
	if v.funcID != "203A_0424" {
		t.Fatalf("向量檔是 %s，不是 203A_0424", v.funcID)
	}
	var match, inZone, inZoneDiff int
	for _, c := range v.cases {
		m := v.mem(c)
		fleetOff, fleetSeg := c.args[0], c.args[1]
		poolOff, poolSeg := m.u16(dgroupSegForTest, 0xCFB4), m.u16(dgroupSegForTest, 0xCFB6)

		count := int(m.u8(fleetSeg, fleetOff+0x0F))
		sf := StrategicFleet{Fleet: &Fleet{}, Ability40: int(m.u8(fleetSeg, fleetOff+0x40))}
		trueMain := 0
		slot := fleetOff + 0x10
		for i := 0; i < count; i++ {
			idx := m.u16(fleetSeg, slot)
			ship := poolOff + idx*78
			g := FleetGroup{Kind: int(m.u8(poolSeg, ship+4)), Ships: int(m.u16(poolSeg, ship+8))}
			sf.Fleet.Composition = append(sf.Fleet.Composition, g)
			if g.Kind == 0 || g.Kind == 1 {
				trueMain += g.Ships
			}
			slot += 4
		}

		got := strategicAttack(sf)
		want := int(c.ret & 0xFFFF)
		zone := trueMain >= 65536 || trueMain*sf.Ability40/100 >= 65536
		if zone {
			inZone++
		}
		if got == want {
			match++
			continue
		}
		if !zone {
			t.Errorf("case %d 溢位區以外不符：remake=%d 原版=%d 槽數=%d 艘數合計=%d 能力=%d",
				c.id, got, want, count, trueMain, sf.Ability40)
			continue
		}
		inZoneDiff++
	}
	t.Logf("案數 %d，相符 %d，溢位區案數 %d（其中不符 %d）", len(v.cases), match, inZone, inZoneDiff)
}

// dgroupSegForTest 是 dosgolem 載入位置（映像在 0349）下的 DGROUP 段。
const dgroupSegForTest = 0x54B3
