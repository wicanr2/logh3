package rules

import "testing"

// TestResetRepairShipsAgainstOriginalVectors 重放原版 1AAB:1286 的行為紀錄（decomp/），
// 比對 ResetRepairShips。
//
//	LOGH3_DECOMP_REPAIR_RESET=/src/workplace/decomp/diff/1AAB_1286/<批次>/vec/vectors.txt \
//	  tools/go.sh test ./internal/rules -run TestResetRepairShipsAgainstOriginalVectors -v
//
// 原版 1AAB:1286 有兩個階段（規格 126、docs/internals/67）：第 1 階段把修護艦的存量重設成現有艘數，
// 第 2 階段用這個存量補同艦隊其他艦艇的缺額。remake 目前只有第 1 階段，
// 所以有缺額的案例會不符。測試不把這件事當失敗，而是分開計數並寫進 log；
// 第 2 階段實作之後，把 strict 改成 true 就會變成嚴格比對。
//
// 只比對良構案例：編制組數不超過 12、池索引互異且小於 352。向量裡被突變成任意狀態的記憶體
// （組數到 255、池索引指到池外、別名）原版會讀寫池以外的記憶體，remake 的型別表達不了，不納入。
func TestResetRepairShipsAgainstOriginalVectors(t *testing.T) {
	const strict = false

	v := loadDecompVectors(t, "LOGH3_DECOMP_REPAIR_RESET")
	if v.funcID != "1AAB_1286" {
		t.Fatalf("向量檔是 %s，不是 1AAB_1286", v.funcID)
	}
	var wellFormed, skipped, same, stage2Diff, other int
	for _, c := range v.cases {
		before := v.mem(c)
		after := v.mem(c)
		for a, b := range c.writes {
			after.b[a] = b
		}
		fleetOff, fleetSeg := c.args[0], c.args[1]
		poolOff, poolSeg := before.u16(dgroupSegForTest, 0xCFB4), before.u16(dgroupSegForTest, 0xCFB6)

		count := int(before.u8(fleetSeg, fleetOff+0x0F))
		if count > 12 {
			skipped++
			continue
		}
		seen := map[uint16]bool{}
		ok := true
		ships := make([]uint16, count) // 每組在池裡的位移
		for i := 0; i < count; i++ {
			idx := before.u16(fleetSeg, fleetOff+0x10+uint16(4*i))
			if idx >= 352 || seen[idx] {
				ok = false
				break
			}
			seen[idx] = true
			ships[i] = poolOff + idx*78
		}
		if !ok {
			skipped++
			continue
		}
		stock := func(m *decompMem, off uint16) int {
			return int(m.u16(poolSeg, off+0x0E)) | int(m.u16(poolSeg, off+0x10))<<16
		}
		f := &Fleet{Region: 7}
		for _, off := range ships {
			f.Composition = append(f.Composition, FleetGroup{
				Kind:      int(before.u8(poolSeg, off+4)),
				Ships:     int(before.u16(poolSeg, off+8)),
				RepairCap: int(before.u16(poolSeg, off+0xC)),
				Missiles:  stock(before, off),
			})
		}
		ResetRepairShips([]*Fleet{f}, 7)

		wellFormed++
		match, stage2 := true, false
		for i, off := range ships {
			g := f.Composition[i]
			if g.Ships != int(after.u16(poolSeg, off+8)) || g.Missiles != stock(after, off) {
				match = false
			}
			if g.Kind != resupplyRepairKind && after.u16(poolSeg, off+8) != before.u16(poolSeg, off+8) {
				stage2 = true // 原版補了其他艦艇的艘數
			}
		}
		switch {
		case match:
			same++
		case stage2:
			stage2Diff++
		default:
			other++
		}
	}
	t.Logf("案數 %d，良構 %d（略過 %d）；與原版完全相符 %d，原版多做第 2 階段而不符 %d，其他不符 %d",
		len(v.cases), wellFormed, skipped, same, stage2Diff, other)
	if other > 0 {
		t.Errorf("有 %d 案不符且原版沒有補其他艦艇的艘數（不是第 2 階段造成的差異）", other)
	}
	if strict && wellFormed != same {
		t.Errorf("良構案例中有 %d 案與原版不符", wellFormed-same)
	}
}
