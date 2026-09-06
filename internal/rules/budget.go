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

	// ShipPct 是切給船艦建造費的百分比。
	ShipPct int
}

// Military 是預估軍事費。
func (b *Budget) Military() int { return b.Treasury + b.Tax }

// ShipCost 是船艦建造費，無條件捨去。
func (b *Budget) ShipCost() int { return b.Military() * b.ShipPct / 100 }

// Rest 是剩餘軍事費。
func (b *Budget) Rest() int { return b.Military() - b.ShipCost() }
