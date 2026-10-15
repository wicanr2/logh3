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

	"github.com/wicanr2/logh3/internal/audio"
	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/presentation"
	"github.com/wicanr2/logh3/internal/ui"
)

type app struct {
	g                 *game.Game
	pal               [gfx.Colors]color.RGBA
	screen            *ebiten.Image
	music             musicPlayback
	musicOriginal     *audio.Library
	musicHD           *audio.Library
	musicMT32         *audio.Library
	hdMusic           bool
	soundFont         string
	soundFontMenu     bool
	soundFontError    string
	preferencesPath   string
	ruleSettingsError string
	cursor            *ui.Cursor
	musicState        game.State
	effects           effectPlayback
	buf               []byte
	hd                *presentation.Renderer
	original          bool
	hdScreen          *ebiten.Image
	hdUploaded        uint64
	frameTrace        *os.File
	animationTime     time.Time
	toolbar           toolbar

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
	// 117：公開確認退出後，由 Ebitengine 正常收尾。
	if a.g.Quit() {
		return ebiten.Termination
	}
	// 正式播放器完成才允許下一音效及圖段首格（規格105）。
	// 只看名義時長會在更新追趕時關掉仍有樣本的播放器。
	a.g.SetSoundEffectPlaying(a.effects != nil && a.effects.Busy())
	// 時間往前走一格；Game統一推進動畫、標題等待及回合演出。
	a.g.Tick(a.animationStep(time.Now(), a.g.State() == game.StateEngagement))
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
	a.updateInput()
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
	fb := a.frame()
	if a.shot != "" {
		var err error
		if a.hd != nil && !a.original {
			err = a.hd.WritePNG(a.shot, fb, a.g.Palette())
		} else {
			err = fb.WritePNG(a.shot, a.paletteSpec())
		}
		if err != nil {
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
	started := time.Now()
	fb := a.frame()
	frameTime := time.Since(started)
	if a.hd != nil && !a.original {
		frame := a.hd.Render(fb, a.g.Palette())
		if a.hdUploaded != a.hd.Generation() {
			a.hdScreen.WritePixels(frame.Pix)
			a.hdUploaded = a.hd.Generation()
		}
		screen.DrawImage(a.hdScreen, nil)
		a.drawToolbar(screen, fb.InterfaceTheme)
		a.traceFrame(fb, started, frameTime)
		return
	}
	// 調色盤每格問一次：開頭動畫那一段是自己的一組灰階。
	a.setPalette(a.g.Palette())
	for i, v := range fb.Pix {
		c := a.pal[v]
		a.buf[i*4], a.buf[i*4+1], a.buf[i*4+2], a.buf[i*4+3] = c.R, c.G, c.B, 255
	}
	a.screen.WritePixels(a.buf)
	options := &ebiten.DrawImageOptions{}
	if a.hd != nil {
		options.GeoM.Scale(presentation.Scale, presentation.Scale)
	}
	screen.DrawImage(a.screen, options)
	a.drawToolbar(screen, fb.InterfaceTheme)
	a.traceFrame(fb, started, frameTime)
}

// Only the display mode changes. Layout, mouse mapping and Game stay intact.
func (a *app) toggleGraphics() bool {
	if a.hd == nil {
		return false
	}
	a.original = !a.original
	return true
}

func (a *app) Layout(int, int) (int, int) {
	if a.hd != nil {
		return gfx.Width * presentation.Scale, gfx.Height * presentation.Scale
	}
	return gfx.Width, gfx.Height
}

func (a *app) gamePosition(x, y int) (int, int) {
	if a.hd == nil {
		return x, y
	}
	divide := func(v int) int {
		if v < 0 {
			return (v - presentation.Scale + 1) / presentation.Scale
		}
		return v / presentation.Scale
	}
	return divide(x), divide(y)
}

var buildVersion = "development"

func main() {
	showVersion := flag.Bool("version", false, "顯示封包版本後結束")
	orig := flag.String("orig", "org_game/extracted", "原版檔案目錄（玩家自備）")
	data := flag.String("data", "docs/internals/data", "畫面規格與資料目錄")
	scale := flag.Float64("scale", 1, "視窗放大倍率")
	graphics := flag.String("graphics", "original", "圖片顯示：original 或 hd")
	artDir := flag.String("art", "assets/hd", "HD 圖片目錄")
	fontStyle := flag.String("font", "classic", "HD字型：original原版點陣、classic原版強化、serif思源宋體、sans思源黑體")
	musicQuality := flag.String("music-quality", "original", "配樂音色：original 或 hd；HD曲庫放在music/hd")
	soundFont := flag.String("soundfont", "", "音色庫：gs 或 mt32；未指定沿玩家偏好")
	frameTrace := flag.String("frame-trace", "", "可選顯示性能JSONL收據；僅觀測，不改遊戲狀態")
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
	repairShips := flag.Bool("repair-ship-repairable", false,
		"可選規則：修護艦可被其他修護艦修理（原版不行；預設關閉，規格125）")
	musicDir := flag.String("music", "",
		"配樂與音效OGG目錄；空值停用音訊。\n"+
			"    玩家自備原版，配樂用 tools/music_to_ogg.sh 匯入；\n"+
			"    音效用 bash tools/sound_to_ogg.sh <新批次名稱> 匯入。\n"+
			"    曲庫及完整28段SFX_00.ogg至SFX_27.ogg放同一目錄；發行包不內嵌原版素材。")
	flag.Parse()
	if *showVersion {
		fmt.Println(buildVersion)
		return
	}
	explicit := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { explicit[f.Name] = true })
	if executable, err := os.Executable(); err == nil {
		packagedDefaults(executable, explicit, orig, data, artDir, musicDir, graphics)
	}
	fmt.Fprintln(os.Stderr, "啟動中：正在載入遊戲與圖片，請稍候。")

	load := func(report func(int, string)) (*app, error) {
		report(1, "載入原版資料")
		res, err := bootstrap.Load(*orig, *data)
		if err != nil {
			return nil, err
		}
		// 走腳本的時候不放動畫：腳本是照「開機就是主選單」寫的，
		// 前面多一張畫面整條序列就對不上收據。
		report(2, "載入開場動畫")
		if *intro && *script == "" {
			if err := bootstrap.LoadIntro(res, *orig); err != nil {
				return nil, err
			}
		} else {
			report(2, "略過開場動畫")
		}
		report(3, "載入存檔")
		if *saveDir != "" {
			if err := bootstrap.UseSaveDir(res, *saveDir); err != nil {
				return nil, err
			}
		}
		report(4, "準備初始局")
		g, err := game.New(res)
		if err != nil {
			return nil, err
		}
		if err := g.SetAILevel(*aiLevel); err != nil {
			return nil, err
		}
		g.SetRepairShipsRepairable(*repairShips)
		a := &app{
			g:         g,
			cursor:    res.Cursor,
			toolbar:   toolbar{font: res.Font},
			buf:       make([]byte, gfx.Width*gfx.Height*4),
			shot:      *shot,
			shotBin:   *shotBin,
			stateOut:  *stateOut,
			shotAfter: *shotAfter,
			script:    *script,
		}
		if *frameTrace != "" {
			a.frameTrace, err = os.OpenFile(*frameTrace, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
			if err != nil {
				return nil, err
			}
		}
		if *graphics != "original" && *graphics != "hd" {
			a.close()
			return nil, fmt.Errorf("圖片顯示模式無效：%s", *graphics)
		}
		a.original = *graphics == "original"
		report(5, "載入HD圖像與字型")
		_, artErr := os.Stat(filepath.Join(*artDir, "face"))
		if *graphics == "hd" || artErr == nil {
			a.hd, err = presentation.LoadWithProgress(*artDir, res.Faces, func(done, total int) {
				report(5, fmt.Sprintf("載入人物圖像 %d / %d", done, total))
			})
			if err != nil {
				if *graphics == "hd" {
					a.close()
					return nil, err
				}
				fmt.Fprintln(os.Stderr, "F6：", err)
			}
		}
		if a.hd != nil {
			if err := a.hd.SetFontStyle(*fontStyle); err != nil {
				a.close()
				return nil, err
			}
			res.Faces.ArtworkPrefix = "face"
			// 固定 Ebitengine 2.8.8 的最終縮放契約：HD 內部已完成圖片取樣，
			// 縮小視窗不再對細字與框線套第二次線性濾波。
			if a.hd.ThemeEnabled() {
				fmt.Fprintf(os.Stderr, "HD：%d/%d 個人物圖索引；A原配置介面與高解析文字\n", a.hd.Count(), len(res.Faces.Entries))
			} else {
				fmt.Fprintf(os.Stderr, "HD：%d/%d 個人物圖索引已載入；其他圖像沿原素材\n", a.hd.Count(), len(res.Faces.Entries))
			}
		}
		// 配樂：**要求了才開音訊**。缺檔就不播，但要說出來——安靜與
		// 「沒接播放層」長得一樣，不講的話沒有人分得出是哪一種
		// （AGENTS.md：缺檔就 skip，不用代用品）。
		report(6, "準備音樂與音效")
		if *musicDir != "" {
			lib := audio.NewLibrary(*musicDir)
			a.musicOriginal = lib
			hdLibrary := audio.NewHDLibrary(filepath.Join(*musicDir, "hd"))
			if len(hdLibrary.Missing()) == 0 {
				a.musicHD = hdLibrary
			}
			mt32Library := audio.NewHDLibrary(filepath.Join(*musicDir, "mt32"))
			if len(mt32Library.Missing()) == 0 {
				a.musicMT32 = mt32Library
			}
			sounds := audio.NewSoundLibrary(*musicDir)
			musicAvailable := lib.Have() > 0 || a.musicHD != nil || a.musicMT32 != nil
			var audioContext *eaudio.Context
			deviceAvailable := hasAudioDevice()
			if deviceAvailable && (musicAvailable || sounds.Have() == 28) {
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
				}
			} else {
				fmt.Fprintf(os.Stderr, "音效：%d/28段或裝置不可用，未啟用；來源轉檔見規格104\n", sounds.Have())
			}
			switch {
			case !deviceAvailable:
				fmt.Fprintln(os.Stderr, "配樂：這台機器沒有可用的音效裝置，"+
					"這一趟沒有音樂")
			case !musicAvailable:
				fmt.Fprintf(os.Stderr, "配樂：%s 裡一首都沒有，這一趟沒有音樂。"+
					"轉檔看 tools/music_to_ogg.sh\n", lib.Dir())
			default:
				if miss := lib.Missing(); len(miss) > 0 {
					fmt.Fprintf(os.Stderr, "配樂：%d/%d 首，缺 %v\n",
						lib.Have(), len(audio.Tracks), miss)
				}
				a.music = newMusicPlayer(audioContext, lib)
			}
		} else {
			report(6, "略過音樂與音效")
		}
		if *musicQuality != "original" && *musicQuality != "hd" {
			a.close()
			return nil, fmt.Errorf("配樂音色無效：%s", *musicQuality)
		}
		if *musicQuality == "hd" && !a.setHDMusic(true) {
			fmt.Fprintln(os.Stderr, "HD配樂不完整，保留原版曲庫。")
		}
		a.preferencesPath = filepath.Join(*saveDir, "preferences.json")
		options, optionsErr := readRulePreferences(a.preferencesPath)
		if optionsErr != nil {
			a.ruleSettingsError = optionsErr.Error()
			fmt.Fprintln(os.Stderr, "規則偏好：", optionsErr)
		} else {
			g.SetRuleOptions(options)
		}
		choice, prefErr := readSoundFontPreference(a.preferencesPath)
		if prefErr != nil {
			a.soundFontError = prefErr.Error()
			fmt.Fprintln(os.Stderr, "音色偏好：", prefErr)
		}
		if *soundFont != "" {
			choice = *soundFont
		}
		if choice != "" {
			if err := a.selectSoundFont(choice, false); err != nil {
				if *soundFont != "" {
					a.close()
					return nil, err
				}
				a.soundFontError = err.Error()
				fmt.Fprintln(os.Stderr, "音色偏好：", err)
			}
		}

		a.setPalette(g.Palette())
		report(7, "載入完成")
		return a, nil
	}

	var a *app
	var err error
	if *script != "" {
		a, err = load(func(int, string) {})
		if err != nil {
			die(err)
		}
		a.prepareWindow(*scale)
		err = ebiten.RunGame(a)
		a.close()
	} else {
		startup := newStartup(load, *scale)
		err = ebiten.RunGame(startup)
		a = startup.game
		startup.close()
	}
	if err != nil && err != ebiten.Termination && a != nil && a.music != nil &&
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
