package game

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// 原版與remake的保持時間不同；依089明示等價輸入映射，驗同結果的完整畫面。
func TestLeftHoldMatchesFreshOracle(t *testing.T) {
	for _, sample := range []struct {
		name, hash string
		reach      func(*testing.T) *Game
		x, y       int
		steps      int
	}{
		{"tab-300000", "edbd2bec6acd24323119e2463036fdace17dbf19f87b09e78e5c6c82cc212d44",
			func(t *testing.T) *Game { return reachListTab(t, listTabAdmiral) }, 617, 388, 5},
		{"star-300000", "dba52021cee100a1ce356723a5c8a4541a61710a0a2184df6c2094b0f427c754",
			func(t *testing.T) *Game { return reachListScreen(t, 0) }, 615, 390, 2},
		{"tax-3000000", "0680a94e9660236f0ba90961dd30872c6208f1ca66bf0ea6c40fbe04f6e15770",
			reachTaxPlan, 471, 289, 2},
		{"pct-3000000", "17cca75c0b6c007883cfa944425a2dfb30e0619547c9fd05b0bd4cdff1bce11e",
			reachTaxPlan, 471, 329, 2},
	} {
		t.Run(sample.name, func(t *testing.T) {
			path := filepath.Join(root, "workplace", "oracle", "hold60-r58-"+sample.name+"-b", "screen.bin")
			want, err := os.ReadFile(path)
			if os.IsNotExist(err) {
				t.Skip("缺第58輪原版保持收據；用tools/left_hold_oracle.sh重生")
			}
			if err != nil || len(want) != 307200 || fmt.Sprintf("%x", sha256.Sum256(want)) != sample.hash {
				t.Fatalf("原版收據無效或漂移：%v", err)
			}
			g := sample.reach(t)
			g.Click(sample.x, sample.y)
			if sample.name == "tax-3000000" || sample.name == "pct-3000000" {
				g.LeftHold(sample.x, sample.y, leftPlanPeriod)
			} else {
				g.LeftHold(sample.x, sample.y, leftListDelay+leftListPeriod*time.Duration(sample.steps-2))
			}
			g.LeftRelease()
			fb := g.Frame()
			if !bytes.Equal(fb.Pix, want) {
				diff := 0
				for i := range want {
					if want[i] != fb.Pix[i] {
						diff++
					}
				}
				t.Fatalf("完整640×480差%d像素", diff)
			}
			if base := os.Getenv("LOGH3_LEFT_HOLD_COMPARE_OUT"); base != "" {
				out := filepath.Join(base, sample.name)
				if err := os.Mkdir(out, 0755); err != nil {
					t.Fatal(err)
				}
				if err := fb.WriteIndex(filepath.Join(out, "screen.bin")); err != nil {
					t.Fatal(err)
				}
				if err := fb.WritePNG(filepath.Join(out, "screen.png"), g.res.Menu.Palette()); err != nil {
					t.Fatal(err)
				}
				if err := g.SaveSlot(out, 0); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
