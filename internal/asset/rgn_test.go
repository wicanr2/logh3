package asset

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRegionsRebuildLiveMap 是地圖層的完整對拍：底圖 ＋ 32 個星域遮罩
// ＋ 每區的顏色，重建出來要與**記憶體裡那份**（遊戲自己塗好的）逐像素相同。
func TestRegionsRebuildLiveMap(t *testing.T) {
	dir := origDir(t)
	bit, err := os.ReadFile(filepath.Join(dir, "STRMAP.BIT"))
	if err != nil {
		t.Skip("找不到 STRMAP.BIT，跳過")
	}
	rgnData, err := os.ReadFile(filepath.Join(dir, "STRMAP.RGN"))
	if err != nil {
		t.Skip("找不到 STRMAP.RGN，跳過")
	}
	liveData, err := os.ReadFile("../../workplace/strmap-live.bin")
	if err != nil {
		t.Skip("沒有 strmap-live.bin，跳過（probe -dump-mem 可重建）")
	}
	const tw, th, cols = 16, 20, 20
	base, err := DecodeSheet(bit, tw, th, cols)
	if err != nil {
		t.Fatal(err)
	}
	live, err := DecodeSheet(liveData, tw, th, cols)
	if err != nil {
		t.Fatal(err)
	}
	regions, err := OpenRGN(rgnData, tw, th)
	if err != nil {
		t.Fatal(err)
	}
	if len(regions) != 32 {
		t.Fatalf("解出 %d 個區域，預期 32", len(regions))
	}
	out := &Image{W: base.W, H: base.H, Pix: append([]uint8(nil), base.Pix...)}
	for i, r := range regions {
		// 顏色是遊戲狀態（某星域屬於誰），這裡從已經塗好的那份反推，
		// 順便驗證「一個區域只有一種顏色」——不成立就是遮罩解錯了。
		colour, ok := regionColour(r, base, live, tw, th)
		if !ok {
			t.Fatalf("第 %d 個區域在原版畫面上不是單一顏色", i)
		}
		r.Paint(out, base, colour, tw, th)
	}
	bad := 0
	for i := range out.Pix {
		if out.Pix[i] != live.Pix[i] {
			bad++
		}
	}
	if bad != 0 {
		t.Fatalf("重建的地圖與原版有 %d / %d 個像素不同", bad, len(out.Pix))
	}
}

// regionColour 從已著色的那份反推區域顏色，並要求它是單一顏色。
func regionColour(r Region, base, live *Image, tw, th int) (uint8, bool) {
	var colour uint8
	var seen bool
	for y := 0; y < r.BH*th; y++ {
		gy := r.BY*th + y
		for x := 0; x < r.BW*tw; x++ {
			gx := r.BX*tw + x
			if gx >= base.W || gy >= base.H || !r.Covers(x, y, tw, th) {
				continue
			}
			if base.Pix[gy*base.W+gx] != 0 {
				continue
			}
			c := live.Pix[gy*live.W+gx]
			if !seen {
				colour, seen = c, true
			} else if c != colour {
				return 0, false
			}
		}
	}
	return colour, seen
}
