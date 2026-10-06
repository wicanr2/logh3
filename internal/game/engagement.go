package game

import (
	"encoding/hex"
	"time"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 交戰演出——打行星那一種（`docs/re/324`、規格 056）。
//
// 原版戰場回合處理的打行星那一步（`20B4:0222` 起）對每一顆有人打的行星：
// `1DD6:0152` 開場（`ANI_PNL1.A98` 底圖、三個樣式 0 的文字視窗、換戰鬥曲）→
// `20B4:0486` 打（同時放方塊動畫、印「擊毀」）→ `1DD6:026B` 收場（換回平時曲）。
// 行星防衛攻擊（`20B4:08B1`，`docs/re/331`）的時間線以計時器 tick 計（底圖畫下去
// 是 tick 0，間隔 I ＝ `ds:8F10[動畫速度]` ＝ 8／4／0）：I 行星 `P?D0` 第 0 格、2I 艦隊
// `??D0` 第 0 格、3I–12I 行星 `P?AB` 十格、13I 起艦隊 `??D{0,1,2}` 整段 → 印「擊毀」與
// 反擊後的艘數 → 停頓「訊息速度」tick（`3765:006C`）→ 收場。播放中**不讀輸入**：
// 左／右鍵排進佇列，到停頓時才讀、讀到就提早收場。動畫效果關掉時四次畫圖都跳過、
// 直接到「擊毀」與停頓。
//
// 既有防衛標題分支的量表已依規格094接入；戰場方塊動畫沒畫（`321F:008A`／`0897`）。
// 舊演出收據仍排除量表與方塊，新的量表測試獨立逐像素核對。
// 艦隊對艦隊、要塞砲、登陸艦攻擊（`20B4:0681`）那幾種的圖沒接。

// engagement 是一場演出要印的數。
type engagement struct {
	planet, fleet int
	// defence／ships 是演出開始那一刻（反擊之前）的值。
	defence, defenceMax int
	ships, shipsMax     int
	destroyed           int
	label               screen.EngagementLabel
	// planetSide／fleetSide 是行星與艦隊畫在哪個視窗（0 左、1 右，`3958:0041` 的 `+0`）；
	// planetKind 是段名第二個字元（P 行星、F 要塞、Z 伊謝爾倫）、fleetName 是艦隊段名前兩個字元。
	planetSide, fleetSide int
	planetKind            byte
	fleetName             [2]byte
	gauges                *rules.PlanetDefenceGauges
}

// queueEngagements 把這一回合打行星的結果排成演出（一顆行星一場：原版
// `20B4:0222` 逐顆行星開場；同一顆被幾支打就在同一場裡逐支印——這裡先
// 一支一場，假設待驗）。回 true ＝ 有得演。
func (g *Game) queueEngagements(hits []rules.PlanetAssaultHit) bool {
	if g.res.Engagement == nil || g.res.EngagementPanel == nil {
		return false
	}
	g.engQueue = g.engQueue[:0]
	for _, h := range hits {
		if h.Planet < 0 || h.Planet >= len(g.battlePlanets) {
			continue
		}
		p := g.battlePlanets[h.Planet]
		after := g.battleFleetShips(h.Fleet)
		e := engagement{
			planet: h.Planet, fleet: h.Fleet,
			defence: p.Defence + h.Damage, defenceMax: p.DefenceMax,
			ships: after + h.Counter, shipsMax: g.battleFleetShipsMax(h.Fleet),
			destroyed: h.Counter,
		}
		// 中間那格的標題：原版四種字串（防衛攻擊／登陸艦攻擊／占領成功／占領失敗）
		// 各在什麼情形印沒讀完；收據那一場是「只挨反擊」→ 防衛攻擊。其餘假設待驗。
		switch {
		case h.Captured:
			e.label = screen.EngagementCaptured
		case h.Damage > 0:
			e.label = screen.EngagementLanding
		default:
			e.label = screen.EngagementDefence
		}
		if e.label == screen.EngagementDefence && h.DefenceGauges != nil {
			copy := *h.DefenceGauges
			e.gauges = &copy
		}
		g.engQueue = append(g.engQueue, e)
	}
	if len(g.engQueue) == 0 {
		return false
	}
	for i := range g.engQueue {
		g.fillEngagementNames(&g.engQueue[i])
	}
	g.startEngagement()
	g.state = StateEngagement
	return true
}

// fillEngagementNames 算這一場的段名與左右（`20B4:1370`／`1269` 組的記錄，`docs/re/331` §1.4）。
func (g *Game) fillEngagementNames(e *engagement) {
	if e.planet < 0 || e.planet >= len(g.battlePlanets) {
		return
	}
	p := g.battlePlanets[e.planet]
	faction := 0
	admiral := -1
	for _, f := range g.battleFleets {
		if f.Index == e.fleet {
			faction = f.Faction
		}
	}
	admiral = g.fleetAdmiral(e.fleet)
	e.fleetSide = faction
	e.planetSide = p.Owner
	if p.Owner != 0 && p.Owner != 1 { // 中立的行星畫在艦隊的對面
		e.planetSide = 1 - faction
	}
	e.planetKind = 'P'
	if p.Fortress {
		e.planetKind = 'F'
		if g.res.Engagement != nil && g.battleRegion == g.res.Engagement.Animation.FortressRegion {
			e.planetKind = 'Z'
		}
	}
	e.fleetName = [2]byte{"EF"[faction%2], 0xDD}
	if g.res.Engagement != nil {
		if codes, err := hex.DecodeString(g.res.Engagement.Animation.FlagshipCodes); err == nil &&
			admiral >= 0 && admiral < len(codes) {
			e.fleetName[1] = codes[admiral]
		}
	}
}

// startEngagement 把佇列最前面那一場從 tick 0 開始。
func (g *Game) startEngagement() {
	g.engTick, g.engTickRemainder, g.engClick = 0, 0, false
}

// engagementPicture 是時間線上的一次畫圖：哪一段、哪一格、哪個視窗、第幾 tick。
type engagementPicture struct {
	segment string
	frame   int
	side    int
	tick    int
}

// engagementTimeline 回這一場的畫圖序列與「擊毀」那一刻的 tick（`20B4:08B1`）。
func (g *Game) engagementTimeline(e engagement) (pics []engagementPicture, result int) {
	if g.res.Comani == nil || g.res.Engagement == nil || g.display.Anim == 0 {
		return nil, 0
	}
	interval := 8
	if iv := g.res.Engagement.Animation.Intervals; g.display.AnimSpeed >= 0 && g.display.AnimSpeed < len(iv) {
		interval = iv[g.display.AnimSpeed]
	}
	tick := 0
	add := func(seg string, frames []int, side int) {
		for _, f := range frames {
			tick += interval
			pics = append(pics, engagementPicture{segment: seg, frame: f, side: side, tick: tick})
		}
	}
	planetD0 := string([]byte{'P', e.planetKind, 'D', '0'})
	planetAB := string([]byte{'P', e.planetKind, 'A', 'B'})
	fleetD0 := string([]byte{e.fleetName[0], e.fleetName[1], 'D', '0'})
	add(planetD0, []int{0}, e.planetSide)
	add(fleetD0, []int{0}, e.fleetSide)
	add(planetAB, g.comaniFrames(planetAB), e.planetSide)
	// 損失百分比（`20B4:103D` 設的 `ds:A768`）：< 20 → D0、< 40 → D1、其餘 D2；0 → 只重畫 D0 第 0 格。
	pct := 0
	if e.ships > 0 {
		pct = e.destroyed * 100 / e.ships
		if pct == 0 && e.destroyed > 0 {
			pct = 1
		}
	}
	switch {
	case pct == 0:
		add(fleetD0, []int{0}, e.fleetSide)
	default:
		sub := byte('2')
		if pct < 20 {
			sub = '0'
		} else if pct < 40 {
			sub = '1'
		}
		seg := string([]byte{e.fleetName[0], e.fleetName[1], 'D', sub})
		add(seg, g.comaniFrames(seg), e.fleetSide)
	}
	return pics, tick
}

// comaniFrames 回一段的全部格序號（段不存在回 nil）。
func (g *Game) comaniFrames(seg string) []int {
	fr, err := g.res.Comani.Frames(seg)
	if err != nil {
		return nil
	}
	out := make([]int, len(fr))
	for i := range out {
		out[i] = i
	}
	return out
}

// engagementTickDuration 是一個計時器 tick 的長度。
// TimerTick 是原版一個計時器 tick 的長度（`3765:006C` 等待數的單位）。
// 腳本化的驗收（`cmd/gin3 -script` 的 `tick N`）用它推進時間。
func (g *Game) TimerTick() time.Duration { return g.engagementTickDuration() }

func (g *Game) engagementTickDuration() time.Duration {
	if g.res.Engagement != nil && g.res.Engagement.Animation.TickNS > 0 {
		return time.Duration(g.res.Engagement.Animation.TickNS)
	}
	return 17164 * time.Microsecond
}

// engagementTick 讓演出往前走 n 個計時器 tick：擊毀之後停頓滿「訊息速度」或停頓時
// 佇列裡有按鍵就收場。
func (g *Game) engagementTick(n int) {
	for ; n > 0 && g.state == StateEngagement; n-- {
		g.engTick++
		g.engagementMaybeEnd()
	}
}

// engagementMaybeEnd 看停頓該不該結束（`3765:006C`）。
func (g *Game) engagementMaybeEnd() {
	if len(g.engQueue) == 0 {
		return
	}
	_, result := g.engagementTimeline(g.engQueue[0])
	if g.engTick < result {
		return
	}
	if g.engClick || g.engTick >= result+g.display.MsgSpeed {
		g.finishEngagement()
	}
}

// battleFleetShips 回這一支現在的艘數合計（船艦總數的分子）。
func (g *Game) battleFleetShips(fleet int) int {
	n := 0
	for _, u := range g.battleUnits {
		if u.Fleet == fleet {
			n += u.Ships
		}
	}
	return n
}

// battleFleetShipsMax 回這一支的滿編艘數合計（船艦總數的分母；假設待驗，
// 收據那一支剛好滿編 22000/22000）。
func (g *Game) battleFleetShipsMax(fleet int) int {
	n := 0
	for _, u := range g.battleUnits {
		if u.Fleet == fleet {
			n += u.FullShips
		}
	}
	return n
}

// engagementView 把目前這一場排成畫面要印的東西。
func (g *Game) engagementView() *screen.EngagementView {
	if len(g.engQueue) == 0 {
		return nil
	}
	e := g.engQueue[0]
	if e.planet < 0 || e.planet >= len(g.battlePlanets) {
		return nil
	}
	p := g.battlePlanets[e.planet]
	v := &screen.EngagementView{
		Faction: p.Owner, PlanetName: g.battlePlanetName(e.planet),
		Defence: e.defence, DefenceMax: e.defenceMax,
		Ships: e.ships, ShipsMax: e.shipsMax,
		Label: e.label, Destroyed: e.destroyed,
	}
	if _, result := g.engagementTimeline(e); g.engTick >= result {
		v.ShowResult = true
		v.Ships = e.ships - e.destroyed
	}
	if e.gauges != nil {
		v.Gauges = &screen.EngagementGaugePair{
			FleetFaction: e.fleetSide, Defence: e.defence,
			Integrity: e.gauges.BeforeIntegrity, Attack: e.gauges.BeforeAttack,
		}
		if v.ShowResult {
			v.Gauges.Integrity, v.Gauges.Attack = e.gauges.AfterIntegrity, e.gauges.AfterAttack
		}
	}
	// 艦隊那格：自軍或索敵等級 3 的才印編號與提督（`ds:4801` 那句「第？艦隊
	// ？？？？？？？？？」是給看不清的），門檻照右下角敵艦隊分頁的規則
	// （`docs/re/319`；假設待驗）。
	for _, f := range g.battleFleets {
		if f.Index != e.fleet {
			continue
		}
		known := f.Faction == g.faction
		if !known {
			for _, u := range g.battleUnits {
				if u.Fleet == e.fleet && u.Detail >= 3 {
					known = true
					break
				}
			}
		}
		if known {
			v.FleetKnown, v.FleetNumber = true, f.Number
			if i := g.fleetAdmiral(e.fleet); i >= 0 && i < len(g.res.AdmiralNames) {
				v.AdmiralName = codesBig5(g.res.AdmiralNames[i])
			}
		}
	}
	return v
}

// fleetAdmiral 回戰略層那一支艦隊的提督索引；找不到回 −1。
func (g *Game) fleetAdmiral(index int) int {
	for _, f := range g.fleets {
		if f.Index == index {
			return f.Admiral
		}
	}
	return -1
}

// battlePlanetName 回戰場上第 i 顆行星的名字（Big5）。
func (g *Game) battlePlanetName(i int) []byte {
	if g.res.Planets == nil || g.battleRegion < 0 {
		return nil
	}
	start, n := g.res.Planets.Range(g.battleRegion)
	if i < 0 || i >= n {
		return nil
	}
	return planetNameBytes(g.res.Planets.Record(start + i))
}

// engagementClick 是演出期間的左鍵／右鍵：排進佇列（播放中不讀），停頓時讀到就收場
// （`docs/re/331` §三）。
func (g *Game) engagementClick() {
	g.engClick = true
	g.engagementMaybeEnd()
}

// finishEngagement 收場：下一場從頭開始，沒有了就回戰場做完那一回合。
func (g *Game) finishEngagement() {
	if len(g.engQueue) > 0 {
		g.engQueue = g.engQueue[1:]
	}
	g.startEngagement()
	if len(g.engQueue) > 0 {
		return
	}
	g.state = StateBattle
	g.finishBattleTurn()
}

// drawEngagement 畫演出：底圖與三個視窗（沒有游標）。
func (g *Game) drawEngagement() *gfx.Framebuffer {
	if g.fleetCombat.active {
		return g.drawFleetCombatScene()
	}
	fb := gfx.NewFramebuffer(0)
	if g.res.Engagement == nil {
		return fb
	}
	g.res.Engagement.Draw(fb, g.res.Font, g.res.FrameTiles, g.res.EngagementPanel, g.engagementView())
	g.drawEngagementPictures(fb)
	return fb
}

// drawEngagementPictures 把時間線上到現在為止的每一格依序疊上兩個視窗（覆寫格與 XOR 格
// 都照 `internal/asset.BMGFrame.Draw`），從底圖重算，所以任何 tick 都畫得出來。
func (g *Game) drawEngagementPictures(fb *gfx.Framebuffer) {
	if len(g.engQueue) == 0 || g.res.Comani == nil || g.res.Engagement == nil {
		return
	}
	pics, _ := g.engagementTimeline(g.engQueue[0])
	win := g.res.Engagement.Animation.Windows
	var planes asset.Planes
	drawn := false
	for _, pc := range pics {
		if pc.tick > g.engTick || pc.side < 0 || pc.side >= len(win) {
			break
		}
		fr, err := g.res.Comani.Frames(pc.segment)
		if err != nil || pc.frame >= len(fr) {
			continue
		}
		if !drawn {
			planes, drawn = asset.PlanesFrom(fb.Pix), true
		}
		_ = fr[pc.frame].Draw(planes, asset.BMGWindowBase(win[pc.side][0], win[pc.side][1]))
	}
	if !drawn {
		return
	}
	pix := planes.Pixels()
	for _, w := range win {
		x0, y0 := w[0]*8, 40+w[1]*4
		for y := y0; y < y0+102 && y < gfx.Height; y++ {
			fb.WriteRow(x0, y, pix[y*gfx.Width+x0:y*gfx.Width+x0+256])
		}
	}
}
