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

func TestPhezzanReportPagesMatchOracle(t *testing.T) {
	root := filepath.Join("..", "..")
	p, err := LoadPhezzanReportPanel(filepath.Join(root, "docs/internals/data/phezzan-report-panel.json"))
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
	facesRaw, err := os.ReadFile(filepath.Join(root, "org_game/extracted/FACE.AMG"))
	if err != nil {
		t.Skip("缺 FACE.AMG")
	}
	faces, err := asset.OpenAMG(facesRaw)
	if err != nil {
		t.Fatal(err)
	}
	face, err := faces.Decode(6) // 帝國軍事專任參謀，WORK +0x37。
	if err != nil {
		t.Fatal(err)
	}

	for page, batch := range []string{"phezzan-report-page1", "phezzan-report-page2b"} {
		scr, err := os.ReadFile(filepath.Join(root, "workplace/oracle", batch, "screen.bin"))
		if err != nil || len(scr) != 640*480 {
			t.Skip("缺費沙報告收據")
		}
		fb := gfx.NewFramebuffer(0)
		p.Draw(fb, fnt, tiles, "owner_0", page, face)
		bad, total := 0, 0
		for y := 160; y < 300; y++ {
			for x := 112; x < 528; x++ {
				total++
				if fb.At(x, y) != scr[y*640+x] {
					bad++
				}
			}
		}
		if bad != 0 {
			t.Errorf("第 %d 頁 overlay 差 %d／%d 像素", page+1, bad, total)
		}
	}
}

func TestPhezzanReportDataHasFourTwoPageVariants(t *testing.T) {
	root := filepath.Join("..", "..")
	p, err := LoadPhezzanReportPanel(filepath.Join(root, "docs/internals/data/phezzan-report-panel.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"owner_0", "owner_1", "opponent_0", "opponent_1"} {
		for page := 0; page < 2; page++ {
			if raw, ok := p.Page(key, page); !ok || len(raw) == 0 {
				t.Errorf("%s 第 %d 頁缺資料", key, page+1)
			}
		}
		if _, ok := p.Page(key, 2); ok {
			t.Errorf("%s 不應有第三頁", key)
		}
	}
}
