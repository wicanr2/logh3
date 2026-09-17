package bootstrap_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"runtime/debug"
	"testing"

	"github.com/wicanr2/logh3/internal/game"
)

// 每一個命令選項都要有作用、在裡面怎麼點都不能 crash（使用者 2026-09-17）。
//
// 命令列三個選單（戰略命令／一覽表／系統指令）共 13 項：從命令階段開選單、點那一項，斷言狀態真的換了；
// 進去之後把 40×24 個文字格的中心逐一點一次（左鍵），每點一下畫一次畫面，被帶回命令階段就重新進入。
// 列出「格子怎麼點畫面都沒變」的項目。

type barItem struct {
	name       string
	menuX      int // 選單鈕
	row        int // 第幾項
	wantMenu   game.State
	wantScreen game.State
}

var barItems = []barItem{
	{"待命提督", 352, 0, game.StateCmdMenu, game.StateStandbyAdmirals},
	{"情報收集", 352, 1, game.StateCmdMenu, game.StateInfo},
	{"徵稅計劃", 352, 2, game.StateCmdMenu, game.StateTaxPlan},
	{"生產計劃", 352, 3, game.StateCmdMenu, game.StateProductionPlan},
	{"星系資料", 448, 0, game.StateListMenu, game.StateStarList},
	{"艦隊資料", 448, 1, game.StateListMenu, game.StateFleetList},
	{"提督資料", 448, 2, game.StateListMenu, game.StateAdmiralListView},
	{"敵將資料", 448, 3, game.StateListMenu, game.StateEnemyList},
	{"取進度", 544, 1, game.StateSysMenu, game.StateLoadGame},
	{"存進度", 544, 2, game.StateSysMenu, game.StateSaveGame},
	{"設定", 544, 3, game.StateSysMenu, game.StateSettings},
	{"結束", 544, 4, game.StateSysMenu, game.StateQuitConfirm},
}

func frameHash(g *game.Game) [32]byte { return sha256.Sum256(g.Frame().Pix) }

// backToCommand 右鍵與點地圖空白處，最多 12 下回到命令階段。
func backToCommand(g *game.Game) bool {
	for i := 0; i < 12 && g.State() != game.StateCommandPhase; i++ {
		if i%2 == 0 {
			g.RightClick(100, 460)
		} else {
			g.Click(100, 460)
		}
	}
	return g.State() == game.StateCommandPhase
}

func enter(g *game.Game, it barItem) game.State {
	g.Click(it.menuX, 125)
	if g.State() != it.wantMenu {
		return g.State()
	}
	g.Click(it.menuX+8, 125+20*it.row)
	return g.State()
}

// knownDead 是已登記、還沒接的項目（issue）。這裡的項目開始有反應，測試就失敗，提醒回頭關 issue。
var knownDead = map[string]string{
	"徵稅計劃": "#64", "生產計劃": "#64", "結束": "#65",
}

func TestEveryCommandItemWorks(t *testing.T) {
	starts := monkeyStarts[:1]
	if os.Getenv("LOGH3_TOUR_ALL") != "" {
		starts = monkeyStarts[:3] // 帝國、同盟、對戰
	}
	for _, st := range starts {
		t.Run(st.name, func(t *testing.T) {
			g := loadGame(t)
			feed(g, st.script)
			if st.name == "對戰" {
				// 交棒視窗之後：關掉月初報告，回到命令階段。
				backToCommand(g)
			}
			if g.State() != game.StateCommandPhase {
				t.Skipf("起點腳本之後在 %v，不是命令階段", g.State())
			}
			var dead []string
			for _, it := range barItems {
				func() {
					cur := ""
					defer func() {
						if p := recover(); p != nil {
							t.Fatalf("%s：panic（%s，狀態 %v）：%v\n%s", it.name, cur, g.State(), p, debug.Stack())
						}
					}()
					if !backToCommand(g) {
						t.Fatalf("%s 之前回不到命令階段（停在 %v）", it.name, g.State())
					}
					if got := enter(g, it); got != it.wantScreen {
						t.Errorf("%s：開選單點第 %d 項之後在 %v，應該是 %v", it.name, it.row, got, it.wantScreen)
						return
					}
					base := frameHash(g)
					changed := 0
					for cy := 0; cy < 24; cy++ {
						for cx := 0; cx < 40; cx++ {
							x, y := cx*16+8, cy*20+10
							cur = fmt.Sprintf("點 (%d,%d)", x, y)
							before := g.State()
							g.Click(x, y)
							if g.State() != before || frameHash(g) != base {
								changed++
							}
							if it.wantScreen == game.StateQuitConfirm && g.State() == game.StateMainMenu {
								return // 「結束」真的結束了，這一局不能再用
							}
							if g.State() != it.wantScreen {
								if !backToCommand(g) {
									t.Fatalf("%s：%s 之後回不到命令階段（停在 %v）", it.name, cur, g.State())
								}
								if enter(g, it) != it.wantScreen {
									t.Fatalf("%s：重新進入失敗（停在 %v）", it.name, g.State())
								}
							}
							base = frameHash(g)
						}
					}
					if changed == 0 {
						dead = append(dead, it.name)
					}
					t.Logf("%s：960 格裡 %d 格點了有變化", it.name, changed)
				}()
			}
			seen := map[string]bool{}
			for _, d := range dead {
				seen[d] = true
				if issue, ok := knownDead[d]; ok {
					t.Logf("%s 怎麼點都沒反應（已登記 %s）", d, issue)
				} else {
					t.Errorf("%s 怎麼點都沒反應，而且沒有登記", d)
				}
			}
			for d, issue := range knownDead {
				if !seen[d] {
					t.Errorf("%s 開始有反應了——回頭核對 %s 能不能關", d, issue)
				}
			}
		})
	}
}
