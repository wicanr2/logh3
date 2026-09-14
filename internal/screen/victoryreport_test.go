package screen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

func TestVictoryReportMatchesOracle(t *testing.T) {
	root := filepath.Join("..", "..")
	p, err := LoadVictoryReportPanel(filepath.Join(root, "docs/internals/data/victory-report-panel.json"))
	if err != nil {
		t.Fatal(err)
	}
	fnt, err := font.Load(filepath.Join(root, "org_game/extracted/GRAPH.IMG"))
	if err != nil {
		t.Skip("缺 GRAPH.IMG")
	}
	blk, err := os.ReadFile(filepath.Join(root, "org_game/extracted/FRAME.BLK"))
	if err != nil {
		t.Skip("缺 FRAME.BLK")
	}
	tiles, err := ui.LoadFrameTiles(blk)
	if err != nil {
		t.Fatal(err)
	}
	faceRaw, err := os.ReadFile(filepath.Join(root, "org_game/extracted/FACE.AMG"))
	if err != nil {
		t.Skip("缺 FACE.AMG")
	}
	faces, err := asset.OpenAMG(faceRaw)
	if err != nil {
		t.Fatal(err)
	}
	face, err := faces.Decode(6)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name, batch  string
		won, regions bool
	}{
		{name: "win", batch: "victory-win-trace", won: true, regions: true},
		{name: "loss", batch: "victory-loss-trace", won: false},
	} {
		t.Run(c.name, func(t *testing.T) {
			scr, err := os.ReadFile(filepath.Join(root, "workplace/oracle", c.batch, "screen.bin"))
			if err != nil || len(scr) != 640*480 {
				t.Skip("缺勝負報告收據")
			}
			regions := 0
			if c.regions {
				regions = 11
			}
			fb := gfx.NewFramebuffer(0)
			p.Draw(fb, fnt, tiles, face, 0, c.won, 796, 1, 1, regions, [2]int{})
			bad, total := 0, 0
			for y := 140; y < 380; y++ {
				for x := 160; x < 512; x++ {
					total++
					if fb.At(x, y) != scr[y*640+x] {
						bad++
					}
				}
			}
			if bad != 0 {
				t.Errorf("勝負報告 overlay 差 %d／%d 像素；樣本 got/want (160,140)=%d/%d (176,160)=%d/%d (500,360)=%d/%d",
					bad, total, fb.At(160, 140), scr[140*640+160],
					fb.At(176, 160), scr[160*640+176], fb.At(500, 360), scr[360*640+500])
			}
		})
	}
}
