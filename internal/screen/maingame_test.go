package screen

import (
	"sort"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// TestMainGameMatchesOracle 是遊戲本體各畫面的同狀態對拍。
//
// 這些畫面共用同一個結構（底板 ＋ 星域地圖 ＋ 疊上去的視窗），
// 所以同一個繪製器換一份規格就能畫——差別只在文字、頭像與熱區。
func TestMainGameMatchesOracle(t *testing.T) {
	for _, c := range []struct{ spec, batch string }{
		{"screen-main-game", "key1"},
		{"screen-ship-order", "ship1"},
		{"screen-build-plan", "plan1"},
	} {
		t.Run(c.batch, func(t *testing.T) { checkGameScreen(t, c.spec, c.batch) })
	}
}


// gameScreenAssets 從原版檔案重建畫面需要的資源。缺檔就 skip，不做代用品。
func gameScreenAssets(t *testing.T) (*font.Font, *asset.Image, *asset.AMG, *ui.Cursor) {
	t.Helper()
	root := "../.."
	fontPath := filepath.Join(root, "org_game/extracted/GRAPH.IMG")
	bitPath := filepath.Join(root, "org_game/extracted/STRMAP.BIT")
	rgnPath := filepath.Join(root, "org_game/extracted/STRMAP.RGN")
	regionsPath := filepath.Join(root, "docs/internals/data/strmap-regions.json")
	facePath := filepath.Join(root, "org_game/extracted/FACE.AMG")
	for _, p := range []string{fontPath, bitPath, rgnPath, regionsPath, facePath} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過同狀態對拍（不用代用品）", p)
		}
	}
	fnt, err := font.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	bit, err := os.ReadFile(bitPath)
	if err != nil {
		t.Fatal(err)
	}
	rgn, err := os.ReadFile(rgnPath)
	if err != nil {
		t.Fatal(err)
	}
	colours, err := loadRegionColours(regionsPath)
	if err != nil {
		t.Fatal(err)
	}
	mapImg, err := asset.BuildStarMap(bit, rgn, colours, 16, 20, 20)
	if err != nil {
		t.Fatal(err)
	}
	faceData, err := os.ReadFile(facePath)
	if err != nil {
		t.Fatal(err)
	}
	faces, err := asset.OpenAMG(faceData)
	if err != nil {
		t.Fatal(err)
	}
	cur, err := ui.LoadCursor(filepath.Join(root, "docs/internals/data/cursor.json"))
	if err != nil {
		t.Fatal(err)
	}
	return fnt, mapImg, faces, cur
}

func checkGameScreen(t *testing.T, specName, batch string) {
	t.Helper()
	root := "../.."
	specPath := filepath.Join(root, "docs/internals/data/"+specName+".json")
	oracle := filepath.Join(root, "workplace/oracle/"+batch+"/screen.bin")
	fontPath := filepath.Join(root, "org_game/extracted/GRAPH.IMG")
	bitPath := filepath.Join(root, "org_game/extracted/STRMAP.BIT")
	rgnPath := filepath.Join(root, "org_game/extracted/STRMAP.RGN")
	regionsPath := filepath.Join(root, "docs/internals/data/strmap-regions.json")
	for _, p := range []string{specPath, oracle, fontPath, bitPath, rgnPath, regionsPath} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過同狀態對拍（不用代用品）", p)
		}
	}
	s, err := LoadMainGameSpec(specPath)
	if err != nil {
		t.Fatal(err)
	}
	fnt, err := font.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	// 地圖層從原版檔案重建：底圖 ＋ 區域遮罩 ＋ 每區顏色。
	// **不用 workplace 裡那份 dump**——那是一次性產物，測試不該靠它。
	bit, err := os.ReadFile(bitPath)
	if err != nil {
		t.Fatal(err)
	}
	rgn, err := os.ReadFile(rgnPath)
	if err != nil {
		t.Fatal(err)
	}
	colours, err := loadRegionColours(regionsPath)
	if err != nil {
		t.Fatal(err)
	}
	mapImg, err := asset.BuildStarMap(bit, rgn, colours, 16, 20, 20)
	if err != nil {
		t.Fatal(err)
	}
	faceData, err := os.ReadFile(filepath.Join(root, "org_game/extracted/FACE.AMG"))
	if err != nil {
		t.Skip("缺 FACE.AMG，跳過")
	}
	faces, err := asset.OpenAMG(faceData)
	if err != nil {
		t.Fatal(err)
	}
	cur, err := ui.LoadCursor(filepath.Join(root, "docs/internals/data/cursor.json"))
	if err != nil {
		t.Fatal(err)
	}
	fb, missing := MainGame(s, fnt, mapImg, faces, cur)
	want, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	bad := 0
	for i := range want {
		if want[i] != fb.Pix[i] {
			bad++
		}
	}
	if bad != 0 {
		_ = os.WriteFile(filepath.Join(root, "workplace/remake-"+batch+".bin"), fb.Pix, 0o644)
		t.Fatalf("%s：與原版不同的像素 %d / %d（畫不出來的字碼 %d 個）",
			batch, bad, len(want), len(missing))
	}
}

// loadRegionColours 讀各星域在 796 年 1 月的所屬勢力色。
func loadRegionColours(path string) ([]uint8, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Regions []struct {
			Colour uint8 `json:"colour_796_01"`
		} `json:"regions"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	out := make([]uint8, len(doc.Regions))
	for i, r := range doc.Regions {
		out[i] = r.Colour
	}
	return out, nil
}

// TestBuildPlanNumbersRedrawIdentically 釘住「重畫數字」這條路：
// 把開局的值原樣畫回去，畫面必須與原版逐位元組相同。
// 這條過了，改預算之後的畫面才有意義——不然分不清是規則錯了還是重畫錯了。
func TestBuildPlanNumbersRedrawIdentically(t *testing.T) {
	root := "../.."
	specPath := filepath.Join(root, "docs/internals/data/screen-build-plan.json")
	oracle := filepath.Join(root, "workplace/oracle/plan1/screen.bin")
	fontPath := filepath.Join(root, "org_game/extracted/GRAPH.IMG")
	for _, p := range []string{specPath, oracle, fontPath} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過", p)
		}
	}
	s, err := LoadMainGameSpec(specPath)
	if err != nil {
		t.Fatal(err)
	}
	fnt, err := font.Load(fontPath)
	if err != nil {
		t.Fatal(err)
	}
	fb, _ := MainGame(s, fnt, nil, nil, nil)
	before := append([]uint8(nil), fb.Pix...)
	s.DrawNumbers(fb, fnt,
		[]int{20, 10, 18, 10, 12, 15, 15},
		[]int{3037, 1518, 2733, 1518, 1822, 2277, 2277},
		[]int{3790, 1260, 6830, 630, 4550, 3250, 1030},
		[]int{1, 0, 1, 0, 1, 1, 1}, 0, 36)
	bad := 0
	for i := range before {
		if before[i] != fb.Pix[i] {
			bad++
		}
	}
	if bad != 0 {
		t.Fatalf("把開局的值原樣重畫，卻與原本差了 %d 個像素", bad)
	}
}

// TestBuildPlanStatesMatchOriginal 拿改過記憶體的原版局面對拍：
// 把同一組百分比餵給 remake，整張 640×480 必須逐位元組相同。
//
// 這些局面手點到不了（例如七列各 1%、某一列吃滿 100%），是用 `-poke`
// 把百分比陣列直接寫進原版記憶體造出來的（`docs/re/11`）。點擊序列與
// plan1 完全一樣，所以游標位置也一樣，可以整張比。
//
// 清單在 `workplace/explore/manifest.json`——那是一次性證據，
// 不在 Git 裡；缺檔就跳過，不做代用品。
func TestBuildPlanStatesMatchOriginal(t *testing.T) {
	root := "../.."
	dir := filepath.Join(root, "workplace/explore")
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Skip("缺 workplace/explore/manifest.json，跳過")
	}
	var mf struct {
		Cost   int                            `json:"cost"`
		States map[string][rules.BuildPlanRows]int `json:"states"`
	}
	if err := json.Unmarshal(raw, &mf); err != nil {
		t.Fatal(err)
	}
	s, err := LoadMainGameSpec(filepath.Join(root, "docs/internals/data/screen-build-plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	fnt, mapImg, faces, cur := gameScreenAssets(t)
	names := make([]string, 0, len(mf.States))
	for n := range mf.States {
		names = append(names, n)
	}
	sort.Strings(names)
	checked := 0
	for _, name := range names {
		want, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Skipf("缺 %s，跳過", name)
		}
		fb, _ := MainGame(s, fnt, mapImg, faces, cur)
		DrawBuildPlan(s, fb, fnt, &rules.BuildPlan{Cost: mf.Cost, Pct: mf.States[name]})
		if len(want) != len(fb.Pix) {
			t.Fatalf("%s 是 %d bytes，畫布是 %d", name, len(want), len(fb.Pix))
		}
		bad, fx, fy := 0, -1, -1
		for i := range want {
			if want[i] != fb.Pix[i] {
				if bad == 0 {
					fx, fy = i%gfx.Width, i/gfx.Width
				}
				bad++
			}
		}
		if bad != 0 {
			t.Errorf("%s：差 %d / %d 像素，第一處在 (%d,%d)", name, bad, len(want), fx, fy)
		}
		checked++
	}
	t.Logf("同狀態對拍 %d 個局面，各 %d 像素", checked, gfx.Width*gfx.Height)
}
