package rules

import (
	"encoding/binary"
	"testing"
)

func stockRecord(faction, kind, n int) []byte {
	r := make([]byte, RegionRecord)
	o := regionShips + faction*regionShipKinds*4 + kind*4
	binary.LittleEndian.PutUint32(r[o:o+4], uint32(n))
	return r
}

// TestTakeRegionShips 測 `2703:1433` 的三種情況（`docs/re/215`）。
func TestTakeRegionShips(t *testing.T) {
	// 存量夠 → 取滿編，剩下的留著。
	r := stockRecord(1, 2, 1000)
	if got := TakeRegionShips(r, 1, 2, 120); got != 120 {
		t.Errorf("存量 1000、滿編 120 取到 %d，預期 120", got)
	}
	if left := RegionShips(r, 1)[2]; left != 880 {
		t.Errorf("扣完剩 %d，預期 880", left)
	}

	// 存量不足 → 取存量，扣成 0。
	r = stockRecord(0, 5, 30)
	if got := TakeRegionShips(r, 0, 5, 120); got != 30 {
		t.Errorf("存量 30、滿編 120 取到 %d，預期 30", got)
	}
	if left := RegionShips(r, 0)[5]; left != 0 {
		t.Errorf("扣完剩 %d，預期 0", left)
	}

	// 空的再取一次 → 0，而且不會變成負數。
	if got := TakeRegionShips(r, 0, 5, 120); got != 0 {
		t.Errorf("空存量取到 %d，預期 0", got)
	}
	if left := RegionShips(r, 0)[5]; left != 0 {
		t.Errorf("再取一次之後剩 %d，預期 0", left)
	}

	// 兩個陣營的存量互不影響。
	r = stockRecord(0, 1, 500)
	TakeRegionShips(r, 0, 1, 100)
	if got := RegionShips(r, 1)[1]; got != 0 {
		t.Errorf("扣帝國卻動到同盟：%d", got)
	}

	// 界外一律回 0、不動記錄。
	for _, c := range []struct{ f, k, full int }{
		{2, 0, 100}, {-1, 0, 100}, {0, 7, 100}, {0, -1, 100}, {0, 0, 0},
	} {
		r := stockRecord(0, 0, 999)
		if got := TakeRegionShips(r, c.f, c.k, c.full); got != 0 {
			t.Errorf("界外 %+v 取到 %d，預期 0", c, got)
		}
		if left := RegionShips(r, 0)[0]; left != 999 {
			t.Errorf("界外 %+v 動到存量：%d", c, left)
		}
	}
}
