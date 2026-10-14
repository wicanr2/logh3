package game

import (
	"fmt"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

type fortressEngagementStage struct {
	scene       rules.FortressGunScene
	previous    *screen.FortressFrameView
	base        screen.FortressFrameView
	schedule    planetEngagementStage
	noticePhase int
}

func (g *Game) queueFortressEngagements(scenes []rules.FortressGunScene) bool {
	if len(scenes) == 0 || g.res.Engagement == nil || g.res.EngagementPanel == nil {
		return false
	}
	if g.res.Engagement.FortressFrame == nil || g.res.FleetPresentation == nil || g.res.FleetSceneSprites == nil {
		g.saveErr = fmt.Errorf("缺144要塞演出資料")
		return false
	}
	g.engQueue = nil
	for _, s := range scenes {
		g.engQueue = append(g.engQueue, engagement{planet: s.Planet, fleet: s.Before.Fleet.Index,
			gun: &fortressEngagementStage{scene: s, schedule: planetEngagementStage{frameApplied: -1}}})
	}
	g.fortressEngagement = true
	g.state = StateEngagement
	g.startEngagement()
	return true
}

func (g *Game) currentFortressNotice() *fortressEngagementStage {
	if !g.fortressEngagement || g.state != StateBattle || len(g.engQueue) == 0 {
		return nil
	}
	s := g.engQueue[0].gun
	if s == nil || s.scene.Unit >= 0 {
		return nil
	}
	return s
}

func (g *Game) startFortressEngagement() {
	s := g.engQueue[0].gun
	if s.scene.Unit < 0 {
		g.state = StateBattle
		s.schedule.resultTick = max(1, g.display.MsgSpeed)
		return
	}
	b := s.scene.Before
	u := b.Units[s.scene.Unit]
	planet := g.planetSnapshotView(b, s.scene.Planet, false).Planet
	visibility := screen.FleetVisibility{PlayerSide: uint8(g.faction), SideControl: [2]uint8{0, 1}}
	if g.versus {
		visibility.Mode = 2
		visibility.SideControl = [2]uint8{0, 0}
	}
	var name []byte
	if a := g.fleetAdmiral(u.Fleet); a >= 0 && a < len(g.res.AdmiralNames) {
		name = codesBig5(g.res.AdmiralNames[a])
	}
	if u.Kind < 0 || u.Kind >= len(g.res.ShipKindNames) {
		g.saveErr = fmt.Errorf("缺要塞目標艦種名稱")
		return
	}
	s.base = screen.FortressFrameView{Planet: planet,
		Target: screen.FleetUnitView{Side: uint8(u.Faction), Squadron: uint8(u.Squadron), Detail: uint8(u.Detail), FleetNumber: uint16(b.Fleet.Number),
			AdmiralName: name, KindName: g.res.ShipKindNames[u.Kind], Ships: uint16(u.Ships), ShipsMax: uint16(u.FullShips), Visibility: visibility},
		Gauges: screen.EngagementGaugePair{FleetFaction: u.Faction, Integrity: b.Integrity, Defence: b.Planet.Defence, Attack: b.Attack}, Destroyed: uint16(s.scene.Hit.Lost)}
	if s.previous != nil {
		s.base.Scene = s.previous.Scene
		s.base.ClearSides = append([]uint8(nil), s.previous.ClearSides...)
		oldSide := s.previous.Target.Side
		found := false
		for _, side := range s.base.ClearSides {
			found = found || side == oldSide
		}
		if !found {
			s.base.ClearSides = append(s.base.ClearSides, oldSide)
		}
		for _, p := range s.previous.Pictures {
			if p.Side != oldSide {
				s.base.Pictures = append(s.base.Pictures, p)
			}
		}
	}
	interval, tick := g.fleetSceneInterval(), 0
	add := func(seg string, side int, all bool, cue int) {
		if g.display.Anim == 0 {
			return
		}
		if g.res.Comani == nil {
			g.saveErr = fmt.Errorf("缺要塞動畫庫")
			return
		}
		frames, err := g.res.Comani.Frames(seg)
		if err != nil || len(frames) == 0 {
			g.saveErr = fmt.Errorf("要塞圖段%X：%v", []byte(seg), err)
			return
		}
		n := 1
		if all {
			n = len(frames)
		}
		first := len(s.schedule.pictures)
		cueTick := tick // 3958在第一格等待前呼叫音效。
		for i := 0; i < n; i++ {
			tick += interval
			s.schedule.pictures = append(s.schedule.pictures, engagementPicture{segment: seg, side: side, frame: i, tick: tick})
		}
		s.schedule.groups = append(s.schedule.groups, planetPictureGroup{seg, side, first, len(s.schedule.pictures)})
		if all && cue >= 0 && g.soundEffects.enabled {
			s.schedule.sounds = append(s.schedule.sounds, fleetSceneSound{id: cue, tick: cueTick, pictureAt: first})
		}
	}
	kind := byte('F')
	if planet.Kind == 2 {
		kind = 'Z'
	}
	prefix := string([]byte{'P', kind})
	target, err := g.fleetSceneName(s.scene.Unit)
	if err != nil {
		g.saveErr = err
		return
	}
	add(prefix+"D0", 1-u.Faction, false, -1)
	add(string(target[:])+"D0", u.Faction, false, -1)
	s.schedule.lowerTick = tick
	if s.scene.First {
		suffix := "M"
		if 1-u.Faction == 1 {
			suffix = "F"
		}
		cue := g.res.Engagement.FortressFrame.Cues[b.Planet.Owner*3+planet.Kind]
		add(prefix+"A"+suffix, 1-u.Faction, true, cue)
	}
	column := 0
	if u.Ships > 0 {
		pct := s.scene.Hit.Lost * 100 / u.Ships
		if pct >= 40 {
			column = 2
		} else if pct >= 20 {
			column = 1
		}
	}
	cue := -1
	if s.scene.Hit.Lost > 0 && g.soundEffects.enabled {
		cue, err = g.res.FleetSound.Cue(u.Faction, u.Kind, column, u.Squadron == 0, true)
		if err != nil {
			g.saveErr = err
		}
	}
	add(string(target[:])+fmt.Sprintf("D%d", column), u.Faction, s.scene.Hit.Lost > 0, cue)
	s.schedule.resultTick = tick
	g.processPlanetSounds(&s.schedule)
}

func (g *Game) fortressEngagementView(s *fortressEngagementStage) screen.FortressFrameView {
	v := s.base
	v.Pictures = append([]screen.FleetScenePicture(nil), s.base.Pictures...)
	for _, group := range s.schedule.groups {
		last := -1
		for i := group.first; i < group.last; i++ {
			if s.schedule.pictures[i].tick > g.engTick {
				break
			}
			last = s.schedule.pictures[i].frame
		}
		if last >= 0 {
			v.Pictures = append(v.Pictures, screen.FleetScenePicture{Side: uint8(group.side), Segment: group.segment, LastFrame: last})
		}
	}
	if g.engTick >= s.schedule.lowerTick {
		v.Scene = g.planetLowerSnapshot(s.scene.Before)
	}
	if g.engTick >= s.schedule.resultTick {
		v.Target.Ships = uint16(s.scene.After.Units[s.scene.Unit].Ships)
		v.Gauges.Integrity = s.scene.After.Integrity
		v.ShowResult = true
		v.Scene = g.planetLowerSnapshot(s.scene.After)
	}
	return v
}

func (g *Game) drawFortressEngagement(s *fortressEngagementStage) *gfx.Framebuffer {
	applied := 0
	for _, p := range s.schedule.pictures {
		if p.tick <= g.engTick {
			applied++
		}
	}
	result, lower := g.engTick >= s.schedule.resultTick, g.engTick >= s.schedule.lowerTick
	c := &s.schedule
	if c.frame != nil && c.frameApplied == applied && c.frameResult == result && c.frameLower == lower {
		return c.frame
	}
	fb, err := g.res.Engagement.RenderFortressFrame(g.fortressEngagementView(s), g.res.FleetPresentation,
		g.res.EngagementPanel, g.res.FleetSceneSprites, g.res.Comani, g.res.Font, g.res.FrameTiles)
	if err != nil {
		g.saveErr = err
		return gfx.NewFramebuffer(0)
	}
	c.frame, c.frameApplied, c.frameResult, c.frameLower = fb, applied, result, lower
	return fb
}

func (g *Game) fortressNoticeMaybeEnd() {
	s := g.currentFortressNotice()
	if s == nil {
		return
	}
	if !g.engClick && g.engTick < s.schedule.resultTick {
		return
	}
	if s.noticePhase == 0 {
		s.noticePhase = 1
		g.engTick = 0
		g.engClick = false
		return
	}
	g.finishEngagement()
}

func (g *Game) fortressNoticeClick() { g.engClick = true; g.fortressNoticeMaybeEnd() }

func (g *Game) drawFortressNotice(fb *gfx.Framebuffer) {
	s := g.currentFortressNotice()
	if s == nil {
		return
	}
	p := s.scene.Before.Planet
	if s.noticePhase == 0 {
		g.res.Engagement.DrawFortressNotice(fb, g.res.Font, g.res.FrameTiles, p.Owner, g.battlePlanetName(s.scene.Planet))
		return
	}
	a, b := rules.FortressRange(p.DefenceMax)
	screen.DrawFortressRange(g.res.Battle, fb, p.X+fortressCentreBias, p.Y+fortressCentreBias, a, b, p.GunAim, g.battleView)
}

type FortressPlaybackObservation struct {
	Planet, Unit, Tick, ResultTick, NoticePhase int
	First, Result                               bool
	Lost, ShipsBefore, ShipsAfter               int
}

func (g *Game) FortressPlaybackObservation() *FortressPlaybackObservation {
	if !g.fortressEngagement || len(g.engQueue) == 0 || g.engQueue[0].gun == nil {
		return nil
	}
	s := g.engQueue[0].gun
	v := &FortressPlaybackObservation{Planet: s.scene.Planet, Unit: s.scene.Unit, Tick: g.engTick, ResultTick: s.schedule.resultTick, NoticePhase: s.noticePhase, First: s.scene.First, Result: g.engTick >= s.schedule.resultTick, Lost: s.scene.Hit.Lost}
	if s.scene.Unit >= 0 {
		v.ShipsBefore = s.scene.Before.Units[s.scene.Unit].Ships
		v.ShipsAfter = s.scene.After.Units[s.scene.Unit].Ships
	}
	return v
}
