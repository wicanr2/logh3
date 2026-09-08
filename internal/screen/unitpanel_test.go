package screen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

// admiralName 從 keireki.json 讀第 i 名提督的姓名字碼。
func admiralName(t *testing.T, root string, i int) []uint16 {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/keireki.json"))
	if err != nil {
		t.Skipf("缺 keireki.json，跳過：%v", err)
	}
	var doc struct {
		Records []struct {
			Index     int      `json:"index"`
			NameCodes []uint16 `json:"name_codes"`
		} `json:"records"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	for _, r := range doc.Records {
		if r.Index == i {
			return r.NameCodes
		}
	}
	t.Fatalf("提督表沒有第 %d 筆", i)
	return nil
}

// TestUnitPanelMatchesOracle 把分隊面板的八列逐位元組對收據
// （`docs/re/118`）。
//
// 兩份收據的值都從原版記憶體查過（艦隊記錄、艦艇記錄、艦種列），
// 這裡直接餵進去比版面：
//
//	s3cmdfl3  第 28 艦隊旗艦（艦種 0，提督 52），艦隊士氣值 99、行動 移動
//	s3cmdsq3  第 31 艦隊第 1 分隊（艦種 1），部隊士氣值 98、艦隊行動 戰鬥
func TestUnitPanelMatchesOracle(t *testing.T) {
	root := "../.."
	s, err := LoadUnitPanelSpec(filepath.Join(root, "docs/internals/data/unit-panel.json"))
	if err != nil {
		t.Fatal(err)
	}
	fd, err := LoadFleetDetailSpec(filepath.Join(root, "docs/internals/data/fleet-detail.json"))
	if err != nil {
		t.Fatal(err)
	}
	fnt, err := font.Load(filepath.Join(root, "org_game/extracted/GRAPH.IMG"))
	if err != nil {
		t.Skipf("缺 GRAPH.IMG，跳過：%v", err)
	}
	cases := []struct {
		batch string
		view  UnitPanelView
	}{
		{"s3cmdfl3", UnitPanelView{
			Faction: 0, FleetNo: 13, Flagship: true,
			Admiral: admiralName(t, root, 52), KindName: fd.ShipKinds[0].Codes,
			Ships: 2977, Morale: 99, FleetMorale: true,
			ActionName: s.FleetActions[0].Codes, ActionInk: s.FleetActionInkOf(0),
			Beam: 17862, Missile: 11908, Fighter: 26793,
			MissileInk: s.WeaponInk.Full, FighterInk: s.WeaponInk.Low,
		}},
		{"s3cmdsq3", UnitPanelView{
			Faction: 0, FleetNo: 16, UnitNo: 2, KindName: fd.ShipKinds[1].Codes,
			Ships: 2950, Morale: 98,
			ActionName: s.FleetActions[1].Codes, ActionInk: s.FleetActionInkOf(1),
			Beam: 17700, Missile: 8850, Fighter: 23600,
			MissileInk: s.WeaponInk.Part, FighterInk: s.WeaponInk.Low,
		}},
	}
	// 比的是文字那一塊（八個字元格寬）：框由 DrawBattleStatusFrame 畫，另外對過。
	x0, y0 := s.Cell.X, s.Cell.Y
	w, h := 8*s.Cell.Col, s.Cell.Rows*s.Cell.Row
	for _, c := range cases {
		t.Run(c.batch, func(t *testing.T) {
			want, err := os.ReadFile(
				filepath.Join(root, "workplace/oracle", c.batch, "screen.bin"))
			if err != nil || len(want) < gfx.Width*gfx.Height {
				t.Skipf("缺收據 %s，跳過（不用代用品）", c.batch)
			}
			fb := gfx.NewFramebuffer(0)
			// 面板的底色是狀態欄的填色 9——先鋪起來再畫字。
			fb.FillRect(x0, y0, w, h, 9)
			DrawUnitPanel(s, fb, fnt, c.view)
			diff, first := 0, [2]int{-1, -1}
			for yy := y0; yy < y0+h; yy++ {
				for xx := x0; xx < x0+w; xx++ {
					if fb.Pix[yy*gfx.Width+xx] != want[yy*gfx.Width+xx] {
						if diff == 0 {
							first = [2]int{xx, yy}
						}
						diff++
					}
				}
			}
			t.Logf("面板 (%d,%d) %d×%d：%d／%d 個像素不同", x0, y0, w, h, diff, w*h)
			if diff != 0 {
				t.Errorf("第一個不同的點 (%d,%d)：remake %d，原版 %d",
					first[0], first[1],
					fb.Pix[first[1]*gfx.Width+first[0]],
					want[first[1]*gfx.Width+first[0]])
			}
		})
	}
}
