package game

import (
	"bytes"
	"encoding/hex"
	"sort"
	"time"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
	"github.com/wicanr2/logh3/internal/ui"
)

// 規格087：每一個觀測者的情報先更新，圖示到逐筆展示才更新。
type spottingObserver struct {
	faction, x, y, radius int
	fleet, planet         int
	human                 bool
}

type battleSpottingState struct {
	active, header bool
	stage          int // 0查詢、1已發現、2標頭等待、3來源置鏡、4標亮入口、5逐筆等待、6候選置鏡
	observers      []spottingObserver
	next           int
	source         spottingObserver
	refs           []rules.PreparationUnit
	found          []int
	item           int
	highlight      int // refs索引；-1無標亮
	ticks          int
	rem            time.Duration
	text           []byte
	sprites        map[int]int        // SHIP池索引→+2F，不由當幀Detail重算
	directions     map[int]int        // FLEET索引→+6A
	view           screen.BattleView  // 已重畫的鏡頭；RAM注入不會自行重畫原版背景。
	cameraGoal     *screen.BattleView // flag1開始步進後保留精確目標，不再套起始近距門檻。
	after          func()
}

func (g *Game) beginBattleSpotting(after func()) {
	s := battleSpottingState{active: true, highlight: -1, after: after,
		refs: g.preparationUnits(), sprites: map[int]int{}, directions: map[int]int{}, view: g.battleView}
	if g.battleDisplayKnown {
		s.view = g.battleDisplayedView
	}
	seats := [2]int{g.faction, 1 - g.faction}
	humans := [2]bool{true, g.versus}
	if g.slot != nil && len(g.slot.Work) >= 0x13 {
		w := g.slot.Work
		if w[0x11] < 2 && w[0x12] < 2 && w[0x11] != w[0x12] {
			seats = [2]int{int(w[0x11]), int(w[0x12])}
			humans = [2]bool{w[0x0D] == 0, w[0x0E] == 0}
		}
	}
	for _, r := range s.refs {
		u := r.Unit
		n := rules.BattleUnitSprite(*u, g.faction)
		if g.slot != nil {
			if raw := g.slot.ShipAt(r.Pool); len(raw) == rules.ShipRecord {
				n = int(raw[0x2F])
			}
		}
		s.sprites[r.Pool] = n
	}
	for _, bf := range g.battleFleets {
		d := rules.DirectionIndex(bf.Facing)
		if g.slot != nil && bf.Index >= 0 && (bf.Index+1)*rules.FleetRecord <= len(g.slot.Fleet) {
			d = int(g.slot.Fleet[bf.Index*rules.FleetRecord+0x6A])
		}
		s.directions[bf.Index] = d
	}
	for seat, faction := range seats {
		for _, bf := range g.battleFleets {
			if bf.Faction != faction {
				continue
			}
			for _, r := range s.refs {
				u := r.Unit
				if u.Fleet != bf.Index || (u.Squadron != 0 && !(u.Squadron < 0 && bf.ScoutMode != 0xFF)) {
					continue
				}
				radius := rules.FlagshipVision
				if u.Kind == 7 {
					radius = rules.ScoutVision
				}
				s.observers = append(s.observers, spottingObserver{faction, u.X, u.Y, radius, bf.Index, -1, humans[seat]})
			}
		}
		for _, p := range g.battlePlanets {
			if p.Owner == faction {
				s.observers = append(s.observers, spottingObserver{faction, p.X, p.Y,
					rules.PlanetQueryRadius(p.Defence), -1, p.Index, humans[seat]})
			}
		}
	}
	g.battleSpotting = s
	g.battlePhase = screen.BattlePhaseSpotting
}

func (g *Game) spottingTick(d time.Duration) {
	s := &g.battleSpotting
	if s.stage == 2 || s.stage == 5 {
		s.rem += d
		for period := g.engagementTickDuration(); s.rem >= period; s.rem -= period {
			s.ticks--
		}
		if s.ticks > 0 {
			return
		}
	}
	g.continueBattleSpotting()
}

// 每次返回都是可重畫的正式狀態。CPU一路向前，不建立假報告。
func (g *Game) continueBattleSpotting() {
	s := &g.battleSpotting
	switch s.stage {
	case 0:
		for s.next < len(s.observers) {
			s.source = s.observers[s.next]
			s.next++
			units := make([]rules.BattleUnit, len(s.refs))
			for i, r := range s.refs {
				units[i] = *r.Unit
			}
			o := s.source
			s.found = rules.ObserveSpotting(units, o.faction, o.x, o.y, o.radius)
			for i, r := range s.refs {
				*r.Unit = units[i]
				g.saveSpottingUnit(r)
			}
			if len(s.found) == 0 || !o.human {
				continue
			}
			s.item, s.stage = 0, 1
			return
		}
		s.active = false
		after := s.after
		s.after = nil
		if after != nil {
			after()
		}
	case 1:
		s.text, s.header = g.spottingHeader(s.source), true
		s.ticks, s.rem, s.stage = g.display.MsgSpeed, 0, 2
	case 2:
		s.header, s.stage = false, 3
		fallthrough
	case 3:
		if g.spottingCamera(s.source.x, s.source.y, true) {
			return
		}
		sort.Slice(s.found, func(i, j int) bool {
			a, b := s.refs[s.found[i]].Unit, s.refs[s.found[j]].Unit
			if uint8(a.Fleet) != uint8(b.Fleet) {
				return uint8(a.Fleet) < uint8(b.Fleet)
			}
			return uint8(a.Squadron) < uint8(b.Squadron)
		})
		g.spottingNextItem()
	case 4:
		r := s.refs[s.highlight]
		u := r.Unit
		s.sprites[r.Pool] = rules.BattleCachedSprite(*u, false, s.sprites[r.Pool])
		if u.Squadron == 0 {
			d := rules.BlankSprite
			if u.Detail == rules.SpottingMax {
				if bf := g.battleFleet(u.Fleet); bf != nil {
					d = rules.DirectionIndex(bf.Facing)
				}
			}
			s.directions[u.Fleet] = d
			if g.slot != nil && u.Fleet >= 0 && (u.Fleet+1)*rules.FleetRecord <= len(g.slot.Fleet) {
				g.slot.Fleet[u.Fleet*rules.FleetRecord+0x6A] = byte(d)
			}
		}
		g.saveSpottingUnit(r)
		s.ticks, s.rem, s.stage = 6, 0, 5
	case 5:
		s.highlight = -1
		s.item++
		g.spottingNextItem()
	case 6:
		g.spottingNextItem()
	}
}

func (g *Game) spottingNextItem() {
	s := &g.battleSpotting
	if s.item >= len(s.found) {
		s.stage = 0
		return
	}
	u := s.refs[s.found[s.item]].Unit
	// 381指定16×16物件的非可見旗標；不能以中心點是否在框內替代。
	x, y := u.X-g.battleView.Col*16, u.Y-g.battleView.Row*20
	if x <= -16 || x >= 448 || y <= -16 || y >= 360 {
		s.stage = 6
		g.spottingCamera(u.X, u.Y, false)
		return // 1588:0437：每次完整重畫先顯示，再判下一步或建立標亮。
	}
	s.highlight, s.stage = s.found[s.item], 4
}

// 1588:01A3，先近距閾值，再遠距直接定位；其他位置沿原版八方位單步。
func (g *Game) spottingCamera(x, y int, repeat bool) bool {
	if repeat && g.battleSpotting.cameraGoal != nil {
		goal := *g.battleSpotting.cameraGoal
		if g.battleView == goal {
			g.battleSpotting.cameraGoal = nil
			return false
		}
		g.spottingCameraStep(goal.Col*16, goal.Row*20)
		return true
	}
	px, py := g.battleView.Col*16, g.battleView.Row*20
	tx, ty := (x-216)/16*16, (y-172)/20*20
	clipped := tx < 0 || tx >= 1600 || ty < 0 || ty >= 1680
	tx, ty = max(0, min(1600, tx)), max(0, min(1680, ty))
	dx, dy := tx-px, ty-py
	abs := func(v int) int {
		if v < 0 {
			return -v
		}
		return v
	}
	if !clipped && abs(dx) < 128 && abs(dy) < 60 {
		return false
	}
	if abs(dx) >= 448 || abs(dy) >= 360 {
		g.battleView = screen.BattleView{Col: max(0, min(100, (x-224)/16)), Row: max(0, min(84, (y-180)/20))}
		g.battleSpotting.view = g.battleView
		return !repeat
	}
	if dx == 0 && dy == 0 {
		return false
	}
	if repeat {
		goal := screen.BattleView{Col: tx / 16, Row: ty / 20}
		g.battleSpotting.cameraGoal = &goal
	}
	g.spottingCameraStep(tx, ty)
	return true
}

func (g *Game) spottingCameraStep(tx, ty int) {
	g.battleView.Col, g.battleView.Row, _ = rules.ViewScrollStep(g.battleView.Col, g.battleView.Row,
		rules.Direction8(rules.AtanTable256(tx-g.battleView.Col*16, ty-g.battleView.Row*20)))
	g.battleSpotting.view = g.battleView
}

func (g *Game) saveSpottingUnit(r rules.PreparationUnit) {
	u := r.Unit
	u.Attack &^= 1
	if u.Seen {
		u.Attack |= 1
	}
	if g.slot == nil {
		return
	}
	raw := g.slot.ShipAt(r.Pool)
	if len(raw) != rules.ShipRecord {
		return
	}
	raw[0x16] = (raw[0x16] &^ 1) | byte(u.Attack&1)
	raw[0x17], raw[0x18] = byte(u.Detail), byte(u.DetailHold)
	if n, ok := g.battleSpotting.sprites[r.Pool]; ok {
		raw[0x2F] = byte(n)
	}
}

func (g *Game) spottingHeader(o spottingObserver) []byte {
	if o.faction < 0 || o.faction >= len(g.res.FactionNames) {
		return nil
	}
	faction := g.res.FactionNames[o.faction]
	tail, _ := hex.DecodeString("aabaafc1bcc4") // DS:4A0B／4A16：的索敵。
	if o.fleet >= 0 {
		f := g.battleFleet(o.fleet)
		if f == nil || g.res.TurnWindow == nil {
			return nil
		}
		raw := append([]byte{1}, []byte("[5,1,")...)
		raw = append(raw, faction...)
		raw = append(raw, 0xB2, 0xC4) // 第
		// 2526:04B1呼叫3982:0073：一位數全形，兩位數半形，共兩個bytes。
		raw = append(raw, g.res.TurnWindow.FleetNumber(f.Number)...)
		raw = append(raw, 0xC4, 0xA5, 0xB6, 0xA4) // 艦隊
		return append(raw, tail...)
	}
	if g.res.Planets == nil || g.res.PlanetInfo == nil {
		return nil
	}
	rec := g.res.Planets.Record(o.planet)
	if len(rec) < 19 {
		return nil
	}
	kind := g.res.PlanetInfo.Lines.Planet.Bytes
	if rec[18] == 1 {
		kind = g.res.PlanetInfo.Lines.Fortress.Bytes
	}
	kindRaw, _ := hex.DecodeString(kind)
	name := rec[:17]
	if n := bytes.IndexByte(name, 0); n >= 0 {
		name = name[:n]
	}
	raw := append([]byte(nil), faction...)
	raw = append(raw, kindRaw...)
	raw = append(raw, name...)
	return append(raw, tail...)
}

func (g *Game) drawSpottingHeader(fb *gfx.Framebuffer) {
	if !g.battleSpotting.header || g.res.FrameTiles == nil || g.res.Font == nil {
		return
	}
	(ui.TileFrame{Tiles: g.res.FrameTiles, Style: 1, Fill: 9}).Draw(fb, 4, 7, 22, 5)
	b := ui.NewTextBuffer(20, 3, 0x97)
	b.Write(g.battleSpotting.text)
	ui.DrawTextBuffer(fb, g.res.Font, g.res.FrameTiles, b, 5, 8)
}

func (g *Game) cachedBattleSprite(u rules.BattleUnit) int {
	if n, ok := g.battleSpotting.sprites[g.battleUnitPool(u)]; ok {
		return n
	}
	return battleSprite(u, g.faction)
}
