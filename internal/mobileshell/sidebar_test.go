package mobileshell

import (
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
)

// 側欄按鈕不壓到畫面上的原版區塊、不超出畫布；「退回」在兩種版面同一個位置；
// 命令列三鈕只在戰略版面；數字鍵盤只在防拷問答。
func TestSideButtons(t *testing.T) {
	cmds := []sideCmd{{codes: []uint16{1, 2, 3, 4}, gx: 380, gy: 130}, {codes: []uint16{1, 2, 3}, gx: 476, gy: 130}, {codes: []uint16{1, 2, 3, 4}, gx: 572, gy: 130}}
	for _, lw := range []int{1066, 853} {
		for _, strategic := range []bool{false, true} {
			for _, cp := range []bool{false, true} {
				if strategic && cp {
					continue // 防拷問答不會是戰略版面
				}
				lay := computeLayout(lw, strategic)
				bs := sideButtons(lay, stateView{copyProt: cp, battle: !cp && !strategic}, cmds)
				var back, cmd, digit, scroll int
				for _, b := range bs {
					if b.x < 0 || b.x+b.w > lw || b.y < 0 || b.y+b.h > gfx.Height {
						t.Errorf("寬 %d：按鈕超出畫布 %+v", lw, b)
					}
					for _, r := range lay.regions {
						w, h := float64(r.src.Dx())*r.scale, float64(r.src.Dy())*r.scale
						if float64(b.x+b.w) > r.x && float64(b.x) < r.x+w && float64(b.y+b.h) > r.y && float64(b.y) < r.y+h {
							t.Errorf("寬 %d 戰略 %v：按鈕壓到原版區塊 %+v", lw, lay.strategic, b)
						}
					}
					switch b.act {
					case actBack:
						back++
						if b.x != colX || b.y != backY {
							t.Errorf("退回位置 %d,%d", b.x, b.y)
						}
					case actGameClick:
						cmd++
					case actDigit:
						digit++
					case actScroll:
						scroll++
					}
				}
				wantCmd := 0
				if lay.strategic {
					wantCmd = 3
				}
				wantDigit := 0
				if cp {
					wantDigit = 10
				}
				wantScroll := 0
				if !cp && !strategic {
					wantScroll = 4
				}
				if scroll != wantScroll {
					t.Errorf("寬 %d 戰略 %v：方向鈕 %d 顆，預期 %d", lw, lay.strategic, scroll, wantScroll)
				}
				if back != 1 || cmd != wantCmd || digit != wantDigit {
					t.Errorf("寬 %d 戰略 %v 防拷 %v：退回 %d、命令 %d、數字 %d", lw, lay.strategic, cp, back, cmd, digit)
				}
			}
		}
	}
	if bs := sideButtons(computeLayout(640, false), stateView{copyProt: true}, cmds); len(bs) != 0 {
		t.Errorf("4:3 沒有側欄，卻回了 %d 顆", len(bs))
	}
}
