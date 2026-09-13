package screen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

func TestBattleDataScrollHitboxesMatchOriginal(t *testing.T) {
	s, err := LoadBattleDataSpec(filepath.Join("../..", "docs/internals/data/battle-data.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		x, y, want int
	}{
		{434, 120, -1}, {447, 139, -1},
		{434, 140, 0}, {447, 379, 0},
		{434, 380, 1}, {447, 399, 1},
		{433, 120, 0}, {448, 399, 0}, {434, 400, 0},
	} {
		if got := s.BattleDataScrollAt(tc.x, tc.y); got != tc.want {
			t.Errorf("(%d,%d)：方向 %d，預期 %d", tc.x, tc.y, got, tc.want)
		}
	}
}

// 收據 `s3data` 那一局表上的字（`docs/re/111` 逐格讀出來的）。
// 這裡是**繪製**的對拍：內容當固定輸入，證明版面與字的落點對得上原版；
// 內容怎麼從戰場狀態算出來由 internal/game 那一側負責。
const (
	b5Di    = 0xB2C4 // 第
	b5Jian  = 0xC4A5 // 艦
	b5Dui   = 0xB6A4 // 隊
	b5Bu    = 0xB3A1 // 部
	b5Slash = 0xA1FE // ／（原版用 A1FE，不是 python big5 編出來的 A241）
)

var s3dataRows = [][]BattleDataRun{
	{
		{X: 48, Ink: 6, Codes: []uint16{b5Di, '1', '3', b5Jian, b5Dui}},
		{X: 128, Codes: []uint16{0xB5CE, 0xA95A, 0xB1F6, 0xAFF7}}, // 舒坦梅茲
		{X: 272, Codes: []uint16{0xA2D0}},                         // Ｂ
		{X: 304, Codes: []uint16{'3', '0', '0', '0'}},
		{X: 336, Codes: []uint16{b5Slash}},
		{X: 352, Codes: []uint16{'3', '0', '0', '0'}},
		{X: 408, Codes: []uint16{'1', '0', '0'}},
	},
}

func init() {
	// 第２–第８部隊：前五個 Ｂ／3000，後三個 Ｃ／4000。
	for i := 2; i <= 8; i++ {
		kind, n := uint16(0xA2D0), "3000" // Ｂ
		if i >= 6 {
			kind, n = 0xA2D1, "4000" // Ｃ
		}
		digits := func(s string) []uint16 {
			out := make([]uint16, len(s))
			for j := 0; j < len(s); j++ {
				out[j] = uint16(s[j])
			}
			return out
		}
		s3dataRows = append(s3dataRows, []BattleDataRun{
			{X: 128, Codes: []uint16{b5Di, uint16(0xA2B0 + i - 1), b5Bu, b5Dui}},
			{X: 272, Codes: []uint16{kind}},
			{X: 304, Codes: digits(n)},
			{X: 336, Codes: []uint16{b5Slash}},
			{X: 352, Codes: digits(n)},
			{X: 408, Codes: []uint16{'1', '0', '0'}},
		})
	}
}

// TestBattleDataMatchesOracle 把「資料」畫面逐位元組對收據 `s3data`。
//
// 這張畫面是不透明的一層（與同狀態底圖 `s3data0` 比，只有 x 16–462、
// y 60–419 這一塊變了，`docs/re/111`），所以畫在空白畫布上就能直接比。
// 原版自己畫的游標壓在表上——游標吃**滑鼠**座標，收據那一下是 (240,220)。
func TestBattleDataMatchesOracle(t *testing.T) {
	root := "../.."
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle/s3data/screen.bin"))
	if err != nil || len(want) < gfx.Width*gfx.Height {
		t.Skipf("缺收據 s3data，跳過（不用代用品）")
	}
	s, err := LoadBattleDataSpec(filepath.Join(root, "docs/internals/data/battle-data.json"))
	if err != nil {
		t.Fatal(err)
	}
	fnt, err := font.Load(filepath.Join(root, "org_game/extracted/GRAPH.IMG"))
	if err != nil {
		t.Skipf("缺 GRAPH.IMG，跳過：%v", err)
	}
	cur, err := ui.LoadCursor(filepath.Join(root, "docs/internals/data/cursor.json"))
	if err != nil {
		t.Fatal(err)
	}
	fb := gfx.NewFramebuffer(0)
	if m := DrawBattleData(s, fb, fnt, 0, s3dataRows); len(m) > 0 {
		t.Fatalf("缺字 %04X", m)
	}
	cur.Draw(fb, 240, 220)
	w := s.Window
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
	t.Logf("視窗 (%d,%d) %d×%d：%d／%d 個像素不同", w.X, w.Y, w.W, w.H, diff, w.W*w.H)
	if diff != 0 && os.Getenv("LOGH3_DUMP") != "" {
		_ = fb.WriteIndex(filepath.Join(root, "workplace/battledata-remake.bin"))
	}
	if diff != 0 {
		t.Errorf("第一個不同的點 (%d,%d)：remake %d，原版 %d", first[0], first[1],
			fb.Pix[first[1]*gfx.Width+first[0]], want[first[1]*gfx.Width+first[0]])
	}
}

// 收據 `s3data_planet` 的行星表。名字的顏色是**擁有者的陣營色**
// （帝國 6、同盟 5，`docs/re/111`）：雷奧米爾是帝國的，另外兩顆是同盟的。
// 不是自己的行星，補給那一欄整段印死字串 `M*:F* *******`。
func planetRows() [][]BattleDataRun {
	half := func(s string) []uint16 {
		out := make([]uint16, len(s))
		for i := 0; i < len(s); i++ {
			out[i] = uint16(s[i])
		}
		return out
	}
	row := func(ink uint8, name []uint16, econ, def, tail string) []BattleDataRun {
		return []BattleDataRun{
			{X: 48, Ink: ink, Codes: name},
			{X: 208, Codes: half(econ)},
			{X: 272, Codes: half(def)},
			{X: 320, Codes: half(tail)},
		}
	}
	return [][]BattleDataRun{
		row(6, []uint16{0xB970, 0xB6F8, 0xA6CC, 0xBAB8}, "4215", "3300", "Mo:Fo xxxxxxx"),
		row(5, []uint16{0xC1C2, 0xA9D4, 0xBC77}, "3403", "3800", "M*:F* *******"),
		row(5, []uint16{0xB946, 0xA7C6, 0xA751}, "5870", "4700", "M*:F* *******"),
	}
}

// TestBattleDataTabsMatchOracle 把另外兩個分頁也逐位元組對收據。
//
// 框、列距、字的落點三張完全相同，換掉的是表頭那一行與內容
// （`docs/re/111`）；小方塊填色 2 的那一格跟著分頁走。
func TestBattleDataTabsMatchOracle(t *testing.T) {
	for _, tc := range []struct {
		batch  string
		tab    int
		mx, my int
		rows   [][]BattleDataRun
	}{
		{"s3data_planet", 1, 144, 50, planetRows()},
		{"s3data_enemy", 2, 224, 50, nil},
	} {
		t.Run(tc.batch, func(t *testing.T) {
			root := "../.."
			want, err := os.ReadFile(filepath.Join(root, "workplace/oracle", tc.batch, "screen.bin"))
			if err != nil || len(want) < gfx.Width*gfx.Height {
				t.Skipf("缺收據 %s，跳過（不用代用品）", tc.batch)
			}
			s, err := LoadBattleDataSpec(filepath.Join(root, "docs/internals/data/battle-data.json"))
			if err != nil {
				t.Fatal(err)
			}
			fnt, err := font.Load(filepath.Join(root, "org_game/extracted/GRAPH.IMG"))
			if err != nil {
				t.Skipf("缺 GRAPH.IMG，跳過：%v", err)
			}
			cur, err := ui.LoadCursor(filepath.Join(root, "docs/internals/data/cursor.json"))
			if err != nil {
				t.Fatal(err)
			}
			fb := gfx.NewFramebuffer(0)
			if m := DrawBattleData(s, fb, fnt, tc.tab, tc.rows); len(m) > 0 {
				t.Fatalf("缺字 %04X", m)
			}
			cur.Draw(fb, tc.mx, tc.my)
			w := s.Window
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
			t.Logf("%s：%d／%d 個像素不同", tc.batch, diff, w.W*w.H)
			if diff != 0 && os.Getenv("LOGH3_DUMP") != "" {
				_ = fb.WriteIndex(filepath.Join(root, "workplace", tc.batch+"-remake.bin"))
			}
			if diff != 0 {
				t.Errorf("第一個不同的點 (%d,%d)：remake %d，原版 %d", first[0], first[1],
					fb.Pix[first[1]*gfx.Width+first[0]], want[first[1]*gfx.Width+first[0]])
			}
		})
	}
}
