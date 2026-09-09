package rules

import (
	"path/filepath"
	"testing"
)

// TestBlockMaskIsAttrDat 釘住「圖塊遮罩就是 `ATTR.DAT`」（`docs/re/262`）：
// 那 128 個位元組逐一相同，執行期躺在 `ds:9BB4`。
//
// 本來這張表是從執行期記憶體倒的，而且只倒了 64 筆——`Blocked` 對
// 超出範圍的圖塊回「擋住」，所以圖塊 64–127（實際上全是 `0x0F` ＝ 可通行）
// 會被當成不能走。已量到的戰場地圖最大只用到圖塊 55，所以行為沒有差異，
// 但來源換成原版檔案之後就不再依賴那份一次性傾印。
func TestBlockMaskIsAttrDat(t *testing.T) {
	terrain, err := LoadBattleTerrain(filepath.Join("..", "..",
		"docs", "internals", "data", "battle-terrain.json"))
	if err != nil {
		t.Fatal(err)
	}
	attr := readOrig(t, "ATTR.DAT")
	if len(attr) != 128 {
		t.Fatalf("ATTR.DAT 有 %d bytes，要 128", len(attr))
	}
	if len(terrain.BlockMask) != len(attr) {
		t.Fatalf("block_mask 有 %d 筆，ATTR.DAT 有 %d 個位元組",
			len(terrain.BlockMask), len(attr))
	}
	for i := range attr {
		if terrain.BlockMask[i] != attr[i] {
			t.Fatalf("第 %d 個位元組：block_mask %02X，ATTR.DAT %02X",
				i, terrain.BlockMask[i], attr[i])
		}
	}
}
