package rules

import (
	"path/filepath"
	"testing"
)

// TestFormationStationMatchesOracle 拿收據 `s3fltmove` 的九個分隊對隊形算式
// （`docs/re/126`）。
//
// 那一份收據裡第 28 艦隊在 (1330,941)、朝向 163，九個分隊的槽索引與原版
// 寫進 `+0x28`／`+0x2A` 的目的地都從記憶體讀出來，逐一列在下面。
func TestFormationStationMatchesOracle(t *testing.T) {
	s, err := LoadFormationSlots(filepath.Join("../..",
		"docs/internals/data/battle-formation-slots.json"))
	if err != nil {
		t.Skipf("缺槽位表，跳過：%v", err)
	}
	if s.Count != 41 || s.Centre != 20 {
		t.Fatalf("槽位表不對：%d 個、中心 %d", s.Count, s.Centre)
	}
	if x, y := s.Offset(20); x != 0 || y != 0 {
		t.Errorf("槽 20 應該是 (0,0)，是 (%d,%d)", x, y)
	}

	const fx, fy, facing = 1330, 941, 163
	cases := []struct{ slot, wantX, wantY int }{
		{20, 1330, 941},
		{18, 1282, 983},
		{19, 1306, 962},
		{21, 1354, 920},
		{22, 1378, 899},
		{37, 1348, 1010},
		{39, 1396, 969},
		{28, 1327, 986},
		{30, 1375, 944},
	}
	for _, c := range cases {
		gx, gy := s.Station(fx, fy, facing, c.slot)
		if gx != c.wantX || gy != c.wantY {
			t.Errorf("槽 %d：算出 (%d,%d)，原版是 (%d,%d)",
				c.slot, gx, gy, c.wantX, c.wantY)
		}
	}
}

// TestRotate256Rounding 釘住旋轉的捨去：原版是「加 512 再算術右移 10 位」。
//
// 換成截斷會與收據對不上（`docs/re/126` 量過：五、六個分隊差一個像素），
// 所以這裡確認兩者真的不同——不然這條測試等於沒在測。
func TestRotate256Rounding(t *testing.T) {
	diff := 0
	for a := 0; a < BattleTurn; a++ {
		for _, p := range [][2]int{{0, -64}, {-64, -32}, {48, 16}} {
			gx, _ := Rotate256(p[0], p[1], a)
			tx := (p[0]*cos256(a) - p[1]*sin256(a)) / 1024
			if gx != tx {
				diff++
			}
		}
	}
	if diff == 0 {
		t.Fatal("四捨五入與截斷沒有任何差別，測試沒有鑑別力")
	}
	t.Logf("768 組樣本裡有 %d 組與截斷不同", diff)
}
