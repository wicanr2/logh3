package asset

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIntroDecoder 解完 `IBM16.350` 全部 374 格。
//
// 三個判準，互相獨立（`docs/re/72`）：
//
//   - 每一格吃掉的位元組數與 `IBMFLS16.350` 的長度表逐一相符
//     （`Next` 自己會檢查），整個檔案吃光；
//   - 累加的 XOR 在**第 97–102 格與第 244–251 格整個歸零**——四個 plane
//     共 72,640 個位元組全是 0。那是兩個場景之間的黑幕，走法只要錯一格
//     就不可能歸零；
//   - 中間那幾格不是空的（不然「歸零」沒有意義）。
func TestIntroDecoder(t *testing.T) {
	orig := os.Getenv("LOGH3_ORIG")
	if orig == "" {
		orig = filepath.Join("..", "..", "org_game", "extracted")
	}
	data, err := os.ReadFile(filepath.Join(orig, "IBM16.350"))
	if err != nil {
		t.Skipf("沒有原版素材，跳過：%v", err)
	}
	lengths, err := os.ReadFile(filepath.Join(orig, "IBMFLS16.350"))
	if err != nil {
		t.Skipf("沒有原版素材，跳過：%v", err)
	}
	d, err := NewIntroDecoder(data, lengths)
	if err != nil {
		t.Fatal(err)
	}
	if d.Frames() != 374 {
		t.Fatalf("格數 %d，應該是 374", d.Frames())
	}
	blank := map[int]bool{}
	for f := 0; ; f++ {
		ok, err := d.Next()
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			if f != 374 {
				t.Fatalf("只解了 %d 格", f)
			}
			break
		}
		if allZero(d.Planes) {
			blank[f] = true
		}
	}
	for _, f := range []int{97, 98, 99, 100, 101, 102, 244, 245, 246, 247, 248, 249, 250, 251} {
		if !blank[f] {
			t.Errorf("第 %d 格的累加應該整個歸零", f)
		}
	}
	if len(blank) != 14 {
		t.Errorf("歸零的格有 %d 個，應該正好是那 14 個：%v", len(blank), blank)
	}
}

func allZero(planes [IntroPlanes][]uint8) bool {
	for _, p := range planes {
		for _, v := range p {
			if v != 0 {
				return false
			}
		}
	}
	return true
}

// TestIntroFirstFrameMatchesOracle 對原版量到的第一格。
//
// 收據 `introseq`（`-watch 27616-2BD05`，整個 plane 0 的緩衝區）錄到的
// 前 12 筆寫入就是第 0 格 plane 0 的全部內容（`docs/re/108`）：
// 從 di ＝ 9,406 起每次 +80，值依序是 C0 80 20 30 70 10 30 08 14 0C 02 06。
// 原始位元組也對得起來——`00 00 29 3F` 跳 10,559 次正好落到 9,406。
func TestIntroFirstFrameMatchesOracle(t *testing.T) {
	orig := os.Getenv("LOGH3_ORIG")
	if orig == "" {
		orig = filepath.Join("..", "..", "org_game", "extracted")
	}
	data, err := os.ReadFile(filepath.Join(orig, "IBM16.350"))
	if err != nil {
		t.Skipf("沒有原版素材，跳過：%v", err)
	}
	lengths, err := os.ReadFile(filepath.Join(orig, "IBMFLS16.350"))
	if err != nil {
		t.Skipf("沒有原版素材，跳過：%v", err)
	}
	d, err := NewIntroDecoder(data, lengths)
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := d.Next(); err != nil || !ok {
		t.Fatalf("第 0 格解不出來：%v", err)
	}
	want := map[int]uint8{}
	for i, v := range []uint8{0xC0, 0x80, 0x20, 0x30, 0x70, 0x10, 0x30, 0x08, 0x14, 0x0C, 0x02, 0x06} {
		want[9406+i*IntroPlaneStride] = v
	}
	for i, got := range d.Planes[0] {
		if w, ok := want[i]; ok {
			if got != w {
				t.Errorf("plane 0 的 di=%d 是 %02X，收據是 %02X", i, got, w)
			}
			continue
		}
		if got != 0 {
			t.Errorf("plane 0 的 di=%d 應該是 0，卻是 %02X", i, got)
		}
	}
}
