package rules

import "testing"

// TestFleetSlotsTable 釘住 `297B:0209` 的九段（`docs/re/214`），
// 含「4 之後跳到 6、沒有 5」那一格。
func TestFleetSlotsTable(t *testing.T) {
	for _, c := range []struct{ level, want int }{
		{0, 1}, {1, 1},
		{2, 2}, {5, 2},
		{6, 3}, {15, 3},
		{16, 4}, {19, 4},
		{20, 6}, {23, 6},
		{24, 7}, {28, 7},
		{29, 8}, {33, 8},
		{34, 9}, {38, 9},
		{39, 10}, {41, 10}, {255, 10},
	} {
		if got := FleetSlots(c.level); got != c.want {
			t.Errorf("等級 %d → %d 組，預期 %d", c.level, got, c.want)
		}
	}
	// 值域裡不會出現 5。
	for level := 0; level <= 255; level++ {
		if FleetSlots(level) == 5 {
			t.Fatalf("等級 %d 算出 5 組——原版那張表沒有 5", level)
		}
	}
}

// TestFleetSlotsMatchesReceipt 收據 `turn260mem` 的三支（`docs/re/209`）。
func TestFleetSlotsMatchesReceipt(t *testing.T) {
	for _, c := range []struct{ admiral, level, groups int }{
		{47, 20, 6}, {52, 18, 4}, {55, 16, 4},
	} {
		if got := FleetSlots(c.level); got != c.groups {
			t.Errorf("提督 %d（等級 %d）→ %d 組，收據是 %d",
				c.admiral, c.level, got, c.groups)
		}
	}
}
