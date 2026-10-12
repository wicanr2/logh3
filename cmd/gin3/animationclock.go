package main

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Ebitengine drops catch-up updates after long drawing stalls. The native
// combat timer must still receive real elapsed time, rather than slow down
// with HD rendering. Scripted probes retain their deterministic fixed step.
func (a *app) animationStep(now time.Time, engagement bool) time.Duration {
	fixed := time.Second / time.Duration(ebiten.TPS())
	if a.script != "" || !engagement {
		a.animationTime = time.Time{}
		return fixed
	}
	if a.animationTime.IsZero() {
		a.animationTime = now
		return fixed
	}
	elapsed := now.Sub(a.animationTime)
	a.animationTime = now
	return elapsed
}
