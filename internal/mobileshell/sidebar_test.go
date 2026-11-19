package mobileshell

import (
	"testing"

	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/gfx"
)

// 側欄按鈕只畫在原版畫面兩側，不壓到中央 640×480；數字鍵盤只在防拷問答出現。
func TestSideButtonsStayInBands(t *testing.T) {
	for _, w := range []int{1066, 853} { // 20:9、16:9
		left, right := gameOrigin(w), gameOrigin(w)+gfx.Width
		for _, st := range []game.State{game.StateMainMenu, game.StateCopyProtection, game.StateCommandPhase} {
			bs := sideButtons(w, st)
			digits := 0
			for _, b := range bs {
				if b.x < 0 || b.x+b.w > w || b.y < 0 || b.y+b.h > gfx.Height {
					t.Errorf("寬 %d 狀態 %d：按鈕超出畫布 %+v", w, st, b)
				}
				if b.x+b.w > left && b.x < right {
					t.Errorf("寬 %d 狀態 %d：按鈕壓到原版畫面 %+v", w, st, b)
				}
				if b.act == actDigit {
					digits++
				}
			}
			want := 0
			if st == game.StateCopyProtection {
				want = 10
			}
			if digits != want {
				t.Errorf("寬 %d 狀態 %d：數字鍵 %d 顆，預期 %d", w, st, digits, want)
			}
		}
	}
	if bs := sideButtons(640, game.StateCopyProtection); len(bs) != 0 {
		t.Errorf("4:3 沒有側欄，卻回了 %d 顆", len(bs))
	}
}
