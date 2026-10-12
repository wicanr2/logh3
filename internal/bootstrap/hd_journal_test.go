package bootstrap_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/presentation"
)

// 以正常公開玩家輸入核顯示紀錄，不注入畫面或取代正式renderer。
func TestHDJournalNormalPlayerScreens(t *testing.T) {
	for faction := 0; faction < 2; faction++ {
		t.Run(fmt.Sprint(faction), func(t *testing.T) {
			g := loadGame(t)
			steps := []struct {
				script string
				state  game.State
			}{
				{"", game.StateMainMenu},
				{fmt.Sprintf("click 300 %d", 210+20*faction), game.StateCopyProtection},
				{"type 138;enter", game.StateMainGame},
				{"click 448 350", game.StateBudgetPlan},
				{"click 487 329", game.StateBudgetPlan},
				{"click 463 370", game.StateMainGame},
				{"click 448 370", game.StateShipOrder},
				{"click 448 370", game.StateCommandPhase},
				{"click 544 125;click 560 185", game.StateSettings},
				{"rclick 600 400;click 544 125;click 560 145", game.StateLoadGame},
				{"click 300 282", game.StateLoadFaction},
				{"click 318 208", game.StateCommandPhase},
			}
			for _, step := range steps {
				feed(g, step.script)
				if g.State() != step.state {
					t.Fatalf("%q: %v", step.script, g.State())
				}
				f := g.Frame()
				if got := gfx.ReplayIndexed(f.Commands); !bytes.Equal(got, f.Pix) {
					for i := range got {
						if got[i] != f.Pix[i] {
							t.Fatalf("%v: 原色重播差異 %d,%d：%d != %d", g.State(), i%gfx.Width, i/gfx.Width, got[i], f.Pix[i])
						}
					}
				}
			}
		})
	}
}

func TestHDFactionThemeNormalLoadAndSaveIsolation(t *testing.T) {
	res, err := bootstrap.Load(filepath.Join(root, "org_game/extracted"), filepath.Join(root, "docs/internals/data"))
	if err != nil {
		t.Skip(err)
	}
	hd, err := presentation.Load(filepath.Join(root, "assets/hd"), res.Faces)
	if err != nil {
		t.Fatal(err)
	}
	for faction, id := range []string{"empire", "alliance"} {
		g := loadGame(t)
		if g.Frame().InterfaceTheme != "" {
			t.Fatal("主選單應保持中性")
		}
		feed(g, fmt.Sprintf("click 300 %d", 210+20*faction))
		if g.Frame().InterfaceTheme != "" {
			t.Fatal("防拷頁不應沿未開局陣營換色")
		}
		feed(g, "type 138;enter;click 448 350")
		if g.Frame().InterfaceTheme != id {
			t.Fatal("新局主題與實際陣營不同")
		}
		// 沿玩家路徑關閉月初與訂造報告，正式遊戲才允許保存。
		feed(g, "click 463 370;click 448 370;click 448 370")
		if g.State() != game.StateCommandPhase {
			t.Fatal("未完成月初正常流程", g.State())
		}
		before, after := t.TempDir(), t.TempDir()
		if err := g.SaveSlot(before, 0); err != nil {
			t.Fatal(err)
		}
		fb := g.Frame()
		pixels := bytes.Clone(fb.Pix)
		pal := g.Palette()
		hd.Render(fb, pal)
		if !bytes.Equal(pixels, fb.Pix) || g.Palette() != pal || !bytes.Equal(gfx.ReplayIndexed(fb.Commands), fb.Pix) {
			t.Fatal("HD主題改寫原畫面或色盤")
		}
		if err := g.SaveSlot(after, 0); err != nil {
			t.Fatal(err)
		}
		for _, stem := range []string{"ADMIRAL", "FLEET", "SHIP", "WORK", "STELLAR", "PLANET"} {
			a, e := os.ReadFile(filepath.Join(before, stem+".DT0"))
			if e != nil {
				t.Fatal(e)
			}
			b, e := os.ReadFile(filepath.Join(after, stem+".DT0"))
			if e != nil {
				t.Fatal(e)
			}
			if !bytes.Equal(a, b) {
				t.Fatal(stem, "因主題顯示而改變")
			}
		}
		// 同一個遊戲由公開S1讀檔選另一陣營，Frame不可保留上一方主題。
		feed(g, "click 544 125;click 560 145;click 300 282")
		feed(g, fmt.Sprintf("click 318 %d", 208+20*(1-faction)))
		if g.Faction() != 1-faction || g.Frame().InterfaceTheme == id {
			t.Fatalf("公開讀檔未切換陣營主題：%v %d %s", g.State(), g.Faction(), g.Frame().InterfaceTheme)
		}
	}
}

func TestHDFactionThemeVersusHandover(t *testing.T) {
	g := loadGame(t)
	feed(g, "click 300 270;type 138;enter;click 300 322;click 318 248")
	if g.State() != game.StateVersusInput || g.Frame().InterfaceTheme != "empire" {
		t.Fatal("載入雙人缺帝國主題", g.State())
	}
	feed(g, "click 432 270;click 544 125;click 560 125")
	if g.State() != game.StateVersusInput || g.Frame().InterfaceTheme != "alliance" {
		t.Fatal("交棒後沒有切換同盟主題", g.State())
	}
}

func BenchmarkHDOriginalLayoutBudget(b *testing.B) {
	res, err := bootstrap.Load(filepath.Join(root, "org_game/extracted"), filepath.Join(root, "docs/internals/data"))
	if err != nil {
		b.Skip(err)
	}
	hd, err := presentation.Load(filepath.Join(root, "assets/hd"), res.Faces)
	if err != nil {
		b.Fatal(err)
	}
	res.Faces.ArtworkPrefix = "face"
	g, err := game.New(res)
	if err != nil {
		b.Fatal(err)
	}
	feed(g, "click 300 230;type 138;enter;click 448 350")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hd.Render(g.Frame(), g.Palette())
	}
}
