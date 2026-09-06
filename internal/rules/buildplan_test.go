package rules

import "testing"

// newPlan 是原版 796 年 1 月的開局：建造費 15,185，七列加起來剛好 100%。
func newPlan() *BuildPlan {
	return &BuildPlan{Cost: 15185, Pct: [BuildPlanRows]int{20, 10, 18, 10, 12, 15, 15}}
}

// TestAmountsMatchOriginal 釘住原版畫面上的金額。
func TestAmountsMatchOriginal(t *testing.T) {
	b := newPlan()
	want := []int{3037, 1518, 2733, 1518, 1822, 2277, 2277}
	for i, w := range want {
		if got := b.Amount(i); got != w {
			t.Fatalf("第 %d 列 %d%%：金額 %d，原版是 %d", i, b.Pct[i], got, w)
		}
	}
	if b.RestPct() != 0 {
		t.Fatalf("開局的餘額是 %d%%，原版是 0%%", b.RestPct())
	}
}

// TestArrowsMatchOriginal 走一遍實測過的操作序列：
// 右欄按兩下 20% → 16%，左欄按一下回到 18%。
func TestArrowsMatchOriginal(t *testing.T) {
	b := newPlan()
	if !b.Dec(0) || !b.Dec(0) {
		t.Fatal("減不下去")
	}
	if b.Pct[0] != 16 || b.Amount(0) != 2429 || b.RestPct() != 4 {
		t.Fatalf("減兩次後 %d%% / %d / 餘 %d%%，原版是 16%% / 2429 / 餘 4%%",
			b.Pct[0], b.Amount(0), b.RestPct())
	}
	if !b.Inc(0) {
		t.Fatal("加不上去")
	}
	if b.Pct[0] != 18 || b.Amount(0) != 2733 || b.RestPct() != 2 {
		t.Fatalf("再加一次後 %d%% / %d / 餘 %d%%，原版是 18%% / 2733 / 餘 2%%",
			b.Pct[0], b.Amount(0), b.RestPct())
	}
}

// TestIncBlockedWhenFull 釘住那個看起來像壞掉的行為：開局餘額 0%，
// 左欄箭頭按下去**不該有任何動靜**。原版就是這樣，害我以為熱區找錯了。
func TestIncBlockedWhenFull(t *testing.T) {
	b := newPlan()
	if b.Inc(0) {
		t.Fatal("餘額 0% 還加得上去")
	}
	if b.Pct[0] != 20 {
		t.Fatalf("沒加成卻改了值：%d%%", b.Pct[0])
	}
}

// TestDecBlockedAtZero：0% 不能再減。
func TestDecBlockedAtZero(t *testing.T) {
	b := &BuildPlan{Cost: 15185}
	if b.Dec(0) {
		t.Fatal("0% 還減得下去")
	}
}

// TestRestAmountMatchesOriginal 釘住餘額金額的六個實測狀態。
// 最後兩個是直接改記憶體造出來的局面（改了七列的百分比），
// 沒有它們就分不出好幾個都對得上前四點的公式（`docs/re/11`）。
func TestRestAmountMatchesOriginal(t *testing.T) {
	cases := []struct {
		pct  [BuildPlanRows]int
		want int
	}{
		{[BuildPlanRows]int{20, 10, 18, 10, 12, 15, 15}, 36},   // 餘 0%
		{[BuildPlanRows]int{18, 10, 18, 10, 12, 15, 15}, 340},  // 餘 2%
		{[BuildPlanRows]int{16, 10, 18, 10, 12, 15, 15}, 644},  // 餘 4%
		{[BuildPlanRows]int{14, 10, 18, 10, 12, 15, 15}, 948},  // 餘 6%
		{[BuildPlanRows]int{20, 10, 16, 10, 12, 15, 15}, 340},  // 餘 2%，改的是第 2 列
		{[BuildPlanRows]int{1, 1, 1, 1, 1, 1, 1}, 14178},       // 餘 93%，改記憶體造的
	}
	for _, c := range cases {
		b := &BuildPlan{Cost: 15185, Pct: c.pct}
		if got := b.RestAmount(); got != c.want {
			t.Fatalf("餘額 %d%%：%d，原版是 %d", b.RestPct(), got, c.want)
		}
	}
}
