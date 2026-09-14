package screen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

func loadFlipPanel(t *testing.T) *FlipReportPanel {
	t.Helper()
	p, err := LoadFlipReportPanel(filepath.Join("..", "..",
		"docs", "internals", "data", "flip-report-panel.json"))
	if err != nil {
		t.Fatalf("讀版面：%v", err)
	}
	return p
}

// 版面的數字就是收據上量到的（`docs/spec/034`、`docs/re/246`、`248`）。
func TestFlipReportPanelLayout(t *testing.T) {
	p := loadFlipPanel(t)
	if p.Name.X != 336 || p.Name.Y != 182 {
		t.Errorf("星系名 ＝ (%d,%d)，預期 (336,182)", p.Name.X, p.Name.Y)
	}
	// 字色跟著新的擁有者走，三格都量到（chain7／chain3／chain6）。
	if want := []uint8{6, 5, 7}; len(p.Name.InkByOwner) != 3 ||
		p.Name.InkByOwner[0] != want[0] || p.Name.InkByOwner[1] != want[1] ||
		p.Name.InkByOwner[2] != want[2] {
		t.Errorf("星系名的字色 ＝ %v，預期 %v", p.Name.InkByOwner, want)
	}
	if p.Message.X != 336 || p.Message.Y != 202 || p.Message.Ink != 7 {
		t.Errorf("訊息 ＝ (%d,%d) 色 %d，預期 (336,202) 色 7",
			p.Message.X, p.Message.Y, p.Message.Ink)
	}
	if p.LineStep != 20 {
		t.Errorf("列距 ＝ %d，預期 20", p.LineStep)
	}
	if p.Face.X != 544 || p.Face.Y != 180 || p.Face.W != 72 || p.Face.H != 93 {
		t.Errorf("頭像 ＝ (%d,%d) %dx%d，預期 (544,180) 72x93",
			p.Face.X, p.Face.Y, p.Face.W, p.Face.H)
	}
	// 「星系」接在名字後面，與名字不同色（原版 `ds:085D` 兩個顏色控制碼）。
	if len(p.Suffix.Codes) != 2 || p.Suffix.Ink != 7 {
		t.Errorf("`星系` ＝ %v 色 %d，預期兩個字、色 7", p.Suffix.Codes, p.Suffix.Ink)
	}
	for _, k := range []string{"taken_by_us", "lost_to_enemy",
		"our_grip_lost", "we_partly_landed", "rebellion"} {
		m, ok := p.Messages[k]
		if !ok || len(m.Lines) == 0 || len(m.Lines[0]) == 0 {
			t.Errorf("缺 %s 那一句", k)
		}
	}
	// 「已被敵軍占領。」七個字（收據 chain3 畫的就是這句）。
	if got := p.Messages["lost_to_enemy"].Lines; len(got) != 1 || len(got[0]) != 7 {
		t.Errorf("`已被敵軍占領。` ＝ %v，預期一行七個字", got)
	}
}

// 收據 `rbx292` 畫的是「亞爾迪巴朗星系／發生叛亂。」；它與易主
// 報告共用 `3826:0307` 的版面，但訊息來源是 ds:086F（`docs/spec/046`）。
func TestRebellionReportTextMatchesOracle(t *testing.T) {
	root := filepath.Join("..", "..")
	scr, err := os.ReadFile(filepath.Join(root, "workplace", "oracle", "rbx292", "screen.bin"))
	if err != nil || len(scr) != 640*480 {
		t.Skip("缺收據 rbx292，跳過（不用代用品）")
	}
	fnt, err := font.Load(filepath.Join(root, "org_game/extracted/GRAPH.IMG"))
	if err != nil {
		t.Skip("缺 GRAPH.IMG，跳過")
	}
	p := loadFlipPanel(t)
	// 亞爾迪巴朗；Big5 字碼由原版畫面文字直接轉得。
	name := []uint16{43208, 47800, 44413, 42202, 44756}
	fb := gfx.NewFramebuffer(9)
	p.Draw(fb, fnt, name, 0, "rebellion")

	bad, total := 0, 0
	for y := 178; y < 220; y++ {
		for x := 334; x < 510; x++ {
			total++
			if fb.At(x, y) != scr[y*640+x] {
				bad++
			}
		}
	}
	if bad != 0 {
		t.Errorf("叛亂報告文字與收據差 %d／%d 個像素，預期 0", bad, total)
	}
}

// 頭像逐像素對收據：`FACE.AMG` 第 13 張（＝ `ds:8F4D[帝國]`）畫在
// `(544,180)`，取左上 72×93（`docs/re/248`）。
func TestFlipReportFaceMatchesOracle(t *testing.T) {
	root := filepath.Join("..", "..")
	scr, err := os.ReadFile(filepath.Join(root, "workplace", "oracle", "chain3", "screen.bin"))
	if err != nil || len(scr) != 640*480 {
		t.Skip("缺收據 chain3，跳過（不用代用品）")
	}
	raw, err := os.ReadFile(filepath.Join(root, "org_game", "extracted", "FACE.AMG"))
	if err != nil {
		t.Skip("缺 FACE.AMG，跳過")
	}
	a, err := asset.OpenAMG(raw)
	if err != nil {
		t.Fatalf("讀 FACE.AMG：%v", err)
	}
	img, err := a.Decode(13)
	if err != nil {
		t.Fatalf("解第 13 張：%v", err)
	}
	p := loadFlipPanel(t)
	bad := 0
	for y := 0; y < p.Face.H; y++ {
		for x := 0; x < p.Face.W; x++ {
			if img.Pix[y*img.W+x] != scr[(p.Face.Y+y)*640+p.Face.X+x] {
				bad++
			}
		}
	}
	if bad != 0 {
		t.Errorf("頭像與收據差 %d 個像素，預期 0", bad)
	}
}

// 文字逐像素對兩份收據：
//
//	chain3  菲列亞星系／已被敵軍占領。              （被敵方打下來）
//	chain6  奇霍伊薩星系／由於受到敵方艦隊入侵…    （自己的變成無主）
//
// 後者是把星域 5 兩顆行星裡的一顆 poke 成同盟，讓那一格變混雜——原版
// 依行星重算就成了無主（`docs/re/17`）。這一份把 `our_grip_lost` 從
// 假設變成量到的（`docs/spec/034`）。
func TestFlipReportNeutralMessageMatchesOracle(t *testing.T) {
	root := filepath.Join("..", "..")
	scr, err := os.ReadFile(filepath.Join(root, "workplace", "oracle", "chain6", "screen.bin"))
	if err != nil || len(scr) != 640*480 {
		t.Skip("缺收據 chain6，跳過（不用代用品）")
	}
	fnt, err := font.Load(filepath.Join(root, "org_game/extracted/GRAPH.IMG"))
	if err != nil {
		t.Skip("缺 GRAPH.IMG，跳過")
	}
	p := loadFlipPanel(t)
	// 星域 5「奇霍伊薩」，字碼取自那一份傾印的星域記錄開頭。
	name := []uint16{43359, 49230, 42476, 49860}

	fb := gfx.NewFramebuffer(9)
	p.Draw(fb, fnt, name, 2, "our_grip_lost") // 變成無主

	bad, total := 0, 0
	x0, x1, y0, y1 := 999, -1, 999, -1
	for y := 178; y < 240; y++ {
		for x := 334; x < 510; x++ {
			total++
			if fb.At(x, y) != scr[y*640+x] {
				bad++
				if x < x0 {
					x0 = x
				}
				if x > x1 {
					x1 = x
				}
				if y < y0 {
					y0 = y
				}
				if y > y1 {
					y1 = y
				}
			}
		}
	}
	if bad != 0 {
		t.Errorf("三行文字與收據差 %d／%d 個像素，預期 0（差異範圍 x %d..%d y %d..%d）",
			bad, total, x0, x1, y0, y1)
	}
}

// 打下敵方星域：收據 `chain7` 把星域 29 兩顆行星都 poke 成帝國，
// 畫的是「流卡斯星系／已被我方占領。」，星系名是色 6
// （新擁有者 ＝ 帝國）。這一份把 `taken_by_us` 與色 6 一起釘住。
func TestFlipReportTakenMessageMatchesOracle(t *testing.T) {
	root := filepath.Join("..", "..")
	scr, err := os.ReadFile(filepath.Join(root, "workplace", "oracle", "chain7", "screen.bin"))
	if err != nil || len(scr) != 640*480 {
		t.Skip("缺收據 chain7，跳過（不用代用品）")
	}
	fnt, err := font.Load(filepath.Join(root, "org_game/extracted/GRAPH.IMG"))
	if err != nil {
		t.Skip("缺 GRAPH.IMG，跳過")
	}
	p := loadFlipPanel(t)
	// 星域 29「流卡斯」。
	name := []uint16{44153, 42340, 46261}

	fb := gfx.NewFramebuffer(9)
	p.Draw(fb, fnt, name, 0, "taken_by_us") // 落到帝國手上

	bad, total := 0, 0
	for y := 178; y < 220; y++ {
		for x := 334; x < 470; x++ {
			total++
			if fb.At(x, y) != scr[y*640+x] {
				bad++
			}
		}
	}
	if bad != 0 {
		t.Errorf("文字與收據差 %d／%d 個像素，預期 0", bad, total)
	}
}

// 敵方星域變成無主：收據 `chain8` 把星域 15 兩顆行星裡的一顆 poke 成帝國，
// 那一格變混雜就成了無主，畫的是「艾爾•法西爾星系／部分行星我軍已成功
// 進駐。」。這一份把最後一句 `we_partly_landed` 也釘住（`docs/re/249`）。
func TestFlipReportPartlyLandedMatchesOracle(t *testing.T) {
	root := filepath.Join("..", "..")
	scr, err := os.ReadFile(filepath.Join(root, "workplace", "oracle", "chain8", "screen.bin"))
	if err != nil || len(scr) != 640*480 {
		t.Skip("缺收據 chain8，跳過（不用代用品）")
	}
	fnt, err := font.Load(filepath.Join(root, "org_game/extracted/GRAPH.IMG"))
	if err != nil {
		t.Skip("缺 GRAPH.IMG，跳過")
	}
	p := loadFlipPanel(t)
	// 星域 15「艾爾•法西爾」。
	name := []uint16{42723, 47800, 41285, 43627, 42728, 47800}

	fb := gfx.NewFramebuffer(9)
	p.Draw(fb, fnt, name, 2, "we_partly_landed") // 敵方的變成無主

	bad, total := 0, 0
	for y := 178; y < 220; y++ {
		for x := 334; x < 525; x++ {
			total++
			if fb.At(x, y) != scr[y*640+x] {
				bad++
			}
		}
	}
	if bad != 0 {
		t.Errorf("文字與收據差 %d／%d 個像素，預期 0", bad, total)
	}
}

// 文字逐像素對收據：收據 `chain3` 畫的是「菲列亞星系／已被敵軍占領。」。
//
// 只比報告那兩行佔的矩形——底板與視窗框由命令階段那張畫面負責
// （`docs/spec/005`，已經逐位元組對過）。
func TestFlipReportTextMatchesOracle(t *testing.T) {
	root := filepath.Join("..", "..")
	scr, err := os.ReadFile(filepath.Join(root, "workplace", "oracle", "chain3", "screen.bin"))
	if err != nil || len(scr) != 640*480 {
		t.Skip("缺收據 chain3，跳過（不用代用品）")
	}
	fnt, err := font.Load(filepath.Join(root, "org_game/extracted/GRAPH.IMG"))
	if err != nil {
		t.Skip("缺 GRAPH.IMG，跳過")
	}
	p := loadFlipPanel(t)

	// 星域 4「菲列亞」，字碼取自收據那一份傾印的星域記錄開頭。
	name := []uint16{46561, 42563, 43208}

	// 報告畫在視窗內容上，底色 9（收據量到的）。
	fb := gfx.NewFramebuffer(9)
	p.Draw(fb, fnt, name, 1, "lost_to_enemy") // 落到同盟手上

	bad, total := 0, 0
	for y := 178; y < 220; y++ {
		for x := 334; x < 470; x++ {
			total++
			if fb.At(x, y) != scr[y*640+x] {
				bad++
			}
		}
	}
	if bad != 0 {
		t.Errorf("報告文字與收據差 %d／%d 個像素，預期 0", bad, total)
	}
}
