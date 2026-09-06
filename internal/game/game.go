// Package game 是玩家路徑：輸入進來、狀態往前走、畫面出去。
//
// 這一層**不碰渲染後端**。畫面永遠是 640×480 的色號陣列，所以同一份
// 邏輯可以無頭跑對拍，也可以接上任何顯示層。
package game

import (
	"fmt"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/screen"
	"github.com/wicanr2/logh3/internal/ui"
)

// State 是玩家走到哪一個畫面。
type State int

const (
	// StateMainMenu 是開機後的主選單（帝國／同盟／對戰／取進度）。
	StateMainMenu State = iota

	// StateCopyProtection 是手冊查表的防拷問答。
	StateCopyProtection

	// StateMainGame 是遊戲本體。
	StateMainGame

	// StateShipOrder 是點了主畫面熱區後開出來的船艦訂造報告。
	StateShipOrder

	// StateBuildPlan 是次月份船艦建造計劃（有預算分配的可操作介面）。
	StateBuildPlan
)

// Resources 是玩家自備的原版檔案解出來的東西。**這個套件不內嵌任何原版資料。**
type Resources struct {
	Font     *font.Font
	Faces    *asset.AMG
	Cursor   *ui.Cursor
	Menu     *screen.MenuSpec
	CopyProt *screen.CopyProtSpec
	MainGame  *screen.MainGameSpec
	ShipOrder *screen.MainGameSpec
	BuildPlan *screen.MainGameSpec
	StarMap  *asset.Image
	Admiral  []byte // ADMIRAL.DT*，防拷問答的答案在裡面
}

// AdmiralRecord 是 `ADMIRAL.DT*` 一筆記錄的長度。
const AdmiralRecord = 28

// occupyOffset 是「占領能力」在記錄裡的位移（`docs/internals/08`）。
const occupyOffset = 11

// Game 是一局遊戲的狀態。
type Game struct {
	res   *Resources
	state State

	// 主選單：滑鼠位置與反白到哪一列。
	mouseX, mouseY int
	highlight      int

	// 防拷問答：已經打了什麼、答案是多少。
	typed  []byte
	answer uint8
	tries  int
}

// New 從主選單開始。
func New(res *Resources) (*Game, error) {
	if res.Font == nil || res.Menu == nil || res.CopyProt == nil || res.MainGame == nil {
		return nil, fmt.Errorf("缺少必要的資源（字型與三個畫面的規格）")
	}
	g := &Game{res: res, highlight: -1, mouseX: -1, mouseY: -1}
	if err := g.pickQuestion(); err != nil {
		return nil, err
	}
	return g, nil
}

// pickQuestion 決定防拷問答的題目與答案。
//
// 題目就是規格裡那張照片的編號——原版問的是「照片上那位提督的占領能力」，
// 答案是 `ADMIRAL.DT*` 同一筆記錄的 +11（`docs/internals/08`）。
func (g *Game) pickQuestion() error {
	i := g.res.CopyProt.Face.Index
	off := i*AdmiralRecord + occupyOffset
	if off >= len(g.res.Admiral) {
		return fmt.Errorf("ADMIRAL 資料只有 %d bytes，取不到第 %d 筆", len(g.res.Admiral), i)
	}
	g.answer = g.res.Admiral[off]
	return nil
}

// State 回目前在哪個畫面。
func (g *Game) State() State { return g.state }

// Typed 回防拷問答已經打進去的字串（畫面上還沒畫，見下方的未確認事項）。
func (g *Game) Typed() string { return string(g.typed) }

// MoveMouse 移動游標。主選單會跟著換反白的那一列。
func (g *Game) MoveMouse(x, y int) {
	g.mouseX, g.mouseY = x, y
	if g.state == StateMainMenu {
		g.highlight = screen.RowAt(g.res.Menu, y)
	}
}

// Click 按一下左鍵。
//
// **主選單只從事件回呼收按下**（`docs/re/08`）——那是原版的實作細節，
// 在這裡的意義是：點擊要落在某一列上才算數，不是點到視窗內就好。
func (g *Game) Click(x, y int) {
	g.MoveMouse(x, y)
	switch g.state {
	case StateMainMenu:
		if g.highlight >= 0 {
			g.state = StateCopyProtection
			g.typed = nil
		}
	case StateMainGame:
		// 可點區域是熱區清單，不是「視窗內就好」（`docs/re/09`）。
		if g.res.MainGame.HitTest(x, y) == "ship-order" && g.res.ShipOrder != nil {
			g.state = StateShipOrder
		}
	case StateShipOrder:
		if g.res.ShipOrder.HitTest(x, y) == "build-plan" && g.res.BuildPlan != nil {
			g.state = StateBuildPlan
		}
	}
}

// Key 送一個按鍵。數字會累積，Enter 送出答案。
//
// 答對就進遊戲本體。**答錯會怎樣還沒逆向出來**（原版的行為未確認），
// 這裡只是清掉重來並記一次，不要把它當成原版規則。
func (g *Game) Key(c byte) {
	if g.state != StateCopyProtection {
		return
	}
	switch {
	case c >= '0' && c <= '9':
		if len(g.typed) < 5 {
			g.typed = append(g.typed, c)
		}
	case c == 8 && len(g.typed) > 0: // Backspace
		g.typed = g.typed[:len(g.typed)-1]
	case c == '\r' || c == '\n':
		var v int
		for _, d := range g.typed {
			v = v*10 + int(d-'0')
		}
		if len(g.typed) > 0 && v == int(g.answer) {
			g.state = StateMainGame
			return
		}
		g.tries++
		g.typed = nil
	}
}

// Frame 畫出目前的畫面。
func (g *Game) Frame() *gfx.Framebuffer {
	switch g.state {
	case StateCopyProtection:
		face, _ := g.res.Faces.Decode(g.res.CopyProt.Face.Index)
		fb, _ := screen.CopyProtection(g.res.CopyProt, g.res.Font, face, string(g.typed))
		return fb
	case StateMainGame:
		fb, _ := screen.MainGame(g.res.MainGame, g.res.Font, g.res.StarMap,
			g.res.Faces, g.res.Cursor)
		return fb
	case StateShipOrder:
		fb, _ := screen.MainGame(g.res.ShipOrder, g.res.Font, g.res.StarMap,
			g.res.Faces, g.res.Cursor)
		return fb
	case StateBuildPlan:
		fb, _ := screen.MainGame(g.res.BuildPlan, g.res.Font, g.res.StarMap,
			g.res.Faces, g.res.Cursor)
		return fb
	default:
		fb, _ := screen.MainMenuAt(g.res.Menu, g.res.Font, g.highlight,
			g.res.Cursor, g.mouseX, g.mouseY)
		return fb
	}
}

// Answer 回這一題的答案。給測試與除錯用——正式玩不會叫它。
func (g *Game) Answer() uint8 { return g.answer }

// Mouse 回目前的游標位置。
func (g *Game) Mouse() (int, int) { return g.mouseX, g.mouseY }
