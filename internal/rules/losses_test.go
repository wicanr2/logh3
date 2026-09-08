package rules

import "testing"

// 原版 s3fight 那一場，同盟第 13 艦隊九個分隊的現有艘數與艦種，
// 以及八個回合累積下來掉的數（`docs/re/86`）。
func TestDistributeLossesMatchesShape(t *testing.T) {
	alive := []int{3000, 3000, 3000, 3000, 3000, 4000, 4000, 2000, 2000}
	kinds := []int{0, 0, 0, 0, 0, 2, 2, 3, 3}
	got := DistributeLosses(alive, kinds, 215, true)
	// 原版量到的（組 5 多掉 10，是八個回合累積的取整差）。
	want := []int{23, 23, 23, 23, 23, 30, 30, 15, 15}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("第 %d 支算出掉 %d，原版量到 %d（全部 %v）", i, got[i], want[i], got)
			break
		}
	}
	sum := 0
	for _, n := range got {
		sum += n
	}
	if sum > 215 {
		t.Errorf("分出去 %d 艘，超過總損失 215", sum)
	}
}

// 艦種 7 不參與扣損，也不算進分母。
func TestDistributeLossesSkipsKind7(t *testing.T) {
	got := DistributeLosses([]int{1000, 1000}, []int{0, 7}, 100, true)
	if got[1] != 0 {
		t.Errorf("艦種 7 掉了 %d 艘，應該是 0", got[1])
	}
	if got[0] != 100 {
		t.Errorf("剩下那一支掉 %d 艘，總損失全部該落在它身上", got[0])
	}
}

// 留一艘的上限。
func TestDistributeLossesSparesOne(t *testing.T) {
	got := DistributeLosses([]int{10}, []int{0}, 100, true)
	if got[0] != 9 {
		t.Errorf("掉了 %d 艘，該留一艘（掉 9）", got[0])
	}
	got = DistributeLosses([]int{10}, []int{0}, 100, false)
	if got[0] != 10 {
		t.Errorf("不留一艘時掉了 %d 艘，該全掉", got[0])
	}
}
