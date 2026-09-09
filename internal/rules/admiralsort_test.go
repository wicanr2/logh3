package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func admiralRec(merit, command, operate int) []byte {
	r := make([]byte, AdmiralRecord)
	r[admiralMerit], r[admiralCommand], r[admiralOperate] =
		byte(merit), byte(command), byte(operate)
	return r
}

// TestCompareAdmiralsForFleet 測 `2E54:063C` 的三段降冪（`docs/re/201`）。
func TestCompareAdmiralsForFleet(t *testing.T) {
	for _, c := range []struct {
		name string
		a, b []byte
		want int
	}{
		{"等級高的在前", admiralRec(29, 80, 80), admiralRec(20, 140, 140), -1},
		{"等級低的在後", admiralRec(20, 140, 140), admiralRec(29, 80, 80), 1},
		{"同級比統率力", admiralRec(20, 120, 80), admiralRec(20, 110, 140), -1},
		{"再同比運用", admiralRec(20, 120, 130), admiralRec(20, 120, 100), -1},
		{"三段全同", admiralRec(20, 120, 130), admiralRec(20, 120, 130), 0},
	} {
		if got := CompareAdmiralsForFleet(c.a, c.b); got != c.want {
			t.Errorf("%s：得到 %d，預期 %d", c.name, got, c.want)
		}
	}
	// 記錄太短不該 panic。
	if got := CompareAdmiralsForFleet(nil, nil); got != 0 {
		t.Errorf("空記錄應該回 0，得到 %d", got)
	}
}

// TestSortAdmiralsForFleetOnOriginalData 拿原版的提督表排一次，確認
// 排完之後三段鍵單調不遞增。缺原版檔就 skip。
func TestSortAdmiralsForFleetOnOriginalData(t *testing.T) {
	path := filepath.Join("../..", "org_game/extracted/ADMIRAL.DTX")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("缺 %s，跳過（不用代用品）", path)
	}
	all := ParseAdmirals(data)
	var idx []int
	for _, a := range all {
		if a.Faction == 0 {
			idx = append(idx, a.Index)
		}
	}
	if len(idx) < 2 {
		t.Fatalf("帝國只有 %d 名提督", len(idx))
	}
	rec := func(i int) []byte { return all[i].Record }
	SortAdmiralsForFleet(idx, rec)
	for k := 1; k < len(idx); k++ {
		if CompareAdmiralsForFleet(rec(idx[k-1]), rec(idx[k])) > 0 {
			t.Fatalf("第 %d 與第 %d 名的順序反了", k-1, k)
		}
	}
	// 第一名的等級要是全體最高。
	best := 0
	for _, i := range idx {
		if v := AdmiralLevel(rec(i)); v > best {
			best = v
		}
	}
	if got := AdmiralLevel(rec(idx[0])); got != best {
		t.Errorf("排第一的等級 ＝ %d，全體最高是 %d", got, best)
	}
}
