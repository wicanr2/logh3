package game

import (
	"bytes"
	"reflect"
	"slices"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/screen"
)

// Resources are immutable after bootstrap. Cache the complete visual inputs,
// without using the elapsed tick as an image key or changing any timing gate.
type engagementTimelineKey struct {
	bmg                   *asset.BMG
	spec                  *screen.EngagementSpec
	interval              int
	planetKind            byte
	fleetName             [2]byte
	planetSide, fleetSide int
	ships, destroyed      int
}

type engagementTimelineCache struct {
	key      engagementTimelineKey
	pictures []engagementPicture
	result   int
}

type engagementFrameCache struct {
	view     *screen.EngagementView
	pictures []engagementPicture
	frame    *gfx.Framebuffer
}

func (g *Game) activeEngagementPictures() []engagementPicture {
	if len(g.engQueue) == 0 {
		return nil
	}
	pictures, _ := g.engagementTimeline(g.engQueue[0])
	for i, p := range pictures {
		if p.tick > g.engTick {
			return pictures[:i]
		}
	}
	return pictures
}

func (g *Game) cachedEngagementFrame(view *screen.EngagementView, pictures []engagementPicture) *gfx.Framebuffer {
	c := g.engFrameCache
	if c != nil && reflect.DeepEqual(view, c.view) && slices.Equal(pictures, c.pictures) {
		return c.frame
	}
	return nil
}

func (g *Game) saveEngagementFrame(view *screen.EngagementView, pictures []engagementPicture, frame *gfx.Framebuffer) {
	var detached *screen.EngagementView
	if view != nil {
		v := *view
		v.PlanetName, v.AdmiralName = bytes.Clone(v.PlanetName), bytes.Clone(v.AdmiralName)
		if v.Gauges != nil {
			gauges := *v.Gauges
			v.Gauges = &gauges
		}
		detached = &v
	}
	g.engFrameCache = &engagementFrameCache{detached, append([]engagementPicture(nil), pictures...), frame}
}
