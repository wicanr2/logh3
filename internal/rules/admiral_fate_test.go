package rules

import "testing"

func TestResolveMissingSpecialAdmiralSurvives(t *testing.T) {
	data := make([]byte, 82*AdmiralRecord)
	rec := data[:AdmiralRecord]
	rec[0] = 0x80
	rec[1] = 0
	rnd := NewRand(0) // 第一次 Next() == 0，必定通過 0x2000。
	got := ResolveMissingAdmirals(data, [2]bool{true, false}, rnd)
	if len(got) != 1 || got[0] != (AdmiralFate{Admiral: 0, Faction: 0}) {
		t.Fatalf("結果 %+v，預期提督 0 生還", got)
	}
	if rec[0] != 0xFF || rec[0x0A] != byte(CapitalRegion(0)) {
		t.Fatalf("生還記錄 status=%02X region=%02X", rec[0], rec[0x0A])
	}
}

func TestResolveMissingAdmiralDeathCleanup(t *testing.T) {
	data := make([]byte, 82*AdmiralRecord)
	rec := data[AdmiralRecord : 2*AdmiralRecord]
	rec[0] = 0x80
	rec[1] = 0
	rec[2], rec[3], rec[4], rec[5] = 1, 2, 3, 4
	rec[9] = 0x12
	rec[0x0B] = 0xFF
	// 原版 LCG：第一次 117 < 0x0444；第二次 2817 % 255 == 12 <= 45。
	got := ResolveMissingAdmirals(data, [2]bool{true, false}, NewRand(0x60F))
	if len(got) != 1 || got[0] != (AdmiralFate{Admiral: 1, Faction: 0, Dead: true}) {
		t.Fatalf("結果 %+v，預期提督 1 死亡", got)
	}
	if rec[0] != 0 || rec[9] != 0x92 {
		t.Fatalf("死亡記錄 status=%02X flags=%02X", rec[0], rec[9])
	}
	for o := 2; o < 6; o++ {
		if rec[o] != 0 {
			t.Fatalf("死亡記錄 +%02X=%02X，預期清零", o, rec[o])
		}
	}
}

func TestResolveMissingAdmiralKeepsUnresolved(t *testing.T) {
	data := make([]byte, 2*AdmiralRecord)
	// seed 1 的第一次 Next() 是 16838，不通過一般提督門檻。
	data[AdmiralRecord] = 0x80
	got := ResolveMissingAdmirals(data, [2]bool{true, false}, NewRand(1))
	if len(got) != 0 || data[AdmiralRecord] != 0x80 {
		t.Fatalf("未通過者被改動：結果 %+v status=%02X", got, data[AdmiralRecord])
	}
}
