package game

import (
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

// 戰略階段的耗損與全滅（`0374:0608(星域)`，#57、`docs/re/336` §九、規格 054）。
//
// 逐格掃描裡這回合沒有戰鬥的格子，在打行星與被動效果之前跑：艦隊表順序逐支，停在這一格、陣營不是擁有者、
// 士氣低落（`+0x00` bit0）的扣兵；第 0 組扣到 0 艘就印「…已經全滅。」、等完才拿掉那一支（報告在
// `297B:034A` 之前），否則把 0 艘的組移掉。

// attritionFrom 從艦隊表第 i 支接著跑這一格；停在全滅報告上回 true（看完由報告接回同一個索引）。
func (g *Game) attritionFrom(region, i int, next func()) bool {
	rec := g.regionRecord(region)
	if rec == nil {
		return false
	}
	owner := int(rec[regionOwnerOff])
	for ; i < len(g.fleets); i++ {
		f := &g.fleets[i]
		if f.Region != region || f.Faction == owner || f.Flags&rules.FleetFlagLowMorale == 0 {
			continue
		}
		adm := g.admiralRecord(f.Admiral)
		command := 0
		if len(adm) > 0x0B {
			command = int(adm[0x0B])
		}
		score, gone := rules.StrategicAttrition(f, command)
		for side := 0; side < 2; side++ {
			g.battleScore[side] += score[side]
			g.campaignScore[side] += score[side]
		}
		if !gone {
			rules.RemoveEmptyGroups(f)
			continue
		}
		idx := i
		rep := g.annihilationReport(f)
		if g.startTurnReports([]turnReport{rep}, func() {
			g.annihilateFleet(idx)
			if !g.attritionFrom(region, idx, next) {
				next()
			}
		}) {
			return true
		}
		g.annihilateFleet(idx)
		i--
	}
	return false
}

// annihilationReport ＝ `0374:073F` 的報告：「<陣營>第Ｎ艦隊／提督名軍階／已經全滅。」＋提督頭像。
// 原版格式字串開頭多一個 `\x03`、畫成花紋；使用者裁決修成正常文字（規格 059 刻意差異）。
func (g *Game) annihilationReport(f *rules.Fleet) turnReport {
	p := g.res.TurnWindow
	if p == nil {
		return turnReport{face: f.Admiral}
	}
	adm := g.admiralRecord(f.Admiral)
	var b []byte
	b = append(b, ui.Sprintf(p.ReportFormat("annihilated_prefix"), g.factionNameColoured(f.Faction))...)
	b = append(b, p.FleetNumber(f.Number)...)
	b = append(b, ui.Sprintf(p.ReportFormat("annihilated_body"),
		codesBig5(g.admiralNameCodes(f.Admiral)),
		codesBig5(g.rankCodes(rules.RankOf(byteAt(adm, admiralMeritOff)))))...)
	b = append(b, p.ReportFormat("annihilated_tail")...)
	return turnReport{text: b, face: f.Admiral}
}

// annihilateFleet ＝ `0374:0815`–`0882`：艦艇作廢（不還存量）＋ `297B:034A`：提督與副官失蹤（`28E0:057F(…, 0)`）、
// 艦隊記錄清掉、星域 `+0x28` 與 `ds:8F49` 減一（remake 拿掉那一筆再重數）、`090A:0E6D` 重算。
func (g *Game) annihilateFleet(i int) {
	if i < 0 || i >= len(g.fleets) {
		return
	}
	f := g.fleets[i]
	g.mutableAdmirals()
	rules.MarkAdmiralMissing(g.admiralRecord(f.Admiral))
	if f.Adjutant >= 0 {
		rules.MarkAdmiralMissing(g.admiralRecord(f.Adjutant))
	}
	// 存檔是從 `g.slot` 寫出去的（`SaveSlot`），只改執行期切片的話那一支在存檔裡
	// 還活著，讀回來就復活（#78）。走的是戰場那一條共用的出口
	// （`rules.(*Slot).AnnihilateFleet`：`+0x04`／`+0x05` ← FF、編制與索敵那幾筆
	// 艦艇 `+0x01` ← FF）。
	if g.slot != nil {
		g.slot.AnnihilateFleet(f.Index)
	}
	g.fleets = append(g.fleets[:i], g.fleets[i+1:]...)
	if len(g.res.Stellar) > 0 {
		rules.RecountRegionFleets(g.res.Stellar, g.fleets)
	}
	g.recomputeUnrest(g.turn)
}
