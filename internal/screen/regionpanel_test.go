package screen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

// TestRegionPanelMatchesOracle 是星域摘要面板的同狀態對拍（規格 `docs/spec/001`）。
//
// 每一份收據都是「在戰略地圖介面對星圖按一次右鍵」，所以整張畫面 ＝
// 命令階段底板 ＋ 游標在按下去的位置 ＋ 左下角那塊摘要。三樣都由 remake
// 自己算出來，然後**整張 640×480 逐位元組**比對。
//
// `pnl*` 那幾份是 `-poke` 把星域 0 的未安定度或艦隊數改掉造出來的局面
// （CLAUDE.md：要某個局面就改記憶體，不要靠亂數）；注入的內容寫在下表，
// 與收據 meta.txt 的 `-poke` 旗標一致。
func TestRegionPanelMatchesOracle(t *testing.T) {
	for _, c := range []struct {
		batch      string
		mx, my     int
		unrest     int  // <0 ＝ 用量到的那一組
		own, enemy int
		whole      bool // 整張都比（保留給游標乾淨的收據；目前一份都沒有）
		panelOnly  bool // 只比左下角那塊（底層是別的局面，地圖與右側面板都不一樣）
		// slot 是存檔槽位（`PLANET.DT?`／`STELLAR.DT?`，`docs/re/41`）。
		// 空的就用 `PLADAT.DAT` ＋ 上面那幾欄湊出來的開局狀態。
		slot string
	}{
		{"hovA", 60, 60, -1, 0, 0, false, false, ""},
		{"hovB", 150, 120, -1, 0, 0, false, false, ""},
		{"hovC", 250, 200, -1, 0, 0, false, false, ""},
		{"rcback", 300, 200, -1, 0, 0, false, false, ""},
		{"sec1", 23, 72, -1, 0, 0, false, false, ""},
		{"sec5", 33, 117, -1, 0, 0, false, false, ""},
		{"sec7", 91, 102, -1, 0, 0, false, false, ""},
		{"sec9", 66, 180, -1, 0, 0, false, false, ""},
		{"pnlU000", 12, 12, 0, 0, 0, false, false, ""},
		{"pnlU040", 12, 12, 40, 0, 0, false, false, ""},
		{"pnlU041", 12, 12, 41, 0, 0, false, false, ""},
		{"pnlU100", 12, 12, 100, 0, 0, false, false, ""},
		{"pnlU150", 12, 12, 150, 0, 0, false, false, ""},
		{"pnlU200", 12, 12, 200, 0, 0, false, false, ""},
		{"pnlFLT", 12, 12, -1, 12, 99, false, false, ""},
		// S3「巴米利恩會戰」右鍵選星域 25（干達巴爾）。這一份的擁有者、
		// 行星數與艦隊數都不是開局的值，直接讀那一格存檔（`docs/re/41`）。
		// 底層是 796 年 1 月的底板，所以只比左下角那塊。
		{"s3rgn", 230, 12, -1, 0, 0, false, true, "A"},
	} {
		c := c
		t.Run(c.batch, func(t *testing.T) {
			checkRegionPanel(t, c.batch, c.mx, c.my, c.unrest, c.own, c.enemy,
				c.whole, c.panelOnly, c.slot)
		})
	}
}

// drawCapital796_01 把右側面板畫成 796 年 1 月的狀態（量到的值）。
func drawCapital796_01(t *testing.T, fb *gfx.Framebuffer, fnt *font.Font,
	region *RegionPanel, planets *rules.Planets) {
	t.Helper()
	root := "../.."
	cap, err := LoadCapitalPanel(filepath.Join(root, "docs/internals/data/capital-panel.json"))
	if err != nil {
		t.Fatal(err)
	}
	rec := panelRecord(t, 0, -1, 0, 0)
	rec[0x1A], rec[0x1B] = byte(82302&0xFF), byte(82302>>8&0xFF)
	rec[0x1C], rec[0x1D] = byte(82302>>16), 0
	rec[0x26] = 32
	cap.Draw(fb, fnt, region, rules.SummariseCapital(rec, planets, 0, 0, 0, 796, 45592, 10341), 0)
}

// panelRecord 湊出面板要的執行期星域記錄。
//
// 名字從 `STELLAR.DTA` 讀（四個劇本的名稱區一模一樣），未安定度取量到的
// 那一組——**載入時未安定度怎麼算出來的還沒解**（`docs/re/24`），
// 所以這一欄只能用量到的值。艦隊數 796 年 1 月全是 0。
func panelRecord(t *testing.T, region, unrest, own, enemy int) []byte {
	t.Helper()
	root := "../.."
	stellar, err := os.ReadFile(filepath.Join(root, "org_game/extracted/STELLAR.DTA"))
	if err != nil {
		t.Skipf("缺 STELLAR.DTA，跳過同狀態對拍（不用代用品）")
	}
	rec := make([]byte, rules.RegionRecord)
	copy(rec, stellar[region*rules.RegionRecord:(region+1)*rules.RegionRecord])
	if unrest < 0 {
		unrest = newGameUnrest(t, region)
	}
	rec[0x7D] = byte(unrest)
	rec[0x28], rec[0x2B] = byte(own), byte(enemy)
	return rec
}

// newGameUnrest 算 796 年 1 月的未安定度。
//
// 整張星域表都是從原版檔案重建的（`rules.NewGameRegions`，`docs/re/44`）：
// 擁有者與行星數讀 `PLADAT.DAT`、經濟力與成長率照成長表算、民心由擁有者
// 決定、稅率照 `090A:0CC0` 那一支算，最後才是未安定度（`docs/spec/006`）。
// 難度參數是量到的（收據 `newload`）。
func newGameUnrest(t *testing.T, region int) int {
	t.Helper()
	root := "../.."
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(root, "org_game/extracted", name))
		if err != nil {
			t.Skipf("缺 %s，跳過同狀態對拍（不用代用品）", name)
		}
		return b
	}
	planets, err := rules.LoadPlanets(read("PLADAT.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	table, err := rules.LoadGrowthTable(filepath.Join(root,
		"docs/internals/data/growth-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	graph, err := rules.LoadStarGraph(filepath.Join(root,
		"docs/internals/data/star-graph.json"))
	if err != nil {
		t.Fatal(err)
	}
	out := rules.NewGameRegions(read("STELLAR.DTX"), planets, read("ADMIRAL.DTX"),
		table, graph, rules.UnrestParams{
			Turn: 0, Ruler: [2]int{16, 60}, Blend: [2]int{2, 2}, Rate: [2]int{2, 2},
			Table: []int{170, 150, 130, 90, 60},
			Base:  []int{50, 75, 100, 150, 200},
		})
	return int(out[region*rules.RegionRecord+rules.RegionUnrest])
}

// savedSlot 讀一格存檔的行星表與星域記錄。
//
// 存檔就是執行期那幾張表原樣寫出去的（`docs/re/41`），載入走
// `rules.LoadSlot`。未安定度那一欄原版載入時整份重算，算式解出來了
// （`docs/re/43`、規格 `docs/spec/006`），所以這裡照算——面板上的
// 治安是**算出來的**，不是抄量到的值。算式要的執行期參數在
// [`unrest-at-load.json`](../../docs/internals/data/unrest-at-load.json)。
func savedSlot(t *testing.T, slot string, pladat []byte, region int) ([]byte, []byte) {
	t.Helper()
	root := filepath.Join("../..", "org_game/extracted")
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Skipf("缺 %s，跳過同狀態對拍（不用代用品）", name)
		}
		return b
	}
	sl, err := rules.LoadSlot(rules.SlotFiles{
		Stellar:      read("STELLAR.DT" + slot),
		Planet:       read("PLANET.DT" + slot),
		Admiral:      read("ADMIRAL.DT" + slot),
		Fleet:        read("FLEET.DT" + slot),
		Ship:         read("SHIP.DT" + slot),
		Work:         read("WORK.DT" + slot),
		PlanetHeader: pladat,
	})
	if err != nil {
		t.Fatal(err)
	}
	graph, err := rules.LoadStarGraph(filepath.Join("../..",
		"docs/internals/data/star-graph.json"))
	if err != nil {
		t.Fatal(err)
	}
	sl.RecomputeUnrest(graph, unrestParams(t, slot))
	return sl.Planets.Data, sl.Region(region)
}

// unrestParams 讀未安定度算式要的執行期參數。
//
// 難度那兩張表與統治者編號不在存檔裡，只能量（`docs/re/43`）。
func unrestParams(t *testing.T, slot string) rules.UnrestParams {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("../..",
		"docs/internals/data/unrest-at-load.json"))
	if err != nil {
		t.Fatal(err)
	}
	var d struct {
		Slots []struct {
			Slot   string `json:"slot"`
			Params struct {
				Turn  int    `json:"turn_word_8F18"`
				Ruler [2]int `json:"ruler_8F4F"`
				Blend [2]int `json:"blend_8FD4"`
				Rate  [2]int `json:"rate_8FF8"`
				Table []int  `json:"table_4B28"`
				Base  []int  `json:"base_147E"`
			} `json:"params"`
		} `json:"slots"`
	}
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	for _, sl := range d.Slots {
		if sl.Slot == slot {
			return rules.UnrestParams{
				Turn: sl.Params.Turn, Ruler: sl.Params.Ruler,
				Blend: sl.Params.Blend, Rate: sl.Params.Rate,
				Table: sl.Params.Table, Base: sl.Params.Base,
			}
		}
	}
	t.Fatalf("unrest-at-load.json 裡沒有第 %s 格的參數", slot)
	return rules.UnrestParams{}
}

func checkRegionPanel(t *testing.T, batch string, mx, my, unrest, own, enemy int,
	whole, panelOnly bool, slot string) {
	t.Helper()
	root := "../.."
	specPath := filepath.Join(root, "docs/internals/data/screen-command-phase.json")
	panelPath := filepath.Join(root, "docs/internals/data/region-panel.json")
	oracle := filepath.Join(root, "workplace/oracle/"+batch+"/screen.bin")
	planetPath := filepath.Join(root, "org_game/extracted/PLADAT.DAT")
	rgnPath := filepath.Join(root, "org_game/extracted/STRMAP.RGN")
	for _, p := range []string{specPath, panelPath, oracle, planetPath, rgnPath} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過同狀態對拍（不用代用品）", p)
		}
	}
	s, err := LoadMainGameSpec(specPath)
	if err != nil {
		t.Fatal(err)
	}
	panel, err := LoadRegionPanel(panelPath)
	if err != nil {
		t.Fatal(err)
	}
	fnt, mapImg, faces, cur := gameScreenAssets(t)

	// 游標搬到按下去的地方，而且整隻都看得見——底板的收據 `main1e` 是
	// 游標被右側面板蓋掉只剩兩行的局面，那個裁切不適用。
	s.Mouse.X, s.Mouse.Y = mx, my
	s.Mouse.Clip = nil
	s.Mouse.Patches = nil

	rgn, err := os.ReadFile(rgnPath)
	if err != nil {
		t.Fatal(err)
	}
	regions, err := asset.OpenRGN(rgn, 16, 20)
	if err != nil {
		t.Fatal(err)
	}
	region := asset.RegionAt(regions, mx, my, 16, 20)
	if region < 0 {
		t.Fatalf("(%d,%d) 不在任何星域裡", mx, my)
	}
	pl, err := os.ReadFile(planetPath)
	if err != nil {
		t.Fatal(err)
	}
	rec := panelRecord(t, region, unrest, own, enemy)
	if slot != "" {
		pl, rec = savedSlot(t, slot, pl, region)
	}
	planets, err := rules.LoadPlanets(pl)
	if err != nil {
		t.Fatal(err)
	}
	sum := rules.Summarise(rec, planets, region, 0)

	fb, _ := MainGame(s, fnt, mapImg, faces, cur)
	if slot == "" {
		// 右側面板同樣是算出來的（`docs/spec/002`）：這幾份收據都是
		// 796 年 1 月，值與 `main1e` 相同。
		drawCapital796_01(t, fb, fnt, panel, planets)
	}
	panel.Draw(fb, fnt, sum, 0)

	want, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	if panelOnly {
		bad := 0
		for y := 340; y < 420; y++ {
			for x := 0; x < 320; x++ {
				if want[y*640+x] != fb.Pix[y*640+x] {
					bad++
				}
			}
		}
		if bad != 0 {
			_ = os.WriteFile(filepath.Join(root, "workplace/remake-"+batch+".bin"), fb.Pix, 0o644)
			t.Fatalf("%s（星域 %d）：左下角那塊與原版不同的像素 %d", batch, region, bad)
		}
		return
	}
	// 游標那一塊不比。原版在右鍵之後**只把游標補回隔列**，要看到完整的
	// 游標得在最後多移動一次滑鼠（`docs/re/09`，底板規格就是為了這個
	// 挑 `main1c` 而不是 `main1`）。這幾份收據的最後一個動作就是右鍵，
	// 所以游標是半畫的——那是另一件事，不在這份規格裡。
	cx, cy := mx+cur.Offset[0], my+cur.Offset[1]
	bad, skipped := 0, 0
	for i := range want {
		x, y := i%640, i/640
		if !whole && x >= cx && x < cx+cur.Width && y >= cy && y < cy+cur.Height {
			skipped++
			continue
		}
		if want[i] != fb.Pix[i] {
			bad++
		}
	}
	if bad != 0 {
		_ = os.WriteFile(filepath.Join(root, "workplace/remake-"+batch+".bin"), fb.Pix, 0o644)
		t.Fatalf("%s（星域 %d）：與原版不同的像素 %d / %d（游標那塊 %d 點沒比）",
			batch, region, bad, len(want), skipped)
	}
}

var _ = font.GlyphW
var _ = ui.Frame{}

// TestCapitalPanelMatchesOracle 是右側面板的同狀態對拍（規格 `docs/spec/002`）。
//
// 兩個月份：`main1e` 是 796 年 1 月 1 日，`turnnext` 是按了十一次「實行」
// 之後的 796 年 2 月 1 日。同一份規格換一組狀態就得畫出兩張不同的畫面
// ——這是「文字真的是算出來的」最直接的證據。
func TestCapitalPanelMatchesOracle(t *testing.T) {
	for _, c := range []struct {
		batch                    string
		turn                     int
		military, taxTotal       int
		economy, taxRate, unrest int
		whole                    bool // 整張都比
		noCursor                 bool // 這一份的游標被面板蓋掉了，不畫
		growth                   int  // 成長率（0 ＝ 用 STELLAR.DTA 的值）
		fleets                   [2]int
		mouse                    [2]int // 非零就把游標搬到這裡
	}{
		{"main1e", 0, 45592, 10341, 82302, 32, 36, true, false, 0, [2]int{}, [2]int{}},
		// 按過一次「實行」：只有日期變成 1 月 4 日，錢與經濟力都沒動。
		// 三份收據（`turn260`／`turn320`／`turn450`，步數不同）逐位元組相同。
		{"turn260", 1, 45592, 10341, 82302, 32, 36, false, true, 0, [2]int{}, [2]int{}},
		{"turn320", 1, 45592, 10341, 82302, 32, 36, false, true, 0, [2]int{}, [2]int{}},
		{"turn450", 1, 45592, 10341, 82302, 32, 36, false, true, 0, [2]int{}, [2]int{}},
		{"turnnext", 10, 42339, 10413, 83304, 32, 36, false, false, 0, [2]int{}, [2]int{}},
		// 從主選單「取進度」載進 S2「諸神的黃昏」（`docs/re/38`）：
		// 798 年 11 月 19 日 ＝ 第 346 回合，數字與開局差很遠，
		// 而且首都有兩支艦隊。
		{"s2empire", 346, 150000, 7280, 86275, 17, 0, false, false, 35,
			[2]int{2, 0}, [2]int{318, 168}},
		// S3「巴米利恩會戰」：799 年 3 月 1 日 ＝ 第 380 回合。首都的
		// 未安定度 13（治安 93%）、艦隊 0：0、稅率 2.2%——
		// `89759 × 22 / 1000 = 1974` 與畫面相同（`docs/re/38`）。
		// 經濟力、稅率、成長率、未安定度都與量到的執行期星域表相同
		// （[`state-s3-799-03.json`](../../docs/internals/data/state-s3-799-03.json)）。
		{"s3rgn", 380, 160000, 13131, 89759, 22, 13, false, false, 45,
			[2]int{0, 0}, [2]int{230, 12}},
	} {
		c := c
		t.Run(c.batch, func(t *testing.T) {
			checkCapitalPanel(t, c.batch, c.turn, c.military, c.taxTotal,
				c.economy, c.taxRate, c.unrest, c.whole, c.noCursor,
				c.growth, c.fleets, c.mouse)
		})
	}
}

func checkCapitalPanel(t *testing.T, batch string, turn, military, taxTotal,
	economy, taxRate, unrest int, whole, noCursor bool,
	growth int, fleets, mouse [2]int) {
	t.Helper()
	root := "../.."
	oracle := filepath.Join(root, "workplace/oracle/"+batch+"/screen.bin")
	want, err := os.ReadFile(oracle)
	if err != nil {
		t.Skipf("缺 %s，跳過同狀態對拍（不用代用品）", oracle)
	}
	panelPath := filepath.Join(root, "docs/internals/data/capital-panel.json")
	cap, err := LoadCapitalPanel(panelPath)
	if err != nil {
		t.Fatal(err)
	}
	region, err := LoadRegionPanel(filepath.Join(root, "docs/internals/data/region-panel.json"))
	if err != nil {
		t.Fatal(err)
	}
	fnt, mapImg, faces, cur := gameScreenAssets(t)
	s, err := LoadMainGameSpec(filepath.Join(root, "docs/internals/data/screen-command-phase.json"))
	if err != nil {
		t.Fatal(err)
	}
	pl, err := os.ReadFile(filepath.Join(root, "org_game/extracted/PLADAT.DAT"))
	if err != nil {
		t.Skip("缺 PLADAT.DAT，跳過")
	}
	planets, err := rules.LoadPlanets(pl)
	if err != nil {
		t.Fatal(err)
	}
	// 首都那一筆記錄：名字從檔案讀，會變的那幾格填量到的值。
	rec := panelRecord(t, 0, unrest, fleets[0], fleets[1])
	if growth != 0 {
		rec[0x1E] = byte(growth)
	}
	rec[0x1A] = byte(economy)
	rec[0x1B] = byte(economy >> 8)
	rec[0x1C], rec[0x1D] = byte(economy>>16), byte(economy>>24)
	rec[0x26] = byte(taxRate)
	sum := rules.SummariseCapital(rec, planets, 0, 0, turn, 796, military, taxTotal)

	if mouse[0] != 0 || mouse[1] != 0 {
		s.Mouse.X, s.Mouse.Y = mouse[0], mouse[1]
		s.Mouse.Clip = nil
		s.Mouse.Patches = nil
	}
	if noCursor {
		// 按過「實行」之後面板整個重畫，把游標蓋掉了——原版那一格是面板
		// 的填色，不是游標。
		s.Mouse = nil
	}
	fb, _ := MainGame(s, fnt, mapImg, faces, cur)
	cap.Draw(fb, fnt, region, sum, 0)

	// `main1e` 整張比：那份收據沒有選中任何星域（左下角是空的），
	// 游標也就在規格量到的位置。`turnnext` 只比右側面板——它是十一次
	// 「實行」跑出來的，游標與左下角都在別的狀態。
	x0, y0, x1, y1 := 0, 0, 640, 480
	if !whole {
		x0, y0, y1 = 320, 40, 300
	}
	bad := 0
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			// 這一份的最後一個動作是點擊，游標只補回一半（`docs/re/09`）
			// ——它壓在星圖右緣，有幾個像素落進右側面板裡。不比那一塊。
			if mouse[0] != 0 && x >= mouse[0] && x < mouse[0]+16 &&
				y >= mouse[1]+40 && y < mouse[1]+40+24 {
				continue
			}
			// 命令列第三個按鈕（系統指令）在這幾份收據裡是**按下去**的狀態
			// ——最後一次點擊就落在它上面，底緣的立體邊跟著反過來（7 px）。
			// 那是命令列的事，不在這份規格裡。
			if noCursor && x >= 542 && x < 611 && y >= 136 && y < 142 {
				continue
			}
			if want[y*640+x] != fb.Pix[y*640+x] {
				bad++
			}
		}
	}
	if bad != 0 {
		_ = os.WriteFile(filepath.Join(root, "workplace/remake-"+batch+"-cap.bin"), fb.Pix, 0o644)
		t.Fatalf("%s：右側面板與原版不同的像素 %d", batch, bad)
	}
}
