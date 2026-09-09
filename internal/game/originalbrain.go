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
		// 三項都是**那個陣營自己的**：軍事費、總稅收與 `Blend` 在
		// remake 也是每陣營一格（原版 `ds:8F52`／`ds:8F6A`／`ds:8FD4`）。
		// 總稅收在 AI 回合開頭跟著稅率重算過（`docs/spec/030`），
		// 維持費 ＝ Σ 各艦隊的 `+0x0A`（`docs/re/218`）。
		//
		// ⚠ 電腦那一方的**造船支出**還沒有模型，月底結算讓它跑與玩家
		// 同一套規則（`docs/spec/030` 末節）。
		//
		// `tight` ＝ `ds:8FD4[陣營] >= 3`：AI 回合開頭按軍事費設好的
		// `Blend`（`docs/spec/030`），換算過去就是軍事費不到 10000。
		Money: func(faction int) (tax, upkeep int, tight bool) {
			return g.taxTotal[faction], g.fleetUpkeep(faction),
				g.blend[faction] >= rules.BlendTightBudget
		},
		Admirals: func(faction int) []ai.AdmiralView {
			return admiralCandidates(all, g.fleets, faction)
		},
		CanForm: g.canFormFleet,
		Form:    g.formFleet,
		// 提督沒有副官時從 `ds:5F78` 填（原版 `2D76:0611`，`docs/re/219`）
		// ——副官會把艦隊能力的 `max` 拉高，維持費跟著變。
		Adjutant: func(i int) int {
			if a := rules.Adjutant(g.admiralRecord(i)); a >= 0 {
				return a
			}
			return g.res.DefaultAdjutants.For(i)
		},
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
		Region: rec,

		Faction:      faction,
		Kinds:        g.res.ShipKinds,
		Tables:       g.res.Composition,
		AdmiralLevel: rules.AdmiralLevel(adm),
		AdmiralKind:  rules.AdmiralAbility(adm, 0x08) & 0x0F,
		ClassKey:     classKey,
	})
}

// fleetUpkeep 回一個陣營的艦隊維持費 ＝ 原版 `ds:8F72[陣營]`
// ＝ 那一方**在用**的每一支艦隊的 `+0x0A` 加總（`docs/re/211`、
// `docs/re/218`）。
//
// 只有帶編制的艦隊算得出來；從存檔載進來的艦隊 `Composition` 是空的，
// 那幾支目前算 0。
func (g *Game) fleetUpkeep(faction int) int {
	if g.res.ShipKinds == nil {
		return 0
	}
	sum := 0
	for _, f := range g.fleets {
		if f.Faction != faction {
			continue
		}
		sum += rules.FleetUpkeep(g.res.ShipKinds, faction, f.Composition, f.Operations)
	}
	return sum
}

// fleetOperations 回一支新艦隊的 `+0x47`（營運）＝ `max(提督, 副官)` 的
// `ADMIRAL` `+14`（`docs/re/218`）。
func (g *Game) fleetOperations(admiral, adjutant int) int {
	v := rules.AdmiralAbility(g.admiralRecord(admiral), 0x0E)
	if adjutant >= 0 {
		if a := rules.AdmiralAbility(g.admiralRecord(adjutant), 0x0E); a > v {
			v = a
		}
	}
	return v
}

// formFleet 把一支新艦隊真的編出來 ＝ 原版 `2D76:0611` 的後半：
// 產生編制（**從星域存量扣掉艦艇**）、算 `+0x47`（營運）與 `+0x0A`
// （維持費）。編不出任何一組就回 false。
func (g *Game) formFleet(faction int, nf *ai.NewFleet) bool {
	comp := g.newComposition(faction, nf.Region, nf.Admiral, nf.ClassKey)
	if len(comp) == 0 {
		return false
	}
	nf.Composition = comp
	nf.Operations = g.fleetOperations(nf.Admiral, nf.Adjutant)
	nf.Upkeep = rules.FleetUpkeep(g.res.ShipKinds, faction, comp, nf.Operations)
	return true
}
