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
	// lastW 是上一次 Draw 的邏輯畫布寬，觸控座標換算用。
	lastW int

	style     *sideStyle
	sideCache map[string]*ebiten.Image
	sidePal   [gfx.Colors]color.RGBA
}

// sideBand 是原版畫面左右兩條的底色（深藍，接近原版視窗外框的色系）。
var sideBand = color.RGBA{0, 0, 0x30, 0xff}

// NewShell 建一個還沒設定目錄的殼。
func NewShell() *Shell { return &Shell{} }

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
		x, y := g.MousePosition()
		g.RightClick(x, y)
	}
	// M1 第一步：輕觸＝移過去再左鍵（觸控板模式與側欄在後續步驟）。
	for _, t := range taps {
		s.tap(g, int(t[0]*float64(s.lastW)), int(t[1]*float64(gfx.Height)))
	}
	if !native {
		for _, id := range inpututil.AppendJustReleasedTouchIDs(nil) {
			tx, ty := inpututil.TouchPositionInPreviousTick(id)
			s.tap(g, tx, ty)
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
		s.tap(g, x, y)
	}
}

// tap 是一次輕觸：側欄按鈕優先，其次是原版畫面（移過去再左鍵）。
func (s *Shell) tap(g *game.Game, x, y int) {
	if b, ok := sideButtonAt(sideButtons(s.lastW, g.State()), x, y); ok {
		runSide(g, b)
		return
	}
	if gx, gy, ok := s.gamePoint(x, y); ok {
		g.Click(gx, gy)
	}
}

// Layout 用固定的邏輯畫布高度 480，寬度隨螢幕比例，原版畫面置中。
func (s *Shell) Layout(outsideW, outsideH int) (int, int) {
	return LogicalSize(outsideW, outsideH)
}

// LogicalSize 回邏輯畫布：高 480，寬依螢幕比例但不小於 640。
func LogicalSize(outsideW, outsideH int) (int, int) {
	if outsideW <= 0 || outsideH <= 0 {
		return gfx.Width, gfx.Height
	}
	w := outsideW * gfx.Height / outsideH
	if w < gfx.Width {
		w = gfx.Width
	}
	return w, gfx.Height
}

// gameOrigin 回原版畫面在邏輯畫布上的左上角（水平置中）。
func gameOrigin(logicalW int) int { return (logicalW - gfx.Width) / 2 }

// gamePoint 把邏輯畫布座標換成原版 640×480 座標；落在畫面外回 ok=false。
func (s *Shell) gamePoint(x, y int) (int, int, bool) {
	gx := x - gameOrigin(s.lastW)
	if gx < 0 || gx >= gfx.Width || y < 0 || y >= gfx.Height {
		return 0, 0, false
	}
	return gx, y, true
}

// Draw 把原版畫面畫在中央。
func (s *Shell) Draw(screen *ebiten.Image) {
	s.lastW = screen.Bounds().Dx()
	s.mu.Lock()
	g, loading, err := s.g, s.loading, s.loadErr
	s.mu.Unlock()
	screen.Fill(sideBand) // 左右兩條的底色；側欄控制之後畫在這裡
	if g == nil {
		// 載入中維持黑畫面；失敗塗暗紅，錯誤內容在 logcat（原版字型的訊息畫面在後續步驟）。
		_ = loading
		if err != nil {
			screen.Fill(color.RGBA{0x60, 0, 0, 0xff})
		}
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
	fb := g.Frame()
	for i, v := range fb.Pix {
		c := s.pal[v]
		s.buf[i*4], s.buf[i*4+1], s.buf[i*4+2], s.buf[i*4+3] = c.R, c.G, c.B, 255
	}
	s.screen.WritePixels(s.buf)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(gameOrigin(s.lastW)), 0)
	screen.DrawImage(s.screen, op)
	s.drawSide(screen, sideButtons(s.lastW, g.State()), s.pal)
}
