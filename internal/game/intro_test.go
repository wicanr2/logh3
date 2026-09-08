package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/screen"
)

// loadIntro 把開頭動畫接上去；缺素材就 skip。
func loadIntro(t *testing.T, res *Resources) {
	t.Helper()
	orig := filepath.Join(root, "org_game/extracted")
	data, err := os.ReadFile(filepath.Join(orig, "IBM16.350"))
	if err != nil {
		t.Skipf("缺原版素材，跳過：%v", err)
	}
	lengths, err := os.ReadFile(filepath.Join(orig, "IBMFLS16.350"))
	if err != nil {
		t.Skipf("缺原版素材，跳過：%v", err)
	}
	rgb, err := os.ReadFile(filepath.Join(orig, "BU16IBM.RGB"))
	if err != nil {
		t.Skipf("缺原版素材，跳過：%v", err)
	}
	dec, err := asset.NewIntroDecoder(data, lengths)
	if err != nil {
		t.Fatal(err)
	}
	pal, err := screen.IntroPalette(rgb)
	if err != nil {
		t.Fatal(err)
	}
	res.Intro, res.IntroPal = dec, pal
}

// TestNewWithoutIntroStartsAtMainMenu 釘住「沒有動畫素材就直接進主選單」。
//
// 對拍收據與 README 的截圖腳本都是照這條路徑寫的，接上動畫會整條位移。
func TestNewWithoutIntroStartsAtMainMenu(t *testing.T) {
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	if g.State() != StateMainMenu {
		t.Fatalf("開機在 %v，預期主選單", g.State())
	}
}

// TestIntroAdvancesOneFramePerInterval 釘住播放速度與收尾。
//
// 一格 125.3 ms 是量出來的（73 個計時器中斷 × PIT 除數 0x0800，
// `docs/re/110`）。走到 IntroFrameInterval 才換下一格，不到不換。
func TestIntroAdvancesOneFramePerInterval(t *testing.T) {
	res := load(t)
	loadIntro(t, res)
	g, err := New(res)
	if err != nil {
		t.Fatal(err)
	}
	if g.State() != StateIntro {
		t.Fatalf("開機在 %v，預期開頭動畫", g.State())
	}
	if g.IntroFrame() != 0 {
		t.Fatalf("開機在第 %d 格，預期第 0 格", g.IntroFrame())
	}
	// 差一奈秒還不換。
	g.Tick(screen.IntroFrameInterval - 1)
	if g.IntroFrame() != 0 {
		t.Fatalf("還沒到一格就換了（第 %d 格）", g.IntroFrame())
	}
	g.Tick(1)
	if g.IntroFrame() != 1 {
		t.Fatalf("一格之後在第 %d 格，預期第 1 格", g.IntroFrame())
	}
	// 一次跳十格也要跟上（掉格時不會愈播愈慢）。
	g.Tick(10 * screen.IntroFrameInterval)
	if g.IntroFrame() != 11 {
		t.Fatalf("跳十格之後在第 %d 格，預期第 11 格", g.IntroFrame())
	}
	// 畫面是 640×480，動畫畫在垂直置中的位置。
	fb := g.Frame()
	if len(fb.Pix) == 0 {
		t.Fatal("畫不出畫面")
	}
	// 調色盤要換成動畫那一組。
	if g.Palette() != res.IntroPal {
		t.Fatal("開頭動畫沒有用自己的調色盤")
	}
}

// TestIntroSkipsToMainMenu 釘住「按什麼都跳過」。
func TestIntroSkipsToMainMenu(t *testing.T) {
	for _, tc := range []struct {
		name string
		do   func(*Game)
	}{
		{"左鍵", func(g *Game) { g.Click(300, 240) }},
		{"右鍵", func(g *Game) { g.RightClick(300, 240) }},
		{"按鍵", func(g *Game) { g.Key('\r') }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := load(t)
			loadIntro(t, res)
			g, err := New(res)
			if err != nil {
				t.Fatal(err)
			}
			tc.do(g)
			if g.State() != StateMainMenu {
				t.Fatalf("跳過之後在 %v，預期主選單", g.State())
			}
			if g.Palette() != res.Menu.Palette() {
				t.Fatal("跳過之後調色盤沒有換回主選單那一組")
			}
		})
	}
}

// TestIntroEndsAtMainMenu 釘住「374 格播完自己進主選單」。
func TestIntroEndsAtMainMenu(t *testing.T) {
	res := load(t)
	loadIntro(t, res)
	frames := res.Intro.Frames()
	g, err := New(res)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < frames+1 && g.State() == StateIntro; i++ {
		g.Tick(screen.IntroFrameInterval)
	}
	if g.State() != StateMainMenu {
		t.Fatalf("播完之後在 %v，預期主選單", g.State())
	}
	if g.IntroFrame() != frames-1 {
		t.Fatalf("最後停在第 %d 格，預期第 %d 格", g.IntroFrame(), frames-1)
	}
}
