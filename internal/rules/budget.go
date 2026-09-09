package rules

// Budget 是次月份預算計劃：錢從哪來、切多少給造船。
//
// 三條關係都直接對得上原版畫面上的數字（`docs/re/12`）：
//
//	預估軍事費 ＝ 國庫 ＋ 預估徵稅額      50,000 ＋ 10,741 ＝ 60,741
//	船艦建造費 ＝ 預估軍事費 × 百分比      60,741 × 25% ＝ 15,185.25 → 15,185
//	剩餘軍事費 ＝ 預估軍事費 − 船艦建造費  60,741 − 15,185 ＝ 45,556
//
// 建造計劃畫面的「船艦建造費」就是這裡算出來的，不是外生常數。
type Budget struct {
	// Treasury 是國庫（畫面上沒直接印出來，記憶體在線性 5DA82）。
	Treasury int

	// Tax 是預估徵稅額。
	Tax int

	// ShipPct 是切給船艦建造費的百分比 ＝ 原版 `ds:8F92[陣營]`
	// （每個陣營一個 byte）。
	ShipPct int

	// Upkeep 是本回合的艦隊維持費（`ds:8F72[陣營]`）、MoveCost 是本回合的
	// 移動經費（`ds:8F82[陣營]`）。**兩項都只影響船艦建造費的基數**，
	// 不影響畫面上的預估軍事費（`docs/re/256`）。
	Upkeep, MoveCost int
}

// Military 是預估軍事費。畫面上印的就是這個數，**不扣維持費與移動經費**。
func (b *Budget) Military() int { return b.Treasury + b.Tax }

// shipBase 是船艦建造費的基數（原版 `090A:0E88`）：
//
//	軍事費 ＋ 總稅收 − 本回合維持費 − 本回合移動經費，負數夾成 0
func (b *Budget) shipBase() int {
	v := b.Military() - b.Upkeep - b.MoveCost
	if v < 0 {
		return 0
	}
	return v
}

// ShipCost 是船艦建造費，無條件捨去。
//
// 原版在基數超過 `0x0147AE14`（＝ 2³¹ ÷ 100）時先除再乘，避開 32 位元
// 溢位；這裡照抄那條分支（`docs/re/256`）。
func (b *Budget) ShipCost() int {
	v := b.shipBase()
	if v >= 0x0147AE14 {
		return v / 100 * b.ShipPct
	}
	return v * b.ShipPct / 100
}

// Rest 是剩餘軍事費。
func (b *Budget) Rest() int { return b.Military() - b.ShipCost() }
