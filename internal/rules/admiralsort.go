package rules

import "sort"

// 提督的能力位移（`docs/internals/08`）。名稱來自提督詳細資料那張畫面。
const (
	admiralCommand = 0x0B // +11 統率力
	admiralOperate = 0x17 // +23 運用
)

// AdmiralAbility 回一筆提督記錄的某一項能力值（位移就是 `docs/internals/08`
// 表上的那個數）。超出記錄長度回 0。
func AdmiralAbility(record []byte, offset int) int {
	if offset < 0 || len(record) <= offset {
		return 0
	}
	return int(record[offset])
}

// CompareAdmiralsForFleet 是原版 `2E54:063C` 的比較函式（`docs/re/201`）：
// AI 在首都新編艦隊之前拿它把提督排序。
//
//	+0x06 等級 → +0x0B 統率力 → +0x17 運用
//
// 三段**都是大的排前面**；`a` 排在 `b` 前面回 -1，反之回 1，全部相同回 0。
func CompareAdmiralsForFleet(a, b []byte) int {
	for _, off := range []int{admiralMerit, admiralCommand, admiralOperate} {
		x, y := AdmiralAbility(a, off), AdmiralAbility(b, off)
		switch {
		case x > y:
			return -1
		case x < y:
			return 1
		}
	}
	return 0
}

// SortAdmiralsForFleet 照 `2E54:063C` 的順序把提督編號排好。
//
// `record` 回一個編號的提督記錄（原版是 `ds:B77C + 編號 × 28`）。
// **穩定排序**：三段都相同時保持原本的順序，與原版那支比較器回 0 時
// 由呼叫端的排序決定次序一致。
func SortAdmiralsForFleet(indices []int, record func(i int) []byte) {
	sort.SliceStable(indices, func(x, y int) bool {
		return CompareAdmiralsForFleet(record(indices[x]), record(indices[y])) < 0
	})
}

// DefaultAdjutants 是原版 `ds:5F78` 的預設副官表：新編艦隊時提督的
// `+0x07` 是 `0xFF` 就從這裡填（`2D76:0611`，`docs/re/219`）。
// 82 格裡只有六格不是 `0xFF`。
type DefaultAdjutants map[int]int

// LoadDefaultAdjutants 讀 `docs/internals/data/default-adjutants.json`。
func LoadDefaultAdjutants(path string) (DefaultAdjutants, error) {
	var raw struct {
		Pairs []struct {
			Admiral  int `json:"admiral"`
			Adjutant int `json:"adjutant"`
		} `json:"pairs"`
	}
	if err := readJSON(path, &raw); err != nil {
		return nil, err
	}
	out := make(DefaultAdjutants, len(raw.Pairs))
	for _, p := range raw.Pairs {
		out[p.Admiral] = p.Adjutant
	}
	return out, nil
}

// For 回某位提督的預設副官；沒有回 -1。
func (d DefaultAdjutants) For(admiral int) int {
	if v, ok := d[admiral]; ok {
		return v
	}
	return -1
}
