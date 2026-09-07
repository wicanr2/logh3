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

// TestArrowsMatchOriginal 走一遍實測過的操作序列。箭頭一次 1%。
func TestArrowsMatchOriginal(t *testing.T) {
	b := newPlan()
	for i := 0; i < 4; i++ {
		if !b.Dec(0) {
			t.Fatalf("第 %d 次減不下去", i+1)
		}
	}
	if b.Pct[0] != 16 || b.Amount(0) != 2429 || b.RestPct() != 4 || b.RestAmount() != 644 {
		t.Fatalf("減四次後 %d%% / %d / 餘 %d%% %d，原版是 16%% / 2429 / 餘 4%% 644",
			b.Pct[0], b.Amount(0), b.RestPct(), b.RestAmount())
	}
	if !b.Inc(0) || !b.Inc(0) {
		t.Fatal("加不上去")
	}
	if b.Pct[0] != 18 || b.Amount(0) != 2733 || b.RestPct() != 2 || b.RestAmount() != 340 {
		t.Fatalf("再加兩次後 %d%% / %d / 餘 %d%% %d，原版是 18%% / 2733 / 餘 2%% 340",
			b.Pct[0], b.Amount(0), b.RestPct(), b.RestAmount())
	}
}

// TestIncBlockedWhenFull 釘住那個看起來像壞掉的行為：開局餘額 0%，
// 左欄箭頭按下去**不該有任何動靜**。
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

// TestRestAmountMatchesOriginal 釘住餘額金額。
//
// 後面那一批是直接把百分比陣列寫進原版記憶體造出來的局面
// （`docs/re/12`）。手點只能造出餘額 0–6%，那個範圍裡好幾組公式
// 都對得上；單列獨佔預算的局面才分得出來。
func TestRestAmountMatchesOriginal(t *testing.T) {
	cases := []struct {
		name string
		pct  [BuildPlanRows]int
		want int
	}{
		{"開局", [BuildPlanRows]int{20, 10, 18, 10, 12, 15, 15}, 36},
		{"餘 1%", [BuildPlanRows]int{19, 10, 18, 10, 12, 15, 15}, 188},
		{"餘 2%", [BuildPlanRows]int{18, 10, 18, 10, 12, 15, 15}, 340},
		{"餘 4%", [BuildPlanRows]int{16, 10, 18, 10, 12, 15, 15}, 644},
		{"餘 6%", [BuildPlanRows]int{14, 10, 18, 10, 12, 15, 15}, 948},
		{"改第 2 列", [BuildPlanRows]int{20, 10, 16, 10, 12, 15, 15}, 340},
		{"改第 6 列", [BuildPlanRows]int{20, 10, 18, 10, 12, 15, 12}, 498},
		{"兩列都改", [BuildPlanRows]int{17, 10, 18, 10, 12, 15, 12}, 954},
		{"全部 1%", [BuildPlanRows]int{1, 1, 1, 1, 1, 1, 1}, 14178},
		{"全部 0%", [BuildPlanRows]int{}, 15185},
		{"第 0 列吃滿", [BuildPlanRows]int{100}, 1},
	}
	for _, c := range cases {
		b := &BuildPlan{Cost: 15185, Pct: c.pct}
		if got := b.RestAmount(); got != c.want {
			t.Errorf("%s（餘 %d%%）：%d，原版是 %d", c.name, b.RestPct(), got, c.want)
		}
	}
}

// TestUnitPrices 用「只給一列預算」的實測局面釘住七個單價。
func TestUnitPrices(t *testing.T) {
	// 一列獨佔 15% 與 20% 時的餘額，逐列量的。
	at15 := [BuildPlanRows]int{12913, 12917, 12909, 12929, 12909, 12910, 12919}
	at20 := [BuildPlanRows]int{12153, 12149, 12149, 12161, 12149, 12154, 12149}
	for i := 0; i < BuildPlanRows; i++ {
		for _, c := range []struct{ pct, want int }{{15, at15[i]}, {20, at20[i]}} {
			b := &BuildPlan{Cost: 15185}
			b.Pct[i] = c.pct
			if got := b.RestAmount(); got != c.want {
				t.Errorf("只有第 %d 列 %d%%：餘額 %d，原版是 %d", i, c.pct, got, c.want)
			}
		}
	}
}

// TestPaidMatchesOracle 釘住「已支付 船艦建造費」那一格。
//
// 796 年 1 月：預估軍事費 60741 的 25% ＝ 建造費 15185，照開局的百分比
// 分下去，各列只買得起整數艘，實際花掉 **15149**——與畫面相同
// （`docs/internals/27`、`docs/spec/003`）。零頭 36 退回。
func TestPaidMatchesOracle(t *testing.T) {
	b := BuildPlan{Cost: 15185, Pct: OpeningPct}
	if got := b.Paid(); got != 15149 {
		t.Errorf("實際花掉 %d，實測 15149", got)
	}
	if got := b.RestAmount(); got != 36 {
		t.Errorf("退回 %d，預期 36（15185 − 15149）", got)
	}
}
