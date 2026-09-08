package rules

import "testing"

// TestDateMatchesOracle：回合計數器 → 年月日。
//
// 前三筆的計數器是從執行期記憶體讀出來的（`ds:8F18`，傾印 m1／m4／m7），
// 其餘是畫面上讀到的日期回推：`turn10` 按了六次「實行」是 1 月 19 日、
// `turn12b` 九次是 1 月 28 日、`turnnext` 十次之後是 2 月 1 日。
func TestDateMatchesOracle(t *testing.T) {
	for _, c := range []struct {
		turn             int
		year, month, day int
		note             string
	}{
		{0, 796, 1, 1, "開局（m1）"},
		{1, 796, 1, 4, "一次實行（m4）"},
		{2, 796, 1, 7, "兩次實行（m7）"},
		{6, 796, 1, 19, "六次（turn10）"},
		{9, 796, 1, 28, "九次（turn12b）"},
		{10, 796, 2, 1, "十次（turnnext）"},
		{120, 797, 1, 1, "滿一年"},
	} {
		y, m, d := Date(c.turn, 796)
		if y != c.year || m != c.month || d != c.day {
			t.Errorf("第 %d 回合 → %d年%d月%d日，預期 %d年%d月%d日（%s）",
				c.turn, y, m, d, c.year, c.month, c.day, c.note)
		}
	}
}
