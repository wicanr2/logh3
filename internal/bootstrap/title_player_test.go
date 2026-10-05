package bootstrap

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/screen"
)

func titlePlayerResources(t *testing.T) (*game.Resources, *game.Game) {
	t.Helper()
	orig := filepath.Join("..", "..", "org_game", "extracted")
	for _, name := range []string{"GIN3.COM", "IBM16.350", "IBMFLS16.350", "BU16IBM.RGB", "TITSPT.RGB", "SPTIT.350", "SPTIT.RGB"} {
		if _, err := os.Stat(filepath.Join(orig, name)); os.IsNotExist(err) {
			t.Skipf("缺原版%s", name)
		} else if err != nil {
			t.Fatal(err)
		}
	}
	r, err := Load(orig, filepath.Join("..", "..", "docs", "internals", "data"))
	if err != nil {
		t.Fatal(err)
	}
	if err := LoadIntro(r, orig); err != nil {
		t.Fatal(err)
	}
	g, err := game.New(r)
	if err != nil {
		t.Fatal(err)
	}
	return r, g
}

// 正常bootstrap及公開輸入，沒有狀態注入；完整自然播放兩趟並核對原版頁。
func TestTitlePlayerPath(t *testing.T) {
	r, g := titlePlayerResources(t)
	digest := func(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
	out := os.Getenv("LOGH3_TITLE_RECEIPTS")
	if out != "" {
		if err := os.Mkdir(out, 0755); err != nil {
			t.Fatal(err)
		}
	}
	var events []map[string]any
	record := func(name string) {
		t.Helper()
		fb, pal := g.Frame(), g.Palette()
		var rgb []byte
		for i := 0; i < gfx.Colors; i++ {
			r, gg, b := pal.RGB8(i)
			rgb = append(rgb, r, gg, b)
		}
		events = append(events, map[string]any{"name": name, "state": g.State().String(), "frame": g.IntroFrame(),
			"pixels_sha256": digest(fb.Pix), "rgb16_sha256": digest(rgb)})
		if out != "" {
			if err := fb.WritePNG(filepath.Join(out, name+".png"), pal); err != nil {
				t.Fatal(err)
			}
			if err := fb.WriteIndex(filepath.Join(out, name+".bin")); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(out, name+"-rgb16.bin"), rgb, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	first := append([]byte(nil), g.Frame().Pix...)
	firstPal := g.Palette()
	if g.State() != game.StateIntro || g.IntroFrame() != 0 {
		t.Fatal("正常入口不是動畫第0格")
	}
	record("first")
	toLast := func() {
		t.Helper()
		for frame := 1; frame < r.Intro.Frames(); frame++ {
			g.Tick(screen.IntroFrameIntervalAt(frame))
		}
		if g.State() != game.StateIntro || g.IntroFrame() != 420 {
			t.Fatal("沒有到421格序列末端")
		}
	}
	toLast()
	record("last")
	last := append([]byte(nil), g.Frame().Pix...)
	page := last[screen.IntroScreenTop*gfx.Width : (screen.IntroScreenTop+screen.IntroPageRows)*gfx.Width]
	if digest(page) != "b9b8f179788ec26d9812d5acd63cdd89c294ab59ccb6b1e0a9f0d73340ba0db9" {
		t.Fatal("末格不符原版")
	}
	if r.IntroTailPal == nil || g.Palette() != *r.IntroTailPal {
		t.Fatal("末格沒有切尾段色盤")
	}
	g.Tick(game.IntroTitleDelay - 1)
	if g.State() != game.StateIntro || !bytes.Equal(g.Frame().Pix, last) {
		t.Fatal("末格等待提早結束")
	}
	g.Tick(1)
	if g.State() != game.StateTitle {
		t.Fatal("等待期滿没有到標題")
	}
	record("title")
	title := append([]byte(nil), g.Frame().Pix...)
	page = title[screen.IntroScreenTop*gfx.Width : (screen.IntroScreenTop+screen.IntroPageRows)*gfx.Width]
	if digest(page) != "addde5e0929eb5f99a5430a714b19041a2cbe5530a4720a3641abebbfa3bf8ae" {
		t.Fatal("標題頁不符原版")
	}
	for _, b := range append(append([]byte(nil), title[:screen.IntroScreenTop*gfx.Width]...), title[(screen.IntroScreenTop+screen.IntroPageRows)*gfx.Width:]...) {
		if b != 0 {
			t.Fatal("標題上下補邊非0")
		}
	}
	if g.Palette() != r.TitlePal {
		t.Fatal("標題沒有自己的色盤")
	}
	for _, c := range []byte{'A', '\r', 8} {
		g.Key(c)
	}
	g.RightClick(300, 210)
	g.MoveMouse(600, 340)
	if g.State() != game.StateTitle || !bytes.Equal(title, g.Frame().Pix) || g.Palette() != r.TitlePal {
		t.Fatal("右鍵／鍵盤／移動改了標題")
	}
	record("ignored")
	g.Tick(game.TitleWaitDuration - 1)
	if g.State() != game.StateTitle {
		t.Fatal("標題提早逾時")
	}
	g.Tick(1)
	if g.State() != game.StateIntro || g.IntroFrame() != 0 || !bytes.Equal(g.Frame().Pix, first) || g.Palette() != firstPal {
		t.Fatal("逾時沒有清累加重播")
	}
	record("replay")
	toLast()
	g.Tick(game.IntroTitleDelay)
	if g.State() != game.StateTitle || !bytes.Equal(title, g.Frame().Pix) {
		t.Fatal("第二趟標題不同")
	}
	record("title-second")
	g.Click(300, 210)
	if g.State() != game.StateMainMenu {
		t.Fatal("左鍵没有返回主選單或穿透選陣營")
	}
	g.LeftRelease()
	g.MoveMouse(600, 340)
	record("main")
	g.Click(300, 210)
	if g.State() != game.StateCopyProtection {
		t.Fatal("下一次左鍵不能選陣營")
	}
	for _, c := range []byte{'1', '3', '8', '\r'} {
		g.Key(c)
	}
	if g.State() != game.StateMainGame {
		t.Fatal("標題離場後正常開局失敗")
	}
	record("new-game")
	if out != "" {
		// 與正式cmd/gin3的60 Hz逐次Update相同；無輸入、無狀態注入。
		_, timed := titlePlayerResources(t)
		for tick := 1; tick <= 5480; tick++ {
			timed.Tick(1000000000 / 60)
			if tick == 5000 || tick == 5480 {
				name := "title-60hz"
				want := game.StateTitle
				if tick == 5480 {
					name, want = "replay-60hz", game.StateIntro
				}
				if timed.State() != want {
					t.Fatalf("第%d更新狀態%v", tick, timed.State())
				}
				if err := timed.Frame().WritePNG(filepath.Join(out, name+".png"), timed.Palette()); err != nil {
					t.Fatal(err)
				}
				if err := timed.Frame().WriteIndex(filepath.Join(out, name+".bin")); err != nil {
					t.Fatal(err)
				}
			}
		}
		b, err := json.MarshalIndent(map[string]any{"schema": "logh3.title-player.v1", "entry": "bootstrap.Load + LoadIntro + game.New", "injection": false,
			"intro_title_delay_ns": int64(game.IntroTitleDelay), "title_wait_ns": int64(game.TitleWaitDuration), "events": events}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, "receipt.json"), append(b, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestTitleLargeTickAndSkip(t *testing.T) {
	for _, input := range []string{"left", "right", "key", "time"} {
		t.Run(input, func(t *testing.T) {
			r, g := titlePlayerResources(t)
			if input != "time" {
				switch input {
				case "left":
					g.Click(300, 210)
				case "right":
					g.RightClick(300, 210)
				case "key":
					g.Key('\r')
				}
				if g.State() != game.StateMainMenu {
					t.Fatal("既有片頭跳過改變")
				}
				return
			}
			var total = game.IntroTitleDelay + game.TitleWaitDuration
			for frame := 1; frame < r.Intro.Frames(); frame++ {
				total += screen.IntroFrameIntervalAt(frame)
			}
			g.Tick(total + screen.IntroFrameIntervalAt(1))
			if g.State() != game.StateIntro || g.IntroFrame() != 1 {
				t.Fatalf("大步進丟時間：%v 格%d", g.State(), g.IntroFrame())
			}
		})
	}
}

func TestTitleMissingAndInvalid(t *testing.T) {
	orig := filepath.Join("..", "..", "org_game", "extracted")
	dir := t.TempDir()
	for _, f := range []string{"IBM16.350", "IBMFLS16.350", "BU16IBM.RGB"} {
		b, err := os.ReadFile(filepath.Join(orig, f))
		if os.IsNotExist(err) {
			t.Skip("缺原版素材")
		}
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, f), b, 0644); err != nil {
			t.Fatal(err)
		}
	}
	r := &game.Resources{}
	if err := LoadIntro(r, dir); err != nil || r.Intro == nil || r.Title != nil {
		t.Fatalf("缺標題fallback：%v", err)
	}
	for _, f := range []string{"SPTIT.350", "SPTIT.RGB"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte{0}, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := LoadIntro(&game.Resources{}, dir); err == nil {
		t.Fatal("完整但損壞標題被略過")
	}
}
