package screen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// TestSlotListMatchesOracle 是存檔畫面那 11 列的同狀態對拍（`docs/spec/008`）：
// 把規格裡量好畫死的那 11 行拿掉，改從 `SAVEDATA.IDX` 算出來畫，
// 整張 640×480 要與原版逐位元組相同（游標那一塊除外）。
func TestSlotListMatchesOracle(t *testing.T) {
	root := "../.."
	specPath := filepath.Join(root, "docs/internals/data/screen-load-game.json")
	oracle := filepath.Join(root, "workplace/oracle/load1e/screen.bin")
	idxPath := filepath.Join(root, "org_game/extracted/SAVEDATA.IDX")
	for _, p := range []string{specPath, oracle, idxPath} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過同狀態對拍（不用代用品）", p)
		}
	}
	spec, err := LoadMainGameSpec(specPath)
	if err != nil {
		t.Fatal(err)
	}
	list, err := LoadSlotList(filepath.Join(root,
		"docs/internals/data/save-slot-list.json"))
	if err != nil {
		t.Fatal(err)
	}
	// 規格裡已經沒有那 11 行了——它們由 SlotList 畫。
	idx, err := os.ReadFile(idxPath)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := rules.ParseSaveIndex(idx)
	if err != nil {
		t.Fatal(err)
	}
	fnt, mapImg, faces, cur := gameScreenAssets(t)
	fb, _ := MainGame(spec, fnt, mapImg, faces, cur)
	list.Draw(fb, fnt, entries, rules.Epoch)

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
		_ = os.WriteFile(filepath.Join(root, "workplace/remake-slotlist.bin"), fb.Pix, 0o644)
		t.Fatalf("與原版不同的像素 %d / %d", bad, len(want))
	}
}
