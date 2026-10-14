package game

import (
	"fmt"
	"sort"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// Each stage retains its own rule snapshots and the previous stage's visible
// layers. Ownership after capture never determines the side of an earlier hit.
type planetEngagementStage struct {
	scene                   rules.PlanetAssaultScene
	previous                *screen.PlanetFrameView
	base                    screen.PlanetFrameView
	groups                  []planetPictureGroup
	pictures                []engagementPicture
	sounds                  []fleetSceneSound
	lowerTick, resultTick   int
	frame                   *gfx.Framebuffer
	frameApplied            int
	frameResult, frameLower bool
}

type planetPictureGroup struct {
	segment     string
	side        int
	first, last int
}

func (g *Game) queuePlanetEngagements(scenes []rules.PlanetAssaultScene) bool {
	if g.res.Engagement == nil || g.res.Engagement.PlanetFrame == nil || g.res.EngagementPanel == nil || g.res.FleetSceneSprites == nil {
		g.saveErr = fmt.Errorf("缺少118行星分段演出資料")
		return false
	}
	g.engQueue = nil
	for _, scene := range scenes {
		b := scene.Before
		e := engagement{planet: scene.PlanetIndex, fleet: b.Fleet.Index,
			defence: b.Planet.Defence, defenceMax: b.Planet.DefenceMax,
			destroyed: scene.Destroyed, planetSide: b.Planet.Owner, fleetSide: b.Fleet.Faction,
			stage: &planetEngagementStage{scene: scene, frameApplied: -1}}
		g.fillEngagementNames(&e)
		// fillEngagementNames serves historical diagnostic fixtures; production
		// sides come from the detached snapshot, not the final planet table.
		e.fleetSide = b.Fleet.Faction
		e.planetSide = b.Planet.Owner
		if e.planetSide != 0 && e.planetSide != 1 {
			e.planetSide = 1 - e.fleetSide
		}
		e.ships, e.shipsMax = snapshotShips(b)
		e.label = screen.EngagementDefence
		if scene.Kind == rules.PlanetLandingScene {
			e.label = screen.EngagementLanding
		}
		if scene.Kind == rules.PlanetCaptureScene {
			e.label = screen.EngagementFailed
			if scene.Captured {
				e.label = screen.EngagementCaptured
			}
		}
		g.engQueue = append(g.engQueue, e)
	}
	if len(g.engQueue) == 0 {
		return false
	}
	g.startEngagement()
	g.state = StateEngagement
	return true
}

func snapshotShips(s rules.PlanetAssaultSnapshot) (ships, maximum int) {
	for _, u := range s.Units {
		if u.Fleet == s.Fleet.Index && u.Squadron >= 0 {
			ships += u.Ships
			maximum += u.FullShips
		}
	}
	return
}

func (g *Game) planetSnapshotView(s rules.PlanetAssaultSnapshot, planet int, landing bool) screen.PlanetFrameView {
	visibility := screen.FleetVisibility{PlayerSide: uint8(g.faction), SideControl: [2]uint8{0, 1}}
	if g.versus {
		visibility.Mode = 2
		visibility.SideControl = [2]uint8{0, 0}
	}
	kind := 0
	if s.Planet.Fortress {
		kind = 1
	}
	if s.Planet.Fortress && g.battleRegion == g.res.Engagement.Animation.FortressRegion {
		kind = 2
	}
	ships, maximum := snapshotShips(s)
	v := screen.PlanetFrameView{
		Planet: screen.PlanetTextView{Owner: s.Planet.Owner, Kind: kind, Name: g.battlePlanetName(planet), Defence: s.Planet.Defence, DefenceMax: s.Planet.DefenceMax},
		Fleet:  screen.PlanetFleetTextView{Side: uint8(s.Fleet.Faction), Number: uint16(s.Fleet.Number), Landing: landing, Attack: uint16(s.Attack), Ships: uint16(ships), ShipsMax: uint16(maximum), Visibility: visibility},
		Gauges: screen.EngagementGaugePair{FleetFaction: s.Fleet.Faction, Defence: s.Planet.Defence, Integrity: s.Integrity, Attack: s.Attack},
	}
	if admiral := g.fleetAdmiral(s.Fleet.Index); admiral >= 0 && admiral < len(g.res.AdmiralNames) {
		v.Fleet.AdmiralName = codesBig5(g.res.AdmiralNames[admiral])
	}
	for _, u := range s.Units {
		if u.Fleet == s.Fleet.Index && u.Squadron == 0 {
			v.Fleet.Detail = uint8(u.Detail)
			break
		}
	}
	return v
}

func (g *Game) planetLowerSnapshot(s rules.PlanetAssaultSnapshot) *screen.PlanetSceneView {
	v := &screen.PlanetSceneView{Visibility: screen.FleetVisibility{PlayerSide: uint8(g.faction), SideControl: [2]uint8{0, 1}}, Facing: uint8(s.Fleet.Facing), X: uint16(s.Planet.X), Y: uint16(s.Planet.Y), Owner: s.Planet.Owner, Defence: uint16(s.Planet.Defence)}
	if g.versus {
		v.Visibility.Mode = 2
		v.Visibility.SideControl = [2]uint8{0, 0}
	}
	for i, u := range s.Units {
		if u.Fleet != s.Fleet.Index || u.Squadron < 0 {
			continue
		}
		v.Units = append(v.Units, screen.FleetSceneUnit{FleetProjectionUnit: rules.FleetProjectionUnit{ID: i, Side: uint8(u.Faction), X: uint16(u.X), Y: uint16(u.Y)}, Squadron: uint8(u.Squadron), Detail: uint8(u.Detail), Ships: uint16(u.Ships)})
	}
	sort.SliceStable(v.Units, func(i, j int) bool { return v.Units[i].Squadron < v.Units[j].Squadron })
	return v
}

func (g *Game) startPlanetEngagement() {
	e := &g.engQueue[0]
	s := e.stage
	if s.scene.Kind == rules.PlanetCaptureScene {
		if s.previous != nil {
			s.base = *s.previous
		}
		s.base.Middle = screen.PlanetMiddleView{Label: e.label, LabelColour: 4, ResultColour: 4}
		s.lowerTick = -1
		return
	}
	s.base = g.planetSnapshotView(s.scene.Before, e.planet, s.scene.Kind == rules.PlanetLandingScene)
	colours := g.res.Engagement.PlanetFrame.Colours
	attacker, target := s.scene.Before.Planet.Owner, s.scene.Before.Fleet.Faction
	if s.scene.Kind == rules.PlanetLandingScene {
		attacker, target = target, attacker
	}
	s.base.Middle = screen.PlanetMiddleView{Label: e.label, LabelColour: colours[attacker], ResultColour: colours[target], Destroyed: uint16(s.scene.Destroyed)}
	if s.previous != nil {
		s.base.Pictures = append([]screen.FleetScenePicture(nil), s.previous.Pictures...)
		s.base.Scene = s.previous.Scene
	}
	interval := g.fleetSceneInterval()
	tick := 0
	add := func(segment string, side int, full bool, cue int) {
		if g.display.Anim == 0 {
			return
		}
		frames, err := g.res.Comani.Frames(segment)
		if err != nil || len(frames) == 0 {
			g.saveErr = fmt.Errorf("行星演出圖段%X：%w", []byte(segment), err)
			return
		}
		n := 1
		if full {
			n = len(frames)
		}
		first := len(s.pictures)
		for i := 0; i < n; i++ {
			tick += interval
			s.pictures = append(s.pictures, engagementPicture{segment: segment, side: side, frame: i, tick: tick})
		}
		s.groups = append(s.groups, planetPictureGroup{segment, side, first, len(s.pictures)})
		if full && cue >= 0 && g.soundEffects.enabled {
			s.sounds = append(s.sounds, fleetSceneSound{id: cue, tick: s.pictures[first].tick, pictureAt: first})
		}
	}
	planetPrefix := string([]byte{'P', e.planetKind})
	fleetPrefix := string(e.fleetName[:])
	if s.scene.Kind == rules.PlanetLandingScene {
		landingPrefix := string([]byte{"EF"[e.fleetSide], 'L'})
		add(landingPrefix+"AB", e.fleetSide, false, -1)
		add(planetPrefix+"D0", e.planetSide, false, -1)
		s.lowerTick = tick
		add(landingPrefix+"AB", e.fleetSide, true, 22) // 20B4:0681 record+6, RE436/437.
		suffix, cue := "D0", 3
		if s.scene.Captured {
			suffix, cue = "D2", 2
		}
		add(planetPrefix+suffix, e.planetSide, true, cue)
	} else {
		add(planetPrefix+"D0", e.planetSide, false, -1)
		add(fleetPrefix+"D0", e.fleetSide, false, -1)
		s.lowerTick = tick
		add(planetPrefix+"AB", e.planetSide, true, 17) // 20B4:08B1, RE437 empire/counterfirst.
		pct := 0
		if e.ships > 0 {
			pct = s.scene.Destroyed * 100 / e.ships
			if pct == 0 && s.scene.Destroyed > 0 {
				pct = 1
			}
		}
		column := 0
		if pct >= 40 {
			column = 2
		} else if pct >= 20 {
			column = 1
		}
		cue := -1
		if pct > 0 && g.soundEffects.enabled {
			var err error
			cue, err = g.res.FleetSound.Cue(e.fleetSide, 7, column, true, true)
			if err != nil {
				g.saveErr = err
			}
		}
		add(fleetPrefix+fmt.Sprintf("D%d", column), e.fleetSide, pct > 0, cue)
	}
	s.resultTick = tick
	g.processPlanetSounds(s)
}

// Use the same completed-audio gate as fleet combat. Delaying a cue also delays
// its first frame and the remaining schedule, including the result boundary.
func (g *Game) processPlanetSounds(s *planetEngagementStage) {
	for i := range s.sounds {
		cue := &s.sounds[i]
		if cue.played || cue.tick > g.engTick {
			continue
		}
		if g.soundEffects.remaining > 0 || g.soundEffects.playing {
			d := g.engagementTickDuration()
			delay := max(1, int((g.soundEffects.remaining+d-1)/d))
			for p := cue.pictureAt; p < len(s.pictures); p++ {
				s.pictures[p].tick += delay
			}
			// Both static pictures and the lower snapshot precede every cue.
			// At interval 0 they share a tick but keep that event order.
			for j := i; j < len(s.sounds); j++ {
				s.sounds[j].tick += delay
			}
			s.resultTick += delay
			continue
		}
		cue.played = true
		g.soundEffects.events = append(g.soundEffects.events, cue.id)
		g.soundEffects.remaining = g.res.SoundSamples.Samples[cue.id].Duration()
	}
}

func (g *Game) planetEngagementView(s *planetEngagementStage) screen.PlanetFrameView {
	v := s.base
	if s.scene.Kind == rules.PlanetCaptureScene {
		return v
	}
	v.Pictures = append([]screen.FleetScenePicture(nil), s.base.Pictures...)
	for _, group := range s.groups {
		last := -1
		for i := group.first; i < group.last; i++ {
			if s.pictures[i].tick > g.engTick {
				break
			}
			last = s.pictures[i].frame
		}
		if last >= 0 {
			v.Pictures = append(v.Pictures, screen.FleetScenePicture{Side: uint8(group.side), Segment: group.segment, LastFrame: last})
		}
	}
	if g.engTick >= s.lowerTick {
		v.Scene = g.planetLowerSnapshot(s.scene.Before)
	}
	if g.engTick >= s.resultTick {
		after := g.planetSnapshotView(s.scene.After, s.scene.PlanetIndex, s.scene.Kind == rules.PlanetLandingScene)
		v.Planet.Defence = after.Planet.Defence
		v.Fleet.Ships, v.Fleet.Attack = after.Fleet.Ships, after.Fleet.Attack
		v.Gauges = after.Gauges
		v.Middle.ShowResult = true
		v.Scene = g.planetLowerSnapshot(s.scene.After)
	}
	return v
}

func (g *Game) drawPlanetEngagement(s *planetEngagementStage) *gfx.Framebuffer {
	applied := 0
	for _, p := range s.pictures {
		if p.tick <= g.engTick {
			applied++
		}
	}
	result, lower := g.engTick >= s.resultTick, g.engTick >= s.lowerTick
	if s.frame != nil && s.frameApplied == applied && s.frameResult == result && s.frameLower == lower {
		return s.frame
	}
	v := g.planetEngagementView(s)
	fb, err := g.res.Engagement.RenderPlanetFrame(v, g.res.EngagementPanel, g.res.FleetSceneSprites, g.res.Comani, g.res.Font, g.res.FrameTiles)
	if err != nil {
		g.saveErr = err
		return gfx.NewFramebuffer(0)
	}
	s.frame, s.frameApplied, s.frameResult, s.frameLower = fb, applied, result, lower
	return fb
}
