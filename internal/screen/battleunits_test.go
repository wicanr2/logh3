package screen

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

// 35D1:02AB比較器的remake分支回歸；色塊只用於辨認覆蓋，不是原版素材代用品。
func TestBattleUnitPriorityPrecedesOrdinal(t *testing.T) {
	s, err := LoadBattleSpec("../../docs/internals/data/screen-battle.json")
	if err != nil {
		t.Fatal(err)
	}
	var sprites []*asset.Image
	for _, ink := range []byte{1, 2} {
		p := make([]byte, 256)
		for i := range p {
			p[i] = ink
		}
		sprites = append(sprites, &asset.Image{W: 16, H: 16, Pix: p})
	}
	for _, c := range []struct {
		name      string
		a, b      int
		wantColor byte
	}{{"priority", 0x70, 0x80, 2}, {"reverse", 0x80, 0x70, 1},
		{"ordinal-tie", 0x80, 0x80, 1}, {"unsigned", 0xFF, 0x80, 1}} {
		t.Run(c.name, func(t *testing.T) {
			units := []rules.BattleUnit{
				{Fleet: 12, Squadron: 1, Faction: 0, X: 208, Y: 160, DisplayPriority: c.a},
				{Fleet: 16, Squadron: 1, Faction: 1, X: 208, Y: 160, DisplayPriority: c.b},
			}
			before := append([]rules.BattleUnit(nil), units...)
			fb := gfx.NewFramebuffer(0)
			DrawBattleUnits(s, fb, sprites, units, func(u rules.BattleUnit) int { return u.Faction }, BattleView{})
			x, y := s.Field.X+208, s.Field.Y+160
			if got := fb.Pix[y*640+x]; got != c.wantColor {
				t.Fatalf("重疊像素%d，依比較器應為%d", got, c.wantColor)
			}
			if !reflect.DeepEqual(units, before) {
				t.Fatal("繪製排序更動規則及點選單位的次序")
			}
		})
	}
}

// s3fight 那一場視野內、畫面上比得出圖塊的分隊（`docs/re/85`）。
var origUnits = []struct {
	X, Y, Sprite int
}{
	{1293, 830, 67}, {1311, 856, 74}, {1320, 812, 67}, {1330, 941, 74},
	{1338, 839, 74}, {1347, 794, 67}, {1356, 865, 74}, {1364, 821, 74},
}

// TestBattleUnitsMatchOracle：把那幾個分隊照 SPRITE.SPR 的塊畫上去，
// 每一塊的 16×16 都要與收據逐位元組相同。
//
// 單位會互相疊，所以這裡只比「畫上去的那幾塊本身」——落點錯一個像素
// 就會整塊對不上，所以位置換算仍然被驗到。
func TestBattleUnitsMatchOracle(t *testing.T) {
	root := "../.."
	specPath := filepath.Join(root, "docs/internals/data/screen-battle.json")
	orig := filepath.Join(root, "org_game/extracted")
	oracle := filepath.Join(root, "workplace/oracle/s3fight/screen.bin")
	for _, p := range []string{specPath, oracle, filepath.Join(orig, "SPRITE.SPR")} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過同狀態對拍（不用代用品）", p)
		}
	}
	s, err := LoadBattleSpec(specPath)
	if err != nil {
		t.Fatal(err)
	}
	sprData, err := os.ReadFile(filepath.Join(orig, "SPRITE.SPR"))
	if err != nil {
		t.Fatal(err)
	}
	sprites, err := asset.DecodeTiles(sprData, 16, 16)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	view := BattleView{Col: 66, Row: 30}
	for _, c := range origUnits {
		fb := gfx.NewFramebuffer(0)
		u := rules.BattleUnit{X: c.X, Y: c.Y}
		n := c.Sprite
		DrawBattleUnits(s, fb, sprites, []rules.BattleUnit{u},
			func(rules.BattleUnit) int { return n }, view)
		sx := s.Field.X + c.X - view.Col*s.Field.TileW
		sy := s.Field.Y + c.Y - view.Row*s.Field.TileH
		bad := 0
		for y := sy; y < sy+16; y++ {
			for x := sx; x < sx+16; x++ {
				if want[y*640+x] != fb.Pix[y*640+x] {
					bad++
				}
			}
		}
		if bad != 0 {
			t.Errorf("世界 (%d,%d) 的第 %d 塊：畫面 (%d,%d) 差 %d／256 個像素",
				c.X, c.Y, c.Sprite, sx, sy, bad)
		}
	}
}
