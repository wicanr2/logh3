package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/screen"
	"github.com/wicanr2/logh3/internal/ui"
)

const root = "../.."

// load 把玩家路徑要用的東西全部備好；缺原版或收據就 skip。
func load(t *testing.T) *Resources {
	t.Helper()
	orig := filepath.Join(root, "org_game/extracted")
	need := []string{
		filepath.Join(orig, "GRAPH.IMG"), filepath.Join(orig, "FACE.AMG"),
		filepath.Join(orig, "ADMIRAL.DTX"), filepath.Join(orig, "STRMAP.BIT"),
		filepath.Join(orig, "STRMAP.RGN"),
	}
	for _, p := range need {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過（不用代用品）", p)
		}
	}
	fnt, err := font.Load(filepath.Join(orig, "GRAPH.IMG"))
	if err != nil {
		t.Fatal(err)
	}
	faceData, err := os.ReadFile(filepath.Join(orig, "FACE.AMG"))
	if err != nil {
		t.Fatal(err)
	}
	faces, err := asset.OpenAMG(faceData)
	if err != nil {
		t.Fatal(err)
	}
	admiral, err := os.ReadFile(filepath.Join(orig, "ADMIRAL.DTX"))
	if err != nil {
		t.Fatal(err)
	}
	menu, err := screen.LoadMenuSpec(filepath.Join(root, "docs/internals/data/screen-main-menu.json"))
	if err != nil {
		t.Fatal(err)
	}
	cp, err := screen.LoadCopyProtSpec(filepath.Join(root, "docs/internals/data/screen-copy-protection.json"))
	if err != nil {
		t.Fatal(err)
	}
	mg, err := screen.LoadMainGameSpec(filepath.Join(root, "docs/internals/data/screen-main-game.json"))
	if err != nil {
		t.Fatal(err)
	}
	cur, err := ui.LoadCursor(filepath.Join(root, "docs/internals/data/cursor.json"))
	if err != nil {
		t.Fatal(err)
	}
	bit, err := os.ReadFile(filepath.Join(orig, "STRMAP.BIT"))
	if err != nil {
		t.Fatal(err)
	}
	rgn, err := os.ReadFile(filepath.Join(orig, "STRMAP.RGN"))
	if err != nil {
		t.Fatal(err)
	}
	colours, err := regionColours(filepath.Join(root, "docs/internals/data/strmap-regions.json"))
	if err != nil {
		t.Fatal(err)
	}
	star, err := asset.BuildStarMap(bit, rgn, colours, 16, 20, 20)
	if err != nil {
		t.Fatal(err)
	}
	return &Resources{Font: fnt, Faces: faces, Cursor: cur, Menu: menu,
		CopyProt: cp, MainGame: mg, StarMap: star, Admiral: admiral}
}

func regionColours(path string) ([]uint8, error) {
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

// matches 把目前畫面與 oracle 收據逐位元組比。
func matches(t *testing.T, g *Game, batch string) {
	t.Helper()
	p := filepath.Join(root, "workplace/oracle", batch, "screen.bin")
	want, err := os.ReadFile(p)
	if err != nil {
		t.Skipf("沒有收據 %s，跳過（tools/oracle.sh 可重建）", batch)
	}
	got := g.Frame().Pix
	bad := 0
	for i := range want {
		if want[i] != got[i] {
			bad++
		}
	}
	if bad != 0 {
		t.Fatalf("%s：與原版不同的像素 %d / %d", batch, bad, len(want))
	}
}

// TestPlayerPathReachesGame 走完整條玩家路徑，每一步都要與原版收據
// 逐位元組相同：主選單反白 → 點擊 → 防拷問答 → 輸入答案 → 遊戲本體。
//
// **這是「正常玩家路徑」的收據**，不是 direct-entry：每個畫面都是
// 前一個畫面的輸入帶進去的。
func TestPlayerPathReachesGame(t *testing.T) {
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	if g.State() != StateMainMenu {
		t.Fatal("開場不是主選單")
	}
	g.MoveMouse(300, 170)
	matches(t, g, "r0")

	g.Click(300, 170)
	if g.State() != StateCopyProtection {
		t.Fatal("點了主選單沒有進防拷問答")
	}
	matches(t, g, "cb1")

	if g.Answer() != 138 {
		t.Fatalf("答案是 %d，預期 138", g.Answer())
	}
	for _, c := range []byte{'1', '3', '8', '\r'} {
		g.Key(c)
	}
	if g.State() != StateMainGame {
		t.Fatal("答對了卻沒進遊戲本體")
	}
	matches(t, g, "key1")
}

// TestWrongAnswerStaysOnQuestion：答錯不會放行。
func TestWrongAnswerStaysOnQuestion(t *testing.T) {
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, 170)
	for _, c := range []byte{'1', '2', '3', '\r'} {
		g.Key(c)
	}
	if g.State() != StateCopyProtection {
		t.Fatal("答錯卻進了遊戲")
	}
	if g.Typed() != "" {
		t.Fatalf("答錯之後輸入沒清掉：%q", g.Typed())
	}
}

// TestClickOutsideMenuDoesNothing：點在選單以外不算數。
func TestClickOutsideMenuDoesNothing(t *testing.T) {
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	g.Click(10, 10)
	if g.State() != StateMainMenu {
		t.Fatal("點在選單外卻換了畫面")
	}
}
