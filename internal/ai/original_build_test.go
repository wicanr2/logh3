package ai

import "testing"

// TestBuildGate 測 `2D76:0060` 的三道關卡（`docs/re/202`）。
func TestBuildGate(t *testing.T) {
	for _, c := range []struct {
		name                   string
		existing, tax, upkeep  int
		tight                  bool
		wantN                  int
		wantOK                 bool
	}{
		{"滿 16 支不建", 16, 100000, 0, false, 0, false},
		{"少於三支不看錢", 2, 0, 999999, false, 2, true},
		{"維持費還低 → 建", 5, 10000, 499, false, 5, true},
		{"維持費 × 20 剛好追上 → 不建", 5, 10000, 500, false, 0, false},
		{"打七五折之後追上 → 不建", 5, 10000, 380, true, 0, false},
		{"打七五折仍未追上 → 建", 5, 10000, 374, true, 5, true},
	} {
		n, ok := BuildGate(c.existing, c.tax, c.upkeep, c.tight)
		if n != c.wantN || ok != c.wantOK {
			t.Errorf("%s：得到 (%d, %v)，預期 (%d, %v)", c.name, n, ok, c.wantN, c.wantOK)
		}
	}
}

// TestClassQuotaOrder 測配額差的升冪排序（`2D76:01A5`／`2D76:0251`）。
func TestClassQuotaOrder(t *testing.T) {
	// 十支艦隊 → 應有 4／4／0／1（10×40/100、10×45/100 捨去、10×5/100 捨去、10×10/100）。
	// 現有 4／0／0／1 → 差 0／−4／0／0，最缺的是類別 1。
	if got := ClassQuotaOrder([4]int{4, 0, 0, 1}, 10); got[0] != 1 {
		t.Errorf("最缺的是類別 %d，預期 1（順序 %v）", got[0], got)
	}
	// 全部剛好 → 差全 0 → 穩定排序保持 0 1 2 3。
	if got := ClassQuotaOrder([4]int{4, 4, 0, 1}, 10); got != [4]int{0, 1, 2, 3} {
		t.Errorf("差全 0 時順序 ＝ %v，預期 0 1 2 3", got)
	}
	// 配額是整數除法捨去：五支時應有 2／2／0／0。
	if got := ClassQuotaOrder([4]int{2, 1, 1, 1}, 5); got[0] != 1 {
		t.Errorf("五支時最缺的是類別 %d，預期 1（順序 %v）", got[0], got)
	}
	// 四項相加 40+45+5+10 ＝ 100。
	sum := 0
	for _, p := range ClassQuota {
		sum += p
	}
	if sum != 100 {
		t.Errorf("配額相加 ＝ %d，預期 100", sum)
	}
}
