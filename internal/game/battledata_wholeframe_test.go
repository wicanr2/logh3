package game_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
)

// 正常主選單載入S3同盟，再實行到首戰；不匯入原版RAM或注入戰場。
func TestBattleDataScrollWholeFrameMatchesOracle(t *testing.T) {
	root := filepath.Join("..", "..")
	orig := filepath.Join(root, "org_game", "extracted")
	if _, err := os.Stat(filepath.Join(orig, "GIN3.COM")); os.IsNotExist(err) {
		t.Skip("本機原版不存在")
	} else if err != nil {
		t.Fatal(err)
	}
	names := []string{"zero", "one", "last"}
	wantHashes := []string{
		"8fdb9d1b89c905dc50ed638aa19dd69a63d3949c5519cc8e35790aa1e154eecb",
		"c62da2e5eb1b406bb6b5d418a75d795b9252ac63c82801a86244c197a7941550",
		"ed8e1a575a412db0fc95acffbe124c773bbd7cf3978c32278de4459159db0782",
	}
	oracle := make([][]byte, len(names))
	for i, name := range names {
		b, err := os.ReadFile(filepath.Join(root, "workplace", "oracle", "data67-r57-"+name+"-a", "screen.bin"))
		if os.IsNotExist(err) {
			t.Skip("缺少第57輪原版完整畫面收據；執行tools/battle_data_scroll_oracle.sh")
		}
		if err != nil || len(b) != 640*480 {
			t.Fatalf("原版%s完整畫面：%v，長度%d", name, err, len(b))
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(b)); got != wantHashes[i] {
			t.Fatalf("原版%s收據已漂移：%s", name, got)
		}
		oracle[i] = b
		mem, err := os.ReadFile(filepath.Join(root, "workplace", "oracle", "data67-r57-"+name+"-a", "mem.bin"))
		if err != nil || len(mem) != 1048576 {
			t.Fatalf("原版%s RAM：%v，長度%d", name, err, len(mem))
		}
		wantOffset := []uint16{0, 1, 26}[i]
		if binary.LittleEndian.Uint16(mem[0x5E6B6:]) != wantOffset || binary.LittleEndian.Uint16(mem[0x5E6B8:]) != 40 {
			t.Fatalf("原版%s未處於四十列、偏移%d", name, wantOffset)
		}
	}
	r, err := bootstrap.Load(orig, filepath.Join(root, "docs", "internals", "data"))
	if err != nil {
		t.Fatal(err)
	}
	g, err := game.New(r)
	if err != nil {
		t.Fatal(err)
	}
	click := func(x, y int) {
		g.MoveMouse(x, y)
		g.Click(x, y)
		g.Frame()
	}
	click(300, 270)
	if g.State() == game.StateMainMenu {
		click(300, 270)
	}
	for _, k := range []byte("138\r") {
		g.Key(k)
	}
	if g.State() != game.StateLoadGame {
		t.Fatalf("取進度：%v", g.State())
	}
	click(300, 322)
	click(318, 228)
	if g.State() != game.StateCommandPhase {
		t.Fatalf("S3同盟：%v", g.State())
	}
	click(544, 125)
	click(560, 125)
	ticks := 0
	for ; ticks < 15000 && g.State() != game.StateBattle; ticks++ {
		switch g.State() {
		case game.StateTurnThinking, game.StateTurnWindow, game.StateTurnReport, game.StateEngagement:
			g.Tick(time.Second / 60)
		case game.StateOccupationReport, game.StateRebellionReport, game.StateRegionFlipReport, game.StateAdmiralFateReport, game.StateShipDelivery, game.StatePhezzanReport:
			click(240, 240)
		default:
			t.Fatalf("實行停在%v", g.State())
		}
		g.Frame()
	}
	if g.State() != game.StateBattle || ticks != 808 {
		t.Fatalf("首戰：state=%v ticks=%d", g.State(), ticks)
	}
	// 表頭開窗，避免表列指向改鏡頭；與原版收據相同的落點。
	click(400, 80)
	click(400, 100)
	g.MoveMouse(440, 390)
	for i, name := range names {
		if i == 1 {
			click(440, 390)
		} else if i == 2 {
			for j := 0; j < 26; j++ {
				click(440, 390)
			}
		}
		fb := g.Frame()
		out := t.TempDir()
		if base := os.Getenv("LOGH3_DATA_SCROLL_OUT"); base != "" {
			out = filepath.Join(base, name)
			if err := os.Mkdir(out, 0755); err != nil {
				t.Fatal(err)
			}
		}
		path := filepath.Join(out, "screen.bin")
		if err := fb.WriteIndex(path); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		diff := 0
		for p := range got {
			if got[p] != oracle[i][p] {
				diff++
			}
		}
		if diff != 0 {
			t.Errorf("%s整張640×480畫面差%d像素", name, diff)
		}
		if err := fb.WritePNG(filepath.Join(out, "screen.png"), r.Menu.Palette()); err != nil {
			t.Fatal(err)
		}
		if err := g.SaveSlot(out, 0); err != nil {
			t.Fatal(err)
		}
		t.Logf("%s：完整307200像素零差；正常入口808 tick", name)
	}
	before := g.Frame()
	path := filepath.Join(t.TempDir(), "before.bin")
	if err := before.WriteIndex(path); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	click(440, 390)
	if err := g.Frame().WriteIndex(path); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(b, after) {
		t.Fatalf("捲動上界未夾住：%v", err)
	}
}
