package rules

import (
	"bytes"
	"testing"
)

func TestRegionListFleetCountsBoundary(t *testing.T) {
	rec := make([]byte, RegionRecord)
	copy(rec[0x28:], []byte{254, 0, 255, 7})
	before := bytes.Clone(rec)
	for _, c := range []struct {
		viewer int
		want   [2]byte
	}{{0, [2]byte{254, 7}}, {1, [2]byte{255, 0}}} {
		got, ok := RegionListFleetCounts(rec, c.viewer)
		if !ok || got != c.want {
			t.Fatalf("座位%d：%v／%v，預期%v", c.viewer, got, ok, c.want)
		}
	}
	if !bytes.Equal(before, rec) {
		t.Fatal("顯示讀取改寫了原始資料")
	}
	for _, c := range []struct {
		record []byte
		viewer int
	}{{nil, 0}, {rec[:0x2B], 1}, {rec, -1}, {rec, 2}} {
		if _, ok := RegionListFleetCounts(c.record, c.viewer); ok {
			t.Fatal("無效記錄或座位被接受")
		}
	}
}
