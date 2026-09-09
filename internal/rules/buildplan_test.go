package rules

import "testing"

// newPlan 是原版 796 年 1 月的開局：建造費 15,185，七列加起來剛好 100%。
func newPlan() *BuildPlan {
	return &BuildPlan{Cost: 15185, Pct: [BuildPlanRows]int{20, 10, 18, 10, 12, 15, 15},
		StaffOps: EmpireStaffOps796}
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
		b := &BuildPlan{Cost: 15185, Pct: c.pct, StaffOps: EmpireStaffOps796}
		if got := b.RestAmount(); got != c.want {
			t.Errorf("%s（餘 %d%%）：%d，原版是 %d", c.name, b.RestPct(), got, c.want)
		}
	}
}

// TestUnitPricesFromBaseAndStaff 釘住「基價 ×100÷ 參謀營運」這條算式：
// 796/1 的兩位軍事專任參謀（帝國提督 6 營運 112、同盟提督 79 營運 106）
// 代進去要得到畫面上量到的那十四個單價（`docs/re/265`，14/14）。
func TestUnitPricesFromBaseAndStaff(t *testing.T) {
	for _, c := range []struct {
		faction, ops int
		want         [BuildPlanRows]int
	}{
		{0, EmpireStaffOps796, [BuildPlanRows]int{8, 12, 4, 24, 4, 7, 22}},
		{1, AllianceStaffOps796, [BuildPlanRows]int{9, 14, 5, 23, 3, 7, 24}},
	} {
		for row := 0; row < BuildPlanRows; row++ {
			if got := UnitPrice(c.faction, row, c.ops); got != c.want[row] {
				t.Errorf("陣營 %d 第 %d 列（營運 %d）單價 ＝ %d，畫面量到 %d",
					c.faction, row, c.ops, got, c.want[row])
			}
		}
	}
	// 營運越高越便宜：帝國戰艦基價 10，營運 200 → 5。
	if got := UnitPrice(0, 0, 200); got != 5 {
		t.Errorf("營運 200 的帝國戰艦單價 ＝ %d，預期 5", got)
	}
	// 問不到參謀時退回 100，不是除以零。
	if got, want := UnitPrice(0, 0, 0), BasePrices[0][0]; got != want {
		t.Errorf("營運 0 的單價 ＝ %d，預期退回 %d", got, want)
	}
}

// TestUnitPricesRestAmount 用「只給一列預算」的實測局面釘住七個單價。
func TestUnitPricesRestAmount(t *testing.T) {
	// 一列獨佔 15% 與 20% 時的餘額，逐列量的。
	at15 := [BuildPlanRows]int{12913, 12917, 12909, 12929, 12909, 12910, 12919}
	at20 := [BuildPlanRows]int{12153, 12149, 12149, 12161, 12149, 12154, 12149}
	for i := 0; i < BuildPlanRows; i++ {
		for _, c := range []struct{ pct, want int }{{15, at15[i]}, {20, at20[i]}} {
			b := &BuildPlan{Cost: 15185, StaffOps: EmpireStaffOps796}
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
	b := BuildPlan{Cost: 15185, Pct: OpeningPct, StaffOps: EmpireStaffOps796}
	if got := b.Paid(); got != 15149 {
		t.Errorf("實際花掉 %d，實測 15149", got)
	}
	if got := b.RestAmount(); got != 36 {
		t.Errorf("退回 %d，預期 36（15185 − 15149）", got)
	}
}

// TestAIBuildPctMatchesDumps 釘住 `2D10:0038` 尾段寫回的那張表
// （`docs/re/257`）。同盟那一排有兩份傾印：`monthblend` 與 `s3aicomp` 的
// `ds:8FD6[7..13]` 都是 `30 25 10 05 10 10 10`。
func TestAIBuildPctMatchesDumps(t *testing.T) {
	if got, want := AIBuildPct[1], ([BuildPlanRows]int{30, 25, 10, 5, 10, 10, 10}); got != want {
		t.Errorf("同盟那一排 ＝ %v，傾印是 %v", got, want)
	}
	// 兩排都要剛好 100——原版寫的是七個立即值，加不到 100 就是讀錯了。
	for f, row := range AIBuildPct {
		sum := 0
		for _, v := range row {
			sum += v
		}
		if sum != 100 {
			t.Errorf("陣營 %d 那一排加起來 ＝ %d，預期 100（%v）", f, sum, row)
		}
	}
	sum := 0
	for _, v := range OpeningPct {
		sum += v
	}
	if sum != 100 {
		t.Errorf("開局那一排加起來 ＝ %d，預期 100", sum)
	}
}
