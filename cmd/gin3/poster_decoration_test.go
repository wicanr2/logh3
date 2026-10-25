package main

import (
	"image"
	"image/color"
	"testing"
	"time"
)

func TestPosterBlackBandsActualContentAndCaps(t *testing.T) {
	for _, scale := range []int{1, 3} {
		w, h := 640*scale, 480*scale
		p := make([]byte, w*h*4)
		for y := 40 * scale; y < 416*scale; y++ {
			for x := 0; x < w; x++ {
				p[(y*w+x)*4] = 1
			}
		}
		b := posterBlackBands(p, w*4, w, h)
		want := [2]image.Rectangle{image.Rect(0, 0, w, 40*scale), image.Rect(0, 416*scale, w, h)}
		if b != want {
			t.Fatalf("scale %d: %v want %v", scale, b, want)
		}
		p[(12*scale*w+15)*4+2] = 1
		b = posterBlackBands(p, w*4, w, h)
		if b[0].Max.Y != 12*scale {
			t.Fatalf("painted content covered: %v", b)
		}
		clear(p)
		if got := posterBlackBands(p, w*4, w, h); got != [2]image.Rectangle{} {
			t.Fatalf("empty/cinematic frame decorated: %v", got)
		}
		p[(240*scale*w+15)*4] = 1
		b = posterBlackBands(p, w*4, w, h)
		if b[0].Dy() != 48*scale || b[1].Dy() != 80*scale {
			t.Fatalf("unbounded decoration: %v", b)
		}
		p[0] = 1
		p[len(p)-4] = 1
		if b = posterBlackBands(p, w*4, w, h); b != [2]image.Rectangle{} {
			t.Fatalf("full-height frame decorated: %v", b)
		}
	}
	if b := posterBlackBands([]byte{0}, 1, 640, 480); b != [2]image.Rectangle{} {
		t.Fatal(b)
	}
}

func TestPosterFlagshipKeysUniqueAndYangFirst(t *testing.T) {
	codes := make([]byte, 80)
	for i := 0; i < 80; i++ {
		codes[i] = 0xdd
	}
	codes[0] = 0xb1
	codes[1] = 0xb2
	codes[40+12] = 0xbd
	keys := posterFlagshipKeys(codes)
	if len(keys[0]) != 3 || len(keys[1]) != 2 || keys[0][0] != "45B1" || keys[1][0] != "46BD" {
		t.Fatal(keys)
	}
	if got := posterFlagshipKeys(nil); len(got[0])+len(got[1]) != 0 {
		t.Fatal(got)
	}
}

func TestPosterDecorationClockAndPages(t *testing.T) {
	for frame := 0; frame < 180*8; frame++ {
		elapsed := time.Duration(frame)*time.Second/15 + time.Nanosecond
		phase, page := posterDecorationPhase(elapsed, true)
		if phase != frame%180 || page != frame/180 {
			t.Fatal(frame, phase, page)
		}
		if a, b := posterDecorationPhase(elapsed, false); a != 0 || b != 0 {
			t.Fatal("disabled animation moves")
		}
	}
	for _, count := range []int{12, 16} {
		seen := map[int]bool{}
		for page := 0; page < 4; page++ {
			for slot := 0; slot < 8; slot++ {
				i := posterDecorationIndex(page, slot, count)
				if i < 0 || i >= count {
					t.Fatal(i)
				}
				seen[i] = true
			}
		}
		if len(seen) != count {
			t.Fatal("some flagships never displayed", count, seen)
		}
	}
	if posterDecorationIndex(0, 0, 0) != -1 || posterDecorationIndex(-1, 0, 12) != 4 {
		t.Fatal("invalid empty or previous-page index")
	}
}

func TestPosterFighterCropKeepsTwoCellsSeparate(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 200, 100))
	im.SetNRGBA(30, 40, color.NRGBA{200, 200, 200, 255})
	im.SetNRGBA(160, 70, color.NRGBA{100, 160, 100, 255})
	for side, point := range []image.Point{image.Pt(30, 40), image.Pt(160, 70)} {
		b := posterFighterRect(im, side)
		if !point.In(b) || !b.In(image.Rect(side*100, 0, (side+1)*100, 100)) {
			t.Fatal(side, b)
		}
	}
}
