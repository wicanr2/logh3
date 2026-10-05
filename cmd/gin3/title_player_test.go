package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/screen"
)

func TestTitlePlayerMusic(t *testing.T) {
	orig := filepath.Join("..", "..", "org_game", "extracted")
	if _, err := os.Stat(filepath.Join(orig, "SPTIT.350")); os.IsNotExist(err) {
		t.Skip("缺原版標題")
	}
	r, err := bootstrap.Load(orig, filepath.Join("..", "..", "docs", "internals", "data"))
	if err != nil {
		t.Fatal(err)
	}
	if err := bootstrap.LoadIntro(r, orig); err != nil {
		t.Fatal(err)
	}
	g, err := game.New(r)
	if err != nil {
		t.Fatal(err)
	}
	p := &recordingMusic{}
	a := &app{g: g, music: p}
	sync := func() {
		t.Helper()
		if err := a.syncMusic(); err != nil {
			t.Fatal(err)
		}
	}
	toTitle := func() {
		t.Helper()
		for frame := 1; frame < r.Intro.Frames(); frame++ {
			g.Tick(screen.IntroFrameIntervalAt(frame))
		}
		g.Tick(game.IntroTitleDelay)
		if g.State() != game.StateTitle {
			t.Fatal("自然播放未到標題")
		}
	}
	sync()
	toTitle()
	sync()
	if !reflect.DeepEqual(p.transitions, []string{"OPEN_M"}) {
		t.Fatal("動畫到標題中斷音樂")
	}
	g.Tick(game.TitleWaitDuration)
	sync()
	if !reflect.DeepEqual(p.transitions, []string{"OPEN_M", "<stop>", "OPEN_M"}) {
		t.Fatal("逾時沒有重播OPEN_M", p.transitions)
	}
	toTitle()
	sync()
	g.Click(300, 210)
	sync()
	if !reflect.DeepEqual(p.transitions, []string{"OPEN_M", "<stop>", "OPEN_M", "<stop>"}) {
		t.Fatal("主選單沒有停配樂", p.transitions)
	}
}
