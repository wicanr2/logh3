package game_test

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
)

// 089正常主選單→S3同盟→實行→首戰→資料。保持到上界、向上到0，
// 各自比第57輪固定原版全圖；速度已核准不同，只宣稱同結果畫面。
func TestLeftHoldNormalBattleWholeFrame(t *testing.T) {
	root := filepath.Join("..", "..")
	orig := filepath.Join(root, "org_game", "extracted")
	if _, err := os.Stat(filepath.Join(orig, "GIN3.COM")); os.IsNotExist(err) {
		t.Skip("缺本機原版")
	}
	wants := map[string][]byte{}
	for name, hash := range map[string]string{
		"zero": "8fdb9d1b89c905dc50ed638aa19dd69a63d3949c5519cc8e35790aa1e154eecb",
		"last": "ed8e1a575a412db0fc95acffbe124c773bbd7cf3978c32278de4459159db0782",
	} {
		p := filepath.Join(root, "workplace", "oracle", "data67-r57-"+name+"-a", "screen.bin")
		b, err := os.ReadFile(p)
		if os.IsNotExist(err) {
			t.Skip("缺第57輪原版固定全圖收據")
		}
		if err != nil || len(b) != 307200 || fmt.Sprintf("%x", sha256.Sum256(b)) != hash {
			t.Fatalf("%s收據無效或漂移：%v", name, err)
		}
		wants[name] = b
	}
	r, err := bootstrap.Load(orig, filepath.Join(root, "docs", "internals", "data"))
	if err != nil {
		t.Fatal(err)
	}
	g, err := game.New(r)
	if err != nil {
		t.Fatal(err)
	}
	click := func(x, y int) { g.Click(x, y); g.Frame() }
	click(300, 270)
	if g.State() == game.StateMainMenu {
		click(300, 270)
	}
	for _, key := range []byte("138\r") {
		g.Key(key)
	}
	if g.State() != game.StateLoadGame {
		t.Fatalf("正常取進度在%v", g.State())
	}
	click(300, 322)
	click(318, 228)
	click(544, 125)
	click(560, 125)
	ticks := 0
	for ; ticks < 15000 && g.State() != game.StateBattle; ticks++ {
		switch g.State() {
		case game.StateTurnThinking, game.StateTurnWindow, game.StateTurnReport, game.StateEngagement:
			g.Tick(time.Second / 60)
		case game.StateOccupationReport, game.StateRebellionReport, game.StateRegionFlipReport,
			game.StateAdmiralFateReport, game.StateShipDelivery, game.StatePhezzanReport:
			click(240, 240)
		default:
			t.Fatalf("實行停在%v", g.State())
		}
		g.Frame()
	}
	if g.State() != game.StateBattle || ticks != 808 {
		t.Fatalf("正常首戰在%v，tick %d", g.State(), ticks)
	}
	if base := os.Getenv("LOGH3_LEFT_HOLD_OUT"); base != "" {
		out := filepath.Join(base, "battle-ready")
		if err := os.Mkdir(out, 0755); err != nil {
			t.Fatal(err)
		}
		if err := g.Frame().WriteIndex(filepath.Join(out, "screen.bin")); err != nil {
			t.Fatal(err)
		}
		if err := g.Frame().WritePNG(filepath.Join(out, "screen.png"), r.Menu.Palette()); err != nil {
			t.Fatal(err)
		}
	}
	click(400, 80)
	click(400, 100)
	for _, sample := range []struct {
		name string
		y    int
	}{{"last", r.BattleData.Scroll.Boxes[len(r.BattleData.Scroll.Boxes)-1].Y + 10},
		{"zero", r.BattleData.Scroll.Boxes[0].Y + 10}} {
		click(440, sample.y)
		for i := 0; i < 300; i++ {
			g.Tick(time.Second / 60)
			g.LeftHold(440, sample.y, time.Second/60)
			g.Frame()
		}
		g.LeftRelease()
		g.MoveMouse(440, 390)
		out := t.TempDir()
		if base := os.Getenv("LOGH3_LEFT_HOLD_OUT"); base != "" {
			out = filepath.Join(base, sample.name)
			if err := os.Mkdir(out, 0755); err != nil {
				t.Fatal(err)
			}
		}
		path := filepath.Join(out, "screen.bin")
		if err := g.Frame().WriteIndex(path); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(got, wants[sample.name]) {
			t.Fatalf("%s正常保持全圖不同：%v", sample.name, err)
		}
		if err := g.Frame().WritePNG(filepath.Join(out, "screen.png"), r.Menu.Palette()); err != nil {
			t.Fatal(err)
		}
		if err := g.SaveSlot(out, 0); err != nil {
			t.Fatal(err)
		}
		t.Logf("%s正常保持：307200像素零差", sample.name)
	}
}
