package rules

import "testing"

// TestRecomputeTaxRatesAgainstOriginalVectors 重放原版 090A:0B93（稅率重算）的行為紀錄（decomp/），
// 比對 RecomputeTaxRates。
//
//	LOGH3_DECOMP_TAXRATE=/src/workplace/decomp/diff/090A_0B93/<批次>/vec/vectors.txt \
//	  tools/go.sh test ./internal/rules -run TestRecomputeTaxRatesAgainstOriginalVectors -v
//
// 向量裡的記憶體是被突變過的任意狀態。adapter 照原版的方式從記憶體讀出星域表（ds:A79C 起 32 筆 × 126）、
// 難度基價（ds:147E 起的 dword 表，下標 level << 2）與統治者提督記錄，建出 remake 的輸入，
// 逐星域比對稅率那一格（+0x26）。只比對良構案例：統治者營運（+0x0E）不為 0（原版靠 3CC9:0070 的
// 除以零路徑得 0xFFFFFFFF，remake 的 Go 整數除法會 panic）。
//
// 已知差異（規格 126 同一輪的發現，docs/internals/67）：原版成長率 <= 0 的星域在存入前直接跳到寫入，
// 不經營運、能力與夾限；remake 的 TaxRatePerMille 兩個分支共用後半。測試把成長率 <= 0 的差異
// 與其他差異分開計數；其他差異（成長率 > 0）要求為 0 以外的部分只記進 log，不在這裡裁決。
func TestRecomputeTaxRatesAgainstOriginalVectors(t *testing.T) {
	v := loadDecompVectors(t, "LOGH3_DECOMP_TAXRATE")
	if v.funcID != "090A_0B93" {
		t.Fatalf("向量檔是 %s，不是 090A_0B93", v.funcID)
	}
	var cases, skipped, regions, same, diffNonPos, diffPos, diffOutOfDomain int
	for _, c := range v.cases {
		before := v.mem(c)
		after := v.mem(c)
		for a, b := range c.writes {
			after.b[a] = b
		}
		level := c.args[0]
		side := int(before.u8(dgroupSegForTest, 0x8F1D))
		if side > 1 {
			skipped++
			continue
		}
		rulerOff := 0xB77C + uint16(before.u8(dgroupSegForTest, 0x8F4F+uint16(side)))*28
		ruler := make([]byte, 28)
		for i := range ruler {
			ruler[i] = before.u8(dgroupSegForTest, rulerOff+uint16(i))
		}
		if ruler[admiralRule] == 0 {
			skipped++
			continue
		}
		baseOff := level<<2 + 0x147E
		base := int(before.u16(dgroupSegForTest, baseOff)) | int(before.u16(dgroupSegForTest, baseOff+2))<<16
		if base < 0 || base > 1<<24 {
			skipped++ // remake 用 int，過大的基價在原版是 32 位元回繞
			continue
		}
		records := make([]byte, 32*RegionRecord)
		for i := range records {
			records[i] = before.u8(dgroupSegForTest, 0xA79C+uint16(i))
		}
		cases++
		owner := func(i int) int { return int(records[i*RegionRecord+regionOwner]) }
		RecomputeTaxRates(records, side, base, ruler, owner)
		for i := 0; i < 32; i++ {
			rec := before.b[decompLin(dgroupSegForTest, 0xA79C+uint16(i*RegionRecord)):]
			if int(rec[regionOwner]) != side {
				continue
			}
			regions++
			want := after.u8(dgroupSegForTest, 0xA79C+uint16(i*RegionRecord+regionTaxRate))
			got := records[i*RegionRecord+regionTaxRate]
			if got == want {
				same++
				continue
			}
			switch {
			case RegionGrowth(rec) <= 0:
				diffNonPos++
			case (side == 1 && rec[regionMorale] > 200) || base > 1000:
				// 輸入域外：民心（+0x7C）上限是 200，200 減它在原版是 16 位元回繞；難度基價表只有 50 到 200。
				diffOutOfDomain++
			default:
				diffPos++
				t.Errorf("成長率 > 0、輸入域內不符：case %d 星域 %d level=%d base=%d g=%d 民心=%d 側=%d remake=%d 原版=%d",
					c.id, i, level, base, RegionGrowth(rec), rec[regionMorale], side, got, want)
			}
		}
	}
	t.Logf("案數 %d（略過 %d），比對星域 %d：相符 %d，成長率 <= 0 不符 %d，成長率 > 0 域外不符 %d，成長率 > 0 域內不符 %d",
		cases, skipped, regions, same, diffNonPos, diffOutOfDomain, diffPos)
}
