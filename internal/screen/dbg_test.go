package screen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

func TestDbgNumbers(t *testing.T) {
	root := "../.."
	s, err := LoadMainGameSpec(filepath.Join(root, "docs/internals/data/screen-build-plan.json"))
	if err != nil {
		t.Skip(err)
	}
	if _, err := os.Stat(filepath.Join(root, "org_game/extracted/GRAPH.IMG")); err != nil {
		t.Skip("no font")
	}
	fnt, _ := font.Load(filepath.Join(root, "org_game/extracted/GRAPH.IMG"))
	fb, _ := MainGame(s, fnt, nil, nil, nil)
	before := append([]uint8(nil), fb.Pix...)
	s.DrawNumbers(fb, fnt, []int{20, 10, 18, 10, 12, 15, 15},
		[]int{3037, 1518, 2733, 1518, 1822, 2277, 2277}, 0, 36)
	minx, miny, maxx, maxy := 9999, 9999, -1, -1
	for i := range before {
		if before[i] != fb.Pix[i] {
			x, y := i%gfx.Width, i/gfx.Width
			if x < minx {
				minx = x
			}
			if y < miny {
				miny = y
			}
			if x > maxx {
				maxx = x
			}
			if y > maxy {
				maxy = y
			}
		}
	}
	t.Logf("差異 bbox x %d..%d y %d..%d", minx, maxx, miny, maxy)
	for y := miny; y <= maxy && y < miny+20; y++ {
		lo, hi := -1, -1
		for x := minx; x <= maxx; x++ {
			if before[y*gfx.Width+x] != fb.Pix[y*gfx.Width+x] {
				if lo < 0 {
					lo = x
				}
				hi = x
			}
		}
		if lo >= 0 {
			t.Logf("y=%d x %d..%d  舊=%v 新=%v", y, lo, hi, before[y*gfx.Width+lo:y*gfx.Width+lo+8], fb.Pix[y*gfx.Width+lo:y*gfx.Width+lo+8])
		}
	}
}
