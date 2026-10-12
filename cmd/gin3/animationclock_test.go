package main

import (
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNativeAnimationClockSurvivesDrawingStall(t *testing.T) {
	a := &app{}
	start := time.Unix(123, 0)
	fixed := time.Second / time.Duration(ebiten.TPS())
	if a.animationStep(start, true) != fixed {
		t.Fatal("first combat update did not reset its clock")
	}
	if a.animationStep(start.Add(137*time.Millisecond), true) != 137*time.Millisecond {
		t.Fatal("rendering stall slowed native timer")
	}
	if a.animationStep(start.Add(137*time.Millisecond), true) != 0 {
		t.Fatal("catch-up update counted elapsed time twice")
	}
	if a.animationStep(start.Add(142*time.Millisecond), true) != 5*time.Millisecond {
		t.Fatal("normal elapsed time changed")
	}
	if a.animationStep(start.Add(time.Hour), false) != fixed || !a.animationTime.IsZero() {
		t.Fatal("leaving combat retained elapsed time")
	}
	if a.animationStep(start.Add(2*time.Hour), true) != fixed {
		t.Fatal("re-entering combat included time from another screen")
	}
}

func TestScriptAnimationClockKeepsFixedSteps(t *testing.T) {
	a := &app{script: "click 1 2"}
	for _, now := range []time.Time{time.Unix(1, 0), time.Unix(3600, 0)} {
		if a.animationStep(now, true) != time.Second/time.Duration(ebiten.TPS()) {
			t.Fatal("script timing became dependent on the wall clock")
		}
	}
}
