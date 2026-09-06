package asset

import (
	"os"
	"path/filepath"
	"testing"
)

// origDir 是玩家自備的原版目錄。**找不到就 skip**——安靜的替代品會讓
// 「還沒做完」看起來像做完了。
func origDir(t *testing.T) string {
	t.Helper()
	d := os.Getenv("LOGH3_ORIG")
	if d == "" {
		d = "../../org_game/extracted"
	}
	if _, err := os.Stat(filepath.Join(d, "FACE.AMG")); err != nil {
		t.Skip("找不到原版 FACE.AMG，跳過")
	}
	return d
}

// TestFaceAMGDecodesExactly 釘住兩件會安靜出錯的事：每一張都要**剛好**
// 用完自己那段位元組（少用或多用都表示解碼走偏），而且尺寸一致。
func TestFaceAMGDecodesExactly(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(origDir(t), "FACE.AMG"))
	if err != nil {
		t.Fatal(err)
	}
	a, err := OpenAMG(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Entries) != 83 {
		t.Fatalf("索引筆數 %d，預期 83", len(a.Entries))
	}
	for i := range a.Entries {
		img, err := a.Decode(i)
		if err != nil {
			t.Fatalf("第 %d 張：%v", i, err)
		}
		if img.Used != a.Entries[i].Size {
			t.Fatalf("第 %d 張用掉 %d bytes，該用 %d", i, img.Used, a.Entries[i].Size)
		}
		if img.W != 80 || img.H != 100 {
			t.Fatalf("第 %d 張尺寸 %d×%d，預期 80×100", i, img.W, img.H)
		}
		for _, v := range img.Pix {
			if v > 15 {
				t.Fatalf("第 %d 張出現非 4 位元色號 %d", i, v)
			}
		}
	}
}

// TestFaceAMGMatchesOracle 拿原版畫面當判準：防拷問答那張照片
// （`FACE.AMG` 第 0 張）在收據裡的位置必須**逐像素相同**。
// 沒有收據就 skip——收據是一次性證據，不進 Git。
func TestFaceAMGMatchesOracle(t *testing.T) {
	const receipt = "../../workplace/oracle/face-fresh/screen.bin"
	fb, err := os.ReadFile(receipt)
	if err != nil {
		t.Skip("沒有 face-fresh 收據，跳過（tools/oracle.sh 可重建）")
	}
	data, err := os.ReadFile(filepath.Join(origDir(t), "FACE.AMG"))
	if err != nil {
		t.Fatal(err)
	}
	a, err := OpenAMG(data)
	if err != nil {
		t.Fatal(err)
	}
	img, err := a.Decode(0)
	if err != nil {
		t.Fatal(err)
	}
	const sw, x0, y0 = 640, 144, 192
	bad := 0
	for y := 0; y < img.H; y++ {
		for x := 0; x < img.W; x++ {
			if fb[(y0+y)*sw+x0+x] != img.Pix[y*img.W+x] {
				bad++
			}
		}
	}
	if bad != 0 {
		t.Fatalf("與原版畫面有 %d / %d 個像素不同", bad, img.W*img.H)
	}
}

// TestSpriteSPRTiles 釘住 `SPRITE.SPR` 的版面：75 張 16×16。
// 尺寸猜錯不會報錯——只會讓每一張的內容從上一張的中間開始，
// 圖案看起來「大致對但錯位」。
func TestSpriteSPRTiles(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(origDir(t), "SPRITE.SPR"))
	if err != nil {
		t.Skip("找不到原版 SPRITE.SPR，跳過")
	}
	tiles, err := DecodeTiles(data, 16, 16)
	if err != nil {
		t.Fatal(err)
	}
	if len(tiles) != 75 {
		t.Fatalf("解出 %d 張，預期 75", len(tiles))
	}
	for i, tl := range tiles {
		for _, v := range tl.Pix {
			if v > 15 {
				t.Fatalf("第 %d 張出現非 4 位元色號 %d", i, v)
			}
		}
	}
	// **整除證明不了尺寸**：9600 同樣是 16×20（160 bytes）的整數倍，
	// 解出來是 60 張錯位的圖。16×16 的依據是原版 blit 常式與圖形本身
	// （見 `docs/internals/10`），不是這個整除關係。
	if _, err := DecodeTiles(data, 16, 20); err != nil {
		t.Fatal("16×20 也整除，這個測試不該靠整除判斷尺寸")
	}
}

// TestStrmapSheet 釘住星域地圖：300 塊 16×20 拼成 320×300。
//
// 另外釘一件事：把它貼到原版畫面的 (0,40)，**差異只能是「地圖是 0、
// 畫面不是 0」**。地圖的 0 是留給底下那層勢力著色的，
// 一旦出現「地圖非 0 卻和畫面不同」就表示圖塊排列或位元序解錯了。
func TestStrmapSheet(t *testing.T) {
	dir := origDir(t)
	data, err := os.ReadFile(filepath.Join(dir, "STRMAP.BIT"))
	if err != nil {
		t.Skip("找不到原版 STRMAP.BIT，跳過")
	}
	img, err := DecodeSheet(data, 16, 20, 20)
	if err != nil {
		t.Fatal(err)
	}
	if img.W != 320 || img.H != 300 {
		t.Fatalf("拼出 %d×%d，預期 320×300", img.W, img.H)
	}
	fb, err := os.ReadFile("../../workplace/oracle/key1/screen.bin")
	if err != nil {
		t.Skip("沒有 key1 收據，跳過對位（tools/oracle.sh 可重建）")
	}
	const sw, x0, y0, visible = 640, 0, 40, 140 // 右半被會計報告視窗蓋住
	for y := 0; y < img.H; y++ {
		for x := 0; x < visible; x++ {
			want, got := fb[(y0+y)*sw+x0+x], img.Pix[y*img.W+x]
			if want != got && got != 0 {
				t.Fatalf("(%d,%d) 地圖畫 %d、原版是 %d", x, y, got, want)
			}
		}
	}
}

// TestStrmapLiveMatchesOracle 是最強的那一條：把**記憶體裡**那份
// （已經塗上勢力色的）地圖解出來，貼到原版畫面要逐像素相同。
// 檔案原始內容做不到這件事——那些 0 是留白，遊戲載入後才就地塗掉。
func TestStrmapLiveMatchesOracle(t *testing.T) {
	live, err := os.ReadFile("../../workplace/strmap-live.bin")
	if err != nil {
		t.Skip("沒有 strmap-live.bin，跳過（probe -dump-mem 可重建）")
	}
	fb, err := os.ReadFile("../../workplace/oracle/key1/screen.bin")
	if err != nil {
		t.Skip("沒有 key1 收據，跳過")
	}
	img, err := DecodeSheet(live, 16, 20, 20)
	if err != nil {
		t.Fatal(err)
	}
	const sw, y0, visible = 640, 40, 140
	bad := 0
	for y := 0; y < img.H; y++ {
		for x := 0; x < visible; x++ {
			if fb[(y0+y)*sw+x] != img.Pix[y*img.W+x] {
				bad++
			}
		}
	}
	if bad != 0 {
		t.Fatalf("與原版畫面有 %d / %d 個像素不同", bad, visible*img.H)
	}
}
