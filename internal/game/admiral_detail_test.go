package game

import "testing"

// TestAdmiralDetailMatchesOracle 從「出擊命令」的待命中提督清單點一列，
// 開出那位提督的詳細資料（`docs/re/65`）。
//
// 四張收據點的是不同列：姓名、軍階、頭像、等級、十個能力值與旗艦
// 都得跟著換。第 14 列的錦茲**沒有艦隊**（記錄 `+0x07` ＝ `0xFF`），
// 那一張沒有「旗艦」那一行。
func TestAdmiralDetailMatchesOracle(t *testing.T) {
	for _, c := range []struct {
		y     int
		batch string
	}{
		// y 用收據當初點的那一點——游標畫在滑鼠停的地方，差三個像素
		// 就整隻箭頭對不上。
		{85, "adm1e"},   // 第 0 列 萊因哈特上將
		{105, "adm2e"},  // 第 1 列 吉爾菲艾斯少將
		{122, "adm3e"},  // 第 2 列 米達麥亞少將
		{362, "adm15e"}, // 第 14 列 錦茲少校，沒有旗艦
	} {
		t.Run(c.batch, func(t *testing.T) {
			g := reachCommandPhase(t)
			g.Click(352, 125) // 戰略命令
			g.Click(400, 125) // 出擊命令
			if g.State() != StateStandbyAdmirals {
				t.Fatalf("點了出擊命令之後在 %v", g.State())
			}
			g.Click(400, c.y)
			if g.State() != StateAdmiralDetail {
				t.Fatalf("點了 y=%d 之後在 %v", c.y, g.State())
			}
			matches(t, g, c.batch)
		})
	}
}
