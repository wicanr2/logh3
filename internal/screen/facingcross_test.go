package screen

import (
	"os"
	"path/filepath"
	"testing"
)

// 收據 `s3cmdface`：點了目的地畫面 (200,300) 之後停在選朝向的狀態，
// 那支艦隊的朝向是 190。十字中心在 (207,307) ＝ 點到的座標各 ＋7
// （`docs/re/122`）。
//
// 十字是 XOR 上去的，而且**色 7 在戰場地形上本來就有**——所以不能用
// 「收據上的色 7 都要落在線上」當判準（會被地形吃掉）。這裡改用
// **對照組**：拿同一組線去試別的朝向，真正的朝向要明顯勝出。
func TestFacingCrossMatchesOracle(t *testing.T) {
	const (
		w, cx, cy, facing = 640, 207, 307, 190
		crossInk          = 7
	)
	scr, err := os.ReadFile(filepath.Join("..", "..",
		"workplace/oracle/s3cmdface/screen.bin"))
	if err != nil || len(scr) < w*400 {
		t.Skip("缺收據 s3cmdface，跳過（不用代用品）")
	}
	score := func(f int) (on, total int) {
		seen := map[[2]int]bool{}
		for _, s := range FacingCrossSegments(f) {
			x0, y0, x1, y1 := cx+s[0], cy+s[1], cx+s[2], cy+s[3]
			n := max2(abs(x1-x0), abs(y1-y0))
			for i := 0; i <= n; i++ {
				x := x0 + (x1-x0)*i/n
				y := y0 + (y1-y0)*i/n
				if x < 16 || y < 40 || x >= 464 || y >= 400 || seen[[2]int{x, y}] {
					continue
				}
				seen[[2]int{x, y}] = true
				total++
				if scr[y*w+x] == crossInk {
					on++
				}
			}
		}
		return
	}
	want, wantN := score(facing)
	if wantN == 0 {
		t.Fatal("線整條都在畫面外")
	}
	best := 0
	for f := 0; f < 256; f += 8 {
		if f == facing/8*8 || f == facing {
			continue
		}
		n, _ := score(f)
		if n > best {
			best = n
		}
	}
	t.Logf("朝向 190：%d/%d 個像素是色 7；其他朝向最多只有 %d 個",
		want, wantN, best)
	if want < wantN/2 {
		t.Errorf("真正的朝向只對上 %d/%d，太少", want, wantN)
	}
	if want <= best {
		t.Errorf("朝向 190 對上 %d 個，別的朝向卻有 %d 個——分不出來", want, best)
	}
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}
