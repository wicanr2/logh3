package screen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
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
	}{
		{"hovA", 60, 60, -1, 0, 0, false},
		{"hovB", 150, 120, -1, 0, 0, false},
		{"hovC", 250, 200, -1, 0, 0, false},
		{"rcback", 300, 200, -1, 0, 0, false},
		{"sec1", 23, 72, -1, 0, 0, false},
		{"sec5", 33, 117, -1, 0, 0, false},
		{"sec7", 91, 102, -1, 0, 0, false},
		{"sec9", 66, 180, -1, 0, 0, false},
		{"pnlU000", 12, 12, 0, 0, 0, false},
		{"pnlU040", 12, 12, 40, 0, 0, false},
		{"pnlU041", 12, 12, 41, 0, 0, false},
		{"pnlU100", 12, 12, 100, 0, 0, false},
		{"pnlU150", 12, 12, 150, 0, 0, false},
		{"pnlU200", 12, 12, 200, 0, 0, false},
		{"pnlFLT", 12, 12, -1, 12, 99, false},
	} {
		c := c
		t.Run(c.batch, func(t *testing.T) {
			checkRegionPanel(t, c.batch, c.mx, c.my, c.unrest, c.own, c.enemy, c.whole)
		})
	}
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
		unrest = measuredUnrest(t, region)
	}
	rec[0x7D] = byte(unrest)
	rec[0x28], rec[0x2B] = byte(own), byte(enemy)
	return rec
}

// measuredUnrest 讀 796 年 1 月量到的未安定度。
func measuredUnrest(t *testing.T, region int) int {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("../..", "docs/internals/data/month-rollover-796-01.json"))
	if err != nil {
		t.Fatal(err)
	}
	var d struct {
		Regions []struct {
			Index  int `json:"index"`
			Unrest int `json:"unrest_796_01"`
		} `json:"regions"`
	}
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	for _, r := range d.Regions {
		if r.Index == region {
			return r.Unrest
		}
	}
	t.Fatalf("量到的資料裡沒有星域 %d", region)
	return 0
}

func checkRegionPanel(t *testing.T, batch string, mx, my, unrest, own, enemy int, whole bool) {
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
	planets, err := rules.LoadPlanets(pl)
	if err != nil {
		t.Fatal(err)
	}
	sum := rules.Summarise(panelRecord(t, region, unrest, own, enemy), planets, region, 0)

	fb, _ := MainGame(s, fnt, mapImg, faces, cur)
	panel.Draw(fb, fnt, sum, 0)

	want, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
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
