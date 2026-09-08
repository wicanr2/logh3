package screen

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
)

// TestFormationMarksMatchOracle：「編隊」畫面上每個空格子的 16×16 標記，
// 逐位元組對收據 `s3form`（`docs/internals/46`）。
//
// 收據同時倒出 `ds:D072` 那 41 筆，所以格子的世界座標不必再推——直接
// 讀出來，減視野原點就是畫面位置。原版畫的是 `SPRITE.SPR` 的第 11 塊，
// **色號 0 透空**：那 118 個位置在收據上留著戰場底色 3。
// （「透空」與「0 換成 3」在這份收據上分不開——沒有一個空格子壓在
// 色 6 的格線上。這裡照透空做，因為原版的顯示物件本來就是疊上去的。）
//
// 五個空格子排除在外，各有明確理由：畫面上有東西壓在標記上面。
// 槽 19／24／25 底下坐著分隊（分隊畫在自己的現在位置，不是隊形位置），
// 槽 15／16 被 dosgolem 畫的滑鼠游標壓到。其餘 30 個全比。
func TestFormationMarksMatchOracle(t *testing.T) {
	root := "../.."
	orig := filepath.Join(root, "org_game/extracted")
	dir := filepath.Join(root, "workplace/oracle/s3form")
	for _, p := range []string{
		filepath.Join(dir, "screen.bin"), filepath.Join(dir, "form_tab.bin"),
		filepath.Join(orig, "SPRITE.SPR"),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過同狀態對拍（不用代用品）", p)
		}
	}
	scr, err := os.ReadFile(filepath.Join(dir, "screen.bin"))
	if err != nil {
		t.Fatal(err)
	}
	tab, err := os.ReadFile(filepath.Join(dir, "form_tab.bin"))
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
	if FormationMarkTile >= len(sprites) {
		t.Fatalf("SPRITE.SPR 只有 %d 塊，取不到第 %d 塊", len(sprites), FormationMarkTile)
	}
	mark := sprites[FormationMarkTile]

	// 視野原點：這一局艦隊 31 的旗艦在世界 (1320,812)，畫在 (280,272)。
	const vx, vy = 1040, 540
	const w = 640
	skip := map[int]bool{15: true, 16: true, 19: true, 24: true, 25: true}

	// 戰場底色（`docs/re/82`）。透空的位置在收據上就是它。
	const fieldBackdrop = 3

	checked := 0
	for i := 0; i < 41; i++ {
		if tab[i*10+4] != 0 || skip[i] {
			continue
		}
		sx := int(binary.LittleEndian.Uint16(tab[i*10:])) - vx
		sy := int(binary.LittleEndian.Uint16(tab[i*10+2:])) - vy
		if sx < 0 || sy < 0 || (sy+16)*w > len(scr) {
			continue
		}
		diff, clear := 0, 0
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				got, want := scr[(sy+y)*w+sx+x], mark.Pix[y*16+x]
				switch {
				case want == 0:
					clear++
					if got != fieldBackdrop {
						diff++
					}
				case got != want:
					diff++
				}
			}
		}
		if diff != 0 {
			t.Errorf("槽 %d 在 (%d,%d)：256 個像素有 %d 個不同", i, sx, sy, diff)
		}
		if clear != 118 {
			t.Errorf("槽 %d：圖塊透空的像素 %d 個，第 11 塊是 118 個", i, clear)
		}
		checked++
	}
	if checked != 30 {
		t.Fatalf("比了 %d 個空格子，收據上該有 30 個", checked)
	}
}
