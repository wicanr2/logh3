package rules

import "testing"

// TestBudgetMatchesOriginal 釘住原版 796 年 1 月預算計劃畫面上的四個數字。
func TestBudgetMatchesOriginal(t *testing.T) {
	b := &Budget{Treasury: 50000, Tax: 10741, ShipPct: 25}
	for _, c := range []struct {
		name string
		got  int
		want int
	}{
		{"預估軍事費", b.Military(), 60741},
		{"船艦建造費", b.ShipCost(), 15185},
		{"剩餘軍事費", b.Rest(), 45556},
	} {
		if c.got != c.want {
			t.Errorf("%s：%d，原版是 %d", c.name, c.got, c.want)
		}
	}
	// 建造計劃那一頁的建造費就是這個值。
	if (&BuildPlan{Cost: b.ShipCost()}).Cost != 15185 {
		t.Fatal("建造費沒接起來")
	}
}

// TestBudgetAtOtherPercent 用第二個獨立來源釘住同一組公式：
// 把造船比例往上調一格之後，原版的「796 年 1 月份會計報告」印出
//
//	軍事費餘額 50000／稅收 10741／合計收入 60741／船艦建造費 15792／總計 44949
//
// 這一組數字與建造計劃那一頁無關，是另一個畫面自己算出來的
// （收據 `account1`，`docs/re/13`）。
func TestBudgetAtOtherPercent(t *testing.T) {
	b := &Budget{Treasury: 50000, Tax: 10741, ShipPct: 26}
	for _, c := range []struct {
		name string
		got  int
		want int
	}{
		{"合計收入", b.Military(), 60741},
		{"船艦建造費", b.ShipCost(), 15792},
		{"總計", b.Rest(), 44949},
	} {
		if c.got != c.want {
			t.Errorf("%s：%d，原版是 %d", c.name, c.got, c.want)
		}
	}
}

// TestBudgetAlliance 第三組獨立資料：選「同盟」開局，內政專任參謀
// 邱吾權上校的會計報告印出
//
//	軍事費餘額 50000／稅收 11087／合計收入 61087／船艦建造費 15271／總計 45816
//
// 稅收與帝國的 10741 不同，所以這一組驗的是同一條公式吃不同的輸入
// （收據 `alliance1`，`docs/re/15`）。
func TestBudgetAlliance(t *testing.T) {
	b := &Budget{Treasury: 50000, Tax: 11087, ShipPct: 25}
	for _, c := range []struct {
		name string
		got  int
		want int
	}{
		{"合計收入", b.Military(), 61087},
		{"船艦建造費", b.ShipCost(), 15271},
		{"總計", b.Rest(), 45816},
	} {
		if c.got != c.want {
			t.Errorf("%s：%d，原版是 %d", c.name, c.got, c.want)
		}
	}
}

// TestShipCostMatchesS3Receipt 釘住收據 `monthmoney` 的兩組數字
// （`docs/re/256`）：S3「巴米利恩會戰」開局，把座位 0 poke 成電腦讓月底
// 本體無頭跑完，前後各倒一份 1 MB。
//
//	ds:8F92 ＝ 00 14 → 帝國 0%、同盟 20%
//	帝國 160000 ＋ 13131 − 2342 − 665 ＝ 170124，×0%  ＝ 0
//	同盟  30000 ＋  9748 −  951 − 1133 ＝  37664，×20% ＝ 7532 ＝ ds:8F5A
//
// 舊的算式（不扣維持費與移動經費）在同盟那組會給 39748 × 20% ＝ 7949。
func TestShipCostMatchesS3Receipt(t *testing.T) {
	for _, c := range []struct {
		name string
		b    Budget
		want int
	}{
		{"帝國", Budget{Treasury: 160000, Tax: 13131, Upkeep: 2342, MoveCost: 665, ShipPct: 0}, 0},
		{"同盟", Budget{Treasury: 30000, Tax: 9748, Upkeep: 951, MoveCost: 1133, ShipPct: 20}, 7532},
	} {
		if got := c.b.ShipCost(); got != c.want {
			t.Errorf("%s 的船艦建造費 ＝ %d，收據是 %d", c.name, got, c.want)
		}
	}
	// 基數是負的要夾成 0（原版 `090A:0E88` 的 `jns` 那一路）。
	neg := Budget{Treasury: 100, Tax: 0, Upkeep: 500, MoveCost: 0, ShipPct: 25}
	if got := neg.ShipCost(); got != 0 {
		t.Errorf("基數為負時的建造費 ＝ %d，預期 0", got)
	}
	// 預估軍事費不扣那兩項——畫面上印的是這個數。
	if got, want := neg.Military(), 100; got != want {
		t.Errorf("預估軍事費 ＝ %d，預期 %d", got, want)
	}
}
