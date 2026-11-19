// Package mobileshell 是 Android 版的殼（規格161）。
//
// 桌面殼 cmd/gin3 原地不動：它是 package main，又被大量文件與歷史收據引用，
// 整個搬走屬於大型重構。行動版版面本來就與桌面不同（分層、跳出視窗），
// 所以另起一個殼，直接用 internal/ 下的遊戲、載入與呈現。
//
// 殼放在 internal，不帶建置標籤，桌面也編得到、可以在桌面測試。被 ebitenmobile
// 綁定的 mobile 套件只露出給 Java 用的函式：gobind 會把綁定套件的每個匯出識別字
// 都轉成 Java，回傳兩個值的函式（例如 Layout）無法轉換。
package mobileshell

import (
	"fmt"
	"image/color"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/gfx"
)

// Config 是 Java 端在啟動時傳進來的目錄。
type Config struct {
	Orig string // 原版遊戲檔
	Data string // 畫面規格與資料（docs/internals/data 的發行複本）
	Save string // 存檔目錄（App 私有目錄）
}

// Shell 是 ebiten.Game。目錄還沒設定、或還在載入時畫載入畫面。
type Shell struct {
	mu      sync.Mutex
	cfg     *Config
	loading bool
	loadErr error
	g       *game.Game

	screen *ebiten.Image
	buf    []byte
	pal    [gfx.Colors]color.RGBA

	// back 由 Java 的返回手勢寫入（任何執行緒），Update 消費成一次右鍵。
	back int
	// taps 是 Java 送來的「手指放開」位置（0–1 的比例座標），Update 消費。
	// nativeTouch 為真時只走這條，不看 Ebitengine 的逐格觸控取樣：慢的機器上
	// 一次輕觸可能整個落在兩格之間，取樣看不到（模擬器上實測漏接）。
	taps        [][2]float64
	nativeTouch bool
	// lay 是上一次 Draw 的版面，觸控換算用；u 是畫布像素／邏輯像素。
	lay   screenLayout
	u     float64
	sharp sharp
	// drawn 是上一次真的畫出去的內容摘要；一樣就不重畫（省電，也省模擬器的軟體繪圖）。
	drawn uint64
	// basePal 是側欄用的調色盤：遊戲本體的，不跟開頭動畫的灰階變色。
	basePal    [gfx.Colors]color.RGBA
	basePalSet bool

	style     *sideStyle
	sideCache map[string]*ebiten.Image
	sidePal   [gfx.Colors]color.RGBA
}

// sideBand 是原版畫面左右兩條的底色（深藍，接近原版視窗外框的色系）。
var sideBand = color.RGBA{0, 0, 0x30, 0xff}

// NewShell 建一個還沒設定目錄的殼。畫面不每格清除：內容沒變的格直接沿用上一格。
func NewShell() *Shell {
	ebiten.SetScreenClearedEveryFrame(false)
	return &Shell{}
}

// Configure 設定目錄並在背景開始載入；重複呼叫只認第一次。
func (s *Shell) Configure(cfg Config) {
	s.mu.Lock()
	if s.cfg != nil {
		s.mu.Unlock()
		return
	}
	s.cfg, s.loading = &cfg, true
	s.mu.Unlock()
	go func() {
		start := time.Now()
		g, res, err := loadGame(cfg)
		// stderr 在 Android 上會進 logcat；載入失敗不能只是一片黑。
		if err != nil {
			log.Printf("logh3: 載入失敗：%v", err)
		} else {
			log.Printf("logh3: 載入完成 %v", time.Since(start).Round(time.Millisecond))
		}
		s.mu.Lock()
		s.g, s.loadErr, s.loading = g, err, false
		s.style = newSideStyle(res)
		s.mu.Unlock()
	}()
}

// UseNativeTouch 改由 QueueTap 收輕觸（Android 由 Java 送），不再看逐格取樣。
func (s *Shell) UseNativeTouch() {
	s.mu.Lock()
	s.nativeTouch = true
	s.mu.Unlock()
}

// QueueTap 排一次輕觸；nx、ny 是在遊戲畫面元件上的比例位置（0–1）。任何執行緒可呼叫。
func (s *Shell) QueueTap(nx, ny float64) {
	s.mu.Lock()
	s.taps = append(s.taps, [2]float64{nx, ny})
	s.mu.Unlock()
}

// Back 是 Android 返回手勢：排一次右鍵。
func (s *Shell) Back() {
	s.mu.Lock()
	s.back++
	s.mu.Unlock()
}

func loadGame(cfg Config) (*game.Game, *game.Resources, error) {
	res, err := bootstrap.Load(cfg.Orig, cfg.Data)
	if err != nil {
		return nil, nil, fmt.Errorf("載入原版資料：%w", err)
	}
	if err := bootstrap.LoadIntro(res, cfg.Orig); err != nil {
		return nil, nil, err
	}
	if cfg.Save != "" {
		if err := bootstrap.UseSaveDir(res, filepath.Clean(cfg.Save)); err != nil {
			return nil, nil, err
		}
	}
	g, err := game.New(res)
	return g, res, err
}

// Update 推進遊戲並處理觸控。
func (s *Shell) Update() error {
	s.mu.Lock()
	g, back, taps, native := s.g, s.back, s.taps, s.nativeTouch
	s.back, s.taps = 0, nil
	s.mu.Unlock()
	if g == nil {
		return nil
	}
	if g.Quit() {
		return ebiten.Termination
	}
	g.Tick(time.Second / time.Duration(ebiten.TPS()))
	for i := 0; i < back; i++ {
		backClick(g)
	}
	// M1 第一步：輕觸＝移過去再左鍵（觸控板模式與側欄在後續步驟）。
	for _, t := range taps {
		s.tap(g, t[0]*float64(s.lay.lw), t[1]*float64(gfx.Height))
	}
	if !native && s.u > 0 {
		for _, id := range inpututil.AppendJustReleasedTouchIDs(nil) {
			tx, ty := inpututil.TouchPositionInPreviousTick(id)
			s.tap(g, float64(tx)/s.u, float64(ty)/s.u)
		}
	}
	return nil
}

// TapLogical 在邏輯畫布座標送一次輕觸（桌面預覽與測試用；Android 走觸控事件）。
func (s *Shell) TapLogical(x, y int) {
	s.mu.Lock()
	g := s.g
	s.mu.Unlock()
	if g != nil {
		s.tap(g, float64(x), float64(y))
	}
}

// tap 是一次輕觸（邏輯座標）：側欄按鈕優先，其次是畫面上的原版區塊（換回原版座標後
// 移過去再左鍵）。
func (s *Shell) tap(g *game.Game, x, y float64) {
	if b, ok := sideButtonAt(s.buttons(g), int(x), int(y)); ok {
		runSide(g, b)
		return
	}
	if gx, gy, ok := s.lay.toGame(x, y); ok {
		g.Click(gx, gy)
	}
}

func (s *Shell) buttons(g *game.Game) []sideButton {
	var cmds []sideCmd
	if s.style != nil {
		cmds = s.style.cmds
	}
	return sideButtons(s.lay, stateView{copyProt: g.State() == game.StateCopyProtection, battle: g.State() == game.StateBattle}, cmds)
}

// Layout 回實際像素大小的畫布：外部尺寸（dp）乘裝置縮放。每個區塊在 Draw 裡
// 只縮放一次，點陣字不會先縮到邏輯畫布再被系統放大第二次。
func (s *Shell) Layout(outsideW, outsideH int) (int, int) {
	f := ebiten.Monitor().DeviceScaleFactor()
	if f <= 0 {
		f = 1
	}
	w, h := int(float64(outsideW)*f), int(float64(outsideH)*f)
	if w <= 0 || h <= 0 {
		return gfx.Width, gfx.Height
	}
	return w, h
}

// LogicalWidth 回畫布對應的邏輯寬：高 480，寬依比例但不小於 640。
func LogicalWidth(canvasW, canvasH int) int {
	if canvasW <= 0 || canvasH <= 0 {
		return gfx.Width
	}
	w := canvasW * gfx.Height / canvasH
	if w < gfx.Width {
		w = gfx.Width
	}
	return w
}

// Draw 依版面畫原版區塊與側欄。
func (s *Shell) Draw(screen *ebiten.Image) {
	cw, ch := screen.Bounds().Dx(), screen.Bounds().Dy()
	s.u = float64(ch) / float64(gfx.Height)
	lw := LogicalWidth(cw, ch)
	s.mu.Lock()
	g, err := s.g, s.loadErr
	s.mu.Unlock()
	if g == nil {
		screen.Fill(sideBand)
		// 載入中維持底色；失敗塗暗紅，錯誤內容在 logcat。
		if err != nil {
			screen.Fill(color.RGBA{0x60, 0, 0, 0xff})
		}
		s.lay = computeLayout(lw, false)
		return
	}
	if s.screen == nil {
		s.screen = ebiten.NewImage(gfx.Width, gfx.Height)
		s.buf = make([]byte, gfx.Width*gfx.Height*4)
	}
	p := g.Palette()
	for i := range s.pal {
		r, gg, b := p.RGB8(i)
		s.pal[i] = color.RGBA{r, gg, b, 255}
	}
	switch g.State() {
	case game.StateIntro, game.StateTitle, game.StateEnding:
	default:
		s.basePal, s.basePalSet = s.pal, true
	}
	// 觸控沒有游標：不畫原版的滑鼠指標（觸控板模式時再畫）。
	fb := g.FrameWithoutCursor()
	s.lay = computeLayout(lw, g.PlainStrategic())
	buttons := s.buttons(g)
	sum := frameSum(fb.Pix, s.pal, cw, ch, s.lay, buttons)
	if sum == s.drawn {
		return // 與上一格完全相同：畫面不每格清除，直接沿用
	}
	s.drawn = sum
	for i, v := range fb.Pix {
		c := s.pal[v]
		s.buf[i*4], s.buf[i*4+1], s.buf[i*4+2], s.buf[i*4+3] = c.R, c.G, c.B, 255
	}
	s.screen.WritePixels(s.buf)
	screen.Fill(sideBand)
	for _, r := range s.lay.regions {
		s.sharp.draw(screen, s.screen, r.src, r.x*s.u, r.y*s.u, r.scale*s.u)
	}
	s.drawSide(screen, buttons, s.sidePalette(), s.u)
}

// frameSum 是這一格要畫的內容摘要（FNV-1a）：原版色號、調色盤、畫布大小、版面與按鈕。
func frameSum(pix []uint8, pal [gfx.Colors]color.RGBA, cw, ch int, lay screenLayout, buttons []sideButton) uint64 {
	h := uint64(14695981039346656037)
	mix := func(b byte) { h ^= uint64(b); h *= 1099511628211 }
	for _, v := range pix {
		mix(v)
	}
	for _, c := range pal {
		mix(c.R)
		mix(c.G)
		mix(c.B)
	}
	for _, v := range []int{cw, ch, lay.lw, len(lay.regions), len(buttons)} {
		mix(byte(v))
		mix(byte(v >> 8))
	}
	if lay.strategic {
		mix(1)
	}
	for _, b := range buttons {
		mix(byte(b.x))
		mix(byte(b.y))
		mix(byte(b.act))
	}
	return h
}

// sidePalette 回側欄用的調色盤：看過遊戲本體就用它，否則用主選單的。
func (s *Shell) sidePalette() [gfx.Colors]color.RGBA {
	if s.basePalSet || s.style == nil {
		return s.basePal
	}
	var out [gfx.Colors]color.RGBA
	for i := range out {
		r, g, b := s.style.menuPal.RGB8(i)
		out[i] = color.RGBA{r, g, b, 255}
	}
	return out
}
