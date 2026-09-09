package screen

import (
	"path/filepath"
	"testing"
)

// 收據 `s3msg`：分隊選單選「移動」之後點一個過不去的格子，原版跳
// 「這是無法行動的地方。」（`ds:4A24`）。這裡把訊息框畫出來，
// 與收據逐位元組比框那一塊（`docs/re/137`）。
func TestMessageBoxGeometryMatchesOracle(t *testing.T) {
	root := "../.."
	l, err := LoadCommandMenuSpec(
		filepath.Join(root, "docs/internals/data/battle-command-menus.json"))
	if err != nil {
		t.Fatal(err)
	}
	s, err := LoadMessageBoxSpec(
		filepath.Join(root, "docs/internals/data/message-box.json"))
	if err != nil {
		t.Fatal(err)
	}
	x, y, w, h, ok := s.Box(2, l)
	if !ok {
		t.Fatal("讀不到種類 2 的樣板")
	}
	// 收據 `s3msg`：分隊選單選「移動」之後點一個過不去的格子，原版跳
	// 「這是無法行動的地方。」。框量到 (65,181) 350×118（`docs/re/137`）。
	if x != 65 || y != 181 || w != 350 || h != 118 {
		t.Fatalf("框是 (%d,%d) %d×%d，收據量到 (65,181) 350×118", x, y, w, h)
	}
	// 位置 ＝ (格x×16+1, (格y+2)×20+1)、大小 ＝ (寬×16+30)×(高×20+38)
	// ——大小的公式與命令選單完全相同，只有 y 多兩格。
	g := &l.Layout
	if w != 20*g.CellW+g.FramePadW || h != 4*g.CellH+g.FramePadH {
		t.Error("大小的公式與命令選單不一致")
	}

	// **框線本身的畫法還不同**（原版走 `3603:0006`，不是選單那一支
	// `ui.MenuFrame`）：收據左緣是 `8 F 8 7 F F 3`、拿 MenuFrame 畫出來是
	// `8 3 F F 7 8 3`，元素一樣但排法不同，開頭幾列還有階梯狀的斜切。
	// 所以現在只驗幾何，框線等 `3603:0006` 讀出來再畫（`docs/re/137`）。
}
