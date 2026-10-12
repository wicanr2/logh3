package game

import (
	"bytes"
	"reflect"
	"testing"
)

func TestPlanetEngagementCachePreservesEveryTick(t *testing.T) {
	g := reachEngagementFixture(t)
	_, result := g.engagementTimeline(g.engQueue[0])
	for tick := 0; tick <= result+2; tick++ {
		g.engTick = tick // Diagnostic sampling, not a normal-player receipt.
		cached := g.drawEngagement()
		if g.drawEngagement() != cached {
			t.Fatalf("tick %d did not reuse identical visual input", tick)
		}
		g.engFrameCache = nil
		fresh := g.drawEngagement()
		if !bytes.Equal(cached.Pix, fresh.Pix) || !reflect.DeepEqual(cached.Commands, fresh.Commands) || !reflect.DeepEqual(cached.Artwork, fresh.Artwork) {
			t.Fatalf("tick %d changed after reconstruction", tick)
		}
	}
	// A changed result, name, or gauge must never reuse an old frame.
	before := g.drawEngagement()
	g.engQueue[0].destroyed++
	if g.drawEngagement() == before {
		t.Fatal("changed result did not invalidate the frame")
	}
	old, _ := g.engagementTimeline(g.engQueue[0])
	g.display.AnimSpeed = 1
	changed, _ := g.engagementTimeline(g.engQueue[0])
	if reflect.DeepEqual(old, changed) {
		t.Fatal("changed animation interval reused old timeline")
	}
	g.startEngagement()
	if g.engFrameCache != nil || g.engTimelineCache != nil {
		t.Fatal("new scene retained old caches")
	}
	t.Logf("checked every tick from 0 through %d against fresh native reconstruction", result+2)
}
