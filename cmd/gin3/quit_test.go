package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
)

// 117：公開確認的 Game 旗標必須交到正式 app；取消仍繼續更新。
func TestPlayerPublicQuit(t *testing.T) {
	orig := filepath.Join("..", "..", "org_game", "extracted")
	if _, err := os.Stat(filepath.Join(orig, "GIN3.COM")); os.IsNotExist(err) {
		t.Skip("缺原版資料")
	}
	r, err := bootstrap.Load(orig, filepath.Join("..", "..", "docs", "internals", "data"))
	if err != nil {
		t.Fatal(err)
	}
	g, err := game.New(r)
	if err != nil {
		t.Fatal(err)
	}
	if err := feed(g, "click 300 210;type 138;enter;rclick 300 240;rclick 300 240;click 544 125;click 560 205"); err != nil {
		t.Fatal(err)
	}
	if g.State() != game.StateQuitConfirm {
		t.Fatal("正常選單未到退出確認", g.State())
	}
	g.Click(383, 263)
	// 測試使用無動作腳本分支，避免讀取測試程序的 X11 輸入。
	a := &app{g: g, fed: true, script: "move 300 240"}
	if g.Quit() || g.State() != game.StateCommandPhase {
		t.Fatal("取消未返回命令階段")
	}
	if err := a.Update(); err != nil {
		t.Fatal("取消後 app 結束", err)
	}
	g.Click(544, 125)
	g.Click(560, 205)
	g.Click(432, 263)
	if !g.Quit() {
		t.Fatal("公開確認未設定退出旗標")
	}
	if err := a.Update(); err != ebiten.Termination {
		t.Fatal("正式 app 未結束", err)
	}
}
