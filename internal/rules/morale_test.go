package rules

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDriftRegionMoraleSteps 驗 `056A:0AA5` 的三條分支與兩個邊界：
// 到目標就停、無號 byte 的 255 往下走。
func TestDriftRegionMoraleSteps(t *testing.T) {
	cases := []struct{ owner, morale, want int }{
		{0, 150, 149}, {0, 1, 0}, {0, 0, 0}, {0, 255, 254},
		{1, 50, 51}, {1, 199, 200}, {1, 200, 200}, {1, 201, 200},
		{2, 20, 21}, {2, 99, 100}, {2, 100, 100}, {2, 101, 100},
	}
	recs := make([]byte, len(cases)*RegionRecord)
	for i, c := range cases {
		recs[i*RegionRecord+regionOwner] = byte(c.owner)
		recs[i*RegionRecord+regionMorale] = byte(c.morale)
	}
	DriftRegionMorale(recs)
	for i, c := range cases {
		if got := int(recs[i*RegionRecord+regionMorale]); got != c.want {
			t.Errorf("擁有者 %d 民心 %d：走一格得 %d，要 %d", c.owner, c.morale, got, c.want)
		}
	}
}

// 收據 `mrx295`（`docs/re/295`）：新遊戲在回合 0 的命令階段把 12 格的民心
// 改成偏離目標的值，連按三次「實行」，每次之間倒一份星域表。收據
// `mrx295d`：改的是四格的**擁有者**（行星不動），按一次「實行」。收據
// `mrx295f`：回合計數器改成 9 再按兩次「實行」，**跨過月底**（兩個座位的
// 月底結算都跑完，`2703:03AF` 在 `#280,485,809`）。
// 相鄰兩份傾印之間 `056A:0AA5` 都正好跑一次（收據的 `-regs-at`）。
var moraleDriftReceipts = []struct {
	dir, prefix string
	snaps       []string
}{
	{"mrx295", "mrx295", []string{"r0p", "r1", "r2", "r3"}},
	{"mrx295d", "mrx295d", []string{"r0p", "r1"}},
	{"mrx295f", "mrx295f", []string{"r0p", "r1", "r2"}},
}

// TestDriftRegionMoraleMatchesReceipts 拿前一份傾印跑一次
// `DriftRegionMorale`，32 格的 `+0x7C` 要與下一份逐格相同。
func TestDriftRegionMoraleMatchesReceipts(t *testing.T) {
	moved := 0
	for _, rc := range moraleDriftReceipts {
		var tabs [][]byte
		for _, s := range rc.snaps {
			p := filepath.Join("..", "..", "workplace", "oracle", rc.dir, rc.prefix+"_"+s+".bin")
			b, err := os.ReadFile(p)
			if err != nil {
				t.Skipf("缺收據 %s，跳過（不用代用品）", p)
			}
			if len(b) != RegionCount*RegionRecord {
				t.Fatalf("%s 長度 %d", p, len(b))
			}
			tabs = append(tabs, b)
		}
		for k := 0; k+1 < len(tabs); k++ {
			got := append([]byte(nil), tabs[k]...)
			DriftRegionMorale(got)
			for i := 0; i < RegionCount; i++ {
				o := i*RegionRecord + regionMorale
				if got[o] != tabs[k][o] {
					moved++
				}
				if got[o] != tabs[k+1][o] {
					t.Errorf("%s %s→%s 星域 %d：算出 %d，收據 %d", rc.dir,
						rc.snaps[k], rc.snaps[k+1], i, got[o], tabs[k+1][o])
				}
			}
		}
	}
	// 正對照：收據裡要真的有會動的格子，否則全等只證明「都不動」。
	if moved != 40 {
		t.Errorf("會動的格子 %d 格，收據是 40 格（mrx295 26 ＋ mrx295d 4 ＋ mrx295f 10）", moved)
	}
}
