package rules

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// 收據 `s3form` 在「編隊」畫面開著的時候把 `ds:D072` 那 41 筆整個倒出來
// （`-dump-mem-at 1450000000:lin:61BA2:410:...`）。這裡拿槽位表算一遍，
// 41 格的世界座標必須完全相同（`docs/internals/46`）。
//
// 那一局的艦隊 31 在 (1320,812)、朝向 168；旗標非 0 的六格是已經有分隊
// 站著的，發令那一個分隊自己的格子被放回 0。
func TestFormationTableMatchesOracle(t *testing.T) {
	root := filepath.Join("..", "..")
	tab, err := os.ReadFile(filepath.Join(root, "workplace/oracle/s3form/form_tab.bin"))
	if err != nil || len(tab) < 41*10 {
		t.Skip("缺收據 s3form 的 form_tab.bin，跳過（不用代用品）")
	}
	s, err := LoadFormationSlots(
		filepath.Join(root, "docs/internals/data/battle-formation-slots.json"))
	if err != nil {
		t.Fatal(err)
	}
	const fx, fy, facing = 1320, 812, 168
	if s.Count != 41 {
		t.Fatalf("槽位表有 %d 格，要 41", s.Count)
	}
	for i := 0; i < 41; i++ {
		wantX := int(binary.LittleEndian.Uint16(tab[i*10:]))
		wantY := int(binary.LittleEndian.Uint16(tab[i*10+2:]))
		gotX, gotY := s.Station(fx, fy, facing, i)
		if gotX != wantX || gotY != wantY {
			t.Errorf("槽 %d 算出 (%d,%d)，收據是 (%d,%d)", i, gotX, gotY, wantX, wantY)
		}
	}
	// 槽 20 是中心，旗艦站那裡——座標必須就是艦隊自己的位置。
	if x, y := s.Station(fx, fy, facing, s.Centre); x != fx || y != fy {
		t.Errorf("中心槽算出 (%d,%d)，要 (%d,%d)", x, y, fx, fy)
	}
	// 已經有人的格子：收據上是這六個。
	var taken []int
	for i := 0; i < 41; i++ {
		if tab[i*10+4] != 0 {
			taken = append(taken, i)
		}
	}
	want := []int{20, 21, 28, 29, 30, 38}
	if len(taken) != len(want) {
		t.Fatalf("收據上有人的格子是 %v，要 %v", taken, want)
	}
	for i := range want {
		if taken[i] != want[i] {
			t.Fatalf("收據上有人的格子是 %v，要 %v", taken, want)
		}
	}
}
