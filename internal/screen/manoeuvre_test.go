package screen

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

// 收據 `s3cmdrange`：第 28 艦隊（回轉 15、巡航 30 → 距離 240、朝向 190）
// 停在選點狀態，圓心是旗艦圖塊中心 (298,409)。
//
// 外形是**色 1 XOR 上去**的，所以畫面上的顏色隨底色而定：戰場底色 3
// 得到 2、虛線格線的 6 得到 7、格線的黑洞 0 得到 1。這裡把整條折線
// 依原版的裁切與畫線走一遍，逐點檢查落點的顏色**都能被 XOR 解釋**
// （`docs/re/155`）。
func TestManoeuvreOuterMatchesOracle(t *testing.T) {
	const (
		w, h           = 640, 480
		cx, cy         = 298, 409
		turn, dist, fc = 15, 240, 190
	)
	scr, err := os.ReadFile(filepath.Join("..", "..",
		"workplace/oracle/s3cmdrange/screen.bin"))
	if err != nil || len(scr) < w*h {
		t.Skip("缺收據 s3cmdrange，跳過（不用代用品）")
	}
	// 只畫外形：底圖全部給 0，畫完哪些像素是 ManoeuvreOuterInk 就是
	// 折線亮起來的點（同一點被畫兩次會 XOR 回 0，與原版一致）。
	fb := gfx.NewFramebuffer(0)
	r := ClipRect{X0: 16, Y0: 60, X1: 463, Y1: 419}
	pts := rules.ManoeuvreBoundary(turn, dist, fc)
	if len(pts) < 2 {
		t.Fatal("算不出外形")
	}
	for i := 1; i < len(pts); i++ {
		r.XorLine(fb, cx+pts[i-1][0], cy+pts[i-1][1],
			cx+pts[i][0], cy+pts[i][1], ManoeuvreOuterInk)
	}
	hist := map[uint8]int{}
	var lit int
	for y := r.Y0; y <= r.Y1; y++ {
		for x := r.X0; x <= r.X1; x++ {
			if fb.At(x, y) == 0 {
				continue
			}
			lit++
			c := scr[y*w+x]
			hist[c]++
			// 收據上的顏色 XOR 掉墨色，必須回到戰場上真的有的底色：
			// 3 底色、6 虛線格線、0 格線的洞。
			switch c ^ ManoeuvreOuterInk {
			case 3, 6, 0:
			default:
				t.Errorf("(%d,%d) 收據是色 %d，XOR 回去是 %d，戰場上沒有這個底色",
					x, y, c, c^ManoeuvreOuterInk)
			}
		}
	}
	// 收據對出來的那一組數字（`docs/re/155`）。
	const wantLit, want2, want7, want1 = 487, 482, 3, 2
	if lit != wantLit {
		t.Errorf("折線在戰場裡亮了 %d 個像素，收據對出來的是 %d", lit, wantLit)
	}
	if hist[2] != want2 || hist[7] != want7 || hist[1] != want1 {
		keys := make([]int, 0, len(hist))
		for k := range hist {
			keys = append(keys, int(k))
		}
		sort.Ints(keys)
		got := map[int]int{}
		for _, k := range keys {
			got[k] = hist[uint8(k)]
		}
		t.Errorf("落點的顏色分佈 %v，收據是 {2:%d 7:%d 1:%d}",
			got, want2, want7, want1)
	}
	t.Logf("外形 %d 點、折線亮 %d 個像素：色 2 有 %d、壓在格線上的色 7 有 %d、色 1 有 %d",
		len(pts), lit, hist[2], hist[7], hist[1])
}
