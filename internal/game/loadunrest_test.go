package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// TestLoadSlot8UnrestMatchesReceipt 驗 `loadSlotFromDisk` 的載入當下重：
// 第 8 格的 Blend 與內政專任參謀不同於舊固定值，32 格仍要與原版相同。
func TestLoadSlot8UnrestMatchesReceipt(t *testing.T) {
	res := load(t)
	graph, err := rules.LoadStarGraph(
		filepath.Join("..", "..", "docs", "internals", "data", "star-graph.json"))
	if err != nil {
		t.Skipf("缺星域圖：%v", err)
	}
	res.Graph = graph
	res.UnrestParams = rules.UnrestParams{
		// 這是 bootstrap 的新遊戲／舊 S2、S3 預設；第 8 格必須由 WORK
		// 換成統治者 32／60 與 Blend 2／2。
		Ruler: [2]int{16, 60}, Blend: [2]int{1, 2}, Rate: [2]int{2, 2},
		Table: []int{170, 150, 130, 90, 60},
		Base:  []int{50, 75, 100, 150, 200},
	}
	g, err := New(res)
	if err != nil {
		t.Fatal(err)
	}
	g.loadSlotFromDisk(8)
	if g.slot == nil {
		t.Skip("缺第 8 格原版存檔，跳過（不用代用品）")
	}
	want, err := os.ReadFile(filepath.Join(root, "workplace", "oracle", "load8unrest", "regions.bin"))
	if err != nil {
		t.Skipf("缺 load8unrest 收據，跳過（不用代用品）")
	}
	if len(want) != rules.RegionCount*rules.RegionRecord {
		t.Fatalf("收據長度 %d，不是 %d", len(want), rules.RegionCount*rules.RegionRecord)
	}
	for r := 0; r < rules.RegionCount; r++ {
		got := g.regionRecord(r)[rules.RegionUnrest]
		exp := want[r*rules.RegionRecord+rules.RegionUnrest]
		if got != exp {
			t.Errorf("星域 %d 未安定度 %d，原版載入第 8 格是 %d", r, got, exp)
		}
	}
	if g.blend != [2]int{2, 2} || g.rate != [2]int{2, 2} {
		t.Errorf("載入後 Blend/Rate = %v/%v，要 [2 2]/[2 2]", g.blend, g.rate)
	}
}
