// Package game 是玩家路徑：輸入進來、狀態往前走、畫面出去。
//
// 這一層**不碰渲染後端**。畫面永遠是 640×480 的色號陣列，所以同一份
// 邏輯可以無頭跑對拍，也可以接上任何顯示層。
package game

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
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

	// StateBudgetPlan 是次月份預算計劃（徵稅、軍事費、造船預算比例）。
	StateBudgetPlan

	// StateAdmiralList 是待命中提督資料（疊在建造計劃上的清單）。
	StateAdmiralList

	// StateEnding 是結局畫面（一張 `.VO5` CG 鋪滿畫布）。
	StateEnding

	// 以下是命令階段：月初報告被右鍵關掉之後露出來的戰略地圖介面
	// （`docs/re/16`）。命令列三格各拉出一張下拉選單，選單項目再開一張畫面。

	// StateCommandPhase 是戰略地圖介面本身。
	StateCommandPhase

	// StateCmdMenu／StateListMenu／StateSysMenu 是三張下拉選單開著的樣子。
	StateCmdMenu
	StateListMenu
	StateSysMenu

	// 戰略命令的四項。
	StateStandbyAdmirals
	StateInfo
	StateTaxPlan
	StateProductionPlan

	// 一覽表的四項。
	StateStarList
	StateFleetList
	StateAdmiralListView
	StateEnemyList

	// 系統指令的五項。第一項「實行」不是畫面，是推進日期的動作。
	StateExecute
	StateLoadGame
	StateSaveGame
	StateSettings
	StateQuitConfirm
)

// screenSpecName 是每個命令階段狀態對應的規格名（`docs/internals/data/`）。
// 「實行」不在裡面——它不開畫面。
var screenSpecName = map[State]string{
	StateCommandPhase:    "screen-command-phase",
	StateCmdMenu:         "screen-cmd-menu",
	StateListMenu:        "screen-list-menu",
	StateSysMenu:         "screen-sys-menu",
	StateStandbyAdmirals: "screen-standby-admirals",
	StateInfo:            "screen-info",
	StateTaxPlan:         "screen-tax-plan",
	StateProductionPlan:  "screen-production-plan",
	StateStarList:        "screen-star-list",
	StateFleetList:       "screen-fleet-list",
	StateAdmiralListView: "screen-admiral-list-view",
	StateEnemyList:       "screen-enemy-list",
	StateLoadGame:        "screen-load-game",
	StateSaveGame:        "screen-save-game",
	StateSettings:        "screen-settings",
	StateQuitConfirm:     "screen-quit-confirm",
}

// barRow 是命令列那一排在畫面上的 y（三格都一樣高，20 像素）。
const barRow, rowHeight = 120, 20

// commandBar 是命令列三格與它們拉出來的選單（`docs/internals/23`、
// `data/menu-hotspots.json`）。
//
// 選單項目與母項目**同一個 x**，寬度是該選單裡最長那個標籤的字數 × 16，
// 四（五）個項目共用；第 n 項的 y 是 `barRow + 20n`——所以**第 0 項與命令列
// 那一格重疊**。實測選單開著的時候是選單項目收下那一點。
var commandBar = []struct {
	X, W  int
	Menu  State
	ItemW int
	Items []State
}{
	{X: 352, W: 64, Menu: StateCmdMenu, ItemW: 64, Items: []State{
		StateStandbyAdmirals, StateInfo, StateTaxPlan, StateProductionPlan}},
	{X: 448, W: 64, Menu: StateListMenu, ItemW: 64, Items: []State{
		StateStarList, StateFleetList, StateAdmiralListView, StateEnemyList}},
	{X: 544, W: 64, Menu: StateSysMenu, ItemW: 48, Items: []State{
		StateExecute, StateLoadGame, StateSaveGame, StateSettings, StateQuitConfirm}},
}

// menuOf 回這個狀態是哪一格拉出來的選單；不是選單就回 -1。
func menuOf(s State) int {
	for i, b := range commandBar {
		if b.Menu == s {
			return i
		}
	}
	return -1
}

// Resources 是玩家自備的原版檔案解出來的東西。**這個套件不內嵌任何原版資料。**
type Resources struct {
	Font     *font.Font
	Faces    *asset.AMG
	Cursor   *ui.Cursor
	Menu     *screen.MenuSpec
	CopyProt *screen.CopyProtSpec
	MainGame  *screen.MainGameSpec
	ShipOrder *screen.MainGameSpec
	BuildPlan  *screen.MainGameSpec
	BudgetPlan  *screen.MainGameSpec
	AdmiralList *screen.MainGameSpec

	// Screens 是命令階段那一批畫面的規格，key 是規格名（見 screenSpecName）。
	Screens map[string]*screen.MainGameSpec

	StarMap  *asset.Image

	// Regions 是星域遮罩（`STRMAP.RGN`），用來問「地圖上這一點是哪個星域」。
	Regions []asset.Region

	// StarMaps 是「照某一組星域色重建的地圖」，key 是規格的 `map.colours`
	// 欄位名。空字串那一份就是 StarMap（勢力色）。情報報告用治安色、
	// 星系資料又是另一組，所以不能共用同一張。
	StarMaps map[string]*asset.Image
	Admiral  []byte // ADMIRAL.DT*，防拷問答的答案在裡面

	// RegionPanel 是左下角星域摘要那塊的版面（規格 `docs/spec/001`），
	// CapitalPanel 是右側那塊（`docs/spec/002`）。
	RegionPanel  *screen.RegionPanel
	CapitalPanel *screen.CapitalPanel

	// BudgetPanel 是預算計劃／徵稅計劃那個視窗的數字版面（`docs/spec/003`），
	// ProductionPanel 是生產計劃那個視窗的（`docs/spec/004`）。
	BudgetPanel     *screen.BudgetPanel
	ProductionPanel *screen.ProductionPanel

	// SlotList 是存檔畫面那 11 列的版面，SaveIndex 是 `SAVEDATA.IDX` 的內容
	// （`docs/spec/008`）。那 11 列是算出來的，不在畫面規格裡。
	SlotList  *screen.SlotList
	SaveIndex []rules.SaveIndexEntry

	// WorkTemplate 是新遊戲開局的 `WORK` 區塊（`docs/re/45`）。
	// 新遊戲要存檔就得靠它——`WORK.DTX` 不存在。
	WorkTemplate []byte

	// OrigDir 是原版檔案的目錄。「取進度」要從那裡讀存檔（`docs/spec/007`）。
	OrigDir string

	// Graph 是星域相鄰圖，UnrestParams 是載入時重算未安定度要的難度參數
	// （`docs/spec/006`）。兩者都給了才走得到「取進度」。
	Graph        *rules.StarGraph
	UnrestParams rules.UnrestParams

	// ShipPct 是切給船艦建造費的百分比（開局 25%）。
	ShipPct int

	// StartMilitary 是**開場那次月底結算之前**的軍事費。劇本給的起始值，
	// 從檔案推不出來；796 年 1 月帝國是 50000（監看軍事費看到
	// 50000 → 60741 → 45592，`docs/re/35`）。
	StartMilitary int

	// Stellar 是執行期的星域記錄（一筆 126 bytes），Planets 是 `PLADAT.DAT`。
	// 面板上的數字從這兩份算出來，不是量好畫死的。
	Stellar []byte
	Planets *rules.Planets

	// Growth 是行星成長率的階梯表（`docs/re/33`），月底結算要用。
	Growth *rules.GrowthTable

	// Faction 是玩家陣營：0 帝國、1 同盟。
	Faction int
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

	// pct 是船艦建造費的百分比分配。**建造計劃與生產計劃共用同一份**
	// ——在建造計劃上按箭頭改的，生產計劃那張跟著變（`docs/spec/004`）。
	// 開局是量到的那一組（`rules.OpeningPct`）。
	pct [rules.BuildPlanRows]int

	// plan 是建造計劃畫面當下那一份（建造費用月初報告那一組算）。
	plan *rules.BuildPlan

	// 結局：目前顯示哪一張 CG。
	ending *asset.Image

	// 命令階段：第幾個回合。遊戲只存這一個數，年月日是印的時候算出來的
	// （`docs/re/25`）。開局 0。
	turn int

	// 月初報告被關掉幾層。兩層都關掉才露出戰略地圖介面。
	reportLayer int

	// 手上的軍事費與下個月的預估徵稅額。開場那次結算算出來，之後每個
	// 月底再算一次。
	military, taxTotal int

	// treasury 是這個月的財政（原定軍事費、已支付各項），
	// monthReport 是月初報告那張上的預算。
	treasury    rules.Treasury
	monthReport rules.Budget

	// building 是**上一次結算撥款出去**的那份計劃：買到的船這個月還在造，
	// 下個月底交付（`docs/re/21`）。生產計劃畫面的「建造中」就是它。
	building rules.BuildPlan

	// 目前選中的星域。左下角那塊寫的就是它。**還沒選過是 −1**，
	// 那塊是空的——收據 `main1e` 就是這個狀態（關月初報告的那兩下右鍵
	// 也點在星圖上，但沒有選中任何星域）。
	selectedRegion int

	// slot 是目前這個局面從哪一格存檔來的（新遊戲是 nil）。
	// 存檔寫出要靠它——`WORK` 那 230 bytes 只解出三個欄位，
	// 其餘得原樣抄回去（`docs/spec/009`）。
	slot *rules.Slot

	// menuRow 是主選單點了哪一列（-1 ＝ 還沒點）。防拷答對之後要靠它
	// 決定是開新遊戲還是進存檔畫面。
	menuRow int
}

// 星圖畫在畫面的 (0,40)，320×300，圖塊 16×20（`docs/internals/12`）。
const (
	mapX, mapY = 0, 40
	mapW, mapH = 320, 300
	tileW, tileH = 16, 20
)

// ShowEnding 切到結局畫面，顯示 `.VO5` 解出來的那張圖。
//
// **結局的播放順序還沒逆向出來**（哪一張配哪一句、怎麼疊、
// 換場怎麼走都未解，見 `docs/internals/18`）；這裡只負責把一張畫出來。
func (g *Game) ShowEnding(img *asset.Image) {
	g.ending = img
	g.state = StateEnding
}

// New 從主選單開始。
func New(res *Resources) (*Game, error) {
	if res.Font == nil || res.Menu == nil || res.CopyProt == nil || res.MainGame == nil {
		return nil, fmt.Errorf("缺少必要的資源（字型與三個畫面的規格）")
	}
	g := &Game{res: res, highlight: -1, mouseX: -1, mouseY: -1, selectedRegion: -1,
		menuRow: -1, military: res.StartMilitary, pct: rules.OpeningPct}
	// 原版開機就跑一次月底結算，796 年 1 月的局面是那一次的結果
	// （`docs/re/33`、`docs/re/35`）。
	g.monthEnd()
	if err := g.pickQuestion(); err != nil {
		return nil, err
	}
	return g, nil
}

// LoadSlot 把一格存檔接上來（規格 `docs/spec/005`）。
//
// 存檔存的是**結算之後**的狀態：回合數、軍事費與總稅收都在 `WORK.DT?`
// 裡，所以載進來不再跑一次月底結算——跑了就會把那個月的稅收重複算一遍。
//
// 未安定度是唯一一欄載進來還不對的（原版載入時重算，算式未解，
// `docs/re/24`）。要對畫面的地方得自己 `Slot.SetUnrest` 蓋成量到的值。
func (g *Game) LoadSlot(sl *rules.Slot) {
	g.slot = sl
	g.res.Stellar = sl.Stellar
	g.res.Planets = sl.Planets
	g.turn = sl.Turn
	g.military = sl.Military
	g.taxTotal = sl.TaxTotal
	g.selectedRegion = -1
	g.state = StateCommandPhase
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
			g.menuRow = g.highlight
			g.state = StateCopyProtection
			g.typed = nil
		}
	case StateLoadGame:
		if n := SlotAt(y); n >= 0 {
			g.loadSlotFromDisk(n)
		}
	case StateMainGame:
		// 可點區域是熱區清單，不是「視窗內就好」（`docs/re/09`）。
		switch g.res.MainGame.HitTest(x, y) {
		case "ship-order":
			if g.res.ShipOrder != nil {
				g.state = StateShipOrder
			}
		case "budget-plan":
			if g.res.BudgetPlan != nil {
				g.state = StateBudgetPlan
			}
		}
	case StateShipOrder:
		if g.res.ShipOrder.HitTest(x, y) == "build-plan" && g.res.BuildPlan != nil {
			g.state = StateBuildPlan
			g.plan = &rules.BuildPlan{Cost: g.monthReport.ShipCost(), Pct: g.pct}
		}
	case StateBuildPlan:
		if g.plan == nil {
			return
		}
		act := g.res.BuildPlan.HitTest(x, y)
		var row int
		switch {
		case strings.HasPrefix(act, "inc-"):
			fmt.Sscanf(act, "inc-%d", &row)
			if g.plan.Inc(row) {
				g.pct = g.plan.Pct
			}
		case strings.HasPrefix(act, "dec-"):
			fmt.Sscanf(act, "dec-%d", &row)
			if g.plan.Dec(row) {
				g.pct = g.plan.Pct
			}
		case act == "admiral-list":
			if g.res.AdmiralList != nil {
				g.state = StateAdmiralList
			}
		}
	case StateCommandPhase:
		for _, b := range commandBar {
			if x >= b.X && x < b.X+b.W && y >= barRow && y < barRow+rowHeight {
				if g.commandSpec(b.Menu) != nil {
					g.state = b.Menu
				}
				return
			}
		}
	case StateCmdMenu, StateListMenu, StateSysMenu:
		b := commandBar[menuOf(g.state)]
		for n, item := range b.Items {
			top := barRow + n*rowHeight
			if x >= b.X && x < b.X+b.ItemW && y >= top && y < top+rowHeight {
				if item == StateExecute {
					// 一次「實行」推進一個回合 ＝ 三天（`docs/re/20`、`docs/re/25`）。
					g.turn++
					// 一個月十回合；跨過去就結算（`docs/re/26`、`docs/re/33`）。
					if g.turn%rules.TurnsPerMonth == 0 {
						g.monthEnd()
					}
					g.state = StateCommandPhase
					return
				}
				if g.commandSpec(item) != nil {
					g.state = item
				}
				return
			}
		}
		// 點在選單項目外面就把選單關掉——實測四個方向的界外點都回主畫面。
		g.state = StateCommandPhase
	}
}

// RightClick 按一下右鍵。
//
// 右鍵在這個遊戲裡是「關掉最上面那一層」（`docs/re/16`）：月初報告要按兩下
// 才會露出完整的戰略地圖介面，命令階段開出來的畫面按一下就回到地圖介面。
//
// **座標目前不影響結果**——實測 (300,240) 那一點在報告視窗上與在地圖上
// 都關得掉；哪裡按不算數還沒量過。
func (g *Game) RightClick(x, y int) {
	g.MoveMouse(x, y)
	// 按在星圖上會順便換掉選中的星域，左下角那塊跟著改寫（收據 `rcback`）。
	//
	// **只有已經在戰略地圖介面時才算**：關月初報告的那兩下右鍵也點在
	// 星圖上（`main1e` 的 `-rclicks 300,200` 兩次），而那張畫面的左下角
	// 是空的——所以關報告的右鍵不選星域。
	if g.state == StateCommandPhase && g.res.Regions != nil &&
		x >= mapX && x < mapX+mapW && y >= mapY && y < mapY+mapH {
		if i := asset.RegionAt(g.res.Regions, x-mapX, y-mapY, tileW, tileH); i >= 0 {
			g.selectedRegion = i
		}
	}
	switch g.state {
	case StateMainGame:
		// 月初報告是兩層。第一下關掉上面那層，第二下才到地圖介面。
		g.reportLayer++
		if g.reportLayer >= 2 && g.commandSpec(StateCommandPhase) != nil {
			g.state = StateCommandPhase
		}
	case StateCmdMenu, StateListMenu, StateSysMenu:
		g.state = StateCommandPhase
	default:
		if _, ok := screenSpecName[g.state]; ok && g.state != StateCommandPhase {
			g.state = StateCommandPhase
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
			// 主選單前三列開新遊戲，第四列「取進度」先進存檔畫面
			// （`docs/re/38`、`docs/spec/007`）。
			if g.menuRow == menuRowLoad && g.commandSpec(StateLoadGame) != nil {
				g.state = StateLoadGame
			} else {
				g.state = StateMainGame
			}
			return
		}
		g.tries++
		g.typed = nil
	}
}

// menuRowLoad 是主選單第四列「取進度」（`docs/internals/data/screen-main-menu.json`）。
const menuRowLoad = 3

// 存檔畫面那 11 列的版面（`docs/spec/007`）。三個有內容的槽位量到的
// 落點是 y ＝ 242／262／282（收據 `slot242`／`s2empire`／`s3rgn`），
// 一列 20 px，往上推到第 0 格就是 82。
const (
	slotRowTop    = 82
	slotRowHeight = 20
)

// SlotAt 回存檔畫面上這個 y 落在第幾格；不在任何一列上回 -1。
func SlotAt(y int) int {
	if y < slotRowTop {
		return -1
	}
	n := (y - slotRowTop) / slotRowHeight
	if n >= rules.SaveSlots {
		return -1
	}
	return n
}

// loadSlotFromDisk 讀一格存檔接上遊戲（`docs/spec/005`、`docs/spec/007`）。
//
// 未安定度載入時要重算（`docs/spec/006`），所以讀完就跑一次。
// 缺檔或缺參數就什麼都不做——存檔畫面上那些空的格子點下去也是這樣。
func (g *Game) loadSlotFromDisk(n int) {
	if g.res.OrigDir == "" || g.res.Graph == nil {
		return
	}
	ext := rules.SlotExt(n)
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(g.res.OrigDir, name))
		if err != nil {
			return nil
		}
		return b
	}
	f := rules.SlotFiles{
		Stellar: read("STELLAR.DT" + ext), Planet: read("PLANET.DT" + ext),
		Admiral: read("ADMIRAL.DT" + ext), Fleet: read("FLEET.DT" + ext),
		Ship: read("SHIP.DT" + ext), Work: read("WORK.DT" + ext),
		PlanetHeader: read("PLADAT.DAT"),
	}
	sl, err := rules.LoadSlot(f)
	if err != nil {
		return
	}
	// 算式要的回合數就是這一格存檔的（`ds:8F18`，`docs/re/43`）。
	params := g.res.UnrestParams
	params.Turn = sl.Turn
	sl.RecomputeUnrest(g.res.Graph, params)
	g.res.Admiral = sl.Admiral
	g.LoadSlot(sl)
}

// SaveSlot 把目前的局面寫回一格存檔（規格 `docs/spec/009`）。
//
// **只有從存檔進來的局面存得了**：`WORK.DT?` 那 230 bytes 只解出三個欄位，
// 新遊戲沒有可以照抄的底稿。`dir` 由呼叫端指定，原版目錄一律唯讀。
func (g *Game) SaveSlot(dir string, index int) error {
	if g.slot == nil {
		// 新遊戲：拿量到的開局 `WORK` 當底稿，其餘的表就是目前手上的。
		// remake 沒有模型的那些欄位會停在 796 年 1 月的初值
		// （`docs/spec/009`）。
		if len(g.res.WorkTemplate) < rules.WorkRecord || g.res.Planets == nil {
			return fmt.Errorf("新遊戲要存檔得先給 WorkTemplate（`docs/spec/009`）")
		}
		g.slot = &rules.Slot{
			Stellar: g.res.Stellar, Planets: g.res.Planets,
			Admiral: g.res.Admiral,
			Fleet:   make([]byte, rules.FleetCount*rules.FleetRecord),
			Ship:    make([]byte, rules.ShipCount*rules.ShipRecord),
			Work:    append([]byte(nil), g.res.WorkTemplate...),
		}
		for i := range g.slot.Fleet {
			g.slot.Fleet[i] = 0xFF
		}
		for i := range g.slot.Ship {
			g.slot.Ship[i] = 0xFF
		}
	}
	g.slot.Turn = g.turn
	g.slot.Military = g.military
	g.slot.TaxTotal = g.taxTotal
	if err := g.slot.Save(dir, index); err != nil {
		return err
	}
	if len(g.res.SaveIndex) == 0 {
		return nil
	}
	b, err := os.ReadFile(filepath.Join(g.res.OrigDir, "SAVEDATA.IDX"))
	if err != nil {
		return err
	}
	name := g.res.SaveIndex[index].Name
	if err := rules.WriteSaveIndex(b, index, g.turn, name); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "SAVEDATA.IDX"), b, 0o644)
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
	case StateEnding:
		return screen.Ending(g.ending)
	case StateAdmiralList:
		fb, _ := screen.MainGame(g.res.AdmiralList, g.res.Font, g.res.StarMap,
			g.res.Faces, g.res.Cursor)
		return fb
	case StateBudgetPlan:
		fb, _ := screen.MainGame(g.res.BudgetPlan, g.res.Font, g.res.StarMap,
			g.res.Faces, g.res.Cursor)
		g.drawBudget(fb)
		return fb
	case StateBuildPlan:
		fb, _ := screen.MainGame(g.res.BuildPlan, g.res.Font, g.res.StarMap,
			g.res.Faces, g.res.Cursor)
		if g.plan != nil {
			screen.DrawBuildPlan(g.res.BuildPlan, fb, g.res.Font, g.plan)
		}
		return fb
	default:
		if spec := g.commandSpec(g.state); spec != nil {
			// 從狀態算出來的那幾塊要畫在底板之後、疊上去的視窗之前
			// （`docs/spec/002`）。
			fb, _ := screen.MainGameWithBoard(spec, g.res.Font, g.starMapFor(spec),
				g.res.Faces, g.res.Cursor, g.drawBoardPanels)
			g.drawPanels(fb)
			g.drawBudget(fb)
			g.drawProduction(fb)
			g.drawSlotList(fb)
			return fb
		}
		fb, _ := screen.MainMenuAt(g.res.Menu, g.res.Font, g.highlight,
			g.res.Cursor, g.mouseX, g.mouseY)
		return fb
	}
}

// monthEnd 跑月底結算：逐行星成長與收稅，軍事費 += 總稅收。
//
// **還沒做的**：船艦交付與扣款（`docs/re/21`）、未安定度與成長率的重算
// （`docs/re/33` 只解出骨架）。
func (g *Game) monthEnd() {
	if g.res.Growth == nil || g.res.Planets == nil || len(g.res.Stellar) == 0 {
		return
	}
	before := g.military
	// 1. 逐星域結算，把自己那一份的稅加進軍事費——加完就是「原定軍事費」。
	collected := rules.MonthEnd(g.res.Stellar, g.res.Planets, g.res.Growth, g.res.Faction)
	budget := before + collected
	// 2. 照百分比切給造船，各列只買得起整數艘，零頭退回；扣掉實際花掉的。
	plan := rules.BuildPlan{Cost: budget * g.res.ShipPct / 100, Pct: g.pct}
	paid := plan.Paid()
	g.building = plan
	g.military = budget - paid
	g.treasury = rules.Treasury{Budget: budget, ShipPaid: paid}
	g.monthReport = rules.Budget{Treasury: before, Tax: collected, ShipPct: g.res.ShipPct}
	// 3. 結算完之後重算下個月的預估徵稅額——徵稅計劃那張印的是這個，
	//    與剛剛實收的那筆不是同一個數（`docs/spec/003`）。
	g.taxTotal = g.estimateTax()
}

// estimateTax 回下個月的預估徵稅額 ＝ 自己擁有的星域的徵稅額相加。
//
// 擁有者要**照行星現算**（`docs/re/17`）——星域記錄裡那一格是檔案裡的
// 原值，載入時會被重算，直接拿來用會多算十個星域。
func (g *Game) estimateTax() int {
	sum := 0
	for i := 0; (i+1)*rules.RegionRecord <= len(g.res.Stellar); i++ {
		if g.res.Planets.RegionOwner(i) != g.res.Faction {
			continue
		}
		sum += rules.RegionTax(g.res.Stellar[i*rules.RegionRecord : (i+1)*rules.RegionRecord])
	}
	return sum
}

// drawBoardPanels 畫底板上那幾塊算出來的內容。命令階段那一批畫面共用
// 同一塊底板，右側面板在每一張上都看得到（被疊上去的視窗蓋住的部分除外）。
func (g *Game) drawBoardPanels(fb *gfx.Framebuffer) {
	if _, ok := screenSpecName[g.state]; !ok {
		return
	}
	g.drawCapitalPanel(fb)
}

// drawSlotList 畫存檔畫面那 11 列（`docs/spec/008`）。
func (g *Game) drawSlotList(fb *gfx.Framebuffer) {
	if g.res.SlotList == nil || len(g.res.SaveIndex) == 0 {
		return
	}
	// **只有「取進度」那張**。「存進度」那張的版面不一樣（量到的只有 7 行、
	// 位置也不同），還沒解，先照規格裡量到的畫。
	if g.state == StateLoadGame {
		g.res.SlotList.Draw(fb, g.res.Font, g.res.SaveIndex, rules.Epoch)
	}
}

// drawBudget 畫預算計劃／徵稅計劃的數字（`docs/spec/003`）。
func (g *Game) drawBudget(fb *gfx.Framebuffer) {
	if g.res.BudgetPanel == nil {
		return
	}
	switch g.state {
	case StateTaxPlan:
		b := rules.Budget{Treasury: g.military, Tax: g.taxTotal, ShipPct: g.res.ShipPct}
		g.res.BudgetPanel.Draw(fb, g.res.Font, g.treasury, b)
	case StateBudgetPlan:
		// 月初報告那一張：本月份那一欄還沒有數字（全 0）。
		g.res.BudgetPanel.Draw(fb, g.res.Font, rules.Treasury{}, g.monthReport)
	}
}

// nextPlan 是次月份的船艦建造計劃：預估軍事費照百分比切出來的那一份。
func (g *Game) nextPlan() rules.BuildPlan {
	b := rules.Budget{Treasury: g.military, Tax: g.taxTotal, ShipPct: g.res.ShipPct}
	return rules.BuildPlan{Cost: b.ShipCost(), Pct: g.pct}
}

// drawProduction 畫生產計劃的數字（`docs/spec/004`）。
func (g *Game) drawProduction(fb *gfx.Framebuffer) {
	if g.state != StateProductionPlan || g.res.ProductionPanel == nil {
		return
	}
	g.res.ProductionPanel.Draw(fb, g.res.Font, g.nextPlan(), g.building)
}

// drawPanels 把兩塊「從狀態算出來」的面板畫上去：右側（規格 002）
// 與左下角（規格 001）。**只有戰略地圖介面畫**——疊在上面的那幾張
// 會把這兩塊改寫成自己的內容，它們的規格裡已經有了。
func (g *Game) drawPanels(fb *gfx.Framebuffer) {
	if g.state != StateCommandPhase {
		return
	}
	g.drawRegionPanel(fb)
}

// drawCapitalPanel 畫右側那塊：日期、軍事費、總稅收與自己首都的摘要。
func (g *Game) drawCapitalPanel(fb *gfx.Framebuffer) {
	if g.res.CapitalPanel == nil || g.res.RegionPanel == nil || g.res.Planets == nil {
		return
	}
	r := rules.CapitalRegion(g.res.Faction)
	o := r * rules.RegionRecord
	if o+rules.RegionRecord > len(g.res.Stellar) {
		return
	}
	c := rules.SummariseCapital(g.res.Stellar[o:o+rules.RegionRecord], g.res.Planets,
		r, g.res.Faction, g.turn, rules.Epoch, g.military, g.taxTotal)
	g.res.CapitalPanel.Draw(fb, g.res.Font, g.res.RegionPanel, c, g.res.Faction)
}

// drawRegionPanel 把選中星域的摘要畫進左下角那塊。
//
// **只有戰略地圖介面畫**：疊在上面的那幾張（一覽表、出擊命令…）會把
// 左下角改寫成自己的內容，它們的規格裡已經有了。
func (g *Game) drawRegionPanel(fb *gfx.Framebuffer) {
	if g.state != StateCommandPhase || g.selectedRegion < 0 ||
		g.res.RegionPanel == nil || g.res.Planets == nil {
		return
	}
	o := g.selectedRegion * rules.RegionRecord
	if o+rules.RegionRecord > len(g.res.Stellar) {
		return
	}
	sum := rules.Summarise(g.res.Stellar[o:o+rules.RegionRecord], g.res.Planets,
		g.selectedRegion, g.res.Faction)
	g.res.RegionPanel.Draw(fb, g.res.Font, sum, g.res.Faction)
}

// commandSpec 回這個命令階段狀態的規格；沒有就回 nil。
func (g *Game) commandSpec(s State) *screen.MainGameSpec {
	name, ok := screenSpecName[s]
	if !ok || g.res.Screens == nil {
		return nil
	}
	return g.res.Screens[name]
}

// starMapFor 回這張規格要用的地圖。**不能共用同一張**：情報報告的地圖
// 跟著治安換色、星系資料又是另一組（`strmap-regions.json` 的 colour_fields）。
func (g *Game) starMapFor(spec *screen.MainGameSpec) *asset.Image {
	if spec.Map.Colours != "" && g.res.StarMaps != nil {
		if m := g.res.StarMaps[spec.Map.Colours]; m != nil {
			return m
		}
	}
	return g.res.StarMap
}

// Turn 回目前是第幾個回合（開局 0）。
func (g *Game) Turn() int { return g.turn }

// Military 回手上的軍事費，TaxTotal 回這個月的總稅收。
func (g *Game) Military() int { return g.military }

// Treasury 回這個月的財政（原定軍事費、已支付各項）。
func (g *Game) Treasury() rules.Treasury { return g.treasury }
func (g *Game) TaxTotal() int { return g.taxTotal }

// Date 回遊戲內的日期。紀元年 796 是原版帝國戰役的開局。
func (g *Game) Date() (year, month, day int) { return rules.Date(g.turn, 796) }

// SelectedRegion 回目前選中的星域編號（`STELLAR.DT*` 的筆序）。
//
// **左下角那塊的內容還沒接上**：星系名與陣營從 `STELLAR.DT*` 讀得到，
// 但治安、行星數、艦隊數那幾格的來源還沒解出來（`docs/re/22`），
// 所以畫面仍然照規格畫死的那一份。
func (g *Game) SelectedRegion() int { return g.selectedRegion }

// Plan 回目前的建造計劃預算分配；不在那個畫面就回 nil。
func (g *Game) Plan() *rules.BuildPlan { return g.plan }

// NextPlan 是次月份的船艦建造計劃（生產計劃那張印的就是它）。
func (g *Game) NextPlan() rules.BuildPlan { return g.nextPlan() }

// Answer 回這一題的答案。給測試與除錯用——正式玩不會叫它。
func (g *Game) Answer() uint8 { return g.answer }

// Mouse 回目前的游標位置。
func (g *Game) Mouse() (int, int) { return g.mouseX, g.mouseY }
