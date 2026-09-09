package ai

import (
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// TestBuildGate 測 `2D76:0060` 的三道關卡（`docs/re/202`）。
func TestBuildGate(t *testing.T) {
	for _, c := range []struct {
		name                  string
		existing, tax, upkeep int
		tight                 bool
		wantN                 int
		wantOK                bool
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

// TestPickBuildAdmiral 測 `2D76:02FA` 的四層（`docs/re/203`）。
func TestPickBuildAdmiral(t *testing.T) {
	// 四個人，等級都夠。
	high := func(int) int { return 20 }
	for _, c := range []struct{ r, want int }{
		{0x7FFF, 3}, {0x7000, 3},
		{0x6FFF, 2}, {0x6000, 2},
		{0x5FFF, 1}, {0x4000, 1},
		{0x3FFF, 0}, {0, 0},
	} {
		if got := PickBuildAdmiral(c.r, 4, high); got != c.want {
			t.Errorf("r ＝ %04X 挑第 %d 名，預期第 %d 名", c.r, got, c.want)
		}
	}

	// 人數不夠就往下一層試，不是直接落到第一名。
	if got := PickBuildAdmiral(0x7FFF, 2, high); got != 1 {
		t.Errorf("只有兩個人時 r ＝ 7FFF 挑第 %d 名，預期第 1 名", got)
	}
	// 等級不夠也一樣往下試。
	lowAt3 := func(k int) int {
		if k == 3 {
			return 15
		}
		return 20
	}
	if got := PickBuildAdmiral(0x7FFF, 4, lowAt3); got != 2 {
		t.Errorf("第四名等級 15 時挑第 %d 名，預期第 2 名", got)
	}
	// 全部都不夠格 → 第一名。
	low := func(int) int { return 15 }
	if got := PickBuildAdmiral(0x7FFF, 4, low); got != 0 {
		t.Errorf("都不夠格時挑第 %d 名，預期第 0 名", got)
	}
	if got := PickBuildAdmiral(0x1000, 0, high); got != -1 {
		t.Errorf("空清單應該回 -1，得到 %d", got)
	}
}

// TestPickBuildAdmiralDistribution 用原版的亂數跑一輪，確認四個名次的
// 比例是 1/2、1/4、1/8、1/8（`docs/re/203`）。
func TestPickBuildAdmiralDistribution(t *testing.T) {
	r := rulesRand(t)
	var hit [4]int
	const n = 80000
	for i := 0; i < n; i++ {
		hit[PickBuildAdmiral(r.Next(), 4, func(int) int { return 20 })]++
	}
	want := [4]float64{0.5, 0.25, 0.125, 0.125}
	for k, w := range want {
		got := float64(hit[k]) / n
		if got < w-0.01 || got > w+0.01 {
			t.Errorf("第 %d 名佔 %.4f，預期約 %.3f", k, got, w)
		}
	}
}

// rulesRand 給分布測試一個原版的亂數產生器。
func rulesRand(t *testing.T) *rules.Rand {
	t.Helper()
	return rules.NewRand(0xD3DC167E)
}

// TestClassKeysAreInverse 對 `ds:5F62`／`5F70`／`5F6C`／`5F74` 與
// `ds:5F52`（`rules.FleetClass`）互為反表（`docs/re/204`）。
func TestClassKeysAreInverse(t *testing.T) {
	seen := map[int]bool{}
	for class := 0; class < 4; class++ {
		for _, k := range ClassKeys(class) {
			if got := rules.FleetClass(k); got != class {
				t.Errorf("鍵 %X 在類別 %d 的表裡，但 FleetClass 回 %d", k, class, got)
			}
			if seen[k] {
				t.Errorf("鍵 %X 出現在兩張表裡", k)
			}
			seen[k] = true
		}
	}
	// 16 個鍵一個不漏。
	if len(seen) != 16 {
		t.Errorf("四張表合起來 %d 個鍵，預期 16", len(seen))
	}
}

// TestPickClassKey 測 `2D76:0576` 的取模抽法。
func TestPickClassKey(t *testing.T) {
	// 類別 1 只有 0A／0B 兩個鍵，取模輪流。
	if got := PickClassKey(1, 0); got != 0x0A {
		t.Errorf("r ＝ 0 抽到 %X，預期 0A", got)
	}
	if got := PickClassKey(1, 1); got != 0x0B {
		t.Errorf("r ＝ 1 抽到 %X，預期 0B", got)
	}
	if got := PickClassKey(1, 2); got != 0x0A {
		t.Errorf("r ＝ 2 抽到 %X，預期 0A（取模繞回）", got)
	}
	// 類別 0 有九個鍵。
	if got := PickClassKey(0, 5); got != 0x06 {
		t.Errorf("類別 0、r ＝ 5 抽到 %X，預期 06", got)
	}
	if got := PickClassKey(4, 0); got != -1 {
		t.Errorf("不合法的類別應該回 -1，得到 %d", got)
	}
}
