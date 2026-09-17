package bootstrap_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
)

// 「所有功能選項都要有作用、不能 crash」的驗證（使用者 2026-09-17）。
//
// 用正式入口（`bootstrap.Load`，與 `cmd/gin3` 同一份資源、同一個戰術 AI）從幾個起點出發，以固定種子隨機送
// 左鍵、右鍵、移動、按鍵、時間與「系統指令 → 實行」，**每一步都畫一次畫面**。任何 panic 都印出起點、種子、
// 第幾步與最後 40 個動作，照那一串就能重現。另外統計走到過哪些狀態，列出從沒走到的畫面。
//
// 預設每個起點 × 種子跑 `LOGH3_MONKEY_STEPS`（預設 400，冒煙用）步；深度驗證用 `tools/feature_check.sh`。

const root = "../.."

// lastSaveDir 是最近一次 loadGame 的存檔目錄（檢查「存進度」有沒有真的寫檔）。
var lastSaveDir string

type monkeyStart struct {
	name   string
	script string
}

var monkeyStarts = []monkeyStart{
	{"帝國新局", "move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240"},
	{"同盟新局", "move 300 230; click; type 138; enter; rclick 300 240; rclick 300 240"},
	{"對戰", "move 300 250; click; type 126; enter; click 432 263"},
	{"主選單", ""},
	// 深層畫面：艦隊視窗（`tools/gui_check.sh` 的 fcmd-win）與 S3 第一回合進戰場（`battleTestGame` 的序列，實行之後照計時走過思考中與過場）。
	{"艦隊視窗", "move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 352 125; click 400 125; click 400 85; click 488 330; click 480 350; click 510 350; click 450 308"},
	{"戰場", "move 300 210; click; type 138; enter; rclick 300 240; rclick 300 240; move 600 340; click 544 125; click 560 145; click 300 322; click 318 208; click 544 125; click 560 125; tick 800"},
}

// lcg 是測試自己的決定性亂數（不碰遊戲的亂數序列）。
type lcg uint64

func (r *lcg) next(n int) int {
	*r = *r*6364136223846793005 + 1442695040888963407
	return int((uint64(*r) >> 33) % uint64(n))
}

func loadGame(t *testing.T) *game.Game {
	t.Helper()
	return loadGameSaveDir(t, t.TempDir())
}

// loadGameSaveDir 同 loadGame，但存檔目錄由呼叫端給（走正式的 `bootstrap.UseSaveDir`）。
func loadGameSaveDir(t *testing.T, saveDir string) *game.Game {
	t.Helper()
	orig := filepath.Join(root, "org_game/extracted")
	if _, err := os.Stat(filepath.Join(orig, "GIN3.COM")); err != nil {
		t.Skip("沒有原版素材 org_game/extracted，跳過")
	}
	res, err := bootstrap.Load(orig, filepath.Join(root, "docs/internals/data"))
	if err != nil {
		t.Fatal(err)
	}
	if err := bootstrap.UseSaveDir(res, saveDir); err != nil { // 存進度寫到暫存目錄，原版目錄一律唯讀
		t.Fatal(err)
	}
	lastSaveDir = res.SaveDir
	g, err := game.New(res)
	if err != nil {
		t.Fatal(err)
	}
	g.SetTactic(ai.OriginalTactic{}, false) // 與 cmd/gin3 預設相同
	return g
}

// feed 照 `cmd/gin3 -script` 的語法送一串動作。
func feed(g *game.Game, script string) {
	for _, cmd := range strings.Split(script, ";") {
		f := strings.Fields(strings.TrimSpace(cmd))
		if len(f) == 0 {
			continue
		}
		xy := func() (int, int) {
			if len(f) == 3 {
				x, _ := strconv.Atoi(f[1])
				y, _ := strconv.Atoi(f[2])
				return x, y
			}
			return g.Mouse()
		}
		switch f[0] {
		case "move":
			g.MoveMouse(xy())
		case "click":
			g.Click(xy())
		case "rclick":
			g.RightClick(xy())
		case "type":
			for _, c := range []byte(f[1]) {
				g.Key(c)
			}
		case "enter":
			g.Key('\r')
		case "tick":
			n, _ := strconv.Atoi(f[1])
			for i := 0; i < n && g.State() != game.StateBattle; i++ {
				g.Tick(g.TimerTick())
			}
		}
	}
}

// monkeyAction 產生下一個動作（腳本語法，方便重現）。
func monkeyAction(r *lcg) string {
	x, y := r.next(640), r.next(480)
	switch k := r.next(100); {
	case k < 55:
		return fmt.Sprintf("click %d %d", x, y)
	case k < 67:
		return fmt.Sprintf("rclick %d %d", x, y)
	case k < 75:
		return fmt.Sprintf("move %d %d", x, y)
	case k < 83:
		keys := []string{"1", "3", "8", "0", "a", "\r", "\x1b", "\b", " "}
		return "key " + strconv.Quote(keys[r.next(len(keys))])
	case k < 95:
		return fmt.Sprintf("tick %d", 1+r.next(120))
	default:
		return "execute"
	}
}

func apply(g *game.Game, a string) {
	f := strings.SplitN(a, " ", 2)
	switch f[0] {
	case "key":
		s, _ := strconv.Unquote(f[1])
		g.Key(s[0])
	case "tick":
		n, _ := strconv.Atoi(f[1])
		for i := 0; i < n; i++ {
			g.Tick(g.TimerTick())
		}
	case "execute":
		g.Click(544, 125)
		g.Click(560, 125)
		for i := 0; i < 400 && g.State() != game.StateCommandPhase; i++ {
			g.Tick(g.TimerTick())
		}
	default:
		feed(g, a)
	}
}

func monkeySteps() int {
	if n, err := strconv.Atoi(os.Getenv("LOGH3_MONKEY_STEPS")); err == nil && n > 0 {
		return n
	}
	return 400
}

func TestMonkeyNoPanic(t *testing.T) {
	steps := monkeySteps()
	visited := map[string]int{}
	for _, st := range monkeyStarts {
		for seed := uint64(1); seed <= 3; seed++ {
			name := fmt.Sprintf("%s／種子 %d", st.name, seed)
			t.Run(name, func(t *testing.T) {
				g := loadGame(t)
				feed(g, st.script)
				r := lcg(seed)
				var log []string
				step := -1
				defer func() {
					if p := recover(); p != nil {
						from := len(log) - 40
						if from < 0 {
							from = 0
						}
						t.Fatalf("panic 在第 %d 步（狀態 %v）：%v\n起點腳本：%s\n最後幾個動作：\n  %s\n%s",
							step, g.State(), p, st.script, strings.Join(log[from:], "\n  "), debug.Stack())
					}
				}()
				for step = 0; step < steps; step++ {
					a := monkeyAction(&r)
					log = append(log, a)
					apply(g, a)
					g.Frame()
					visited[g.State().String()]++
				}
			})
		}
	}
	var names []string
	for n := range visited {
		names = append(names, n)
	}
	sort.Strings(names)
	t.Logf("走到過的狀態 %d 種：%s", len(names), strings.Join(names, " "))
	var never []string
	for s := game.State(0); ; s++ {
		n := s.String()
		if strings.HasPrefix(n, "State(") {
			break
		}
		if visited[n] == 0 {
			never = append(never, n)
		}
	}
	t.Logf("沒走到的狀態 %d 種：%s", len(never), strings.Join(never, " "))
}
