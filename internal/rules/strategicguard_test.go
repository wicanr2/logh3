package rules

import (
	"encoding/binary"
	"testing"
)

// 合成資料只驗原池消費契約；原版五案另由固定dosgolem收據驗收。
func TestStrategicGuardLivePool(t *testing.T) {
	p := make([]byte, PlanetRecord)
	p[planetOwner] = 1
	s := &Slot{Ship: make([]byte, 3*ShipRecord)}
	for i, n := range []int{3000, 60000, 20000} {
		r := s.ShipAt(i)
		r[0], r[4] = 1, 0
		binary.LittleEndian.PutUint16(r[8:], uint16(n))
	}
	pools := []int{0}
	guard, err := NewStrategicGuard(s, pools)
	if err != nil || guard(p) != 30 {
		t.Fatal("近距守軍", err)
	}
	pools[0] = 2
	binary.LittleEndian.PutUint16(s.ShipAt(0)[8:], 4000)
	if guard(p) != 40 {
		t.Fatal("讀取過期艘數或外部索引")
	}
	s.ShipAt(0)[0x1e] = 0x80
	if guard(p) != 0 {
		t.Fatal("Out仍提供守軍")
	}
	s.ShipAt(0)[0x1e] = 0
	binary.LittleEndian.PutUint16(s.ShipAt(0)[0x34:], 512)
	if guard(p) != 0 {
		t.Fatal("移遠仍提供守軍")
	}
	guard, err = NewStrategicGuard(s, []int{1, 2})
	if err != nil || guard(p) != 655 {
		t.Fatal("沒有先飽和再除100", err)
	}
	for _, pools := range [][]int{{-1}, {3}} {
		if _, err := NewStrategicGuard(s, pools); err == nil {
			t.Fatal("接受越界清單", pools)
		}
	}
	if _, err := NewStrategicGuard(nil, []int{0}); err == nil {
		t.Fatal("缺原池被當成0守軍")
	}
	guard, err = NewStrategicGuard(nil, nil)
	if err != nil || guard(p) != 0 {
		t.Fatal("空清單", err)
	}
}
