package game

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

func TestPlanetAnimationBasesFollowDisplayedSides(t *testing.T) {
	for _, faction := range []int{0, 1} {
		g := captureSceneGame(t, faction)
		// A previous surface must also be discarded when the later landing
		// craft has no D0 base; keep the other side's own base intact.
		g.engQueue[0].stage.base.Pictures = []screen.FleetScenePicture{{Side: uint8(faction), Segment: "PPD0", LastFrame: 0}}
		for _, tick := range []int{2 * g.fleetSceneInterval(), g.engQueue[0].stage.resultTick} {
			g.engTick = tick
			view := g.planetEngagementView(g.engQueue[0].stage)
			var prefixes [2]string
			for _, p := range view.Pictures {
				prefixes[p.Side] = p.Segment[:2]
			}
			seen := 0
			for _, c := range g.drawEngagement().Commands {
				if c.Kind != gfx.DrawAnimation {
					continue
				}
				side := 0
				if c.X == 376 {
					side = 1
				}
				base, err := g.res.Comani.StaticPicture(prefixes[side])
				if err != nil {
					// Landing craft have AB but no D0 segment. Their own
					// fallback must not borrow the opposite planet's base.
					if c.Key != "" || len(c.Base) != 0 {
						t.Fatal("missing D0 borrowed an opposite-side base", side, c.Key)
					}
					seen++
					continue
				}
				if c.Key != base.ArtworkKey || !bytes.Equal(c.Base, base.Pix) {
					t.Fatalf("faction%d tick%d side%d: wrong HD base %q for displayed prefix %X", faction, tick, side, c.Key, prefixes[side])
				}
				seen++
			}
			if seen != 2 {
				t.Fatal("both animation windows must retain their own base", seen)
			}
		}
	}
}

// Diagnostic rule-to-renderer coverage; normal GUI evidence is recorded
// separately in playtest121 and is not inferred from this injected state.
func captureSceneGame(t *testing.T, faction int) *Game {
	t.Helper()
	g := reachDashiliBattle(t)
	g.faction = faction
	for i := range g.battlePlanets {
		g.battlePlanets[i].Owner = faction
	}
	p := &g.battlePlanets[0]
	p.Owner = 1 - faction
	p.Defence, p.DefenceMax = 1, 1
	chosen := -1
	for i := range g.battleFleets {
		f := &g.battleFleets[i]
		f.X, f.Y = 10000, 10000
		if chosen < 0 && f.Faction == faction {
			chosen = i
			f.X, f.Y = p.X, p.Y
			f.Action = 1
			f.Integrity = 100
			f.Ability3 = 100
		}
	}
	if chosen < 0 {
		t.Fatal("missing faction fleet")
	}
	g.battleRand = rules.NewRand(0x12345678)
	hits := g.planetAssault()
	if len(hits) != 1 || !hits[0].Captured {
		t.Fatalf("diagnostic capture did not occur: %+v", hits)
	}
	if !g.queueEngagements(hits) || len(g.engQueue) != 2 {
		t.Fatalf("missing landing/capture stages: %d, %v", len(g.engQueue), g.saveErr)
	}
	return g
}

func TestPlanetSoundWaitUsesOneClockAndKeepsStaticLower(t *testing.T) {
	g := captureSceneGame(t, 1)
	fleetSoundAssets(t, g)
	g.display.AnimSpeed = 2
	g.soundEffects.remaining = 2 * time.Second
	g.queuePlanetEngagements(g.battleAssaultScenes)
	s := g.engQueue[0].stage
	if len(g.DrainSoundEffects()) != 0 || g.planetEngagementView(s).Scene == nil {
		t.Fatal("fastest mode must draw static lower before waiting for audio")
	}
	elapsed := time.Duration(0)
	for elapsed+g.TimerTick() < 2*time.Second {
		g.Tick(g.TimerTick())
		elapsed += g.TimerTick()
		if len(g.DrainSoundEffects()) != 0 {
			t.Fatal("audio clock advanced twice or cue played before previous sample ended")
		}
	}
	g.SetSoundEffectPlaying(true)
	for i := 0; i < 4; i++ {
		g.Tick(g.TimerTick())
		if len(g.DrainSoundEffects()) != 0 {
			t.Fatal("backend is still playing")
		}
	}
	g.SetSoundEffectPlaying(false)
	var ids []int
	for i := 0; i < 4 && len(ids) == 0; i++ {
		g.Tick(g.TimerTick())
		ids = g.DrainSoundEffects()
	}
	if len(ids) != 1 || ids[0] != 22 {
		t.Fatal("landing cue did not resume exactly once", ids)
	}
}

func TestPlanetAnimationOffRetainsLowerResultAndIndependentWait(t *testing.T) {
	g := captureSceneGame(t, 0)
	g.display.Anim = 0
	g.queuePlanetEngagements(g.battleAssaultScenes)
	s := g.engQueue[0].stage
	v := g.planetEngagementView(s)
	if len(s.pictures) != 0 || len(s.sounds) != 0 || v.Scene == nil || !v.Middle.ShowResult || v.Planet.Owner != 1 {
		t.Fatal("animation off removed rule result/lower or changed owner")
	}
	g.drawEngagement()
	if g.saveErr != nil {
		t.Fatal(g.saveErr)
	}
	g.engagementClick()
	if len(g.engQueue) != 1 || g.engQueue[0].label != screen.EngagementCaptured || g.engClick {
		t.Fatal("landing result click must leave capture waiting")
	}
}

func TestPlanetScenePlayerChainKeepsOwnerAndConsumesOneClick(t *testing.T) {
	for _, faction := range []int{0, 1} {
		t.Run(fmt.Sprint(faction), func(t *testing.T) {
			g := captureSceneGame(t, faction)
			if g.engQueue[0].label != screen.EngagementLanding {
				t.Fatal("landing replaced by final capture label")
			}
			v := g.planetEngagementView(g.engQueue[0].stage)
			if v.Planet.Owner != 1-faction || int(v.Fleet.Side) != faction || !v.Fleet.Landing {
				t.Fatal("final owner leaked into pre-capture view")
			}
			g.drawEngagement()
			if g.saveErr != nil {
				t.Fatal(g.saveErr)
			}
			_, result := g.engagementTimeline(g.engQueue[0])
			g.engTick = result - 1
			g.engagementClick()
			if len(g.engQueue) != 2 {
				t.Fatal("click ended animation early")
			}
			g.engagementTick(1)
			if len(g.engQueue) != 1 || g.engClick || g.engQueue[0].label != screen.EngagementCaptured {
				t.Fatal("click skipped more than one result wait")
			}
			v = g.planetEngagementView(g.engQueue[0].stage)
			if v.Middle.ShowResult || v.Middle.Label != screen.EngagementCaptured || v.Planet.Owner != 1-faction {
				t.Fatal("capture result changed preserved layers or kept destroyed row")
			}
			g.drawEngagement()
			if g.saveErr != nil {
				t.Fatal(g.saveErr)
			}
			g.engagementTick(1)
			if len(g.engQueue) != 1 {
				t.Fatal("queued click also shortened capture wait")
			}
		})
	}
}
