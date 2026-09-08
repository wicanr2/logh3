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

// TestFleetDetailMatchesOracle 把艦隊詳細資料逐位元組對收據 `s3detail`。
//
// 內容是收據上讀出來的固定輸入（`docs/re/112`）：這一條測的是版面與
// 繪製；內容怎麼從戰場狀態算出來由 internal/game 那一側負責。
func TestFleetDetailMatchesOracle(t *testing.T) {
	root := "../.."
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle/s3detail/screen.bin"))
	if err != nil || len(want) < gfx.Width*gfx.Height {
		t.Skipf("缺收據 s3detail，跳過（不用代用品）")
	}
	bd, err := LoadBattleDataSpec(filepath.Join(root, "docs/internals/data/battle-data.json"))
	if err != nil {
		t.Fatal(err)
	}
	s, err := LoadFleetDetailSpec(filepath.Join(root, "docs/internals/data/fleet-detail.json"))
	if err != nil {
		t.Fatal(err)
	}
	fnt, err := font.Load(filepath.Join(root, "org_game/extracted/GRAPH.IMG"))
	if err != nil {
		t.Skipf("缺 GRAPH.IMG，跳過：%v", err)
	}
	faceData, err := os.ReadFile(filepath.Join(root, "org_game/extracted/FACE.AMG"))
	if err != nil {
		t.Skipf("缺 FACE.AMG，跳過：%v", err)
	}
	faces, err := asset.OpenAMG(faceData)
	if err != nil {
		t.Fatal(err)
	}
	face, err := faces.Decode(12)
	if err != nil {
		t.Fatal(err)
	}
	cur, err := ui.LoadCursor(filepath.Join(root, "docs/internals/data/cursor.json"))
	if err != nil {
		t.Fatal(err)
	}
	name := func(list []FleetDetailName, i int) []uint16 { return list[i].Codes }
	v := FleetDetailView{
		Faction:    []uint16{0xABD2, 0xB0EA, 0xAD78}, // 帝國軍
		FactionInk: 6,
		FleetNo:    13,
		Commander:  []uint16{0xB5CE, 0xA95A, 0xB1F6, 0xAFF7, 0xA457, 0xB14E}, // 舒坦梅茲上將
		CmdLevel:   29, CmdExp: 0,
		Adjutant: []uint16{0xA960, 0xA6E8, 0xA4DA, 0xABA2, 0xA4A4, 0xB14E}, // 奈西巴哈中將
		AdjLevel: 24, AdjExp: 0,
		Values:    [12]int{100, 11, 26, 16, 93, 106, 102, 94, 101, 94, 98, 105},
		ValueInks: [12]uint8{2: 5, 3: 5, 7: 5, 8: 5, 10: 5},
		UnitNo:    1,
		KindName:  name(s.ShipKinds, 0), // 戰艦
		Ships:     3000, Full: 3000,
		UnitMorale: 100, TopSpeed: 30, BattleSpeed: 14, Durability: 97,
		Beam:     FleetDetailWeapon{Name: name(s.Weapons, 0), Cur: 18000, Max: 18000},
		Missile:  FleetDetailWeapon{Name: name(s.Weapons, 5), Cur: 12000, Max: 12000},
		Launcher: FleetDetailWeapon{Cur: 6000, Max: 6000},
		Fighter:  FleetDetailWeapon{Name: name(s.Weapons, 10), Cur: 36000, Max: 36000},
		Face:     face,
	}
	fb := gfx.NewFramebuffer(0)
	bd.Window.Frame.frame().Draw(fb, bd.Window.X, bd.Window.Y, bd.Window.W, bd.Window.H)
	if m := DrawFleetDetail(s, fb, fnt, v); len(m) > 0 {
		t.Fatalf("缺字 %04X", m)
	}
	cur.Draw(fb, 200, 90)
	w := bd.Window
	diff, first := 0, [2]int{-1, -1}
	for y := w.Y; y < w.Y+w.H; y++ {
		for x := w.X; x < w.X+w.W; x++ {
			if fb.Pix[y*gfx.Width+x] != want[y*gfx.Width+x] {
				if diff == 0 {
					first = [2]int{x, y}
				}
				diff++
			}
		}
	}
	t.Logf("艦隊詳細資料：%d／%d 個像素不同", diff, w.W*w.H)
	if diff != 0 {
		t.Errorf("第一個不同的點 (%d,%d)：remake %d，原版 %d", first[0], first[1],
			fb.Pix[first[1]*gfx.Width+first[0]], want[first[1]*gfx.Width+first[0]])
		if os.Getenv("LOGH3_DUMP") != "" {
			_ = fb.WriteIndex(filepath.Join(root, "workplace/fleetdetail-remake.bin"))
		}
	}
}
