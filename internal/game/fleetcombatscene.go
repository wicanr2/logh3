package game

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

type fleetCombatSceneState struct {
	own, target                 int
	weapon                      byte
	command                     rules.FleetCombatCommand
	before                      []rules.BattleUnit
	integrity                   [2]int
	pictures                    []engagementPicture
	sounds                      []fleetSceneSound
	attackEnd, resultTick, tick int
	damaged, click              bool
	hit                         rules.FleetCombatHit
}

func (g *Game) fleetDamageCommand(own, target int, weapon byte) (rules.FleetCombatCommand, error) {
	c := rules.FleetCombatCommand{Weapon: int(weapon)}
	a, err := g.fleetCombatRawFleet(g.battleUnits[own].Fleet)
	if err != nil {
		return c, err
	}
	b, err := g.fleetCombatRawFleet(g.battleUnits[target].Fleet)
	if err != nil {
		return c, err
	}
	r, err := g.fleetCombatRecord(target)
	if err != nil {
		return c, err
	}
	for i := range c.AttackAbilities {
		c.AttackAbilities[i] = int(a[0x3d+i])
		c.DefendAbilities[i] = int(b[0x3d+i])
	}
	admiral := g.admiralRec(int(b[4]))
	if len(admiral) != rules.AdmiralRecord {
		return c, fmt.Errorf("交戰目標提督原表缺失")
	}
	c.TargetFacing, c.TargetWeapon, c.TargetAttribute = int(b[0x4a]), int(r[0x2e]), int(admiral[0x12])
	c.TargetComputer = !g.humanSeat(g.battleUnits[target].Faction)
	return c, nil
}

func (g *Game) beginFleetCombatScenes() {
	c := &g.fleetCombat
	var entries []rules.FleetCombatOrderEntry
	var units []int
	for _, fleet := range c.fleets {
		r, err := g.fleetCombatRawFleet(fleet)
		if err != nil {
			g.fleetCombatFail(err)
			return
		}
		admiral := g.admiralRec(int(r[4]))
		if len(admiral) != rules.AdmiralRecord {
			g.fleetCombatFail(fmt.Errorf("交戰提督原表缺失"))
			return
		}
		ops := int(admiral[0x17])
		if r[5] != 255 {
			a := g.admiralRec(int(r[5]))
			if len(a) != rules.AdmiralRecord {
				g.fleetCombatFail(fmt.Errorf("交戰副官原表缺失"))
				return
			}
			ops = max(ops, int(a[0x17]))
		}
		mobility, err := rules.FleetCombatMobility(ops, int(r[15]))
		if err != nil {
			g.fleetCombatFail(err)
			return
		}
		for _, i := range g.fleetCombatUnits(fleet) {
			entries = append(entries, rules.FleetCombatOrderEntry{Unit: g.battleUnits[i], Attribute: int(admiral[0x0b]), Mobility: mobility})
			units = append(units, i)
		}
	}
	order, keys, err := rules.FleetCombatOrder(entries, g.res.ShipKinds, g.battleRand)
	if err != nil {
		g.fleetCombatFail(err)
		return
	}
	for i, key := range keys {
		r, err := g.fleetCombatRecord(units[i])
		if err != nil {
			g.fleetCombatFail(err)
			return
		}
		binary.LittleEndian.PutUint16(r[0x42:], key)
	}
	c.commands = nil
	for _, i := range order {
		c.commands = append(c.commands, units[i])
	}
	c.commandAt = 0
	c.stage = fleetCombatScene
	g.state = StateEngagement
	g.nextFleetCombatScene()
}

func (g *Game) nextFleetCombatScene() {
	c := &g.fleetCombat
	for c.commandAt < len(c.commands) {
		own := c.commands[c.commandAt]
		c.commandAt++
		pool, weapon, err := g.fleetCombatCommand(own)
		if err != nil {
			g.fleetCombatFail(err)
			return
		}
		if weapon == 255 || weapon == 3 {
			continue
		}
		if weapon > 2 {
			g.fleetCombatFail(fmt.Errorf("交戰原池武器無效"))
			return
		}
		target := -1
		for i, u := range g.battleUnits {
			if u.Squadron >= 0 && g.battleUnitPool(u) == pool {
				target = i
				break
			}
		}
		if target < 0 {
			g.fleetCombatFail(fmt.Errorf("交戰目標池沒有分隊"))
			return
		}
		if g.battleUnits[own].Ships == 0 || g.battleUnits[target].Ships == 0 {
			continue
		}
		if g.battleUnits[own].Faction == g.battleUnits[target].Faction {
			g.fleetCombatFail(fmt.Errorf("交戰原池目標陣營無效"))
			return
		}
		command, err := g.fleetDamageCommand(own, target, weapon)
		if err != nil {
			g.fleetCombatFail(err)
			return
		}
		if g.res.FleetPresentation == nil || g.res.FleetSceneSprites == nil || g.res.FleetCommand == nil || g.res.EngagementPanel == nil || g.res.Engagement == nil {
			g.fleetCombatFail(fmt.Errorf("缺艦隊交戰畫面素材"))
			return
		}
		s := &fleetCombatSceneState{own: own, target: target, weapon: weapon, command: command, before: append([]rules.BattleUnit(nil), g.battleUnits...)}
		for _, i := range []int{own, target} {
			u := g.battleUnits[i]
			s.integrity[u.Faction] = g.battleFleet(u.Fleet).Integrity
		}
		c.scene = s
		if err := g.fleetSceneProjection(s); err != nil {
			g.fleetCombatFail(err)
			return
		}
		if err := g.startFleetScenePictures(s); err != nil {
			g.fleetCombatFail(err)
			return
		}
		g.updateFleetScene()
		return
	}
	g.finishFleetCombat()
}

func (g *Game) fleetSceneName(i int) ([2]byte, error) {
	var name [2]byte
	u := g.battleUnits[i]
	if u.Faction < 0 || u.Faction > 1 || u.Kind < 0 || u.Kind > 7 {
		return name, fmt.Errorf("交戰動畫分隊欄位無效")
	}
	name[0] = "EF"[u.Faction]
	name[1] = g.res.FleetCommand.AnimationKind(u.Kind)
	if u.Squadron == 0 {
		codes, err := hex.DecodeString(g.res.Engagement.Animation.FlagshipCodes)
		admiral := g.fleetAdmiral(u.Fleet)
		if err != nil || admiral < 0 || admiral >= len(codes) {
			return name, fmt.Errorf("交戰旗艦動畫名字缺失")
		}
		name[1] = codes[admiral]
	}
	return name, nil
}

func (g *Game) fleetSceneInterval() int {
	if g.display.Anim == 0 {
		return 0
	}
	i := g.display.AnimSpeed
	if i < 0 || i >= len(g.res.Engagement.Animation.Intervals) {
		return 8
	}
	return g.res.Engagement.Animation.Intervals[i]
}

func (g *Game) fleetSceneAdd(s *fleetCombatSceneState, i int, action, suffix byte, all bool, tick *int) error {
	name, err := g.fleetSceneName(i)
	if err != nil {
		return err
	}
	segment := string([]byte{name[0], name[1], action, suffix})
	if g.res.Comani == nil {
		return fmt.Errorf("缺交戰動畫庫")
	}
	frames, err := g.res.Comani.Frames(segment)
	if err != nil {
		return err
	}
	n := 1
	if all {
		n = len(frames)
	}
	for frame := 0; frame < n; frame++ {
		*tick += g.fleetSceneInterval()
		s.pictures = append(s.pictures, engagementPicture{segment: segment, frame: frame, side: g.battleUnits[i].Faction, tick: *tick})
	}
	return nil
}

func (g *Game) startFleetScenePictures(s *fleetCombatSceneState) error {
	if g.display.Anim == 0 {
		return nil
	}
	tick := 0
	for _, i := range []int{s.own, s.target} {
		if err := g.fleetSceneAdd(s, i, 'D', '0', false, &tick); err != nil {
			return err
		}
	}
	if err := g.addFleetSceneSound(s, s.own, int(s.weapon), tick, len(s.pictures), false); err != nil {
		return err
	}
	if err := g.fleetSceneAdd(s, s.own, 'A', g.res.FleetCommand.AnimationWeapon(int(s.weapon)), true, &tick); err != nil {
		return err
	}
	s.attackEnd = tick
	return nil
}

func (g *Game) fleetSceneProjection(s *fleetCombatSceneState) error {
	v, err := g.fleetFrameView(s, false)
	if err != nil {
		return err
	}
	var sources [2][]rules.FleetProjectionUnit
	for side, list := range v.Scene.Sides {
		for _, u := range list {
			sources[side] = append(sources[side], u.FleetProjectionUnit)
		}
	}
	p, err := rules.ProjectFleetScene(sources)
	if err != nil {
		return err
	}
	for _, u := range p.Units {
		r, err := g.fleetCombatRecord(u.ID)
		if err != nil {
			return err
		}
		binary.LittleEndian.PutUint16(r[0x3e:], uint16(u.X))
		binary.LittleEndian.PutUint16(r[0x40:], uint16(u.Y))
		binary.LittleEndian.PutUint16(r[0x44:], uint16(u.Depth))
		visible := uint16(0)
		if u.Visible {
			visible = 1
		}
		binary.LittleEndian.PutUint16(r[0x42:], visible)
	}
	return nil
}

func (g *Game) applyFleetSceneDamage(s *fleetCombatSceneState) error {
	h, err := rules.ApplyFleetCombatDamage(&g.battleUnits[s.own], &g.battleUnits[s.target], g.res.ShipKinds, s.command, g.battleRand)
	if err != nil {
		return err
	}
	s.hit, s.damaged = h, true
	if !h.Skipped {
		g.battleScore[h.Scorer] += h.Score
		g.campaignScore[h.Scorer] += h.Score
		g.battleFleetScore[h.Scorer] += h.Score
		if len(g.workRecord()) >= rules.WorkRecord {
			w := g.mutableWork()
			for side := 0; side < 2; side++ {
				binary.LittleEndian.PutUint32(w[0xd0+side*4:], uint32(g.battleScore[side]))
				binary.LittleEndian.PutUint32(w[0xd8+side*4:], uint32(g.campaignScore[side]))
			}
		}
		r, err := g.fleetCombatRawFleet(g.battleUnits[s.own].Fleet)
		if err != nil {
			return err
		}
		for _, a := range []struct{ index, amount int }{{int(r[4]), h.AdmiralMerit}, {int(r[5]), h.AdjunctMerit}} {
			if a.index == 255 {
				continue
			}
			data := g.mutableAdmirals()
			if a.index < 0 || (a.index+1)*rules.AdmiralRecord > len(data) {
				return fmt.Errorf("交戰功績提督缺失")
			}
			p := data[a.index*rules.AdmiralRecord:]
			binary.LittleEndian.PutUint32(p[2:], binary.LittleEndian.Uint32(p[2:])+uint32(a.amount))
		}
		g.slot.ApplyBattleResult(g.battleUnits)
		// 1261:032B 只刷新受損艦隊；全滅保留 +3C 並設 bit3。
		f := g.battleFleet(g.battleUnits[s.target].Fleet)
		admiral := g.admiralRec(g.fleetAdmiral(f.Index))
		if len(admiral) != rules.AdmiralRecord {
			return fmt.Errorf("交戰完整度提督缺失")
		}
		f.AdmiralLossFactor = int(admiral[0x0b])
		r, err = g.fleetCombatRawFleet(f.Index)
		if err != nil {
			return err
		}
		alive := false
		for _, i := range g.fleetCombatUnits(f.Index) {
			alive = alive || g.battleUnits[i].Ships > 0
		}
		if alive {
			f.Integrity = rules.FleetIntegrity(f, g.battleUnits)
			f.Flag0 = f.Integrity < 50
			r[0] &= ^byte(9)
			if f.Flag0 {
				r[0] |= 1
			}
			r[0x3c] = byte(f.Integrity)
		} else {
			r[0] |= 8
		}
	}
	tick := s.attackEnd
	if g.display.Anim != 0 && h.Lost > 0 {
		pct := h.Lost * 100 / s.before[s.target].Ships
		if pct == 0 {
			pct = 1
		}
		suffix := byte('2')
		if pct < 20 {
			suffix = '0'
		} else if pct < 40 {
			suffix = '1'
		}
		if err := g.addFleetSceneSound(s, s.target, int(suffix-'0'), tick, len(s.pictures), true); err != nil {
			return err
		}
		if err := g.fleetSceneAdd(s, s.target, 'D', suffix, true, &tick); err != nil {
			return err
		}
	}
	s.resultTick = tick
	return nil
}

func (g *Game) updateFleetScene() {
	s := g.fleetCombat.scene
	if s == nil {
		return
	}
	g.processFleetSceneSounds(s)
	if !s.damaged && s.tick >= s.attackEnd {
		if err := g.applyFleetSceneDamage(s); err != nil {
			g.fleetCombatFail(err)
			return
		}
	}
	g.processFleetSceneSounds(s)
	if s.damaged && s.tick >= s.resultTick && (s.click || s.tick >= s.resultTick+g.display.MsgSpeed) {
		g.nextFleetCombatScene()
	}
}

func (g *Game) fleetSceneTick() {
	if s := g.fleetCombat.scene; s != nil {
		s.tick++
		g.updateFleetScene()
	}
}
func (g *Game) fleetSceneClick() {
	if s := g.fleetCombat.scene; s != nil {
		s.click = true
		g.updateFleetScene()
	}
}

func (g *Game) fleetFrameView(s *fleetCombatSceneState, result bool) (screen.FleetFrameView, error) {
	var v screen.FleetFrameView
	units := s.before
	if result {
		units = g.battleUnits
	}
	visibility := screen.FleetVisibility{PlayerSide: uint8(g.faction), SideControl: [2]uint8{0, 1}}
	if g.versus {
		visibility.Mode = 2
		visibility.SideControl = [2]uint8{0, 0}
	}
	v.Scene.Visibility = visibility
	v.Scene.AttackerID = s.own
	v.Scene.TargetID = s.target
	for _, i := range []int{s.own, s.target} {
		u := units[i]
		f := g.battleFleet(u.Fleet)
		if f == nil {
			return v, fmt.Errorf("交戰場景艦隊缺失")
		}
		side := u.Faction
		for _, j := range g.fleetCombatUnits(u.Fleet) {
			x := units[j]
			r, err := g.fleetCombatRecord(j)
			if err != nil {
				return v, err
			}
			weapon := r[0x2e]
			if weapon == 255 {
				weapon = 3
			}
			v.Scene.Sides[side] = append(v.Scene.Sides[side], screen.FleetSceneUnit{FleetProjectionUnit: rules.FleetProjectionUnit{ID: j, Side: uint8(side), X: uint16(x.X), Y: uint16(x.Y)}, Squadron: uint8(x.Squadron), Detail: uint8(x.Detail), Weapon: weapon, Ships: uint16(x.Ships)})
		}
		admiral := g.fleetAdmiral(u.Fleet)
		var name []byte
		if admiral >= 0 && admiral < len(g.res.AdmiralNames) {
			name = codesBig5(g.res.AdmiralNames[admiral])
		}
		if u.Kind < 0 || u.Kind >= len(g.res.ShipKindNames) {
			return v, fmt.Errorf("缺交戰艦種名稱")
		}
		v.Units[side] = screen.FleetUnitView{Side: uint8(side), Squadron: uint8(u.Squadron), Detail: uint8(u.Detail), FleetNumber: uint16(f.Number), AdmiralName: name, KindName: g.res.ShipKindNames[u.Kind], Ships: uint16(u.Ships), ShipsMax: uint16(u.FullShips), Visibility: visibility}
		v.Integrity[side] = s.integrity[side]
		if result {
			v.Integrity[side] = f.Integrity
		}
		v.Scene.Facing[side] = uint8(f.Facing)
	}
	u := units[s.own]
	k, ok := g.res.ShipKinds.Kind(u.Faction, u.Kind)
	if !ok {
		return v, fmt.Errorf("缺交戰艦種表")
	}
	v.Weapon = screen.FleetWeaponView{Side: uint8(u.Faction), Weapon: s.weapon, WeaponTypes: [3]uint8{uint8(k.Byte(10)), uint8(k.Byte(14)), uint8(k.Byte(18))}, ShowResult: result, Destroyed: uint16(s.hit.Lost)}
	for _, p := range s.pictures {
		if p.tick > s.tick {
			break
		}
		if p.frame == 0 {
			v.Pictures = append(v.Pictures, screen.FleetScenePicture{Side: uint8(p.side), Segment: p.segment, LastFrame: 0})
		} else {
			v.Pictures[len(v.Pictures)-1].LastFrame = p.frame
		}
	}
	return v, nil
}

func (g *Game) drawFleetCombatScene() *gfx.Framebuffer {
	s := g.fleetCombat.scene
	if s == nil {
		return gfx.NewFramebuffer(0)
	}
	v, err := g.fleetFrameView(s, s.damaged && s.tick >= s.resultTick)
	if err != nil {
		g.fleetCombatFail(err)
		return gfx.NewFramebuffer(0)
	}
	fb, err := g.res.FleetPresentation.RenderFleetFrame(v, g.res.EngagementPanel, g.res.FleetSceneSprites, g.res.Comani, g.res.Font, g.res.FrameTiles)
	if err != nil {
		g.fleetCombatFail(err)
		return gfx.NewFramebuffer(0)
	}
	return fb
}
