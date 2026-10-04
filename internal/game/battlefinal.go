package game

import (
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

// 戰場一個回合的最後階段：原版第 ⑧ 步 `181B:03F0` → `185C:0001` 的子步 8-1
// （`185C:001D`，規格 062、`docs/re/339` §3.1 與 §九）。
//
//	逐支在這一格的艦隊（`ds:9898` 的順序）：
//	  第 0 組（旗艦那一組）艘數 0 → 全滅
//	      185C:028E  報告（`3826:0405` 臨時視窗，**沒有軍階**，與戰略那條的 0374:073F 不同）
//	      185C:00F0  擦小地圖點、作廢艦艇（含索敵分遣隊那一筆）、297B:034A 提督與副官失蹤
//	  否則
//	      185C:0177  清掉艘數 0 的組
//	      1261:0104  印這一回合積著的士氣（bit2）／全軍覆沒（bit3）報告
//	最後 12AA:03BC 重建戰場表。
//
// 子步8-2由battleOwnershipPhase接續（規格106），8-3由107接續。
// 8-4–8-5（提督升級、原版的結束判定）還沒接，
// 落差列在規格 062 與 `docs/re/339` §六。

// battleFinalPhase 從戰場艦隊清單第 i 支接著跑。停在報告上時回 true——看完由報告
// 自己接回同一個索引，與戰略那條（`attritionFrom`）同一套續接寫法。
func (g *Game) battleFinalPhase(i int, next func()) bool {
	for ; i < len(g.battleFleets); i++ {
		index := g.battleFleets[i].Index
		f := g.fleetByIndex(index)
		if f == nil {
			continue
		}
		if !g.battleFlagshipGone(index) {
			rules.RemoveEmptyGroups(f)
			g.removeEmptyBattleUnits(index)
			resume := i + 1
			if g.startBattleReports(g.moraleReports(f), func() {
				if !g.battleFinalPhase(resume, next) {
					next()
				}
			}) {
				return true
			}
			continue
		}
		idx := i
		if g.startBattleReports([]turnReport{g.battleAnnihilationReport(f)}, func() {
			g.annihilateBattleFleet(idx)
			if !g.battleFinalPhase(idx, next) {
				next()
			}
		}) {
			return true
		}
		g.annihilateBattleFleet(idx)
		i--
	}
	return false
}

// battleFlagshipGone 是 `185C:00B2`：**只看第 0 組**（旗艦那一個分隊）的艘數，
// 不看陣營、不看士氣。找不到那一個分隊也算沒了。
func (g *Game) battleFlagshipGone(index int) bool {
	for _, u := range g.battleUnits {
		if u.Fleet == index && u.Squadron == 0 {
			return u.Ships == 0
		}
	}
	return true
}

// removeEmptyBattleUnits 是 `185C:0177`：艘數 0 的分隊從戰場上拿掉（旗艦那一組留著，
// 走到這裡就表示它還有船）。
func (g *Game) removeEmptyBattleUnits(index int) {
	out := g.battleUnits[:0]
	for _, u := range g.battleUnits {
		if u.Fleet == index && u.Squadron != 0 && u.Ships == 0 {
			continue
		}
		out = append(out, u)
	}
	g.battleUnits = out
}

// battleAnnihilationReport 是 `185C:028E`：`3982:011D` 的「<陣營>第Ｎ艦隊<提督名>」
// ＋ `ds:3D46`「  \n已經全滅。    」。與戰略那條（`0374:073F`）三處不同：**沒有軍階**、
// 視窗是 `3826:0405` 的臨時視窗、**不畫頭像**——第二個參數是 `0xFFFF`，同一個視窗的
// 士氣報告（`1261:0258`／`031C` 傳座位）有頭像（`docs/re/339` §九的實測）。
func (g *Game) battleAnnihilationReport(f *rules.Fleet) turnReport {
	p := g.res.TurnWindow
	if p == nil {
		return turnReport{face: -1}
	}
	var b []byte
	b = append(b, ui.Sprintf(p.ReportFormat("battle_annihilated_prefix"), g.factionNameColoured(f.Faction))...)
	b = append(b, p.FleetNumber(f.Number)...)
	b = append(b, ui.Sprintf(p.ReportFormat("battle_annihilated_body"),
		codesBig5(g.admiralNameCodes(f.Admiral)))...)
	return turnReport{text: b, face: -1}
}

// annihilateBattleFleet 是 `185C:00F0`：作廢那一支的艦艇（編制表那幾筆 ＋ `+0x38`
// 的索敵分遣隊，**不看 `+0x3A`**）、擦掉小地圖上的點，再走與戰略那條相同的
// `297B:034A`（提督與副官失蹤、艦隊記錄拿掉、星域艦隊數減一）。
func (g *Game) annihilateBattleFleet(i int) {
	if i < 0 || i >= len(g.battleFleets) {
		return
	}
	index := g.battleFleets[i].Index
	out := g.battleUnits[:0]
	for _, u := range g.battleUnits {
		if u.Fleet == index {
			continue
		}
		out = append(out, u)
	}
	g.battleUnits = out
	g.battleFleets = append(g.battleFleets[:i], g.battleFleets[i+1:]...)
	for k := range g.fleets {
		if g.fleets[k].Index == index {
			g.annihilateFleet(k)
			break
		}
	}
	// 戰場收尾（`EndBattle`）會從存檔表重讀艦隊，所以也要寫進那一份——
	// 不寫的話這一支會在下一幀活過來。
	if g.slot != nil {
		g.slot.AnnihilateFleet(index)
	}
	// `12AA:03BC(星域)` 重建戰場的兩張表。remake 的表是切片，拿掉那幾筆就等於
	// 重建，所以這裡只把編制完整度重算（打行星讀 `+0x3C`）。原版在整個子步跑完
	// 才重建一次，remake 逐支拿掉時做——**拿掉艦隊是唯一會讓表過期的事**，結果相同。
	rules.RefreshFleetIntegrity(g.battleFleets, g.battleUnits)
}
