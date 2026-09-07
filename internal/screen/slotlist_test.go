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
	fb, _ := MainGameWithBoard(spec, fnt, mapImg, faces, cur, openingBoardPanel(t))
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

// TestSaveBarMatchesOracle 是「存進度」那條紅條的對拍（`docs/spec/010`）：
// 條子跟著滑鼠走，畫在滑鼠停的那一列。收據 `save1e` 的滑鼠在 (560,125)
// ——畫面 y ＝ 165，落在第 1 列。
func TestSaveBarMatchesOracle(t *testing.T) {
	root := "../.."
	specPath := filepath.Join(root, "docs/internals/data/screen-save-game.json")
	oracle := filepath.Join(root, "workplace/oracle/save1e/screen.bin")
	for _, p := range []string{specPath, oracle} {
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
	// 規格裡的 `mouse.y` 是 dosgolem 的滑鼠座標，比畫面座標少 40。
	row := list.SaveRowAt(spec.Mouse.Y + 40)
	if row != 1 {
		t.Fatalf("滑鼠 y=%d 落在第 %d 列，要第 1 列", spec.Mouse.Y, row)
	}
	fnt, mapImg, faces, cur := gameScreenAssets(t)
	fb, _ := MainGameWithBoard(spec, fnt, mapImg, faces, cur, openingBoardPanel(t))
	// 條子畫上去會把那一列蓋掉，原版接著把字重畫在條子上。
	list.DrawSaveBar(fb, row)
	list.DrawSaveRow(fb, fnt, row)
	cur.Draw(fb, spec.Mouse.X, spec.Mouse.Y) // 條子蓋掉游標，原版再畫一次

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
		_ = os.WriteFile(filepath.Join(root, "workplace/remake-savebar.bin"), fb.Pix, 0o644)
		t.Fatalf("與原版不同的像素 %d / %d", bad, len(want))
	}
}
