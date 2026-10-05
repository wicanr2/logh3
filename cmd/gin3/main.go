// Command gin3 是 remake 的可玩入口：Ebiten 開視窗、把 internal/game 的
// 畫面畫出來、把滑鼠與鍵盤送回去。
//
// 遊戲邏輯與繪製都在 internal/，這一層只做三件事：把色號陣列展開成 RGBA、
// 把視窗座標換成遊戲座標、把輸入轉成 game 的動作。**原版素材由玩家自備**
// （`-orig`），本專案不內嵌任何原版檔。
//
//	tools/gui.sh run ./cmd/gin3 -orig /orig
//	tools/gui.sh run ./cmd/gin3 -orig /orig -shot /src/workplace/gui.png -shot-after 8
package main

import (
	"flag"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	eaudio "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/wicanr2/logh3/internal/audio"
	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/gfx"
)

type app struct {
	g          *game.Game
	pal        [gfx.Colors]color.RGBA
	screen     *ebiten.Image
	music      musicPlayback
	musicState game.State
	effects    effectPlayback
	buf        []byte

	shot      string
	shotBin   string
	stateOut  string
	shotAfter int
	ticks     int

	// script 是開場先送進去的輸入（與 cmd/play 同一套語法）。
	// 給無頭驗收用：同一條玩家路徑走完再倒圖，才比得出顯示層有沒有畫錯。
	script string
	fed    bool
}

// musicPlayback 是正常玩家路徑與音訊裝置之間的窄邊界。正式執行放
// musicPlayer；測試放不開裝置的記錄器，仍然能驗「哪個狀態要求哪首曲」。
type musicPlayback interface {
	Play(string) error
	Stop()
	Playing() string
}

// tuneFor 回這個狀態該播哪一首。對照表本身在 `internal/audio`
// （純的、測得到；這個套件連結時要 ALSA，測試須走 GUI image），
// 來源是 `docs/internals/17` §哪個畫面播哪一首 與 `docs/re/274`。
//
// 三格設定從遊戲狀態讀（設定畫面點了就換、存檔帶得走）。**試聽是這裡
// 長出來的**：每一幀都問一次「現在該播哪一首」，換了就換曲——原版
// 選完音源或平時曲會試聽、選戰鬥曲不會，正好對上 `TuneFor` 用得到
// 哪幾格。
func (a *app) tuneFor() string {
	return audio.TuneFor(a.g.State(), a.g.Faction(), a.musicSettings())
}

// musicSettings 把 `Game` 上那三格換成曲庫看得懂的樣子。
func (a *app) musicSettings() audio.Settings {
	m := a.g.MusicSettings()
	return audio.Settings{Device: m.Device, BGM: m.BGM, Battle: m.Battle}
}

// syncMusic 把目前遊戲狀態同步到播放器。獨立成一支是為了讓沒有音效裝置的
// 容器仍能驗正式 app 與 Game 的接線，不用拿合成音訊冒充人耳驗收。
func (a *app) syncMusic() error {
	st := a.g.State()
	// 112：原版標題逾時會重新讀OPEN_M；同曲也必須從頭播放。
	if a.musicState == game.StateTitle && st == game.StateIntro {
		a.music.Stop()
	}
	a.musicState = st
	tune := a.tuneFor()
	if tune == "" {
		a.music.Stop()
		return nil
	}
	return a.music.Play(tune)
}

func (a *app) Update() error {
	// 正式播放器完成才允許下一音效及圖段首格（規格105）。
	// 只看名義時長會在更新追趕時關掉仍有樣本的播放器。
	a.g.SetSoundEffectPlaying(a.effects != nil && a.effects.Busy())
	// 時間往前走一格；Game統一推進動畫、標題等待及回合演出。
	a.g.Tick(time.Second / time.Duration(ebiten.TPS()))
	if a.music != nil {
		if err := a.syncMusic(); err != nil {
			// 配樂出錯不該把遊戲帶走，但**要留下痕跡**——安靜地沒有音樂
			// 與「還沒接播放層」長得一樣。
			fmt.Fprintln(os.Stderr, "配樂：", err)
			a.music = nil
		}
	}
	if !a.fed {
		a.fed = true
		if err := feed(a.g, a.script); err != nil {
			return err
		}
	}
	if a.script != "" {
		// 走腳本的時候不吃真的滑鼠鍵盤——不然 Xvfb 的游標會混進來。
		a.updateEffects()
		a.ticks++
		if (a.shot != "" || a.shotBin != "") && a.ticks >= a.shotAfter {
			return a.dump()
		}
		return nil
	}
	// 視窗就是 640×480 的邏輯畫布，Game 也吃畫面座標，所以直接傳。
	x, y := ebiten.CursorPosition()
	a.g.MoveMouse(x, y)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		a.g.Click(x, y)
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		a.g.LeftHold(x, y, time.Second/time.Duration(ebiten.TPS()))
	} else if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		a.g.LeftRelease()
	}
	switch {
	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight):
		a.g.RightClick(x, y)
	case ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight):
		// 按著沒放：戰場上每一幀再捲一格視野（`docs/re/326`）。
		a.g.RightHold(x, y)
	case inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight):
		a.g.RightRelease()
	}
	for k := ebiten.Key0; k <= ebiten.Key9; k++ {
		if inpututil.IsKeyJustPressed(k) {
			a.g.Key(byte('0' + (k - ebiten.Key0)))
		}
	}
	for k := ebiten.KeyKP0; k <= ebiten.KeyKP9; k++ {
		if inpututil.IsKeyJustPressed(k) {
			a.g.Key(byte('0' + (k - ebiten.KeyKP0)))
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		a.g.Key('\r')
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		a.g.Key(8)
	}
	a.updateEffects()
	a.ticks++
	if (a.shot != "" || a.shotBin != "") && a.ticks >= a.shotAfter {
		return a.dump()
	}
	return nil
}

// dump 把畫面倒出來就結束。`.bin` 是色號陣列，格式與 dosgolem 的
// `-dump-ega` 相同，可以直接與收據逐位元組比。
func (a *app) dump() error {
	a.setPalette(a.g.Palette())
	fb := a.g.Frame()
	if a.shot != "" {
		if err := fb.WritePNG(a.shot, a.paletteSpec()); err != nil {
			return err
		}
	}
	if a.shotBin != "" {
		if err := fb.WriteIndex(a.shotBin); err != nil {
			return err
		}
	}
	if a.stateOut != "" {
		windowW, windowH := ebiten.WindowSize()
		x, y := ebiten.CursorPosition()
		receipt := fmt.Sprintf("state=%s\nwindow=%dx%d\ncursor=%d,%d\ncursor_mode=%d\nupdates=%d\nintro_frame=%d\n",
			a.g.State(), windowW, windowH, x, y, ebiten.CursorMode(), a.ticks, a.g.IntroFrame())
		if err := os.WriteFile(a.stateOut, []byte(receipt), 0o644); err != nil {
			return err
		}
	}
	return ebiten.Termination
}

// setPalette 把 6 位元調色盤展開成畫面用的 RGBA。
func (a *app) setPalette(p gfx.Palette) {
	for i := range a.pal {
		r, g, b := p.RGB8(i)
		a.pal[i] = color.RGBA{r, g, b, 255}
	}
}

// paletteSpec 回畫 PNG 用的調色盤（6 位元，與 gfx 的慣例相同）。
func (a *app) paletteSpec() gfx.Palette {
	var p gfx.Palette
	for i := range p {
		c := a.pal[i]
		p[i] = [3]uint8{c.R >> 2, c.G >> 2, c.B >> 2}
	}
	return p
}

func (a *app) Draw(screen *ebiten.Image) {
	// 調色盤每格問一次：開頭動畫那一段是自己的一組灰階。
	a.setPalette(a.g.Palette())
	fb := a.g.Frame()
	for i, v := range fb.Pix {
		c := a.pal[v]
		a.buf[i*4], a.buf[i*4+1], a.buf[i*4+2], a.buf[i*4+3] = c.R, c.G, c.B, 255
	}
	a.screen.WritePixels(a.buf)
	screen.DrawImage(a.screen, nil)
}

func (a *app) Layout(int, int) (int, int) { return gfx.Width, gfx.Height }

func main() {
	orig := flag.String("orig", "org_game/extracted", "原版檔案目錄（玩家自備）")
	data := flag.String("data", "docs/internals/data", "畫面規格與資料目錄")
	scale := flag.Float64("scale", 1, "視窗放大倍率")
	shot := flag.String("shot", "", "跑幾格之後把畫面倒成 PNG 就結束（無頭驗收用）")
	script := flag.String("script", "", "開場先送進去的輸入，分號分隔："+
		"move X Y／click [X Y]／rclick [X Y]／type 數字／enter／back")
	shotBin := flag.String("shot-bin", "", "同 -shot，但倒的是色號陣列（可與收據逐位元組比）")
	stateOut := flag.String("state-out", "", "倒圖時另寫 GUI 狀態、視窗與邏輯游標收據")
	shotAfter := flag.Int("shot-after", 4, "倒圖前先跑幾格")
	// 預設走已接上的原版戰略 AI；強化版由玩家選用（規格 017／068）。
	intro := flag.Bool("intro", true, "從開頭動畫開始（`OPEN.EXE` 的 374 格）；"+
		"素材不齊就自動跳過")
	saveDir := flag.String("save", defaultSaveDir(), "玩家存檔的目錄（「存進度」寫這裡，原版目錄一律唯讀）")
	aiLevel := flag.Int("ai", -1, "電腦 AI：-1 ＝ 原版；0–3 ＝ 強化版，"+
		"數字越大越兇（會挑目標、集結、包圍）")
	musicDir := flag.String("music", "",
		"配樂與音效OGG目錄；空值停用音訊。\n"+
			"    玩家自備原版，配樂用 tools/music_to_ogg.sh 匯入；\n"+
			"    音效用 bash tools/sound_to_ogg.sh <新批次名稱> 匯入。\n"+
			"    曲庫及完整28段SFX_00.ogg至SFX_27.ogg放同一目錄；發行包不內嵌原版素材。")
	flag.Parse()

	res, err := bootstrap.Load(*orig, *data)
	if err != nil {
		die(err)
	}
	// 走腳本的時候不放動畫：腳本是照「開機就是主選單」寫的，
	// 前面多一張畫面整條序列就對不上收據。
	if *intro && *script == "" {
		if err := bootstrap.LoadIntro(res, *orig); err != nil {
			die(err)
		}
	}
	if *saveDir != "" {
		if err := bootstrap.UseSaveDir(res, *saveDir); err != nil {
			die(err)
		}
	}
	g, err := game.New(res)
	if err != nil {
		die(err)
	}
	if err := g.SetAILevel(*aiLevel); err != nil {
		die(err)
	}
	a := &app{
		g:         g,
		screen:    ebiten.NewImage(gfx.Width, gfx.Height),
		buf:       make([]byte, gfx.Width*gfx.Height*4),
		shot:      *shot,
		shotBin:   *shotBin,
		stateOut:  *stateOut,
		shotAfter: *shotAfter,
		script:    *script,
	}
	// 配樂：**要求了才開音訊**。缺檔就不播，但要說出來——安靜與
	// 「沒接播放層」長得一樣，不講的話沒有人分得出是哪一種
	// （AGENTS.md：缺檔就 skip，不用代用品）。
	if *musicDir != "" {
		lib := audio.NewLibrary(*musicDir)
		sounds := audio.NewSoundLibrary(*musicDir)
		var audioContext *eaudio.Context
		deviceAvailable := hasAudioDevice()
		if deviceAvailable && (lib.Have() > 0 || sounds.Have() == 28) {
			audioContext = eaudio.NewContext(audio.SampleRate)
		}
		if audioContext != nil && sounds.Have() == 28 {
			player := newEffectPlayer(audioContext, sounds)
			if err := player.validate(); err != nil {
				fmt.Fprintln(os.Stderr, "音效：", err)
			} else if err := g.EnableSoundEffects(true); err != nil {
				fmt.Fprintln(os.Stderr, "音效：", err)
			} else {
				a.effects = player
				defer a.effects.Stop()
			}
		} else {
			fmt.Fprintf(os.Stderr, "音效：%d/28段或裝置不可用，未啟用；來源轉檔見規格104\n", sounds.Have())
		}
		switch {
		case !deviceAvailable:
			fmt.Fprintln(os.Stderr, "配樂：這台機器沒有可用的音效裝置，"+
				"這一趟沒有音樂")
		case lib.Have() == 0:
			fmt.Fprintf(os.Stderr, "配樂：%s 裡一首都沒有，這一趟沒有音樂。"+
				"轉檔看 tools/music_to_ogg.sh\n", lib.Dir())
		default:
			if miss := lib.Missing(); len(miss) > 0 {
				fmt.Fprintf(os.Stderr, "配樂：%d/%d 首，缺 %v\n",
					lib.Have(), len(audio.Tracks), miss)
			}
			a.music = newMusicPlayer(audioContext, lib)
		}
	}

	a.setPalette(g.Palette())
	ebiten.SetWindowSize(int(float64(gfx.Width)*(*scale)), int(float64(gfx.Height)*(*scale)))
	ebiten.SetWindowTitle("銀河英雄傳說III SP")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	// 遊戲會依原版素材自行畫游標；若保留作業系統游標，使用者會同時看到
	// 即時的硬體游標與下一更新格才重畫的遊戲游標，視覺上必然不同步。
	ebiten.SetCursorMode(ebiten.CursorModeHidden)
	err = ebiten.RunGame(a)
	if err != nil && err != ebiten.Termination && a.music != nil &&
		strings.Contains(err.Error(), "oto:") {
		// 裝置在但開不起來（被佔用、權限、PulseAudio 沒跑）。
		// 原始訊息只有 ALSA 的字串，看不出與配樂有關。
		die(fmt.Errorf("配樂開不起來：%w\n（用 -music \"\" 關掉配樂就能跑）", err))
	}
	if err != nil && err != ebiten.Termination {
		die(err)
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// feed 把腳本送進遊戲。語法與 cmd/play 相同。
func feed(g *game.Game, script string) error {
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
			x, _ := strconv.Atoi(f[1])
			y, _ := strconv.Atoi(f[2])
			g.MoveMouse(x, y)
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
		case "back":
			g.Key(8)
		case "tick":
			// tick N：推進 N 個原版計時器 tick（回合裡的「思考中」與過場要等計時才換畫面）。
			n, err := strconv.Atoi(f[1])
			if err != nil || len(f) != 2 {
				return fmt.Errorf("tick 要一個整數：%q", cmd)
			}
			for i := 0; i < n; i++ {
				g.Tick(g.TimerTick())
			}
		default:
			return fmt.Errorf("看不懂的指令：%q", cmd)
		}
	}
	return nil
}

// defaultSaveDir 是使用者設定目錄底下的 logh3/save；拿不到設定目錄就不存。
func defaultSaveDir() string {
	d, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(d, "logh3", "save")
}
