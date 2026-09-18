package rules

// 戰略階段的耗損與全滅（`0374:0608(星域)`，#57、`docs/re/336` §九）。
//
// 逐格處理（`0374:03E8`）裡這回合沒有戰鬥的格子才跑：停在這一格、陣營不是擁有者（byte 比，無主 2 也算）、
// `+0x00` bit0（士氣低於 50）的艦隊，每組扣兵（`0374:06BC`），扣完只看第 0 組的艘數——0 就是全滅
// （`0374:073F` 報告、艦艇作廢、`297B:034A`），否則 `0374:0896` 把 0 艘的組移掉。整條沒有亂數。

// StrategicAttrition ＝ `0374:06BC(f)`：現有每一組扣 `min(滿編 ÷ 6, 艘數)`（16 位元無號），交給
// `203A:0593(艦艇, n, NULL)`——扣艘數、對手陣營兩個計分器各加 `n ÷ 20`（每組各自除）、每組扣完重算士氣
// （`1261:0067(f, 0)`，不印）；艦種 7 不扣也不重算。回兩個陣營計分器各加多少，與第 0 組是不是 0 艘
// （`0374:068B`：只看第 0 組）。
func StrategicAttrition(f *Fleet, command int) (score [2]int, annihilated bool) {
	res := &StrategicAssaultResult{}
	sf := StrategicFleet{Fleet: f, Command: command}
	for gi := range f.Composition {
		n := f.Composition[gi].Full / 6
		if n > f.Composition[gi].Ships {
			n = f.Composition[gi].Ships
		}
		strategicShipLoss(sf, gi, n, f.Faction, res)
	}
	return res.Score, len(f.Composition) > 0 && f.Composition[0].Ships == 0
}

// RemoveEmptyGroups ＝ `0374:0896` 的組迴圈：艘數 0 的組作廢（`2AD2:01EF`，不還存量）並移除（`0374:0A70`），
// 其餘往前遞補。語意取自解包後的原碼（與戰場那一支 `185C:0177` 同形）；dosgolem 上那 16 bytes 被遊戲自己的
// 鍵盤常式寫壞（`docs/re/336` §九），執行期不能當對照。`0896` 結尾的 `297B:0546`（能力與速度重算）沒有接。
func RemoveEmptyGroups(f *Fleet) {
	kept := f.Composition[:0]
	for _, g := range f.Composition {
		if g.Ships != 0 {
			kept = append(kept, g)
		}
	}
	f.Composition = kept
}

// MarkAdmiralMissing ＝ `28E0:057F(提督, 0)`：`+0x09 |= 0x80`、`+0x00 ← 0x80`（失蹤）、功績 `+0x02..+0x05 ← 0`。
// 全滅那一支的提督與副官都走這一條（`297B:034A`）；失蹤之後由回合開頭的結局判定決定生還或死亡（`admiral_fate.go`）。
func MarkAdmiralMissing(rec []byte) {
	if len(rec) < 0x0A {
		return
	}
	rec[0x09] |= 0x80
	rec[admiralStatus] = 0x80
	for o := 0x02; o < 0x06; o++ {
		rec[o] = 0
	}
}

// AnnihilateSlotFleet 把存檔表裡第 i 支標成全滅——`297B:034A` 與 `185C:00F0`／`0374:0815`
// 真正寫進表裡的那幾個 byte（`docs/re/339` §九 1 MB 逐位元組對過）：
//
//	艦隊記錄 +0x04（提督）與 +0x05（副官） ← FF。`+0x04` 同時就是「這一格沒在用」，
//	    所以整筆不必填 FF；原版也沒有動 +0x01（陣營）、+0x0F（組數）。
//	編制那幾組與 +0x38 的索敵分遣隊，艦艇記錄 +0x01（擁有者） ← FF ＝ 作廢（`2AD2:01EF`）。
//
// 提督的失蹤旗標走 MarkAdmiralMissing、星域的艦隊數由呼叫端重數。
func (s *Slot) AnnihilateFleet(i int) {
	r := s.fleet(i)
	if r == nil {
		return
	}
	void := func(ship int) {
		if ship == 0xFFFF {
			return
		}
		if rec := s.ShipAt(ship); len(rec) > 1 {
			rec[1] = 0xFF
		}
	}
	for g := 0; g < fleetSlotN; g++ {
		o := fleetSlots + g*4
		void(int(r[o]) | int(r[o+1])<<8)
	}
	void(int(r[fleetScout]) | int(r[fleetScout+1])<<8)
	r[fleetUnused] = 0xFF
	r[fleetAdjutant] = 0xFF
}
