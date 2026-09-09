package game

import (
	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/rules"
)

// 把 `ai.OriginalStrategy` 接上這一局的資料（規格 `docs/spec/029`）。
//
// **這裡不動預設。** `New` 建出來的 `Game` 仍然用 `ai.Original`
// （什麼都不做）；要換是 `SetBrain(g.OriginalBrain())`，那一步是
// `docs/spec/029` 末節那個裁決。

// strategyRandSeed 是戰略 AI 亂數的起始種子（原版 `ds:94DC`）。
//
// **量到了**：收據 `turn260mem` 在第一次 AI 回合之前倒的那一份，
// `ds:94DC` 就是這個值（`docs/re/209`）——而 `docs/re/184` 在另一個劇本
// 量到的呼叫前種子也相同，代表載入之後到第一次 AI 之間沒有東西動過它。
const strategyRandSeed = 0xD3DC167E

// OriginalBrain 回接好線的原版戰略 AI。
//
// 缺星域圖或星域記錄就回 nil——那兩樣是決策層的地基，沒有的話
// `ai.Turn` 也走不動。
func (g *Game) OriginalBrain() ai.Brain {
	if g.res == nil || g.res.Graph == nil || len(g.res.Stellar) == 0 {
		return nil
	}
	if g.strategyRand == nil {
		g.strategyRand = rules.NewRand(strategyRandSeed)
	}
	return ai.OriginalStrategy{
		Contested:   g.regionContested,
		Blocked:     g.regionBlocked,
		EnemyFleets: g.enemyFleets,
		Split:       func(f rules.Fleet) bool { return f.Flags&1 != 0 },
		Ships:       g.fleetShipTotals,
		Routes:      g.res.AdvanceRoutes,
		Rand:        g.strategyRand,
		Support:     g.buildSupport(),
	}
}

// regionRecord 回一筆執行期星域記錄，超出範圍回 nil。
func (g *Game) regionRecord(r int) []byte {
	o := r * rules.RegionRecord
	if r < 0 || o+rules.RegionRecord > len(g.res.Stellar) {
		return nil
	}
	return g.res.Stellar[o : o+rules.RegionRecord]
}

// regionContested ＝ 星域 `+0x13` 的 bit2：這一格有戰鬥（`docs/re/173`）。
func (g *Game) regionContested(r int) bool {
	rec := g.regionRecord(r)
	return rec != nil && rec[0x13]&0x04 != 0
}

// regionBlocked ＝ 星域 `+0x13` 的 `& 0x14`（`docs/re/197`）。
func (g *Game) regionBlocked(r int) bool {
	rec := g.regionRecord(r)
	return rec != nil && rec[0x13]&0x14 != 0
}

// enemyFleets 回敵方在那一格有幾支艦隊（`+0x28`／`+0x29`，`docs/re/199`）。
//
// 「敵方」是相對於**電腦那一方**——`runBrain` 把 `1 - g.faction` 交給
// 決策層，所以敵方就是玩家的陣營。
func (g *Game) enemyFleets(r int) int {
	rec := g.regionRecord(r)
	if rec == nil {
		return 0
	}
	return rules.RegionFleets(rec, g.faction)
}

// fleetShipTotals 回一支艦隊的艘數與主力槽數。
//
// 新遊戲沒有存檔（`g.slot` 是 nil）也沒有艦艇——原版的新遊戲樣板同樣
// 沒有 `SHIP.DTX`（`docs/re/200`），所以兩個都回 0 是對的。
func (g *Game) fleetShipTotals(f rules.Fleet) (ships, mainSlots int) {
	if g.slot == nil {
		return 0, 0
	}
	return g.slot.FleetShipTotals(f.Index)
}

// buildSupport 把階段一（建艦隊）要的東西接起來（`docs/re/202`–`205`）。
// 缺提督表就回 nil ＝ 不建艦隊。
func (g *Game) buildSupport() *ai.BuildSupport {
	if len(g.res.Admiral) < rules.AdmiralRecord {
		return nil
	}
	all := rules.ParseAdmirals(g.res.Admiral)
	return &ai.BuildSupport{
		// ⚠ **`Money` 回的是玩家那一方的數字**：remake 只追一方的
		// 軍事費與稅收（`g.taxTotal`），電腦那一方的財政沒有模型。
		// 單人模式下這是把玩家的稅收拿去當電腦的預算——**已知不忠實**，
		// 而且因為維持費也是 0，那道閘門實際上永遠通過。
		Money: func(faction int) (tax, upkeep int, tight bool) {
			// 總稅收 ＝ ds:8F6A。**艦隊維持費（ds:8F72）remake 還沒算**
			// ——收據上帝國那一欄是 0（`docs/internals/27`），所以這裡
			// 也是 0，閘門的那一項等於永遠通過。
			// `ds:8FD4[陣營]` ＝ 未安定度公式的 `Blend`（0–4，`docs/re/207`）。
			// remake 只在載入時算未安定度用它，還沒有跨回合維護，
			// 所以這裡一律 false ＝ 當成 0（門檻是 3，開局值 2 本來也不觸發）。
			return g.taxTotal, 0, false
		},
		Admirals: func(faction int) []ai.AdmiralView {
			return admiralCandidates(all, g.fleets, faction)
		},
		CanForm:  g.canFormFleet,
		Adjutant: func(i int) int { return rules.Adjutant(g.admiralRecord(i)) },
	}
}

// admiralCandidates 收候選提督 ＝ 原版 `28E0:0103(&清單, 0xFF, 0x25, 0)`
// （`docs/re/208`）。`0x25` 是三個過濾位元：
//
//	bit0  +0x01 == 陣營      同陣營
//	bit2  +0x00 == 0xFF      沒有職務
//	bit5  +0x09 bit7 有設    沒有帶艦隊
//
// 加上底層那道 `+0x00 != 0xFE`。
//
// ⚠ **「沒有帶艦隊」這一項用 remake 自己的艦隊表推，不是讀 `+0x09` bit7**：
// `docs/internals/08` 記著一個待驗——載入時 `+0x09` 被夾成 `0x7F`，
// 最高位一定是 0，那樣就不會有人合格。矛盾沒解之前用艦隊表比較可靠。
func admiralCandidates(all []rules.Admiral, fleets []rules.Fleet, faction int) []ai.AdmiralView {
	led := make(map[int]bool, len(fleets))
	for _, f := range fleets {
		led[f.Admiral] = true
		if f.Adjutant >= 0 {
			led[f.Adjutant] = true
		}
	}
	out := make([]ai.AdmiralView, 0, len(all))
	for _, a := range all {
		if a.Faction != faction || led[a.Index] {
			continue
		}
		// `+0x00`：0xFE 連底層都不收，0xFF 以外的是有職務的。
		if rules.AdmiralAbility(a.Record, 0x00) != 0xFF {
			continue
		}
		out = append(out, ai.AdmiralView{
			Index:   a.Index,
			Level:   rules.AdmiralLevel(a.Record),
			Command: rules.AdmiralAbility(a.Record, 0x0B),
			Ability: [4]int{
				rules.AdmiralAbility(a.Record, 0x14), // 類別 0：+20
				rules.AdmiralAbility(a.Record, 0x13), // 類別 1：占領
				rules.AdmiralAbility(a.Record, 0x12), // 類別 2：防禦
				rules.AdmiralAbility(a.Record, 0x17), // 類別 3：運用
			},
			Kind: rules.AdmiralAbility(a.Record, 0x08) & 0x0F,
			// Region 用提督記錄的 `+0x0A`（所在星域），`canFormFleet` 要看。
			Region: rules.AdmiralAbility(a.Record, 0x0A),
		})
	}
	return out
}

// canFormFleet ＝ 原版 `07F9:023B` ＋ `07F9:02CB`：那一位所在星域裡、
// 他那個艦種的艦艇存量要**不低於艦種表那一列的滿編艘數**
// （`ds:4E3A` 的 word 0，`docs/internals/40`）。
//
// 缺艦種表就回 false ＝ 編不成，不用代用門檻。
func (g *Game) canFormFleet(faction int, a ai.AdmiralView) bool {
	rec := g.regionRecord(a.Region)
	if rec == nil || g.res.ShipKinds == nil {
		return false
	}
	stock := rules.RegionShips(rec, faction)
	if a.Kind < 0 || a.Kind >= len(stock) {
		return false
	}
	k, ok := g.res.ShipKinds.Kind(faction, a.Kind)
	if !ok || k.FullShips <= 0 {
		return false
	}
	return stock[a.Kind] >= k.FullShips
}

// newComposition 產生一支新艦隊的編制（`rules.BuildComposition`，
// 原版 `297B:0066` ＋ `0FF4:11F4`，`docs/re/212`–`docs/re/217`）。
//
// **會就地扣掉那一格的艦艇存量**——原版就是這樣：新艦隊不是憑空變出來的，
// 是把星域的庫存編成部隊（`docs/re/215`）。
//
// 缺艦種表或樣板表就回 nil；那時候艦隊仍然建得出來，只是沒有編制。
func (g *Game) newComposition(faction, region, admiral, classKey int) []rules.FleetGroup {
	rec := g.regionRecord(region)
	if rec == nil || g.res.ShipKinds == nil || g.res.Composition == nil {
		return nil
	}
	adm := g.admiralRecord(admiral)
	if len(adm) < rules.AdmiralRecord {
		return nil
	}
	return rules.BuildComposition(rules.BuildInput{
		Region:       rec,
		RegionOwner:  int(rec[0x12]),
		Faction:      faction,
		Kinds:        g.res.ShipKinds,
		Tables:       g.res.Composition,
		AdmiralLevel: rules.AdmiralLevel(adm),
		AdmiralKind:  rules.AdmiralAbility(adm, 0x08) & 0x0F,
		ClassKey:     classKey,
	})
}
