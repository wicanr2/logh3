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
	"time"

	"github.com/wicanr2/logh3/internal/ai"
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

	// StateAdmiralDetail 是提督的詳細資料：從「出擊命令」的待命中提督
	// 清單點一位提督開出來，下面有經歷／中止／出擊三顆按鈕
	// （`docs/re/65`）。
	StateAdmiralDetail

	// StateAdmiralBio 是提督經歷：詳細資料按「經歷」開出來的人物傳
	// （`docs/internals/15`、`docs/re/65`）。
	StateAdmiralBio

	// StateFleetOrg 是艦隊編成：詳細資料按「出擊」開出來的那一張，
	// 艦隊能力是提督與副官逐項取大（`docs/re/66`、`docs/spec/013`）。
	StateFleetOrg

	// StateFleetNumber 是艦隊編成按「艦隊編號」開出的十六列清單，
	// StateFleetAdjutant 是按「副官」開出的「變更／卸任」兩列小選單
	// （`docs/re/67`）。兩張都是滑鼠停哪一列、哪一列就反白。
	StateFleetNumber
	StateFleetAdjutant

	// StateFleetComposition 是艦隊編成按「確認」開出的**艦隊組成**：
	// 9 × 9 的格盤 ＋ 七列待命部隊與數量 ＋ 陣形／中止／決定
	// （`docs/re/67`、`docs/spec/015`）。
	StateFleetComposition

	// StateFormation 是艦隊組成按「陣形」開出的八列清單：圓錐／紡錘／
	// 半月／Ｕ字／Ｍ字／Ｔ字／分散／密集（`docs/re/70`）。
	StateFormation

	// StateBattle 是戰術畫面：面板 ＋ 星域的戰場地形 ＋ 格線
	// （`docs/internals/34`）。原版在輸入階段不畫單位（`docs/re/82`），
	// 所以這張目前也只有地形。
	StateBattle

	// StateShipDelivery 是月底的「船艦接收報告」：第十次「實行」跨月時
	// 自己跳出來，按「確認」才回到地圖介面（`docs/re/21`、`docs/re/55`）。
	StateShipDelivery

	// StateRegionFlipReport 是星域易主的戰略報告：回合結束逐格重算擁有者
	// 時每有一格易主就跳一則，**擋住後面的路徑表重建與態勢重算**，
	// 滑鼠任一鍵可關、鍵盤不行（`docs/spec/034`、`docs/re/246`）。
	// 畫在命令階段那張底上，所以共用 `screen-command-phase`。
	StateRegionFlipReport

	// StateVersusInput 是對戰模式的交棒視窗：「帝國軍／同盟軍的戰略輸入」
	// 疊在戰略地圖上，按「確認」開始這一方的回合
	// （`docs/re/76`、`docs/spec/016`）。
	StateVersusInput

	// StateVersusAccount 是對戰模式的月初會計報告：內容與單人那張
	// （`StateMainGame`）相同，但底板已經畫了陣營那一行、三顆命令按鈕
	// 與分頁鈕（收據 `vs4`，`docs/spec/016`）。
	StateVersusAccount

	// StateIntro 是開頭動畫（`OPEN.EXE` 的 374 格，`docs/spec/020`）。
	// 排在最後是為了不動既有狀態的編號；玩家路徑上它在主選單**之前**。
	StateIntro

	// StateVersusShipOrder 是對戰模式月初報告的第二張：船艦訂造報告。
	// 內容與單人那張（`StateShipOrder`）相同，底板是交棒之後那一張
	// （收據 `vs7`，`docs/re/260`）。同樣排在最後不動既有編號。
	StateVersusShipOrder
)

// String 回穩定的狀態名稱，供日誌與 GUI 黑箱收據使用。狀態值仍保留給程式
// 比較；收據不用只記一個容易因日後追加狀態而看不懂的整數。
func (s State) String() string {
	names := [...]string{
		"StateMainMenu", "StateCopyProtection", "StateMainGame",
		"StateShipOrder", "StateBuildPlan", "StateBudgetPlan",
		"StateAdmiralList", "StateEnding", "StateCommandPhase",
		"StateCmdMenu", "StateListMenu", "StateSysMenu",
		"StateStandbyAdmirals", "StateInfo", "StateTaxPlan",
		"StateProductionPlan", "StateStarList", "StateFleetList",
		"StateAdmiralListView", "StateEnemyList", "StateExecute",
		"StateLoadGame", "StateSaveGame", "StateSettings",
		"StateQuitConfirm", "StateAdmiralDetail", "StateAdmiralBio",
		"StateFleetOrg", "StateFleetNumber", "StateFleetAdjutant",
		"StateFleetComposition", "StateFormation", "StateBattle",
		"StateShipDelivery", "StateRegionFlipReport", "StateVersusInput",
		"StateVersusAccount", "StateIntro", "StateVersusShipOrder",
	}
	if int(s) >= 0 && int(s) < len(names) {
		return names[s]
	}
	return fmt.Sprintf("State(%d)", s)
}

// screenSpecName 是每個命令階段狀態對應的規格名（`docs/internals/data/`）。
// 「實行」不在裡面——它不開畫面。
var screenSpecName = map[State]string{
	StateCommandPhase:     "screen-command-phase",
	StateCmdMenu:          "screen-cmd-menu",
	StateListMenu:         "screen-list-menu",
	StateSysMenu:          "screen-sys-menu",
	StateStandbyAdmirals:  "screen-standby-admirals",
	StateInfo:             "screen-info",
	StateTaxPlan:          "screen-tax-plan",
	StateProductionPlan:   "screen-production-plan",
	StateStarList:         "screen-star-list",
	StateFleetList:        "screen-fleet-list",
	StateAdmiralListView:  "screen-admiral-list-view",
	StateEnemyList:        "screen-enemy-list",
	StateLoadGame:         "screen-load-game",
	StateSaveGame:         "screen-save-game",
	StateSettings:         "screen-settings",
	StateQuitConfirm:      "screen-quit-confirm",
	StateShipDelivery:     "screen-ship-delivery",
	StateRegionFlipReport: "screen-command-phase",
	StateAdmiralDetail:    "screen-admiral-detail",
	StateAdmiralBio:       "screen-admiral-bio",
	StateFleetOrg:         "screen-fleet-org",
	StateFleetNumber:      "screen-fleet-number",
	StateFleetAdjutant:    "screen-fleet-adjutant",
	StateFleetComposition: "screen-fleet-composition",
	StateFormation:        "screen-formation",
	StateVersusInput:      "screen-versus-input",
	StateVersusAccount:    "screen-versus-account",
	StateVersusShipOrder:  "screen-versus-ship-order",
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
	Font        *font.Font
	Faces       *asset.AMG
	Cursor      *ui.Cursor
	Menu        *screen.MenuSpec
	CopyProt    *screen.CopyProtSpec
	MainGame    *screen.MainGameSpec
	ShipOrder   *screen.MainGameSpec
	BuildPlan   *screen.MainGameSpec
	BudgetPlan  *screen.MainGameSpec
	AdmiralList *screen.MainGameSpec

	// Screens 是命令階段那一批畫面的規格，key 是規格名（見 screenSpecName）。
	Screens map[string]*screen.MainGameSpec

	StarMap *asset.Image

	// StarBit／StarRGN 是星圖的原始位元組（`STRMAP.BIT`／`STRMAP.RGN`），
	// RegionColours 是 `strmap-regions.json` 的 `colour_796_01`。
	// 留著是因為**首都那格的顏色跟著玩家的陣營走**，開局時要照陣營重建
	// 一次勢力色地圖（見 `factionStarMap`）。
	StarBit, StarRGN []byte
	RegionColours    []uint8

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
	// DeliveryPanel 是月底船艦接收報告那個視窗算出來的部分。
	DeliveryPanel *screen.DeliveryPanel

	// ReportPanel 是主畫面會計報告那個視窗算出來的部分（署名、頭像、金額）。
	ReportPanel *screen.ReportPanel

	// FlipReport 是星域易主報告那一頁的版面（`docs/spec/034`）。
	FlipReport *screen.FlipReportPanel

	// Reporter 是報告頭像用的提督編號，＝ 原版的 `ds:8F4D[陣營]`
	// （`docs/re/248`）。**不是統治者**（那個是 `ds:8F4F`）。
	// 量自新遊戲那一份傾印；換一個劇本會不會不同還沒驗。
	Reporter [2]int

	// BuildPlanPanel 是建造計劃那張的「船艦建造費」（同樣是 rules.Budget 算的）。
	BuildPlanPanel *screen.ReportPanel

	// AdmiralPanel 是提督詳細資料那個視窗算出來的部分（`docs/spec/012`）。
	AdmiralPanel *screen.AdmiralPanel

	// FleetPanel 是艦隊編成那個視窗算出來的部分（`docs/spec/013`）。
	FleetPanel *screen.FleetPanel

	// FleetList 是戰略地圖介面右下角那張清單的「艦隊」分頁（`docs/re/70`）。
	FleetList *screen.FleetList

	// CompositionPanel 是艦隊組成那個視窗算出來的部分（`docs/spec/015`），
	// Sprites 是 `SPRITE.SPR` 那 75 張 16 × 16 圖塊（格盤用它）。
	CompositionPanel *screen.CompositionPanel
	Sprites          []*asset.Image

	// BioPanel 是提督經歷那個視窗的內文版面，Biographies 是 82 篇人物傳
	// 的字碼（`docs/internals/15`）。
	BioPanel    *screen.BioPanel
	Biographies [][]uint16

	// Flagships 是兩個陣營各 16 艘旗艦的名字（`docs/re/65`）。
	Flagships *screen.Flagships

	// ProductionPanel 是生產計劃那個視窗的（`docs/spec/004`）。
	BudgetPanel     *screen.BudgetPanel
	ProductionPanel *screen.ProductionPanel

	// SlotList 是存檔畫面那 11 列的版面，SaveIndex 是 `SAVEDATA.IDX` 的內容
	// （`docs/spec/008`）。那 11 列是算出來的，不在畫面規格裡。
	SlotList  *screen.SlotList
	SaveIndex []rules.SaveIndexEntry

	// SaveDir 是存檔要寫去哪。**不會是原版目錄**——原版檔案一律唯讀。
	SaveDir string

	// AdmiralNames 是 82 名提督姓名的 Big5 字碼（`docs/internals/data/keireki.json`）。
	AdmiralNames [][]uint16

	// StarListOrder 是「星系資料」那張清單的列順序
	// （`docs/internals/data/star-list-order.json`）。
	StarListOrder [][]int

	// WorkTemplate 是新遊戲開局的 `WORK` 區塊（`docs/re/45`）。
	// 新遊戲要存檔就得靠它——`WORK.DTX` 不存在。
	WorkTemplate []byte

	// OrigDir 是原版檔案的目錄。「取進度」要從那裡讀存檔（`docs/spec/007`）。
	OrigDir string

	// Battle 是戰場畫面的版面，BattlePanel 是面板背景（`ANI_PNL0.A98`），
	// BattleTiles 是 `BLOCK_2.BLK` 的 128 張圖塊，BattleMaps 是 32 個星域
	// 各一張的戰場地形（`docs/internals/34`、`36`、`37`）。
	// 缺原版檔案時全是 nil，戰場畫面就開不起來。
	Battle      *screen.BattleSpec
	BattlePanel *asset.Image
	BattleTiles []*asset.Image
	BattleMaps  [][]uint8

	// Entries 是戰場入口表（`docs/internals/39`）。
	Entries *rules.BattleEntries

	// BattleMenu 是戰術選單的版面（`docs/re/101`）。
	BattleMenu *screen.BattleMenuSpec

	// CommandMenus 是戰場上五張命令選單的版面與內容（`docs/spec/023`）。
	CommandMenus *screen.CommandMenuSpec

	// UnitPanel 是點到分隊時右側換上的那塊面板（`docs/re/118`）。
	UnitPanel *screen.UnitPanelSpec

	// Formation 是 41 格的隊形槽位表（原版 `ds:59BA`，`docs/re/126`）。
	// 「編隊」讓玩家在上面挑一格（`docs/internals/46`）。
	Formation *rules.FormationSlots

	// Terrain 是戰場地形的通行規則（`docs/re/100`）。戰術層挑集結點時
	// 要用它判「這條直線走不走得通」；nil ＝ 不判，一律當通得過。
	Terrain *rules.BattleTerrain

	// BattleData 是戰術選單「資料」那張畫面的版面（`docs/spec/021`），
	// FleetDetail 是點一列開出來的艦隊詳細資料（`docs/spec/022`）。
	BattleData  *screen.BattleDataSpec
	FleetDetail *screen.FleetDetailSpec

	// MessageBox 是訊息框的版面，FrameTiles 是 `FRAME.BLK` 的外框圖塊
	// （`docs/internals/47`）。缺原版檔案時 FrameTiles 是 nil，訊息框
	// 就不畫——**不用代用品**。
	MessageBox *screen.MessageBoxSpec
	FrameTiles *ui.FrameTiles

	// PlanetInfo 是行星情報視窗的版面（`docs/re/144`）。
	PlanetInfo *screen.PlanetInfoSpec

	// SupplyPanel 是「補給」選對象時右側那個面板（`docs/re/143`）。
	SupplyPanel *screen.SupplyPanelSpec

	// FactionNames 是陣營名的原始位元組（`ds:60F4`）。字串自己帶著換色的
	// 跳脫碼，所以顏色不是版面常數。ShipKindNames 是艦種名（`ds:613A`）。
	FactionNames  [][]byte
	ShipKindNames [][]byte

	// AdvanceRoutes 是原版寫死的進軍路線表（`docs/internals/data/
	// ai-advance-routes.json`，`docs/re/178`）。戰略 AI 的階段四、五要用。
	AdvanceRoutes *ai.AdvanceRoutes

	// DefaultAdjutants 是 `ds:5F78`：新編艦隊時提督沒有副官就從這裡填
	// （`docs/re/219`）。
	DefaultAdjutants rules.DefaultAdjutants

	// Composition 是新編艦隊要用的兩張表（`ds:2BC6` 樣板與 `ds:295C`
	// 艦種替代表，`docs/re/212`／`docs/re/216`）。nil ＝ 建出來的艦隊
	// 只有旗艦那一組。
	Composition *rules.CompositionTables

	// ShipKinds 是艦種屬性表（`docs/internals/40`）。
	ShipKinds *rules.ShipKindTable

	// Intro 是開頭動畫的解碼器（`IBM16.350` ＋ `IBMFLS16.350`），
	// IntroPal 是它的調色盤（`BU16IBM.RGB`）。**nil ＝ 直接從主選單開始**
	// ——測試與對拍腳本都不帶開頭動畫，帶了整條輸入序列就會位移。
	Intro    *asset.IntroDecoder
	IntroPal gfx.Palette

	// Weights 是戰力權重表（原版的 `ds:6032`，`docs/re/79`）。
	// nil ＝ 算不出戰力，AI 退回「每支艦隊算 1」。
	Weights rules.StrengthWeights

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
//
// ⚠ 畫面問的是**占領**，而占領在 `+0x13`（19），`+11` 是統率力。
// 這個常數留著只是為了 `Answer()`——**原版根本不驗答案**
// （`docs/re/75`：打 `000` 照樣進遊戲），所以它不影響任何行為。
const occupyOffset = 11

// Game 是一局遊戲的狀態。
type Game struct {
	res *Resources

	// pendingFlips 是還沒給玩家看的易主報告，afterFlips 是全部看完之後
	// 要接回去的那條路（`docs/spec/034`）。
	pendingFlips []RegionFlip
	afterFlips   func()
	state        State

	// 主選單：滑鼠位置與反白到哪一列。
	mouseX, mouseY int
	highlight      int

	// 防拷問答：已經打了什麼、答案是多少。
	typed  []byte
	answer uint8

	// pct 是船艦建造費的百分比分配。**建造計劃與生產計劃共用同一份**
	// ——在建造計劃上按箭頭改的，生產計劃那張跟著變（`docs/spec/004`）。
	// 開局是量到的那一組（`rules.OpeningPct`）。
	pct [rules.BuildPlanRows]int

	// plan 是建造計劃畫面當下那一份（建造費用月初報告那一組算）。
	plan *rules.BuildPlan

	// 結局：目前顯示哪一張 CG。
	ending *asset.Image

	// 開頭動畫：目前這一格的頁面、播到第幾格、離下一格還差多久。
	introPage  *asset.Image
	introFrame int
	introLeft  time.Duration

	// 命令階段：第幾個回合。遊戲只存這一個數，年月日是印的時候算出來的
	// （`docs/re/25`）。開局 0。
	turn int

	// 月初報告被關掉幾層。兩層都關掉才露出戰略地圖介面。
	reportLayer int

	// 手上的軍事費與下個月的預估徵稅額，**每個陣營一格**（原版的
	// `ds:8F52`／`ds:8F6A` 也是，`docs/internals/27`）。開場那次結算
	// 算出來，之後每個月底再算一次。電腦那一方要有自己的錢，戰略 AI
	// 的預算閘門才擋得住（`docs/spec/030`）。
	military, taxTotal [2]int
	// shipPct 是每個陣營切給造船的百分比 ＝ 原版 `ds:8F92[陣營]`
	// （`docs/re/256`）。新遊戲兩邊都是 `Resources.ShipPct`（25，
	// `63A1:0074` 設的）；電腦那一方每個回合收尾自己改
	// （`docs/re/257`）。**存檔裡的那一格還沒定位**，載入時先沿用開局值。
	shipPct [2]int
	// foePct 是**電腦那一方**切給七個艦種的百分比（原版
	// `ds:8FD6[陣營 × 7]`；玩家那一份就是 `pct`）。電腦在回合收尾寫回
	// 一張固定表（`docs/re/257`）。
	foePct [rules.BuildPlanRows]int
	// shipPctSet 分得開「還沒設過」與「設成 0」——**0% 是合法值**
	// （S3 的帝國就是 0%），不能拿零值當「沒設過」。
	shipPctSet bool
	// paidUpkeep／paidMove 是這個月**累計已付**的艦隊維持費與移動經費
	// （原版 `ds:8F7A`／`ds:8F8A`），月底歸零（`docs/re/255`）。
	// 徵稅計劃「已支付」那一欄印的就是它們（`docs/re/258`）。
	paidUpkeep, paidMove [2]int

	// blend 是未安定度算式的 `Blend`（原版 `ds:8FD4[陣營]`，0–4）。
	// 電腦座位的戰略 AI 每回合按軍事費重設它，一設就重算該陣營所有
	// 星域的稅率（`docs/spec/030`、`docs/internals/55`）。
	blend [2]int

	// rate 是原版 `ds:8FF8[陣營]`：月底真正收稅那一趟（`2703:03AF`）把當下的
	// `blend` 抄一份過來（`docs/re/252`）。只有未安定度的 f1 讀它
	// （`docs/spec/042`）。
	rate [2]int

	// treasury 是這個月的財政（原定軍事費、已支付各項），
	// monthReport 是月初報告那張上的預算。
	treasury    rules.Treasury
	monthReport rules.Budget

	// building 是**上一次結算撥款出去**的那份計劃：買到的船這個月還在造，
	// 下個月底交付（`docs/re/21`）。生產計劃畫面的「建造中」就是它。
	//
	// **每個陣營各一份**，索引就是陣營——原版的 `ds:8F9C` 也是
	// （`docs/internals/27`）。電腦那一方也會交付，少留一份它就永遠
	// 造不出船（`docs/spec/036`）。
	building [2]rules.Building

	// delivered 是月底報告上那一份：跨月的當下交付出去的船。
	delivered               rules.Building
	reportYear, reportMonth int

	// 目前選中的星域：對星圖按右鍵選的那一格。**右下角清單的「艦隊」
	// 分頁看的是它**（`fleetRows`）；左下角那塊看的是 `hoverRegion`，
	// 不是這個。**還沒選過是 −1** ＝ 退回自己的首都。
	selectedRegion int

	// hoverRegion 是左下角那塊現在寫著哪一格星域，−1 ＝ 整塊空的
	// （`docs/re/277`、`docs/spec/001`）。
	//
	// 它是**游標底下**那一格，不是右鍵選中的那一格；而且只在戰略地圖
	// 介面是最上層的時候跟著游標走——選單或視窗一開就凍在那裡。
	// 重算的地方只有 `refreshHoverRegion` 一處。
	hoverRegion int

	// music 是設定畫面上那三行：音源、平時配樂、戰鬥配樂。
	// 原版的 `ds:8F33`／`8F34`／`8F35`，跟著存檔走（`WORK.DT?` 的
	// `+0x1F`／`+0x20`／`+0x21`，`docs/internals/30`）。
	music MusicSettings

	// 原版的 `ds:8F36`／`8F37`／`8F38`（動畫效果／動畫速度／訊息速度），
	// 同樣跟著存檔走（`WORK.DT?` 的 `+0x22`／`+0x23`／`+0x24`，
	// `docs/re/286`）。
	display DisplaySettings

	// saveErr 是上一次存檔的結果（nil ＝ 成功或還沒存過）。
	saveErr error

	// strategyRand 是戰略 AI 的亂數（原版 `ds:94DC`，`docs/re/184`）。
	// **跨回合沿用同一份**——原版的種子是全域的。叛亂判定抽的也是它
	// （`docs/spec/041`），取用一律走 `rand94DC`。
	strategyRand *rules.Rand

	// rebellionRolls／rebellions 是這一回合叛亂判定的逐格結果與叛亂的
	// 星域（`runRebellions`，`docs/spec/041`）。報告畫面還沒有
	// （`strategy-report-screen`），先留給測試與之後的報告用。
	rebellionRolls [rules.RegionCount]rules.RebellionRoll
	rebellions     []int

	// 費沙事件：pending 是還沒在回合尾巴發動的登記；battleResolved 用來
	// 保留原版「戰術戰鬥只有人類拿下才登記」的例外。
	phezzanPending        bool
	phezzanBattleResolved bool
	phezzanRebellions     []int

	// slot 是目前這個局面從哪一格存檔來的（新遊戲是 nil）。
	// 存檔寫出要靠它——`WORK` 那 230 bytes 只解出三個欄位，
	// 其餘得原樣抄回去（`docs/spec/009`）。
	slot *rules.Slot

	// faction 是玩家陣營：0 帝國、1 同盟。主選單前兩列各選一邊
	// （`docs/re/61`）。
	faction int

	// worldStellar／worldPlanets 是星域記錄與行星表的開局值。月底結算
	// 會就地改寫那兩份，重開新遊戲時要先還原（見 `snapshotWorld`）。
	worldStellar, worldPlanets []byte

	// listIndex 是目前清單每一列對應的提督索引（`setListRows` 填）。
	listIndex []int

	// detailAdmiral 是提督詳細資料那張畫面在看哪一位；-1 ＝ 沒有。
	detailAdmiral int

	// selectedRow 是有選取條那種清單目前選在第幾列（艦隊組成的七個艦種）。
	selectedRow int

	// fleets 是已經編成的艦隊，順序就是艦隊編號 − 1
	// （`internal/rules.Fleet`，對應原版的執行期艦隊表）。
	fleets []rules.Fleet

	// starMap 是照玩家陣營重建的勢力色地圖（首都那格的顏色跟著玩家走）。
	// nil 表示還沒建，畫的時候退回 `Resources.StarMap`。
	//
	// plainMap 是同一張但**不塗玩家色**：對戰的交棒視窗用它
	// （`docs/spec/016`）。
	starMap, plainMap *asset.Image

	// menuRow 是主選單點了哪一列（-1 ＝ 還沒點）。防拷答對之後要靠它
	// 決定是開新遊戲還是進存檔畫面。
	menuRow int

	// versus ＝ 主選單選了第 2 列「對戰」：兩個人輪流下命令
	// （`docs/re/76`、`docs/spec/016`）。
	versus bool

	// battleRegion 是戰術畫面看的是哪個星域的地形，battleView 是視野
	// 在整張戰場上的位置（單位是格）。-1 ＝ 還沒進過戰場。
	battleRegion int
	battleView   screen.BattleView
	battleTurn   int
	battleUnits  []rules.BattleUnit
	tactic       ai.Tactic
	autoBattle   bool

	// 結算用的（`docs/spec/018`）：參戰艦隊、亂數、上一回合的交火紀錄，
	// 以及「有一方全沒了」。
	battleFleets []rules.BattleFleet
	// 戰術選單開在游標處：battleMenu 是「開著沒有」，battleMenuX／Y 是框的
	// 左上角（`docs/internals/34`）。
	battleMenu bool

	// battleData ＝ 戰術選單「資料」開著沒有，battleDataTab 是看哪一個
	// 分頁（0 艦隊／1 行星／2 敵艦隊）。
	battleData    bool
	battleDataTab int

	// battleDetail ＝ 艦隊詳細資料那一層開著沒有，battleDetailRow 是
	// 點的是表上第幾列。
	battleDetail             bool
	battleDetailRow          int
	battleMenuX, battleMenuY int

	// battleCmd ＝ 戰場的命令選單開著沒有（`docs/spec/023`）；
	// battleCmdUnit 是點到的分隊在 battleUnits 裡的索引，battleCmdKind
	// 是五張表的哪一張，battleCmdX／Y 是框的左上角。
	battleCmd              bool
	battleCmdUnit          int
	battleCmdPlanet        int // -1 ＝ 這一張不是行星的選單
	battleCmdKind          screen.CommandMenuKind
	battleCmdX, battleCmdY int

	// battleMove ＝ 選了「移動」之後在等玩家點目的地（`docs/spec/024`）；
	// battleMoveUnit 是要移動的分隊。
	battleMove     bool
	battleMoveUnit int

	// battleForm ＝ 選了「編隊」之後在等玩家點隊形格（`docs/internals/46`）；
	// battleFormUnit 是要換位置的分隊。
	battleForm     bool
	battleFormUnit int

	// battleSupply ＝ 選了「補給」或「修理」之後在等玩家點對象
	// （`docs/re/131`）；battleSupplyRepair 分辨是哪一個。
	battleSupply       bool
	battleSupplyUnit   int
	battleSupplyRepair bool
	battleSupplyPlanet int // >= 0 ＝ 這一次是行星的「補給」／「補充」

	// 艦隊選單的「移動」／「戰鬥」是兩段互動（`docs/re/120`–`122`）：
	// 先點目的地，再繞著它轉游標選抵達朝向。
	fleetOrder       FleetOrderStage
	fleetOrderUnit   int
	fleetOrderBattle bool
	fleetOrderX      int
	fleetOrderY      int
	fleetOrderFacing int

	// planetInfoOn ＝ 行星情報視窗開著（`docs/re/144`），planetInfo 是
	// 哪一顆。原版按任一鍵關掉。
	planetInfoOn bool
	planetInfo   int

	// msgOn ＝ 訊息框開著（模態）。msgKind 是樣板種類，msgRaw 是原版資料段
	// 那一串 Big5（控制碼夾在裡面，交給 `ui.TextBuffer` 解；`docs/re/141`）。
	msgOn   bool
	msgKind int
	msgRaw  []byte

	// battleGunOn ＝ 正在替要塞砲指定目標，battleGun 是哪一座
	// （`docs/re/134`）。**用旗標而不是 -1**：`Game` 的零值要是
	// 「沒有在指定」，否則沒走過 enterBattle 的測試會被吃掉點擊。
	battleGunOn bool
	battleGun   int
	// battleGunHits 是上一回合要塞砲打掉的分隊（`docs/spec/028`）；
	// battleAssaults 是上一回合打行星的結果（`docs/spec/019`）。
	battleGunHits  []rules.FortressGunHit
	battleAssaults []rules.PlanetAssaultHit
	// battleGunAim 是射程橢圓現在的傾角（一整圈 256）；瞄準當下由
	// 游標的方位算，按下左鍵才寫進行星記錄的 `+0x50`。
	battleGunAim int

	battleRally   []rules.RallySite
	battlePlanets []rules.BattlePlanet
	battleRand    *rules.Rand
	battleLosses  []rules.BattleLoss
	battleScore   [2]int
	battleOver    bool

	// work／admirals 是這一局**可寫**的 `WORK` 區塊與提督表。
	//
	// 兩者原本都直接讀資源（範本或存檔），但電腦每個回合會重新指派專任
	// 參謀（`docs/re/268`）——那要寫 `WORK` 的 `+0x37` 起那六格與提督記錄
	// 的 `+0x09`。`Resources` 可能被多局共用，所以第一次要寫的時候才複製
	// （`mutableWork`／`mutableAdmirals`）。nil ＝ 這一局還沒有人寫過。
	work     []byte
	admirals []byte

	// brain 是電腦那一方的決策（`internal/ai`、`docs/spec/017`）。
	//
	// **預設是 nil ＝ 按實行時什麼都不做**，與 `ai.Original` 同義
	// （那一支也回 nil）。要原版那套戰略 AI 是 `SetBrain(OriginalBrain())`
	// ——那一步是 `docs/spec/029` 末節的裁決；要強化版就換成 `ai.Focused`。
	brain ai.Brain

	// sides 是對戰模式下兩邊各自的錢與造船計劃，索引就是陣營。
	// 世界資料（星域記錄、行星表）兩邊共用一份——月底的成長是就地
	// 寫回的，一個月只能跑一次。
	//
	// ⚠ **兩邊各自的開局數字是推論，不是量到的**（`docs/spec/016`）：
	// 收據只走到帝國的第一張月初報告。
	sides [2]side
}

// side 是對戰模式下屬於某一個陣營的那幾筆狀態。單人模式用不到——
// 那時 Game 上那幾個欄位就是玩家自己的。
type side struct {
	pct         [rules.BuildPlanRows]int
	treasury    rules.Treasury
	monthReport rules.Budget
	fleets      []rules.Fleet
}

// 星圖畫在畫面的 (0,40)，320×300，圖塊 16×20（`docs/internals/12`）。
const (
	mapX, mapY   = 0, 40
	mapW, mapH   = 320, 300
	tileW, tileH = 16, 20
)

// startNewGame 從開局狀態開始一局，陣營由主選單那一列決定
// （`docs/re/61`）。
//
// 原版開機就跑一次月底結算，796 年 1 月的局面是那一次的結果
// （`docs/re/33`、`docs/re/35`）；換陣營要重跑一次，因為收的稅是
// **自己那一方的星域**加出來的。
// execute 是系統指令的「實行」。
//
// 單人模式一次推進三天（`docs/re/20`、`docs/re/25`）。對戰模式的意思是
// 「我這一方下完了」：帝國按完換同盟、日期不動，同盟按完才推進三天並
// 換回帝國（`docs/re/76`，收據 `vs5`／`vs6`）。
func (g *Game) execute() {
	if g.versus {
		g.sides[g.faction] = g.side()
		next := 1 - g.faction
		g.enterSide(next)
		g.state = StateVersusInput
		if next != 0 {
			return
		}
		// 換回帝國 ＝ 這一輪兩邊都下完了，日期才推進。
	}
	g.advanceTurn()
}

// advanceTurn 推進一個回合 ＝ 三天。
func (g *Game) advanceTurn() {
	g.turn++
	// 電腦那一方先下命令，再讓所有艦隊走。
	// `runBrain` 讀的原版回合計數器是 `g.turn - 1`；它跑完之後原版才把
	// 計數器進位，接著進新回合的月底報告（`docs/re/299` 時間線）。
	g.runBrain()
	// 命令階段那一串的第一支 `056A:09EB` 一進去就先跑 `090A:0E6D`，把 32 格的
	// 未安定度整份重算（`docs/spec/042`）——叛亂的門檻讀的就是這一次的值。
	// 回合計數器要到回合尾巴才進位，remake 的 `g.turn` 在開頭就先加了，
	// 所以這裡的 `ds:8F18` 是 `g.turn − 1`。
	g.recomputeUnrest(g.turn - 1)
	// 每個戰略回合逐支扣艦隊維持費（原版 `056A:09EB` 的
	// `sub [bx+8F52], si`，`docs/re/236`）：**兩邊一起扣**，扣的是那一支
	// 自己陣營的那一格。順序照原版——在 AI 回合之後、艦隊移動之前。
	for f := 0; f < 2; f++ {
		up := g.fleetUpkeep(f)
		g.military[f] -= up
		g.paidUpkeep[f] += up // ds:8F7A，月底歸零
	}
	// 民心往擁有者的目標值走一格（原版 `056A:0AA5`，`docs/spec/040`）：
	// 同一串裡緊接在扣維持費之後、艦隊移動之前。擁有者讀星域記錄的
	// `+0x12`，這時還是上一個回合尾巴重算完的值。
	if g.res != nil {
		rules.DriftRegionMorale(g.res.Stellar)
	}
	// 有命令的艦隊往目的地走一格（`docs/re/77`、`internal/rules.Advance`）。
	// **一回合走一格是假設待驗**——量到的三組移動都只隔一格。
	if g.res.Graph != nil {
		for i := range g.fleets {
			if !g.res.Graph.Advance(&g.fleets[i]) {
				continue
			}
			// 走一格要錢：艦隊記錄的 `+0x0C`，原版 `113A:0729` 逐支從
			// 那一支陣營的軍事費扣（`docs/re/237`）。
			f := g.fleets[i]
			if f.Faction < 0 || f.Faction > 1 {
				continue
			}
			cost := rules.FleetOtherCost(
				g.res.ShipKinds, f.Faction, f.Composition, f.Operations)
			g.military[f.Faction] -= cost
			g.paidMove[f.Faction] += cost // ds:8F8A，月底歸零
		}
	}
	// 叛亂（`docs/spec/041`）：原版命令階段那一串的最後一支 `056A:0469`
	// 逐格判定，排在艦隊移動與重數艦隊（`2703:0853`）之後、逐格開戰之前；
	// 效果（`2703:1908`）在逐格開戰那一圈沒有戰鬥的格子裡套上。兩段之間
	// 互不相干，這裡判完就直接套。亂數與戰略 AI 同一個序列，排在 AI 之後。
	//
	// ⚠ 原版每一格叛亂都跳一則模態報告（`056A:06F3`），remake 還沒有畫面
	// （`strategy-report-screen`）；結果在 `g.rebellions`。
	g.runRebellions()
	// 兩邊的艦隊走到同一個星域就直接切進戰術畫面（`docs/internals/34`
	// 「怎麼走到」：S3 按六次實行，雙方會合在達希利就開戰）。
	// **哪一個星域先打、能不能不打，都還沒解**——同時有多個星域符合時
	// 這裡取編號最小的那一個。
	//
	// 原版 `0374:03E8` 逐格掃，`+0x13` bit2 有設的就開打，**打完接著掃
	// 下一格**；全部打完才走回合結束（`docs/re/251`）。所以這裡進戰鬥
	// 之後由 `EndBattle` 接著掃，不是就此結束這個回合。
	if r := g.contestedRegion(); r >= 0 {
		g.StartBattle(r)
		return
	}
	g.endOfTurn()
}

// endOfTurn 是所有戰鬥都打完之後的那一段（原版主迴圈的 `056A:078E`
// 加上 `ds:8F18++`，`docs/re/251`）：依行星重算星域歸屬、跳易主報告、
// 重建路徑表與每一格的態勢，然後才是跨月結算與下一個畫面。
func (g *Game) endOfTurn() {
	// **有易主就先停在報告上**，後半段等按掉才跑（`docs/spec/034`）。
	if g.runRegionPass(g.finishTurn) {
		return
	}
	g.finishTurn()
}

// finishTurn 是回合結束的尾段：跨月結算與下一個畫面。易主報告會把它
// 延後到報告全部按掉之後（`docs/spec/034`）。
func (g *Game) finishTurn() {
	// 費沙事件在逐格易主報告與擁有者重算之後、勝負／月底結算之前。
	g.runPhezzanEvent()
	// 一個月十回合；跨過去就結算（`docs/re/26`、`docs/re/33`）
	// 並跳出船艦接收報告。
	//
	// ⚠ **對戰模式的跨月沒有量過**（`docs/spec/016`）：這裡照單人模式
	// 只結算目前這一方，另一方的錢停在上個月。要往下走得先有收據。
	month := g.turn%rules.TurnsPerMonth == 0
	if month {
		g.deliverShips()
		g.delivered = g.building[g.faction]
		g.reportYear, g.reportMonth, _ = g.Date()
		// 報告掛的是**剛結束的那個月**，日期已經翻到次月 1 日。
		g.reportMonth--
		if g.reportMonth == 0 {
			g.reportMonth, g.reportYear = 12, g.reportYear-1
		}
		g.monthEnd()
	}
	// 原版 `0654:0034` 的共同尾巴每個回合都跑：照現在的狀態把預估徵稅額
	// 刷一次（`docs/spec/035`）。中途打下一個星域之後，右側面板與徵稅
	// 計劃當回合就要跟著變，不必等到月底。
	g.refreshTaxEstimate()
	// 同一支 `090A:0E6D` 的第三件事是 32 格的未安定度（`2703:02E5` →
	// `2703:0DA1`）。這時回合計數器已經進位，路徑表也在座位回合開頭
	// （`04F8:003E` → `2B99:0003`）照新的擁有者重建過（`docs/spec/042`）。
	g.recomputeUnrest(g.turn)
	if month {
		if g.commandSpec(StateShipDelivery) != nil {
			g.state = StateShipDelivery
			return
		}
	}
	if g.versus {
		g.state = StateVersusInput
		return
	}
	g.state = StateCommandPhase
}

// contestedRegion 回「同時有兩個陣營的艦隊」的星域，沒有就回 -1。
//
// 原版在雙方會合的那一刻直接切進戰術畫面（`docs/internals/34`）。
// 判斷的細節（幾支才算、能不能迴避）還沒解，這裡只看有沒有兩邊都在。
func (g *Game) contestedRegion() int { return g.contestedRegionFrom(0) }

// contestedRegionFrom 從 `from` 起找下一個雙方都有艦隊的星域；沒有回 -1。
//
// 原版 `0374:03E8` 是**逐格往上掃**，打完一格接著掃下一格
// （`docs/re/251`），所以續掃要從上一格的下一個開始，不會回頭。
func (g *Game) contestedRegionFrom(from int) int {
	for r := from; r < rules.RegionCount; r++ {
		var side [2]bool
		for _, f := range g.fleets {
			if f.Region == r && f.Faction >= 0 && f.Faction < 2 {
				side[f.Faction] = true
			}
		}
		if side[0] && side[1] {
			return r
		}
	}
	return -1
}

// SetBrain 換掉電腦那一方的決策。nil ＝ 回到原版（什麼都不做）。
func (g *Game) SetBrain(b ai.Brain) { g.brain = b }

// runBrain 讓電腦那一方下這一回合的命令。
//
// 命令只有一種形狀：**寫目的地並解開移動狀態**，與原版的 `+0x07`／
// `+0x08` 一樣（`docs/internals/33`）。對戰模式兩邊都是人，不跑。
func (g *Game) runBrain() {
	if g.brain == nil || g.versus || g.res.Graph == nil || g.res.Planets == nil {
		return
	}
	foe := 1 - g.faction
	w := ai.World{
		Fleets: g.fleets,
		Graph:  g.res.Graph,
		Owner:  func(r int) int { return g.res.Planets.RegionOwner(r) },
		Strength: func(f rules.Fleet) int {
			if f.Strength > 0 {
				return f.Strength
			}
			return 1
		},
		Value: func(r int) int {
			o := r * rules.RegionRecord
			if o+rules.RegionRecord > len(g.res.Stellar) {
				return 1
			}
			return rules.RegionTax(g.res.Stellar[o : o+rules.RegionRecord])
		},
	}
	// **建艦隊排在下命令前面**：原版 `2D10:0005` 先 `2D76:0002`（建艦隊）
	// 再 `2E54:000C`（那五個階段，`docs/re/175`）。順序反了有兩個後果
	// ——這一回合新編的艦隊拿不到命令，而且亂數的呼叫次序整個錯開
	// （建艦隊每建一支吃兩個，`docs/re/218`）。
	//
	// 原版 AI 回合的第一件事是 `2D10:0319(1)`：按軍事費重設 `Blend`，
	// 一設就重算全國稅率（`docs/spec/030`）。建艦隊的預算閘門看的就是
	// 重算之後的稅收，所以這一步要排在建艦隊前面。
	g.refreshBlend(foe, rules.BlendForFunds(g.military[foe]))
	// 階段一產生的是**新艦隊**不是命令，所以走另一條路
	// （`ai.Builder`，`docs/internals/54`）。**選配**：沒實作就不建。
	if b, ok := g.brain.(ai.Builder); ok {
		g.buildFleets(b.Build(w, foe, g.turn), foe)
		w.Fleets = g.fleets // 新編的要看得到，後面的階段才挑得到它們
	}
	// 建艦隊之後、五個階段之前跑整補（原版 `2D76:0732`，
	// `docs/spec/031`）。**選配**：沒實作就不整補。
	if _, ok := g.brain.(ai.Refitter); ok {
		g.refitFleets(foe)
		w.Fleets = g.fleets
	}
	for _, o := range g.brain.Orders(w, foe, g.turn) {
		if o.Fleet < 0 || o.Fleet >= len(g.fleets) {
			continue
		}
		g.fleets[o.Fleet].Target = o.Target
		g.fleets[o.Fleet].Parked = false
	}
	// `2D10:0038` 的前兩步：重新指派情報與內政專任參謀（`docs/re/268`）。
	// 排在 `Blend` 之前——換了人之後那一步重算稅率才會用到新的那位。
	g.assignStaff(foe)
	// 回合的最後一件事是 `2D10:0319(0)`：月底照軍事費再算一次，
	// 平時把 `Blend` 歸零。兩條都會再重算一次稅率。
	//
	// 原版在這一刻的 `ds:8F18` 還沒進位；remake 已在 `advanceTurn` 開頭
	// 把 `g.turn` 加一，所以原版計數器是 `g.turn - 1`（`docs/re/299`）。
	level := 0
	if (g.turn-1)%rules.TurnsPerMonth == rules.TurnsPerMonth-1 {
		level = rules.BlendForFunds(g.military[foe])
	}
	g.refreshBlend(foe, level)
	// `2D10:0038` 的尾段：重算完稅率之後，照**新的**總稅收設自己的造船
	// 百分比（`docs/re/257`）。順序照原版——`Blend(0)` 會重算稅率，
	// 所以這一步讀到的是重算後的數。
	g.shipPct[foe] = rules.AIShipPct(g.taxTotal[foe])
	if foe == 0 || foe == 1 {
		g.foePct = rules.AIBuildPct[foe]
	}
}

// recomputeAllTaxRates 是原版 `090A:0B5E`：月底在收稅之前把兩個陣營所有
// 星域的稅率重算一遍（各自用自己的 `Blend` 當難度下標）。
//
// 稅率吃星域的成長率，而成長率剛剛被這個月實際的成長改過，所以這一步
// 不做的話稅率會一直停在開局那一份（`docs/spec/038`、`docs/re/267`）。
//
// **無主星域那一支（`090A:0DC0`）還沒讀**——那些格子不進任何一方的國庫，
// 只影響星域面板上顯示的稅率。
func (g *Game) recomputeAllTaxRates() {
	pr := g.res.UnrestParams
	if g.res.Planets == nil || len(g.res.Stellar) == 0 {
		return
	}
	owner := func(i int) int { return g.res.Planets.RegionOwner(i) }
	for f := 0; f < 2; f++ {
		level := g.blend[f]
		if level < 0 || level >= len(pr.Base) {
			continue
		}
		ruler := g.admiralRecord(g.taxRateAdmiral(f))
		if ruler == nil {
			continue
		}
		rules.RecomputeTaxRates(g.res.Stellar, f, pr.Base[level], ruler, owner)
	}
}

// taxRateAdmiral 回稅率算式要乘的那位提督。原版 `090A:0B93` 讀的是
// `ds:8F4F + 陣營`——三個幕僚職的第三格，也就是**內政專任參謀**
// （`docs/re/267`）。
//
// ⚠ **`ds:8F4F` 不是常數，而 remake 目前拿不到它的執行期值。**
// `work-796-01.json` 量自第 71,200,000 步（防拷畫面之前），同盟那一格是 60；
// 進到遊戲之後（第 300,000,000 步，回合 1）已經變成 63，而拿 63 重算回合 10
// 的十三個同盟星域稅率**全中**。但 796/1 的同盟開局畫面 `main1eAlly`
// 要的是 60——兩個值各有收據，所以那一格在載入之後被改過。
//
// **改它的是誰已經查完了**（`docs/re/268`）：電腦在自己回合的尾巴重新指派
// 內政專任參謀（`2D10:0038` → `2D10:0213`），規則本身在
// `rules.PickDomesticStaff`，用收據的提督表重算選出的就是 63。remake 這裡
// 仍然回載入時那一份，因為**戰略 AI 還沒接進 `advanceTurn`**
// （`internal/ai` 的 `Original` 回 `nil`）——整個 AI 回合都沒跑，這一步
// 自然也沒有。接上去之前同盟的月底稅率會偏低約四分之一
// （`docs/spec/038`「不在這一份裡」）。
func (g *Game) taxRateAdmiral(faction int) int {
	if i := rules.StaffAdmiral(g.workRecord(), rules.PostDomestic, faction); i >= 0 {
		return i
	}
	return g.res.UnrestParams.Ruler[faction]
}

// assignStaff 是原版 `2D10:0038` 的前兩步：電腦在自己回合的尾巴重新指派
// **情報專任參謀**（`2D10:0143`）與**內政專任參謀**（`2D10:0213`），
// 順序照原版（`docs/re/268`）。軍事專任參謀那格 AI 不換。
//
// 挑完接著的 `refreshBlend` 就是原版的 `090A:0B93`——換了人稅率要跟著
// 重算，那正是同盟月底稅收對不上的成因。
func (g *Game) assignStaff(faction int) {
	admirals := g.mutableAdmirals()
	if len(admirals) < rules.AdmiralRecord {
		return
	}
	work := g.mutableWork()
	capital := rules.CapitalRegion(faction)
	posts := []struct {
		post int
		pick func(admirals []byte, faction, capital, current int) int
	}{
		{rules.PostIntel, rules.PickIntelStaff},
		{rules.PostDomestic, rules.PickDomesticStaff},
	}
	for _, p := range posts {
		cur := rules.StaffAdmiral(work, p.post, faction)
		if cur < 0 {
			continue
		}
		// 原版先把現任寫回「閒置」（`2D10:023A`），所以現任自己也在候選
		// 清單裡、留任是合法結果。
		if r := g.admiralRecord(cur); r != nil {
			rules.LeaveStaffPost(r)
		}
		next := p.pick(admirals, faction, capital, cur)
		rules.SetStaffAdmiral(work, p.post, faction, next)
		if r := g.admiralRecord(next); r != nil {
			rules.TakeStaffPost(r)
		}
	}
}

// refreshBlend 走原版 `2D10:0319` 的後半：寫下等級，接著用它當下標
// 重算那個陣營所有星域的稅率（`090A:0B93`），最後跑 `090A:0E6D`——預估
// 徵稅額與 32 格的未安定度跟著重算（`docs/spec/042`）。`2D10:0213` 換完
// 內政專任參謀也是 `090A:0B93` → `090A:0E6D`，remake 在 `runBrain` 裡緊接著
// 就是這一支，所以不另外跑。
func (g *Game) refreshBlend(faction, level int) {
	g.blend[faction] = level
	pr := g.res.UnrestParams
	if g.res.Planets == nil || len(g.res.Stellar) == 0 ||
		level < 0 || level >= len(pr.Base) {
		return
	}
	ruler := g.admiralRecord(g.taxRateAdmiral(faction))
	if ruler == nil {
		return
	}
	rules.RecomputeTaxRates(g.res.Stellar, faction, pr.Base[level], ruler,
		func(i int) int { return g.res.Planets.RegionOwner(i) })
	g.taxTotal[faction] = g.estimateTaxFor(faction)
	// AI 回合在回合計數器進位之前，`ds:8F18` ＝ `g.turn − 1`（見 advanceTurn）。
	g.recomputeUnrest(g.turn - 1)
}

// recomputeUnrest 把 32 格的未安定度 `+0x7D` 照現在的局面整份重算——原版
// `2703:0DA1` 逐格跑一次（`docs/spec/042`）。原版有兩個呼叫端：
//
//	090A:0E6D → 2703:02E5   稅率或參謀一變、每次「實行」扣維持費之前、
//	                        每個座位回合開頭（月底那一趟跑兩次）
//	2703:08DE              回合尾巴，重建路徑表與態勢旗標之後
//
// `counter` 是那一刻的回合計數器 `ds:8F18`（算式只用它除以 10 的餘數）。
// 其餘輸入都讀現況：`Blend`／`ds:8FF8`、內政專任參謀（`ds:8F4F`，就是稅率
// 那一位）、星域旗標與民心、行星、`+0x08 == 5` 的艦隊，以及照現在的擁有者
// 建的路徑表（`2B99:0003` 在回合尾巴與玩家座位回合開頭重建，擁有者只在回合
// 尾巴那一段會變，所以兩者相同）。
func (g *Game) recomputeUnrest(counter int) {
	if !g.regionsReady() || len(g.res.UnrestParams.Base) == 0 ||
		len(g.res.UnrestParams.Table) == 0 {
		return
	}
	if counter < 0 {
		counter = 0
	}
	params := g.res.UnrestParams
	params.Turn = counter
	params.Blend = g.blend
	params.Rate = g.rate
	for f := 0; f < 2; f++ {
		params.Ruler[f] = g.taxRateAdmiral(f)
		if params.Blend[f] < 0 || params.Blend[f] >= len(params.Table) ||
			params.Rate[f] < 0 || params.Rate[f] >= len(params.Base) {
			return
		}
	}
	owner := make([]int, rules.RegionCount)
	for i := range owner {
		owner[i] = int(g.regionRecord(i)[regionOwnerOff])
	}
	fleets := g.fleets
	if fleets == nil {
		fleets = []rules.Fleet{} // 非 nil ＝ 用 remake 的艦隊表
	}
	st := &rules.UnrestState{
		Regions: g.res.Stellar, Planets: g.res.Planets, FleetList: fleets,
		Admirals: g.res.Admiral, Graph: g.res.Graph, Params: params,
		Paths: g.res.Graph.UnrestPaths(owner),
	}
	for i := 0; i < rules.RegionCount; i++ {
		// `2703:0DA1` 只讀別格的擁有者與 `+0x13`，不讀別格的 `+0x7D`，
		// 所以逐格寫回與全部算完再寫結果相同。
		g.regionRecord(i)[rules.RegionUnrest] = byte(st.Unrest(i))
	}
}

// buildFleets 把 `ai.Builder` 要的新艦隊編出來。
//
// 原版的上限是每個陣營 16 支（艦隊表每邊 16 個槽，`ds:8F49[陣營] < 16`，
// `docs/internals/33`），而且一位提督只帶一支。編出來的停在原地、
// 沒有移動命令——與玩家走「艦隊編成」建出來的那一支相同（`docs/re/70`）。
func (g *Game) buildFleets(list []ai.NewFleet, faction int) {
	if len(list) == 0 {
		return
	}
	n := 0
	led := make(map[int]bool, len(g.fleets))
	for _, f := range g.fleets {
		if f.Faction == faction {
			n++
			led[f.Admiral] = true
		}
	}
	for _, nf := range list {
		if n >= ai.FleetCap {
			return // 滿了
		}
		if nf.Region < 0 || nf.Region >= rules.RegionCount || led[nf.Admiral] {
			continue
		}
		g.fleets = append(g.fleets, rules.Fleet{
			Index:    len(g.fleets),
			Number:   len(g.fleets) + 1,
			Faction:  faction,
			Admiral:  nf.Admiral,
			Adjutant: nf.Adjutant,
			Region:   nf.Region,
			Target:   rules.NoTarget,
			ClassKey: nf.ClassKey,
			Expected: nf.Expected,
			// 編制與維持費在 `ai.BuildSupport.Form` 那一步就算好了
			// （存量也是在那裡扣的），這裡只是收下來。
			Composition: nf.Composition,
			Groups:      len(nf.Composition),
			Operations:  nf.Operations,
			// `297B:0066` 建立艦隊時寫 `+0x3C = 100`（`docs/re/214`）。
			Morale: rules.NewFleetMorale,
		})
		// 原版 `2D76:0611` 配到艦隊槽時把提督記錄的 `+0x09` 寫成
		// 「bit7 清掉、低七位是槽號」，副官那一筆寫同一個槽
		// （`docs/internals/08`）。**挑專任參謀的候選清單看的就是那個
		// 位元**（`docs/re/268`），少了這一步，帶隊的提督下一輪還會被
		// 挑去坐參謀，候選集就與原版不同。
		g.mutableAdmirals()
		if r := g.admiralRecord(nf.Admiral); r != nil {
			rules.SetFleetSlot(r, n)
		}
		if r := g.admiralRecord(nf.Adjutant); r != nil {
			rules.SetFleetSlot(r, n)
		}
		led[nf.Admiral] = true
		n++
	}
}

// startVersus 開一局對戰。
//
// 兩邊的開局狀態各算一次：`startNewGame` 會先把世界還原成開局值再結算，
// 所以跑兩次不會把成長疊上去——留在世界上的是第二次那一份，而兩次的
// 成長完全相同，差別只在收誰的稅。
//
// ⚠ **兩邊各自的開局數字是推論**：收據只走到帝國的第一張月初報告
// （`docs/spec/016`）。
func (g *Game) startVersus() {
	g.versus = true
	for f := 0; f < 2; f++ {
		g.startNewGame(f)
		g.fleets = nil
		g.sides[f] = g.side()
	}
	g.enterSide(0)
	g.state = StateVersusInput
}

// side 把目前這一方的狀態收起來。
func (g *Game) side() side {
	return side{pct: g.pct,
		treasury: g.treasury, monthReport: g.monthReport, fleets: g.fleets}
}

// enterSide 換到某一方：陣營、地圖顏色與那一方的錢一起換過去。
func (g *Game) enterSide(faction int) {
	g.faction = faction
	g.starMap = g.factionStarMap(faction, true)
	g.plainMap = g.factionStarMap(faction, false)
	s := g.sides[faction]
	g.pct = s.pct
	g.treasury, g.monthReport = s.treasury, s.monthReport
	g.fleets = s.fleets
	// 交棒的當下底板是乾淨的：右側面板只剩陣營那一行與按鈕，
	// 左下角是空的（收據 `vs3`／`vs5`）。
	g.selectedRegion = -1
	g.hoverRegion = -1
	g.detailAdmiral = -1
	g.reportLayer = 0
}

func (g *Game) startNewGame(faction int) {
	g.restoreWorld()
	g.work = nil
	g.phezzanPending = false
	g.phezzanBattleResolved = false
	g.phezzanRebellions = nil
	g.faction = faction
	g.starMap = g.factionStarMap(faction, true)
	g.plainMap = g.factionStarMap(faction, false)
	g.turn = 0
	// 兩個陣營的起始軍事費相同（796/1 量到帝國 50000，`docs/re/35`）。
	g.military = [2]int{g.res.StartMilitary, g.res.StartMilitary}
	g.blend = [2]int{rules.NewGameBlend, rules.NewGameBlend}
	// `ds:8FF8` 開局照新遊戲的量測（兩邊都是 2，`docs/spec/006`）；緊接著的
	// 開機月底結算會再抄一次 `blend`，值相同。
	g.rate = g.res.UnrestParams.Rate
	g.shipPct = [2]int{g.res.ShipPct, g.res.ShipPct}
	g.shipPctSet = true
	g.pct = rules.OpeningPct
	g.foePct = rules.OpeningPct
	// 開新遊戲時原版把兩首曲號清零（簽章 `31 C0 A2 34 8F A2 35 8F CB`，
	// `docs/re/274` §4）——**音源那一格不清**，它是開機偵測的結果。
	g.music.BGM, g.music.Battle = 0, 0
	g.slot = nil
	g.monthEnd()
}

// snapshotWorld 記下星域記錄與行星表的開局值。
//
// 月底結算會**就地改寫**這兩份（行星的經濟力成長、星域的經濟力寫回），
// 所以「重開一次新遊戲」必須先把它們還原——不然第二次結算是在已經
// 長過一次的資料上跑，稅收會多出來（換陣營時多 134）。
func (g *Game) snapshotWorld() {
	g.worldStellar = append([]byte(nil), g.res.Stellar...)
	if g.res.Planets != nil {
		g.worldPlanets = append([]byte(nil), g.res.Planets.Data...)
	}
}

// restoreWorld 把兩份世界資料還原成開局值。
func (g *Game) restoreWorld() {
	if g.worldStellar == nil {
		g.snapshotWorld()
		return
	}
	copy(g.res.Stellar, g.worldStellar)
	if g.res.Planets != nil {
		copy(g.res.Planets.Data, g.worldPlanets)
	}
}

// ShowEnding 切到結局畫面，顯示 `.VO5` 解出來的那張圖。
//
// **結局的播放順序還沒逆向出來**（哪一張配哪一句、怎麼疊、
// 換場怎麼走都未解，見 `docs/internals/18`）；這裡只負責把一張畫出來。
func (g *Game) ShowEnding(img *asset.Image) {
	g.ending = img
	g.state = StateEnding
}

// New 開一局。有開頭動畫的素材就從動畫開始，沒有就直接進主選單。
func New(res *Resources) (*Game, error) {
	if res.Font == nil || res.Menu == nil || res.CopyProt == nil || res.MainGame == nil {
		return nil, fmt.Errorf("缺少必要的資源（字型與三個畫面的規格）")
	}
	g := &Game{res: res, highlight: -1, mouseX: -1, mouseY: -1, selectedRegion: -1,
		hoverRegion: -1, menuRow: -1, detailAdmiral: -1, battleRegion: -1,
		// 音源是開機偵測的結果，不是每局重來的東西：沒有 MPU-401 就退到
		// OPL（`docs/re/274` §3）。remake 不模擬 MIDI 介面，所以與出貨的
		// 原版一樣一律 ＦＭ——既有的執行期傾印 `ds:8F33` 全是 1，收據
		// `settings1e` 上反白的也是「ＦＭ」那一格。
		music: MusicSettings{Device: MusicFM},
		// 另外三行是純軟體設定，開機就有預設值（動畫開啟／普通／90），
		// 不像音源要等偵測（`docs/re/286` §二）。
		display: newDisplaySettings()}
	g.startNewGame(res.Faction)
	// 電腦那一方走原版的戰略 AI（`docs/spec/029` READY，使用者裁決
	// 2026-09-10）。資源不齊時 `OriginalBrain` 自己回 nil ＝ 不下命令。
	g.SetBrain(g.OriginalBrain())
	if err := g.pickQuestion(); err != nil {
		return nil, err
	}
	if res.Intro != nil {
		if err := g.startIntro(); err != nil {
			return nil, err
		}
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
	g.work = nil
	g.phezzanPending = false
	g.phezzanBattleResolved = false
	g.phezzanRebellions = nil
	g.res.Stellar = sl.Stellar
	g.res.Planets = sl.Planets
	g.turn = sl.Turn
	g.military = sl.Military
	g.taxTotal = sl.TaxTotal
	g.blend = sl.Blend
	g.rate = sl.Rate
	// 造船百分比與七個艦種的分配都在 `WORK.DT?` 裡（`ds:8F92`／`ds:8FD6`，
	// `docs/re/256`、`docs/re/257`）。0 是合法值（S3 的帝國就是 0%），
	// 所以整組照抄，不做「零值就換掉」的補救。
	g.shipPct = sl.ShipPct
	g.shipPctSet = true
	g.pct = sl.BuildPct[g.faction]
	g.foePct = sl.BuildPct[1-g.faction]
	// 徵稅計劃「已支付」那一欄也在 `WORK.DT?` 裡（`docs/re/258`）。
	// 建造中的船艦（`ds:8F9C`）：下個月底交付的就是它（`docs/spec/036`）。
	g.building = sl.Building
	g.paidUpkeep = sl.PaidUpkeep
	g.paidMove = sl.PaidMove
	// 設定畫面那三格跟著存檔走（`docs/re/274` §6）：三格出貨存檔載進來
	// 之後換到的曲子就是 `+0x20` 那個號碼。
	g.music = MusicSettings{Device: sl.MusicDevice, BGM: sl.MusicBGM,
		Battle: sl.MusicBattle}
	// 另外三行同一塊、同一條路（`WORK.DT?` 的 `+0x22`–`+0x25`，
	// `docs/re/286` §四）。三格出貨存檔都是 `01 00 5A 00`。
	g.display = DisplaySettings{Anim: sl.AnimEffect, AnimSpeed: sl.AnimSpeed,
		MsgSpeed: sl.MsgSpeed}
	g.treasury = rules.Treasury{Budget: sl.BudgetStart[g.faction],
		ShipPaid: sl.ShipPaid[g.faction]}
	// 存檔裡就有兩邊的艦隊（`FLEET.DT?` ＝ 執行期那張表，`docs/re/41`）。
	// 讀進來之後 remake 才有敵方的艦隊可以畫、可以推進。
	g.fleets = sl.Fleets(g.res.Weights)
	g.selectedRegion = -1
	g.hoverRegion = -1
	g.state = StateCommandPhase
	// 星域記錄與行星表剛剛才換成存檔那一份，戰略 AI 的地基跟著換，
	// 所以重接一次。
	g.SetBrain(g.OriginalBrain())
}

// pickQuestion 決定防拷問答的題目與答案。
//
// 題目就是規格裡那張照片的編號——原版問的是「照片上那位提督的占領能力」。
//
// ⚠ 這裡讀的是 `+11`（統率力），而畫面問的占領在 `+0x19`。之所以沒差，
// 是因為**原版根本不驗答案**（`docs/re/75`）：`Answer()` 只給文件與
// 測試看，不參與判斷。題目是怎麼挑的仍然未解（`docs/re/74`）。
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

// Faction 回玩家陣營：0 帝國、1 同盟。對戰模式下是**目前輪到的那一方**。
func (g *Game) Faction() int { return g.faction }

// Typed 回防拷問答已經打進去的字串（畫面上還沒畫，見下方的未確認事項）。
func (g *Game) Typed() string { return string(g.typed) }

// mouseDY 見 `ui.MouseDY`：**Game 對外一律用畫面座標**，只有要餵給吃滑鼠
// 座標的東西時才減掉它。兩套座標曾經同時存在，症狀是「命中判定對、
// 游標畫在別的地方」——兩邊各錯一次剛好互相抵銷（`docs/re/52`）。
const mouseDY = ui.MouseDY

// MoveMouse 移動游標，座標是**畫面座標**。主選單會跟著換反白的那一列。
//
// 單純的移動會順便重算左下角那塊寫哪一格（`refreshHoverRegion`）；
// 點擊帶著的那次移動**不算**，所以 `Click`／`RightClick` 走 `moveMouse`。
func (g *Game) MoveMouse(x, y int) {
	g.moveMouse(x, y)
	g.refreshHoverRegion()
}

// moveMouse 只搬游標，不碰左下角那塊。
func (g *Game) moveMouse(x, y int) {
	g.mouseX, g.mouseY = x, y
	if g.state == StateMainMenu {
		g.highlight = screen.RowAt(g.res.Menu, y)
	}
	// 艦隊命令選抵達朝向那一段，指標跟著游標繞目的地轉（`docs/re/122`）。
	g.fleetOrderMouse(x, y)
	// 要塞砲的射程橢圓也跟著游標轉（`docs/re/145`）。
	g.fortressAimMouse(x, y)
}

// refreshHoverRegion 重算左下角那塊要寫哪一格星域（`docs/re/277`）。
//
// ⚠ **從別的畫面「回到」地圖介面的那一次輸入不重算。** 原版讓左下角留著
// 上一張畫面在那裡畫的東西：收據 `dec1` 是從艦隊組成按「決定」回地圖，
// 游標停在 (510,310)（星圖外），而那塊寫的是艦隊那幾張畫上去的首都摘要
// ——照「游標底下那一格」重算會把它擦掉，差 2,354 個像素。所以
// `Click`／`RightClick` 只在**輸入之前就已經在地圖介面**時才重算。
//
// **只在戰略地圖介面是最上層的時候重算**：選單或疊上去的視窗開著的時候
// 一個字都不動，所以那幾張畫面留著的是開起來之前的內容（收據 `cmd1e`：
// 選單開著，滑鼠已經搬到右側面板上，左下角照樣寫著艾流塞拉）。
//
// 游標不在星圖上就是 −1 ＝ 整塊只剩外框（收據 `main1e`：滑鼠在 (600,340)，
// 那塊 0 個非底色像素）。
//
// 呼叫點是三個輸入入口的**結尾**——「這次輸入做完之後還在地圖介面嗎」
// 才是判準。左鍵點命令列把選單開起來時，結尾的狀態已經不是地圖介面，
// 所以不重算，凍住的就是點下去之前游標底下那一格。
//
// ⚠ **同盟那條路凍住的內容與這裡相反，機制還沒解**（`docs/re/278`）：
// 同一條輸入序列，帝國凍著游標底下那一格、同盟整塊是空的；再多按一次
// 星圖上的右鍵或多移動一次滑鼠，兩邊又剛好對調。這裡照帝國那邊的證據做。
func (g *Game) refreshHoverRegion() {
	if g.state != StateCommandPhase {
		return
	}
	g.hoverRegion = -1
	if g.res.Regions == nil ||
		g.mouseX < mapX || g.mouseX >= mapX+mapW ||
		g.mouseY < mapY || g.mouseY >= mapY+mapH {
		return
	}
	if i := asset.RegionAt(g.res.Regions, g.mouseX-mapX, g.mouseY-mapY, tileW, tileH); i >= 0 {
		g.hoverRegion = i
	}
}

// moveBeforePress 在按下之前先讓游標「移動到」那個位置一次。
//
// ⚠ **這一步不是多餘的。** 真的滑鼠不可能在同一瞬間移到新位置又按下去
// ——驅動程式一定先送一個移動事件，遊戲因此在處理按下**之前**就看過
// 新座標。左下角那塊摘要跟著游標走、選單一開就凍住（`docs/re/277`），
// 所以「有沒有先移動」決定凍住的是哪一格。
//
// 少了它，遊戲照樣開得了選單、框與內容逐位元組相同，只是「游標經過
// 那裡」帶起來的副作用整批不見，而畫面看起來完全正常。原版 oracle 那側
// 同一件事在 dosgolem 的 `-click-move-lead`（`docs/spec/004` §4.20.1）
// ——它曾經在改寫命令列時掉過，42 份收據因此重跑出另一個畫面
// （`docs/re/282`）。
func (g *Game) moveBeforePress(x, y int) {
	if g.state == StateIntro {
		return // 動畫期間任何輸入都是「跳過」，沒有游標語意
	}
	g.MoveMouse(x, y)
}

// forcedLeftRegion 回「這個畫面自己要寫進左下角的那一格」，−1 ＝ 不寫
// （那就跟著游標，見 `refreshHoverRegion`）。
//
// ⚠ **寫進去之後會留著。** 原版的左下角是**黏的**：提督與艦隊那一族畫面
// 把自己的首都摘要寫上去，按「決定」回到地圖之後那塊**不會被擦掉**
// （收據 `dec1`，游標停在星圖外的 (510,310)，那塊仍是首都摘要）。
// 所以這個值要存進 `hoverRegion`，不能只在繪製的時候套用——只在繪製時
// 套用的話，回到地圖的那一幀就變回「跟著游標」，差 2,354 個像素。
func (g *Game) forcedLeftRegion(s State) int {
	switch s {
	case StateAdmiralListView, StateStandbyAdmirals, StateAdmiralList,
		StateAdmiralDetail, StateAdmiralBio, StateFleetOrg,
		StateFleetNumber, StateFleetAdjutant, StateFleetComposition,
		StateFormation:
		return rules.CapitalRegion(g.faction)
	}
	return -1
}

// Click 按一下左鍵。
//
// **主選單只從事件回呼收按下**（`docs/re/08`）——那是原版的實作細節，
// 在這裡的意義是：點擊要落在某一列上才算數，不是點到視窗內就好。
func (g *Game) Click(x, y int) {
	onMap := g.state == StateCommandPhase
	g.moveBeforePress(x, y)
	g.click(x, y)
	if onMap {
		g.refreshHoverRegion()
	}
	if r := g.forcedLeftRegion(g.state); r >= 0 {
		g.hoverRegion = r
	}
}

func (g *Game) click(x, y int) {
	if g.state == StateIntro {
		g.SkipIntro()
		return
	}
	g.moveMouse(x, y)
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
	case StateSaveGame:
		// 點下去就直接存到那一格，不問名字——原版在唯讀的原版目錄上
		// 會跳「磁碟片載入時有發生錯誤。」（`docs/spec/010`）。
		if n := g.saveRow(); n >= 0 && g.res.SaveDir != "" {
			g.saveErr = g.SaveSlot(g.res.SaveDir, n)
		}
	case StateShipDelivery:
		// 報告只有一個「確認」鈕，點哪裡都收掉——哪裡按不算數還沒量過。
		//
		// **關掉之後接著跳會計報告與訂造報告**（`docs/re/259`）：原版
		// `0654:0034` 在 `0654:0130`（接收）之後依序呼叫 `0654:0486`
		// （會計）與 `0654:02E0`（訂造），收據 `monthacct`／`monthorder`
		// 各一張。對戰模式的跨月沒量過，那一路照舊。
		if !g.versus && g.res.MainGame != nil {
			g.reportLayer = 0
			g.state = StateMainGame
			return
		}
		g.state = StateCommandPhase
	case StateRegionFlipReport:
		// 滑鼠任一鍵關掉一則（收據 `chain4` 右鍵、`chain5` 左鍵）；
		// 兩份都點在地圖上，其他位置沒量過。
		g.dismissFlip()
	case StateVersusInput:
		// 交棒視窗只有「確認」一顆（`docs/re/76`）。點在框裡但熱區外
		// 什麼都不會發生（收據 `vs3`），所以走 `BoxAt`。
		if spec := g.commandSpec(StateVersusInput); spec != nil {
			if w := spec.Windows["overlay"]; w != nil && w.BoxAt(x, y) == 0 {
				g.reportLayer = 0
				if g.commandSpec(StateVersusAccount) != nil {
					g.state = StateVersusAccount
				} else {
					g.state = StateMainGame
				}
			}
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
			p := g.thisPlan()
			g.plan = &p
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
	case StateAdmiralDetail:
		// 下緣三顆按鈕：經歷／中止／出擊（`docs/re/65`）。
		if g.res.AdmiralPanel == nil {
			return
		}
		switch g.res.AdmiralPanel.ButtonAt(x, y) {
		case "bio":
			if g.commandSpec(StateAdmiralBio) != nil {
				g.state = StateAdmiralBio
			}
		case "cancel":
			g.state = StateStandbyAdmirals
		case "sortie":
			if g.commandSpec(StateFleetOrg) != nil {
				g.state = StateFleetOrg
			}
		}
	case StateFleetComposition:
		// 下緣三顆按鈕：陣形／中止／決定（熱區清單 `2A00`）。
		if p := g.res.CompositionPanel; p != nil {
			switch p.ButtonAt(x, y) {
			case "formation":
				if g.commandSpec(StateFormation) != nil {
					g.state = StateFormation
					return
				}
			case "cancel":
				g.state = StateFleetOrg
				return
			case "decide":
				// 艦隊建成，回戰略地圖（收據 `dec1`）。艦隊記錄怎麼填
				// 見 `docs/internals/33`；remake 目前只記「這一位提督
				// 帶了第幾艦隊」。
				if g.commandSpec(StateCommandPhase) != nil {
					g.fleets = append(g.fleets, rules.Fleet{
						Number:   len(g.fleets) + 1,
						Faction:  g.faction,
						Admiral:  g.detailAdmiral,
						Adjutant: rules.Adjutant(g.admiralRecord(g.detailAdmiral)),
						// 應有組數由提督等級決定（原版 `297B:0066` →
						// `297B:0209`，`docs/re/214`）。**玩家這條路與 AI
						// 建艦隊走同一支** `297B:0066`（兩個呼叫者：
						// `0x0814F` 與 `0x2DE35`），所以規則相同。
						Expected: rules.FleetSlots(
							rules.AdmiralLevel(g.admiralRecord(g.detailAdmiral))),
						// 同一支 `297B:0066` 也寫 `+0x3C = 100`。
						Morale: rules.NewFleetMorale,
						// 編成完停在自己的首都，還沒有移動命令
						// （`docs/re/70` 的收據 `dec1` 就是這個狀態）。
						Region: rules.CapitalRegion(g.faction),
						Target: rules.NoTarget,
					})
					// 右下角清單跟著看首都那一格——剛編成的艦隊停在
					// 那裡，回到地圖之後「艦隊」分頁要多出一列
					// （收據 `dec1`，`docs/re/70`）。
					//
					// 左下角那塊不看這個：`dec1` 的游標按完「決定」停在
					// (510,350)，不在星圖上，所以整塊是空的。
					g.selectedRegion = rules.CapitalRegion(g.faction)
					g.state = StateCommandPhase
					return
				}
			}
		}
		// 點七個艦種列的哪一列，選取條就跳到哪一列（收據 `row3`）。
		if spec := g.commandSpec(g.state); spec != nil {
			if w := spec.Windows["overlay"]; w != nil {
				if i := w.SelectRowAt(x, y); i >= 0 {
					g.selectedRow = i
				}
			}
		}
	case StateFleetOrg:
		// 下緣三顆按鈕：艦隊編號／副官／確認（熱區清單 `10BC`，
		// `docs/re/68`）。座標是熱區，比畫出來的框左右各窄 2 像素。
		if g.res.FleetPanel == nil {
			return
		}
		switch g.res.FleetPanel.ButtonAt(x, y) {
		case "fleet-number":
			if g.commandSpec(StateFleetNumber) != nil {
				g.state = StateFleetNumber
			}
		case "adjutant":
			if g.commandSpec(StateFleetAdjutant) != nil {
				g.state = StateFleetAdjutant
			}
		case "confirm":
			if g.commandSpec(StateFleetComposition) != nil {
				g.selectedRow = 0
				g.state = StateFleetComposition
			}
		}
	case StateStandbyAdmirals:
		// 點清單的一列開那位提督的詳細資料（`docs/re/65`）。
		if i := g.listAdmiralAt(y); i >= 0 && g.commandSpec(StateAdmiralDetail) != nil {
			g.detailAdmiral = i
			g.state = StateAdmiralDetail
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
	case StateSettings:
		// 六組都在這裡：音源／平時配樂／戰鬥配樂三組單選（`docs/re/285`）、
		// 訊息速度的上下箭頭、動畫效果與動畫速度，以及「結束」
		// （`docs/re/286`）。「ＭＩＤＩ Ｉ／Ｆ」與 `BUG!` 還沒接。
		g.settingsClick(x, y)
	case StateBattle:
		g.battleClick(x, y)
	case StateCmdMenu, StateListMenu, StateSysMenu:
		b := commandBar[menuOf(g.state)]
		for n, item := range b.Items {
			top := barRow + n*rowHeight
			if x >= b.X && x < b.X+b.ItemW && y >= top && y < top+rowHeight {
				if item == StateExecute {
					g.execute()
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
	onMap := g.state == StateCommandPhase
	g.moveBeforePress(x, y)
	g.rightClick(x, y)
	if onMap {
		g.refreshHoverRegion()
	}
	if r := g.forcedLeftRegion(g.state); r >= 0 {
		g.hoverRegion = r
	}
}

func (g *Game) rightClick(x, y int) {
	if g.state == StateIntro {
		g.SkipIntro()
		return
	}
	g.moveMouse(x, y)
	// 按在星圖上會順便換掉選中的星域（右下角的艦隊分頁看的是它）。
	//
	// **只有已經在戰略地圖介面時才算**：關月初報告的那兩下右鍵也點在
	// 星圖上（`main1e` 的 `-rclicks 300,200` 兩次），而那張畫面右下角
	// 列的是首都的艦隊。左下角那塊不看這個，看的是游標底下那一格
	// （`refreshHoverRegion`，`docs/re/277`）。
	if g.state == StateCommandPhase && g.res.Regions != nil &&
		x >= mapX && x < mapX+mapW && y >= mapY && y < mapY+mapH {
		if i := asset.RegionAt(g.res.Regions, x-mapX, y-mapY, tileW, tileH); i >= 0 {
			g.selectedRegion = i
		}
	}
	switch g.state {
	case StateShipDelivery:
		// 接收報告也是右鍵關掉的（收據 `mreponlyr`：三張報告全部只用右鍵
		// 就依序退場，第三下之後露出地圖介面）。左鍵那一路在 `Click`
		// 裡，落點還沒量過。
		if !g.versus && g.res.MainGame != nil {
			g.reportLayer = 0
			g.state = StateMainGame
			return
		}
		if g.commandSpec(StateCommandPhase) != nil {
			g.state = StateCommandPhase
		}
	case StateMainGame:
		// 月初報告是**一串兩張**（`docs/re/259`）：會計報告 →（右鍵）
		// 船艦訂造報告 →（右鍵）地圖介面。原版在 `0654:0034` 裡依序
		// 呼叫 `0654:0486`（會計）與 `0654:02E0`（訂造），`main1e` 那兩下
		// 右鍵關的就是這兩張。
		if g.res.ShipOrder != nil {
			g.reportLayer = 1
			g.state = StateShipOrder
			return
		}
		g.reportLayer++
		if g.reportLayer >= 2 && g.commandSpec(StateCommandPhase) != nil {
			g.state = StateCommandPhase
		}
	case StateShipOrder:
		// **右鍵一律到地圖介面**，兩條路進來的都一樣（收據 `orderback`：
		// 會計報告點熱區 → 訂造報告 → 右鍵 → 地圖介面，與報告串那一路
		// `openreports` 的第二下右鍵落到同一張畫面）。
		g.reportLayer = 0
		if g.commandSpec(StateCommandPhase) != nil {
			g.state = StateCommandPhase
		}
	case StateVersusAccount:
		// 對戰模式那張（`vs4`）也是兩下右鍵（收據 `vs5`），而**中間那張
		// 就是船艦訂造報告**——收據 `vs7` 的畫面與 `vsreports` 在兩下
		// 右鍵之間倒的那一張逐位元組相同（`docs/re/260`）。
		if g.commandSpec(StateVersusShipOrder) != nil {
			g.reportLayer = 1
			g.state = StateVersusShipOrder
			return
		}
		g.reportLayer++
		if g.reportLayer >= 2 && g.commandSpec(StateCommandPhase) != nil {
			g.state = StateCommandPhase
		}
	case StateFormation:
		g.state = StateFleetComposition
	case StateFleetNumber, StateFleetAdjutant, StateFleetComposition:
		g.state = StateFleetOrg
	case StateAdmiralBio, StateFleetOrg:
		g.state = StateAdmiralDetail
	case StateAdmiralDetail:
		g.state = StateStandbyAdmirals
	case StateBattle:
		// 右鍵關掉最上面那一層：詳細資料 → 「資料」 → 命令選單 → 選單。
		if g.battleDetail {
			g.battleDetail = false
			return
		}
		if g.battleData {
			g.battleData = false
			return
		}
		if g.MoveTargetPending() {
			g.cancelMoveTarget()
			return
		}
		if g.FormationPickPending() {
			g.cancelFormationPick()
			return
		}
		if g.SupplyPickPending() {
			g.cancelSupplyPick()
			return
		}
		if g.FleetOrderStage() != FleetOrderNone {
			g.cancelFleetOrder()
			return
		}
		if g.FortressTargetPending() {
			g.cancelFortressTarget()
			return
		}
		if g.CommandMenuOpen() {
			g.closeCommandMenu()
			return
		}
		g.battleMenu = false
	case StateCmdMenu, StateListMenu, StateSysMenu:
		g.state = StateCommandPhase
	case StateVersusInput:
		// 交棒視窗按右鍵會怎樣**沒有量過**，先什麼都不做——
		// 這一步不能靠猜，猜錯會讓對戰模式跳掉一方的回合。
	default:
		if _, ok := screenSpecName[g.state]; ok && g.state != StateCommandPhase {
			g.state = StateCommandPhase
		}
	}
}

// Key 送一個按鍵。數字會累積，Enter 送出答案。
//
// **原版不驗答案**：打 `000` 照樣進遊戲（收據 `cbwrong`，`docs/re/75`）。
// 所以 Enter 只要求「有打東西」。`Answer()` 留著給文件與測試看，
// 不參與判斷。
func (g *Game) Key(c byte) {
	if g.state == StateIntro {
		g.SkipIntro()
		return
	}
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
		if len(g.typed) > 0 {
			// 主選單前三列開新遊戲，第四列「取進度」先進存檔畫面
			// （`docs/re/38`、`docs/spec/007`）。
			if g.menuRow == menuRowLoad && g.commandSpec(StateLoadGame) != nil {
				g.state = StateLoadGame
			} else if g.menuRow == menuRowVersus && g.commandSpec(StateVersusInput) != nil {
				// 第 2 列「對戰」：兩個人輪流下命令。先跳帝國的交棒視窗
				// （`docs/re/76`、`docs/spec/016`）。
				g.startVersus()
			} else {
				// 第 0 列帝國、第 1 列同盟（`docs/re/61`）。
				//
				// **只有換邊才重開局**：開局的月底結算在 New() 就跑過
				// 一次了，再跑一次會把成長與收稅疊上去。
				want := 0
				if g.menuRow == menuRowAlliance {
					want = 1
				}
				if want != g.faction {
					g.startNewGame(want)
				}
				g.state = StateMainGame
			}
		}
	}
}

// saveRow 回「存進度」畫面上滑鼠停在第幾列。**不存成狀態**——原版一開畫面
// 紅條就在滑鼠底下那一列，所以它是「現在滑鼠在哪」的函式，不是點出來的。
func (g *Game) saveRow() int {
	if g.res.SlotList == nil {
		return -1
	}
	return g.res.SlotList.SaveRowAt(g.mouseY)
}

// SaveErr 回上一次存檔的結果。
func (g *Game) SaveErr() error { return g.saveErr }

// 主選單四列（`docs/internals/data/screen-main-menu.json`）：
// 帝國／同盟／對戰／取進度。
const (
	menuRowEmpire   = 0
	menuRowAlliance = 1
	menuRowVersus   = 2
	menuRowLoad     = 3
)

// 存檔畫面那 11 列的版面（`docs/spec/007`）。畫面上第 0 列在 y ＝ 122，
// 一列 20 px。
//
// **收據裡的點擊座標比這裡少 40**：dosgolem 的虛擬滑鼠座標是畫面座標
// 減 40（`docs/re/09`），所以量到的 242／262／282 對應畫面的 282／302／322。
// `Game` 這一層一律用畫面座標。
const (
	slotRowTop    = 122
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
	// 載入當下那次未安定度重算讀的也是剛載入的 WORK：回合數、Blend、
	// Rate 與兩邊的內政專任參謀（統治者）都不能沿用新遊戲固定值
	// （`docs/spec/006`、`docs/spec/042`）。
	params := g.res.UnrestParams
	params.Turn = sl.Turn
	params.Blend = sl.Blend
	params.Rate = sl.Rate
	for f := 0; f < 2; f++ {
		if i := rules.StaffAdmiral(sl.Work, rules.PostDomestic, f); i >= 0 {
			params.Ruler[f] = i
		}
	}
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
	// `WORK` 尚未全部提升成結構欄位；費沙旗標與動態參謀都直接改可寫
	// 複本，因此存檔前先整段同步，未知位元組仍維持原樣。
	if work := g.workRecord(); len(work) >= rules.WorkRecord {
		g.slot.Work = append(g.slot.Work[:0], work[:rules.WorkRecord]...)
	}
	g.slot.Turn = g.turn
	g.slot.Military = g.military
	g.slot.TaxTotal = g.taxTotal
	g.slot.Blend = g.blend
	g.slot.Rate = g.rate
	g.slot.ShipPct = g.shipPct
	g.slot.BuildPct[g.faction] = g.pct
	g.slot.BuildPct[1-g.faction] = g.foePct
	g.slot.Building = g.building
	g.slot.PaidUpkeep = g.paidUpkeep
	g.slot.PaidMove = g.paidMove
	// 那三格音樂設定。**新遊戲的底稿在這一格上是 0**（`work-796-01.json`
	// 量自第 71,200,000 步，那時候音源偵測還沒跑），而原版存檔的時候
	// `ds:8F33` 已經是 1——三格出貨存檔的 `+0x1F` 都是 1，所以照現在的值寫。
	g.slot.MusicDevice = g.music.Device
	g.slot.MusicBGM = g.music.BGM
	g.slot.MusicBattle = g.music.Battle
	g.slot.AnimEffect = g.display.Anim
	g.slot.AnimSpeed = g.display.AnimSpeed
	g.slot.MsgSpeed = g.display.MsgSpeed
	// 原定軍事費與已付建造費目前只留玩家那一方（`monthEnd` 只填那一格）；
	// 電腦那一方沿用底稿裡的值。
	g.slot.BudgetStart[g.faction] = g.treasury.Budget
	g.slot.ShipPaid[g.faction] = g.treasury.ShipPaid
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
	case StateBattle:
		return g.drawBattle()
	case StateCopyProtection:
		face, _ := g.res.Faces.Decode(g.res.CopyProt.Face.Index)
		fb, _ := screen.CopyProtection(g.res.CopyProt, g.res.Font, face, string(g.typed))
		return fb
	case StateMainGame:
		// 底板上的首都摘要與報告裡的署名／金額都是算出來的
		// （`docs/spec/011`）——規格裡只有版面。
		g.liveCursor(g.res.MainGame)
		fb, _ := screen.MainGameWithBoard(g.res.MainGame, g.res.Font, g.playerStarMap(),
			g.res.Faces, g.res.Cursor, g.drawBoardPanels)
		g.drawReport(fb)
		return fb
	case StateShipOrder:
		g.applyStaffFaces(g.res.ShipOrder)
		fb, _ := screen.MainGameWithBoard(g.res.ShipOrder, g.res.Font, g.playerStarMap(),
			g.res.Faces, g.res.Cursor, g.drawBoardPanels)
		g.drawSpecStaff(fb, g.res.ShipOrder)
		g.drawShipOrder(fb)
		return fb
	case StateIntro:
		return screen.IntroScreen(g.introPage)
	case StateEnding:
		return screen.Ending(g.ending)
	case StateAdmiralList:
		g.setListRows(g.res.AdmiralList)
		g.applyStaffFaces(g.res.AdmiralList)
		fb, _ := screen.MainGameWithBoard(g.res.AdmiralList, g.res.Font, g.playerStarMap(),
			g.res.Faces, g.res.Cursor, g.drawBoardPanels)
		g.drawSpecStaff(fb, g.res.AdmiralList)
		return fb
	case StateBudgetPlan:
		g.applyStaffFaces(g.res.BudgetPlan)
		fb, _ := screen.MainGameWithBoard(g.res.BudgetPlan, g.res.Font, g.playerStarMap(),
			g.res.Faces, g.res.Cursor, g.drawBoardPanels)
		g.drawSpecStaff(fb, g.res.BudgetPlan)
		g.drawBudget(fb)
		return fb
	case StateBuildPlan:
		g.applyStaffFaces(g.res.BuildPlan)
		fb, _ := screen.MainGameWithBoard(g.res.BuildPlan, g.res.Font, g.playerStarMap(),
			g.res.Faces, g.res.Cursor, g.drawBoardPanels)
		g.drawSpecStaff(fb, g.res.BuildPlan)
		if p := g.res.BuildPlanPanel; p != nil {
			p.Draw(fb, g.res.Font, g.monthReport, nil, nil)
		}
		if g.plan != nil {
			screen.DrawBuildPlan(g.res.BuildPlan, fb, g.res.Font, g.plan)
		}
		return fb
	default:
		if spec := g.commandSpec(g.state); spec != nil {
			g.setListRows(spec)
			g.liveCursor(spec)
			g.setHoverRow(spec)
			g.applyStaffFaces(spec)
			if g.state == StateSettings {
				g.applySettings(spec)
			}
			// 從狀態算出來的那幾塊要畫在底板之後、疊上去的視窗之前
			// （`docs/spec/002`）。
			fb, _ := screen.MainGameLayered(spec, g.res.Font, g.starMapFor(spec),
				g.res.Faces, g.res.Cursor, g.drawBoardPanels, g.afterWindow)
			g.drawBudget(fb)
			g.drawProduction(fb)
			g.drawDelivery(fb)
			g.drawSlotList(fb)
			g.drawReport(fb)
			g.drawAdmiralDetail(fb)
			g.drawAdmiralBio(fb)
			g.drawSpecStaff(fb, spec)
			return fb
		}
		fb, _ := screen.MainMenuAt(g.res.Menu, g.res.Font, g.res.FrameTiles,
			g.highlight, g.res.Cursor, g.mouseX, g.mouseY-mouseDY)
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
	// 0. `2703:03AF` 的第一件事：`ds:8FF8[陣營] ← ds:8FD4[陣營]`（`docs/re/252`）。
	//    原版是逐座位各做自己那一格；remake 兩邊一起結算，就兩格一起抄。
	g.rate = g.blend
	// 1. 逐星域結算。成長是就地寫回的，一個月只能跑一次，所以兩個陣營
	//    的徵稅額要在同一趟裡拿。
	collected := rules.MonthEndTotals(g.res.Stellar, g.res.Planets, g.res.Growth,
		g.recomputeAllTaxRates)
	// 2. 照百分比切給造船，各列只買得起整數艘，零頭退回；扣掉實際花掉的。
	//    百分比與各列的分配都是每個陣營自己的（`shipPct`／`buildPctFor`，
	//    原版 `ds:8F92`／`ds:8FD6`）——電腦那一方在回合收尾按自己的總稅收
	//    設百分比，各列則寫回一張固定表（`docs/re/257`）。
	for f := 0; f < 2; f++ {
		// 建造費的基數是「軍事費 ＋ 總稅收 − 維持費 − 移動經費」，
		// 取的是**結算前**的軍事費——原版算 `ds:8F5A` 的那一趟
		// （`090A:0E88`）跑在 `軍事費 += 總稅收` 之前（`docs/re/256`）。
		b := rules.Budget{Treasury: before[f], Tax: collected[f],
			ShipPct: g.factionShipPct(f), Upkeep: g.fleetUpkeep(f),
			MoveCost: g.fleetMoveCost(f)}
		budget := b.Military()
		plan := rules.BuildPlan{Cost: b.ShipCost(), Pct: g.buildPctFor(f),
			Faction: f, StaffOps: g.staffOps(f)}
		paid := plan.Paid()
		g.military[f] = budget - paid
		// **兩個陣營都留一份**：下個月底各自交付到自己首都
		// （`docs/spec/036`）。
		g.building[f] = plan.Building()
		if f != g.faction {
			continue
		}
		g.treasury = rules.Treasury{Budget: budget, ShipPaid: paid}
		g.monthReport = b
	}
	// `ds:8F7A` 與 `ds:8F8A` 在 `0654:0034` 的尾段歸零（`docs/re/255`）。
	g.paidUpkeep = [2]int{}
	g.paidMove = [2]int{}
	// 3. 結算完之後重算下個月的預估徵稅額——徵稅計劃那張印的是這個，
	//    與剛剛實收的那筆不是同一個數（`docs/spec/003`）。
	g.refreshTaxEstimate()
}

// refreshTaxEstimate 重算兩個陣營下個月的預估徵稅額。
//
// 原版的 `090A:0E6D`（`2703:02E5` ＋ `090A:0E88`）**每個座位回合都跑**，
// 而且是冪等的——收據 `monthgate7`／`monthgateB` 各跑三趟，20 筆逐筆
// 累加值完全相同（`docs/spec/035`、`docs/re/255`）。真正改狀態的是月底
// 那一趟旗標 1 的 `2703:03AF`。
func (g *Game) refreshTaxEstimate() {
	for f := 0; f < 2; f++ {
		g.taxTotal[f] = g.estimateTaxFor(f)
	}
}

// estimateTax 回下個月的預估徵稅額 ＝ 自己擁有的星域的徵稅額相加。
//
// 擁有者要**照行星現算**（`docs/re/17`）——星域記錄裡那一格是檔案裡的
// 原值，載入時會被重算，直接拿來用會多算十個星域。
func (g *Game) estimateTax() int { return g.estimateTaxFor(g.faction) }

// estimateTaxFor 是同一件事，但算指定陣營的——原版的 `ds:8F6A` 是
// **每個陣營一格**（`docs/internals/27`），戰略 AI 的建艦隊閘門看的是
// 電腦那一方的（`docs/re/202`）。
func (g *Game) estimateTaxFor(faction int) int {
	if g.res.Planets == nil {
		return 0
	}
	sum := 0
	for i := 0; (i+1)*rules.RegionRecord <= len(g.res.Stellar); i++ {
		if g.res.Planets.RegionOwner(i) != faction {
			continue
		}
		sum += rules.RegionTax(g.res.Stellar[i*rules.RegionRecord : (i+1)*rules.RegionRecord])
	}
	return sum
}

// drawBoardPanels 畫底板上那幾塊算出來的內容。命令階段那一批畫面共用
// 同一塊底板，右側面板在每一張上都看得到（被疊上去的視窗蓋住的部分除外）。
func (g *Game) drawBoardPanels(fb *gfx.Framebuffer) {
	g.drawLeftSummary(fb)
	g.drawFleetRows(fb)
	switch g.state {
	case StateMainGame, StateShipOrder, StateBuildPlan, StateAdmiralList, StateBudgetPlan:
		if c, _, ok := g.capitalSummary(); ok {
			g.res.CapitalPanel.DrawOpening(fb, g.res.Font, g.res.RegionPanel, c, g.faction)
		}
		return
	}
	if _, ok := screenSpecName[g.state]; !ok {
		return
	}
	// 對戰交棒的當下只有陣營那一行與按鈕（`docs/spec/016`）。
	// 接著那張月初報告是「交棒的底板 ＋ 單人那張的首都摘要」——
	// 兩件事都要，所以兩個都畫（收據 `vs4`）。
	if g.state == StateVersusInput || g.state == StateVersusAccount ||
		g.state == StateVersusShipOrder {
		c, _, ok := g.capitalSummary()
		if !ok {
			return
		}
		g.res.CapitalPanel.DrawHandover(fb, g.res.Font, c, g.faction)
		if g.state != StateVersusInput {
			g.res.CapitalPanel.DrawOpening(fb, g.res.Font, g.res.RegionPanel, c, g.faction)
		}
		return
	}
	// 月底結算的當下右側面板的字整塊是空的，只寫「戰略模式處理中」——
	// 算出來的那幾塊不畫，但按鈕與分頁鈕還在。
	if g.state == StateShipDelivery {
		if g.res.CapitalPanel != nil {
			g.res.CapitalPanel.DrawChrome(fb, g.res.Font)
		}
		screen.DrawSettlementScrollArrows(fb)
		return
	}
	// 易主報告是同一塊底板加上兩行字與一張頭像（`docs/spec/034`）。
	if g.state == StateRegionFlipReport {
		if g.res.CapitalPanel != nil {
			g.res.CapitalPanel.DrawChrome(fb, g.res.Font)
		}
		g.drawFlipReport(fb)
		return
	}
	g.drawCapitalPanel(fb)
}

// capitalSummary 回自己首都那一格的摘要，第二個回傳值說算不算得出來。
func (g *Game) capitalSummary() (rules.CapitalSummary, int, bool) {
	if g.res.CapitalPanel == nil || g.res.RegionPanel == nil || g.res.Planets == nil {
		return rules.CapitalSummary{}, 0, false
	}
	r := rules.CapitalRegion(g.faction)
	o := r * rules.RegionRecord
	if o+rules.RegionRecord > len(g.res.Stellar) {
		return rules.CapitalSummary{}, 0, false
	}
	return rules.SummariseCapital(g.res.Stellar[o:o+rules.RegionRecord], g.res.Planets,
		r, g.faction, g.turn, rules.Epoch,
		g.military[g.faction], g.taxTotal[g.faction]), r, true
}

// liveCursor 讓戰略地圖介面的游標跟著滑鼠跑。
//
// 規格裡的 `mouse` 記的是**量那張收據時**滑鼠停在哪，對疊上去的那些畫面
// 是對的（開畫面的那一下點在哪，游標就停在哪）；但地圖介面本身是玩家
// 待最久的地方，滑鼠愛停哪就停哪。按一次「實行」回到地圖，原版的游標
// 在剛剛點「實行」的位置，remake 卻畫在收據當初 `-mouse-at` 搬過去的
// (600,300)——兩隻箭頭，223 個像素。
//
// 月初會計報告也一樣：規格量自 `key1`，游標停在主選單第 0 列（帝國）；
// 點第 1 列開同盟的局，原版的游標低一列。提督詳細資料同理——
// 規格量自點第 0 列那一次，點別列游標就在別處。
//
// 只對這兩張這樣做：疊上去那些畫面的規格還帶著游標的裁切範圍與
// 「游標之後再畫一次」的補丁，那幾件事綁著當初那個位置，還沒模型化。
func (g *Game) liveCursor(spec *screen.MainGameSpec) {
	switch g.state {
	case StateCommandPhase, StateMainGame, StateAdmiralDetail, StateAdmiralBio,
		StateFleetOrg, StateFleetNumber, StateFleetAdjutant,
		StateFleetComposition, StateFormation,
		// 這五張的收據末尾都有一次 `-mouse-at`（搬到 (600,340)），
		// 而畫面上的箭頭跟著搬過去了：`info1e`／`lst1e`／`lst2e`／
		// `lst3e`／`settings1e` 與同盟的五張，十份都是同一個結論。
		StateInfo, StateFleetList, StateAdmiralListView, StateEnemyList,
		StateSettings,
		// 交棒視窗的游標停在剛剛按下確認／實行的地方：`vs3` 在主選單
		// 那一列 (300,250)，`vs5`／`vs6` 在「實行」那一格 (560,125)。
		StateVersusInput, StateVersusAccount:
	default:
		return
	}
	if spec.Mouse == nil || g.mouseY < 0 {
		return
	}
	spec.Mouse.X, spec.Mouse.Y = g.mouseX, g.mouseY-mouseDY
}

// afterWindow 在每一層畫完之後補上那一層算出來的內容。
//
// 艦隊編成那張的姓名與能力值屬於 `overlay` 那一層：等整張畫完再補，
// 疊在上面的艦隊編號清單就會被戳出一排洞（`docs/spec/013`）。
func (g *Game) afterWindow(name string, fb *gfx.Framebuffer) {
	if name == "overlay" {
		g.drawFleetOrg(fb)
		g.drawFleetComposition(fb)
		// 訂造報告那七列要畫在**游標之前**——收據 `vs7` 的游標正好壓在
		// 「戰艦」那一列的數字上，順序反了就差 32 個像素。
		g.drawShipOrder(fb)
	}
}

// drawFleetComposition 畫艦隊組成裡算出來的那幾塊（`docs/spec/015`）。
//
// 「待命部隊」是**該星域**的艦艇艘數除以一單位幾艘（`docs/re/69`）；
// 開局那一支艦隊站在自己的首都。
func (g *Game) drawFleetComposition(fb *gfx.Framebuffer) {
	if g.res.CompositionPanel == nil {
		return
	}
	// 陣形那一層疊在這一張上面，底下這一層照畫。
	if g.state != StateFleetComposition && g.state != StateFormation {
		return
	}
	region := rules.CapitalRegion(g.faction)
	o := region * rules.RegionRecord
	if region < 0 || o+rules.RegionRecord > len(g.res.Stellar) {
		return
	}
	ships := rules.RegionShips(g.res.Stellar[o:o+rules.RegionRecord], g.faction)
	v := screen.CompositionView{
		Ships:    ships,
		PerUnit:  rules.ShipsPerUnit[:],
		Mobility: rules.NewFleetMorale,
		Groups:   g.compositionGroups(ships),
		Slots: rules.FleetSlots(
			rules.AdmiralLevel(g.admiralRecord(g.detailAdmiral))),
	}
	p := g.res.CompositionPanel
	p.Draw(fb, g.res.Font, v)
	g.drawCompositionGrid(fb, v)
}

// compositionGroups 是這張畫面上「已經編成」的那幾組。
//
// 進這張畫面時原版只填了**旗艦那一組**：`297B:00BD` 那一批寫入把
// `+0x0F` 設成 1、`+0x12`／`+0x13`（第 0 組的艦種與隊形槽）設成
// 0 與 20（`docs/re/70`）。艦種取提督 `+0x08` 的低四位、隊形槽取
// 編制樣板第 0 組的 `[3]`（十六個類別鍵都是 20，`docs/re/224`）。
//
// 艘數是 `min(滿編, 星域待命艘數)`（原版 `2AD2:0029` 從待命部隊拿，
// `docs/re/215`）；收據 `fok` 的瓦爾哈拉有 75000 艘戰艦，所以那一組
// 是滿編、字母印色 7。
func (g *Game) compositionGroups(ships []int) []screen.CompositionGroup {
	kind := rules.AdmiralAbility(g.admiralRecord(g.detailAdmiral), 0x08) & 0x0F
	if kind < 0 || kind >= len(rules.ShipsPerUnit) {
		return nil
	}
	full := rules.ShipsPerUnit[kind]
	have := full
	if kind < len(ships) && ships[kind] < full {
		have = ships[kind]
	}
	return []screen.CompositionGroup{{
		Kind: kind, Slot: rules.FlagshipFormationSlot, Ships: have, Full: full,
	}}
}

// drawCompositionGrid 鋪格盤那 9 × 9 的圖塊（`docs/re/67`、`docs/internals/56`）：
// `(欄 + 列)` 為偶數的 41 格才畫，有部隊的那一格畫 `Unit ＋ 艦種`。
func (g *Game) drawCompositionGrid(fb *gfx.Framebuffer, v screen.CompositionView) {
	spec := g.commandSpec(g.state)
	if spec == nil || g.res.Sprites == nil {
		return
	}
	w := spec.Windows["overlay"]
	if w == nil || w.Grid == nil {
		return
	}
	gr := w.Grid
	empty, unit := gr.Empty, gr.Unit
	if g.faction == 1 && gr.UnitAlliance > 0 {
		empty, unit = gr.EmptyAlliance, gr.UnitAlliance
	}
	// 隊形槽 → 欄列 → 圖塊。原版是 41 個顯示物件，各自被
	// `0FF4:0EF1(隊形槽, 艦種)` 換掉圖形（`docs/internals/56`）。
	tiles := map[[2]int]int{}
	for _, u := range v.Groups {
		col, row, ok := g.res.CompositionPanel.GridCell(u.Slot)
		if !ok {
			continue
		}
		tiles[[2]int{col, row}] = unit + u.Kind
	}
	for row := 0; row < gr.Rows; row++ {
		for col := 0; col < gr.Cols; col++ {
			if (col+row)%2 != 0 {
				continue
			}
			tile := empty
			if t, ok := tiles[[2]int{col, row}]; ok {
				tile = t
			}
			if tile < 0 || tile >= len(g.res.Sprites) {
				continue
			}
			img := g.res.Sprites[tile]
			for y := 0; y < img.H; y++ {
				for x := 0; x < img.W; x++ {
					fb.Set(gr.X+col*gr.Tile+x, gr.Y+row*gr.Tile+y, img.Pix[y*img.W+x])
				}
			}
		}
	}
}

// setHoverRow 把「滑鼠停在哪一列」告訴有反白條的那個視窗
// （艦隊編號那 16 列、副官那 2 列，`docs/re/67`）。
//
// 反白條是遊戲畫的，不在規格的 `lines` 裡：換一個滑鼠位置就換一列，
// 量進規格就綁死在收據那一列上了。
func (g *Game) setHoverRow(spec *screen.MainGameSpec) {
	for _, w := range spec.Windows {
		if w.Hover != nil {
			w.SetHoverRow(w.HoverRowAt(g.mouseX, g.mouseY))
		}
		if w.Select != nil {
			// 艦隊組成開畫面時選在第 0 列（收據 `fok`，假設待驗：
			// 只有一張收據，還沒看過怎麼換列）。
			w.SetSelected(g.selectedRow)
		}
	}
}

// drawSlotList 畫存檔畫面那 11 列（`docs/spec/008`）。
func (g *Game) drawSlotList(fb *gfx.Framebuffer) {
	if g.res.SlotList == nil || len(g.res.SaveIndex) == 0 {
		return
	}
	switch g.state {
	case StateLoadGame:
		g.res.SlotList.Draw(fb, g.res.Font, g.res.SaveIndex, rules.Epoch)
	case StateSaveGame:
		// 條子把那一列蓋掉，字要重畫在條子上，游標再蓋一次
		// ——原版的游標永遠在最上面。
		row := g.saveRow()
		g.res.SlotList.DrawSaveBar(fb, row)
		g.res.SlotList.DrawSaveRow(fb, g.res.Font, row)
		if g.res.Cursor != nil {
			g.res.Cursor.Draw(fb, g.mouseX, g.mouseY-mouseDY)
		}
	}
}

// drawBudget 畫預算計劃／徵稅計劃的數字（`docs/spec/003`）。
func (g *Game) drawBudget(fb *gfx.Framebuffer) {
	if g.res.BudgetPanel == nil {
		return
	}
	switch g.state {
	case StateTaxPlan:
		g.res.BudgetPanel.Draw(fb, g.res.Font, g.currentTreasury(), g.nextBudget())
	case StateBudgetPlan:
		// 月初報告那一張：本月份那一欄還沒有數字（全 0）。
		g.res.BudgetPanel.Draw(fb, g.res.Font, rules.Treasury{}, g.monthReport)
	}
}

// nextBudget 是次月份的預算：國庫與預估徵稅額都取現值，維持費與移動經費
// 現算（原版 `090A:0E88` 每回合重算一次，`docs/spec/035`）。
func (g *Game) nextBudget() rules.Budget {
	f := g.faction
	return rules.Budget{Treasury: g.military[f], Tax: g.taxTotal[f],
		ShipPct: g.factionShipPct(f), Upkeep: g.fleetUpkeep(f),
		MoveCost: g.fleetMoveCost(f)}
}

// currentTreasury 是**現在**的財政：原定軍事費與已付的建造費在月底定下來，
// 已付維持費與移動經費是這個月累計到現在的，「本次預定」兩格現算
// （原版 `ds:8F72`／`ds:8F82` 每個回合由 `297B:10DE` 重算，`docs/re/258`）。
func (g *Game) currentTreasury() rules.Treasury {
	t := g.treasury
	f := g.faction
	if f < 0 || f > 1 {
		return t
	}
	t.FleetUpkeep = g.paidUpkeep[f]
	t.Movement = g.paidMove[f]
	t.NextUpkeep = g.fleetUpkeep(f)
	t.NextMovement = g.fleetMoveCost(f)
	return t
}

// buildPctFor 回一個陣營切給七個艦種的百分比。玩家那一份是可以改的
// `pct`（生產計劃那張畫面），電腦那一份由它自己在回合收尾設
// （`docs/re/257`）。對戰模式兩邊都是玩家，一律用 `pct`。
func (g *Game) buildPctFor(faction int) [rules.BuildPlanRows]int {
	if g.versus || faction == g.faction {
		return g.pct
	}
	if !g.shipPctSet {
		return rules.OpeningPct
	}
	return g.foePct
}

// factionShipPct 回一個陣營的造船百分比。`shipPct` 是零值時（direct-entry
// 的測試沒走過 `startNewGame`）沿用開局值，不要靜靜地變成 0%。
func (g *Game) factionShipPct(faction int) int {
	if faction < 0 || faction > 1 || !g.shipPctSet {
		return g.res.ShipPct
	}
	return g.shipPct[faction]
}

// thisPlan 是**這個月**的船艦建造計劃：月底結算算出來的建造費照百分比
// 切出來的那一份。訂造報告與建造計劃這兩張印的都是它。
func (g *Game) thisPlan() rules.BuildPlan {
	return rules.BuildPlan{Cost: g.monthReport.ShipCost(), Pct: g.pct,
		Faction: g.faction, StaffOps: g.staffOps(g.faction)}
}

// nextPlan 是次月份的船艦建造計劃：預估軍事費照百分比切出來的那一份。
func (g *Game) nextPlan() rules.BuildPlan {
	b := g.nextBudget()
	return rules.BuildPlan{Cost: b.ShipCost(), Pct: g.pct,
		Faction: g.faction, StaffOps: g.staffOps(g.faction)}
}

// drawDelivery 畫月底船艦接收報告裡算出來的那幾塊。
func (g *Game) drawDelivery(fb *gfx.Framebuffer) {
	if g.state != StateShipDelivery || g.res.DeliveryPanel == nil {
		return
	}
	name, rank := g.staffAdmiral(rules.PostMilitary)
	g.res.DeliveryPanel.Draw(fb, g.res.Font, g.reportYear, g.reportMonth,
		g.delivered, name, rank)
}

// drawReport 畫主畫面會計報告裡算出來的那幾塊（`docs/spec/011`）：
// 署名的參謀、他的頭像與五個金額。署名的是**內政專任參謀**。
func (g *Game) drawReport(fb *gfx.Framebuffer) {
	if g.state != StateMainGame && g.state != StateVersusAccount {
		return
	}
	if g.res.ReportPanel == nil {
		return
	}
	p := g.res.ReportPanel
	name, rank := g.staffAdmiral(rules.PostDomestic)
	p.Draw(fb, g.res.Font, g.monthReport, name, rank)
	if g.res.Faces == nil {
		return
	}
	i := rules.StaffAdmiral(g.workRecord(), rules.PostDomestic, g.faction)
	img, err := g.res.Faces.Decode(i)
	if err != nil {
		return
	}
	for y := 0; y < img.H; y++ {
		for x := 0; x < img.W; x++ {
			fb.Set(p.Face.X+x, p.Face.Y+y, img.Pix[y*img.W+x])
		}
	}
}

// applyStaffFaces 把規格宣告 `staff_post` 的頭像換成那位幕僚本人。
//
// 規格量自帝國的收據；同一張畫面換一個陣營署名的人就不同，頭像跟著換
// （`docs/spec/011`）。要在畫視窗之前做——頭像是視窗自己畫的。
func (g *Game) applyStaffFaces(spec *screen.MainGameSpec) {
	if spec == nil {
		return
	}
	for _, w := range spec.Windows {
		if w == nil || w.Face == nil || w.Face.StaffPost == nil {
			continue
		}
		if i := rules.StaffAdmiral(g.workRecord(), *w.Face.StaffPost, g.faction); i >= 0 {
			w.Face.Index = i
		}
	}
}

// drawSpecStaff 把規格宣告的署名（姓名 ＋ 軍階）與「兩個陣營印不同字」
// 的那幾行畫上去。
func (g *Game) drawSpecStaff(fb *gfx.Framebuffer, spec *screen.MainGameSpec) {
	if spec == nil {
		return
	}
	for _, w := range spec.Windows {
		if w == nil {
			continue
		}
		w.DrawFactionText(fb, g.res.Font, g.faction)
		if w.Staff == nil {
			continue
		}
		name, rank := g.staffAdmiral(w.Staff.Post)
		if len(name) == 0 {
			continue
		}
		screen.DrawStaffLine(fb, g.res.Font, w.Staff.X, w.Staff.Y, name, rank, w.Staff.Ink)
	}
}

// drawShipOrder 畫「船艦訂造報告」裡算出來的部分：標題的年月與七列的
// 造艦數、部隊數。版面與月底的接收報告**完全相同**，所以共用同一份
// （`docs/internals/data/delivery-panel.json`）；署名與頭像走規格的
// `staff`／`staff_post`。
func (g *Game) drawShipOrder(fb *gfx.Framebuffer) {
	if g.res.DeliveryPanel == nil {
		return
	}
	if g.state != StateShipOrder && g.state != StateVersusShipOrder {
		return
	}
	c, _, ok := g.capitalSummary()
	if !ok {
		return
	}
	p := g.thisPlan()
	g.res.DeliveryPanel.Draw(fb, g.res.Font, c.Year, c.Month, p.Building(), nil, nil)
}

// drawAdmiralDetail 畫提督詳細資料裡跟著提督走的那幾塊（`docs/spec/012`）。
//
// 上半（姓名、軍階、頭像、現在位置、旗艦）在經歷那張畫面也看得到——
// 經歷的視窗只蓋住 y ≥ 200，上面那塊是同一份。下半（等級、經驗值與
// 十個能力值）只有詳細資料那張有。
func (g *Game) drawAdmiralDetail(fb *gfx.Framebuffer) {
	if g.res.AdmiralPanel == nil {
		return
	}
	if g.state != StateAdmiralDetail && g.state != StateAdmiralBio {
		return
	}
	i := g.detailAdmiral
	if i < 0 || (i+1)*rules.AdmiralRecord > len(g.res.Admiral) {
		return
	}
	p := g.res.AdmiralPanel
	rec := g.res.Admiral[i*rules.AdmiralRecord : (i+1)*rules.AdmiralRecord]
	v := screen.AdmiralView{Record: rec, Level: rules.AdmiralLevel(rec)}
	if i < len(g.res.AdmiralNames) {
		v.Name = g.res.AdmiralNames[i]
	}
	v.Rank = g.rankCodes(rules.RankOf(rules.AdmiralLevel(rec)))
	// 待命中的提督站在自己陣營的首都（三張收據都是瓦爾哈拉，假設待驗）。
	if r := rules.CapitalRegion(rules.AdmiralFaction(rec)); r >= 0 {
		o := r * rules.RegionRecord
		if o+rules.RegionRecord <= len(g.res.Stellar) {
			v.Location = rules.RegionName(g.res.Stellar[o : o+rules.RegionRecord])
		}
	}
	if slot := rules.FlagshipSlot(rec, i); slot >= 0 {
		v.Flagship = g.res.Flagships.Name(rules.AdmiralFaction(rec), slot)
	}
	p.Draw(fb, g.res.Font, v, g.state == StateAdmiralDetail)
	if g.res.Faces == nil {
		return
	}
	if img, err := g.res.Faces.Decode(i); err == nil {
		for y := 0; y < img.H; y++ {
			for x := 0; x < img.W; x++ {
				fb.Set(p.Face.X+x, p.Face.Y+y, img.Pix[y*img.W+x])
			}
		}
	}
}

// drawFleetOrg 畫艦隊編成裡跟著提督與副官走的那幾塊（`docs/spec/013`）。
//
// 副官是提督記錄的 `+0x07`（`docs/re/66`）；`0xFF` ＝ 沒有副官，那位也
// 沒有旗艦、帶不了艦隊，走不到這張畫面。
func (g *Game) drawFleetOrg(fb *gfx.Framebuffer) {
	if g.res.FleetPanel == nil {
		return
	}
	// 艦隊編號與副官那兩個小視窗疊在這一張上面，底下這一層照畫。
	switch g.state {
	case StateFleetOrg, StateFleetNumber, StateFleetAdjutant:
	default:
		return
	}
	i := g.detailAdmiral
	rec := g.admiralRecord(i)
	if rec == nil {
		return
	}
	adj := rules.Adjutant(rec)
	arec := g.admiralRecord(adj)
	if arec == nil {
		return
	}
	p := g.res.FleetPanel
	v := screen.FleetView{
		AdmiralRecord:  rec,
		AdjutantRecord: arec,
		AdmiralLevel:   rules.AdmiralLevel(rec),
		AdjutantLevel:  rules.AdmiralLevel(arec),
	}
	if i < len(g.res.AdmiralNames) {
		v.AdmiralName = g.res.AdmiralNames[i]
	}
	if adj < len(g.res.AdmiralNames) {
		v.AdjutantName = g.res.AdmiralNames[adj]
	}
	v.AdmiralRank = g.rankCodes(rules.RankOf(v.AdmiralLevel))
	v.AdjutantRank = g.rankCodes(rules.RankOf(v.AdjutantLevel))
	slot := rules.FlagshipSlot(rec, i)
	if slot >= 0 {
		v.Flagship = g.res.Flagships.Name(rules.AdmiralFaction(rec), slot)
		// 艦隊編號 ＝ 旗艦的欄位 ＋ 1（假設待驗，兩個資料點：`docs/re/66`）。
		v.Number = slot + 1
	}
	p.Draw(fb, g.res.Font, v)
	if g.res.Faces == nil {
		return
	}
	for _, f := range []struct {
		index, x, y int
	}{
		{i, p.Admiral.Face.X, p.Admiral.Face.Y},
		{adj, p.Adjutant.Face.X, p.Adjutant.Face.Y},
	} {
		img, err := g.res.Faces.Decode(f.index)
		if err != nil {
			continue
		}
		for y := 0; y < img.H; y++ {
			for x := 0; x < img.W; x++ {
				fb.Set(f.x+x, f.y+y, img.Pix[y*img.W+x])
			}
		}
	}
}

// admiralRecord 回第 i 位提督在 `ADMIRAL.DT*` 裡的那 28 個位元組；
// 索引不合法（含 `0xFF` ＝ 沒有這個人）回 nil。
func (g *Game) admiralRecord(i int) []byte {
	if i < 0 || (i+1)*rules.AdmiralRecord > len(g.res.Admiral) {
		return nil
	}
	return g.res.Admiral[i*rules.AdmiralRecord : (i+1)*rules.AdmiralRecord]
}

// drawAdmiralBio 把那位提督的人物傳排進經歷視窗（`docs/spec/012`）。
func (g *Game) drawAdmiralBio(fb *gfx.Framebuffer) {
	if g.state != StateAdmiralBio || g.res.BioPanel == nil {
		return
	}
	i := g.detailAdmiral
	if i < 0 || i >= len(g.res.Biographies) {
		return
	}
	g.res.BioPanel.Draw(fb, g.res.Font, g.res.Biographies[i])
	// 內文是在視窗畫完之後才補上去的，會把游標蓋掉一角——原版的游標
	// 壓在字上面，所以再畫一次（收據 `keireki1e`：滑鼠停在第五列的字上）。
	if spec := g.commandSpec(g.state); spec != nil && spec.Mouse != nil && g.res.Cursor != nil {
		g.res.Cursor.Draw(fb, spec.Mouse.X, spec.Mouse.Y)
	}
}

// drawProduction 畫生產計劃的數字（`docs/spec/004`）。
func (g *Game) drawProduction(fb *gfx.Framebuffer) {
	if g.state != StateProductionPlan || g.res.ProductionPanel == nil {
		return
	}
	g.res.ProductionPanel.Draw(fb, g.res.Font, g.nextPlan(), g.building[g.faction])
}

// drawCapitalPanel 畫右側那塊：日期、軍事費、總稅收與自己首都的摘要。
func (g *Game) drawCapitalPanel(fb *gfx.Framebuffer) {
	if g.res.CapitalPanel == nil || g.res.RegionPanel == nil || g.res.Planets == nil {
		return
	}
	c, _, ok := g.capitalSummary()
	if !ok {
		return
	}
	g.res.CapitalPanel.Draw(fb, g.res.Font, g.res.RegionPanel, c, g.faction)
}

// drawFleetRows 畫右下角清單的「艦隊」分頁（`docs/re/70`）。
//
// **畫在底板那一層**：疊上去的視窗要蓋得住它。
func (g *Game) drawFleetRows(fb *gfx.Framebuffer) {
	if g.res.FleetList == nil || len(g.fleets) == 0 {
		return
	}
	rows := g.fleetRows()
	g.res.FleetList.Draw(fb, g.res.Font, rows)
}

// fleetRows 回右下角「艦隊」分頁要列的那幾列。
//
// **只列右側面板那一格裡的艦隊**：載入 S2 之後畫面上寫的是
// 「帝國軍星系瓦爾哈拉」，清單就只有停在瓦爾哈拉的第 5 與第 12 艦隊
// （收據 `s2ai0`）。面板看哪一格，清單就看哪一格。
func (g *Game) fleetRows() []screen.FleetRow {
	region := g.selectedRegion
	if region < 0 {
		region = rules.CapitalRegion(g.faction)
	}
	var rows []screen.FleetRow
	for _, f := range g.fleets {
		if f.Faction != g.faction || f.Region != region {
			continue
		}
		var name []uint16
		if f.Admiral >= 0 && f.Admiral < len(g.res.AdmiralNames) {
			name = g.res.AdmiralNames[f.Admiral]
		}
		// 右邊那個數字目前一律 100——與艦隊組成上的「機動力」相同，
		// 是不是機動力還沒驗（只有一個資料點，`docs/re/70`）。
		rows = append(rows, screen.FleetRow{Number: f.Number, Name: name, Value: 100})
	}
	return rows
}

// drawLeftSummary 把一個星域的摘要畫進左下角那塊。
//
// 提督資料那幾張畫的是**自己的首都**（收據 `lst2e`／`sortie1e`／
// `admirals1` 與同盟的對應張），星系資料畫的是**滑鼠停的那一列**那個
// 星域（`star1e`／`star1eAlly`）。其餘畫面畫的是 `hoverRegion`
// ——地圖介面上游標底下那一格，選單或視窗開起來之前是哪一格就留哪一格
// （`docs/re/277`、`docs/spec/001`）。`−1` ＝ 整塊只剩外框。
//
// **畫在底板那一層**（`drawBoardPanels` 裡）：待命中提督資料那張的視窗會
// 蓋掉左下角一部分，畫在視窗之後會蓋錯；徵稅計劃（`tax1e`）與取進度
// （`load1e`）那幾張也是視窗壓在摘要上面，摘要只從沒被蓋到的地方露出來。
func (g *Game) drawLeftSummary(fb *gfx.Framebuffer) {
	if g.res.RegionPanel == nil || g.res.Planets == nil {
		return
	}
	region := g.hoverRegion
	if r := g.forcedLeftRegion(g.state); r >= 0 {
		region = r
	}
	switch g.state {
	case StateStarList:
		if spec := g.commandSpec(g.state); spec != nil {
			region = g.starListSelected(spec)
		}
	}
	o := region * rules.RegionRecord
	if region < 0 || o+rules.RegionRecord > len(g.res.Stellar) {
		return
	}
	sum := rules.Summarise(g.res.Stellar[o:o+rules.RegionRecord], g.res.Planets,
		region, g.faction)
	g.res.RegionPanel.Draw(fb, g.res.Font, sum, g.faction)
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
	switch spec.Map.Colours {
	case securityColours:
		if m := g.securityStarMap(); m != nil {
			return m
		}
	case starListColours:
		if m := g.starListStarMap(spec); m != nil {
			return m
		}
	}
	if spec.Map.Colours != "" && g.res.StarMaps != nil {
		if m := g.res.StarMaps[spec.Map.Colours]; m != nil {
			return m
		}
	}
	return g.playerStarMap()
}

// playerStarMap 回勢力色那一張：開局時照陣營重建過的優先。
func (g *Game) playerStarMap() *asset.Image {
	// 交棒視窗與月底結算中的地圖不上玩家色（收據 `vs3`／`vs5`、
	// `turnmonth2`；`docs/spec/045`）。
	if (g.state == StateVersusInput || g.state == StateShipDelivery) && g.plainMap != nil {
		return g.plainMap
	}
	if g.starMap != nil {
		return g.starMap
	}
	return g.res.StarMap
}

// securityColours 是戰略情報報告那張地圖用的色表名，
// starListColours 是星系資料那張的。
const (
	securityColours = "colour_796_01_security"
	starListColours = "colour_796_01_star_list"
)

// starListStarMap 依現在的局面重建「星系資料」那張地圖。
//
// 底色是星域的勢力色（`rules.RegionColour`：帝國 1、同盟 12、無主 3），
// **選中那一格另外用陣營的文字色**（`rules.FactionColour`：帝國 6、
// 同盟 5、無主 7）。選中的是滑鼠停的那一列——帝國那張停在第 2 列 ＝
// 卡斯特羅普（無主，色 7），同盟那張同樣第 2 列 ＝ 凱利姆（同盟，色 5），
// 兩張的色號不同正是因為擁有者不同（`docs/re/64`）。
func (g *Game) starListStarMap(spec *screen.MainGameSpec) *asset.Image {
	if g.res.StarBit == nil || g.res.StarRGN == nil || g.res.Planets == nil {
		return nil
	}
	n := len(g.res.Stellar) / rules.RegionRecord
	if n == 0 {
		return nil
	}
	cols := make([]uint8, n)
	for i := range cols {
		cols[i] = rules.RegionColour(g.res.Planets.RegionOwner(i))
	}
	if r := g.starListSelected(spec); r >= 0 && r < n {
		cols[r] = rules.FactionColour(g.res.Planets.RegionOwner(r))
	}
	img, err := asset.BuildStarMap(g.res.StarBit, g.res.StarRGN, cols, 16, 20, 20)
	if err != nil {
		return nil
	}
	return img
}

// starListSelected 回星系清單上滑鼠停的那一列對應的星域；問不到回 -1。
func (g *Game) starListSelected(spec *screen.MainGameSpec) int {
	order := g.starListOrder()
	var l *screen.ListSpec
	for _, w := range spec.Windows {
		if w != nil && w.List != nil {
			l = w.List
		}
	}
	if l == nil || l.Pitch <= 0 || len(order) == 0 {
		return -1
	}
	y := g.mouseY
	if y < 0 && spec.Mouse != nil {
		y = spec.Mouse.Y + mouseDY
	}
	row := (y - l.Y) / l.Pitch
	if y < l.Y || row < 0 || row >= l.Rows || row >= len(order) {
		return -1
	}
	return order[row]
}

// securityStarMap 依現在的局面重建戰略情報報告那張地圖。
//
// 著色規則是 `rules.SecurityMapColour`：**自己這一方**的星域照治安分四段
// 上色，其餘一律色 3（等於底色，看起來是黑的）。所以換一個陣營整張圖的
// 明暗會反過來——`strmap-regions.json` 裡那份 `colour_796_01_security`
// 是帝國視角量的，只當對照。
//
// 分段的除數是**情報專任參謀的 `+0x14`**（帝國 85、同盟 82，`docs/re/63`），
// 所以同一個未安定度在兩邊可能塗不同的顏色。
func (g *Game) securityStarMap() *asset.Image {
	if g.res.StarBit == nil || g.res.StarRGN == nil || g.res.Planets == nil {
		return nil
	}
	n := len(g.res.Stellar) / rules.RegionRecord
	if n == 0 {
		return nil
	}
	div := rules.SecurityDivisor(g.res.Admiral,
		rules.StaffAdmiral(g.workRecord(), rules.PostIntel, g.faction))
	cols := make([]uint8, n)
	for i := range cols {
		rec := g.res.Stellar[i*rules.RegionRecord : (i+1)*rules.RegionRecord]
		own := g.res.Planets.RegionOwner(i) == g.faction
		cols[i] = rules.SecurityMapColour(int(rec[rules.RegionUnrest]), own, div)
	}
	img, err := asset.BuildStarMap(g.res.StarBit, g.res.StarRGN, cols, 16, 20, 20)
	if err != nil {
		return nil
	}
	return img
}

// factionStarMap 依玩家陣營重建勢力色地圖。
//
// `strmap-regions.json` 的 `colour_796_01` 是**帝國視角**量到的一份：
// 瓦爾哈拉（星域 0）是 6，其餘帝國星域 1、同盟 12、無主 3。同盟開局那張
// 收據（`main1eAlly`）裡瓦爾哈拉退回 1、巴拉特（星域 31）變成 5 ——
// 也就是**只有玩家自己的首都有特別色，帝國 6、同盟 5**，正好是
// `rules.FactionColour` 那兩個值（`docs/re/61`）。
//
// 建不起來（缺原始位元組）時回 nil，由 starMapFor 退回載入時那一張。
//
// `tint` ＝ 要不要塗那一格。**對戰的交棒視窗那一刻不塗**（`vs5` 對 `vs3`
// 扣掉游標只差六個字，兩邊的地圖完全一樣）；按下確認之後那張月初報告
// 又塗回來了（`vs4`）。所以它跟著**畫面**走，不是跟著模式走。
func (g *Game) factionStarMap(faction int, tint bool) *asset.Image {
	if g.res.StarBit == nil || g.res.StarRGN == nil || len(g.res.RegionColours) == 0 {
		return nil
	}
	cols := append([]uint8(nil), g.res.RegionColours...)
	for f := 0; f < 2; f++ {
		if r := rules.CapitalRegion(f); r >= 0 && r < len(cols) {
			cols[r] = rules.RegionColour(f)
		}
	}
	if r := rules.CapitalRegion(faction); r >= 0 && r < len(cols) && tint {
		cols[r] = rules.FactionColour(faction)
	}
	img, err := asset.BuildStarMap(g.res.StarBit, g.res.StarRGN, cols, 16, 20, 20)
	if err != nil {
		return nil
	}
	return img
}

// Turn 回目前是第幾個回合（開局 0）。
func (g *Game) Turn() int { return g.turn }

// Military 回手上的軍事費，TaxTotal 回這個月的總稅收。
func (g *Game) Military() int { return g.military[g.faction] }

// Treasury 回這個月的財政（原定軍事費、已支付各項）。
func (g *Game) Treasury() rules.Treasury { return g.currentTreasury() }
func (g *Game) TaxTotal() int            { return g.taxTotal[g.faction] }

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

// setListRows 把一覽表那幾張清單要畫的列算好放進規格。
//
// 名字、軍階與幕僚記號全部從原版資料檔算（`internal/rules`），
// 規格裡只有版面。缺資料就留空——空清單看得出來是缺資料，不是假裝畫完了。
func (g *Game) setListRows(spec *screen.MainGameSpec) {
	if spec == nil {
		return
	}
	spec.ListRows = nil
	g.listIndex = nil
	if g.state == StateStarList {
		g.setStarRows(spec)
		return
	}
	faction, standby, ok := g.listState(g.state)
	if !ok || len(g.res.Admiral) == 0 || len(g.res.AdmiralNames) == 0 {
		return
	}
	staff := rules.StaffAdmirals(g.workRecord(), g.faction)
	isStaff := map[int]bool{}
	for _, i := range staff {
		isStaff[i] = true
	}
	for _, a := range rules.AdmiralList(rules.ParseAdmirals(g.res.Admiral), faction, standby, staff) {
		var name []uint16
		if a.Index < len(g.res.AdmiralNames) {
			name = g.res.AdmiralNames[a.Index]
		}
		spec.ListRows = append(spec.ListRows, screen.ListRow{
			Staff: isStaff[a.Index],
			Name:  name,
			Rank:  rules.RankOf(a.Merit),
		})
		g.listIndex = append(g.listIndex, a.Index)
	}
}

// workRecord 回目前這一局的 `WORK` 區塊：讀進來的存檔優先，
// 沒有就用新遊戲的樣板。
func (g *Game) workRecord() []byte {
	if len(g.work) > 0 {
		return g.work
	}
	if g.slot != nil && len(g.slot.Work) > 0 {
		return g.slot.Work
	}
	return g.res.WorkTemplate
}

// mutableWork 回這一局可寫的 `WORK` 複本，第一次呼叫時才從資源複製。
func (g *Game) mutableWork() []byte {
	if len(g.work) == 0 {
		g.work = append([]byte(nil), g.workRecord()...)
	}
	return g.work
}

// mutableAdmirals 回這一局可寫的提督表。複製之後把 `res.Admiral` 指過去，
// 既有的讀取點（`admiralRecord` 那些）就自動看到同一份，原始陣列不被動到。
func (g *Game) mutableAdmirals() []byte {
	if len(g.admirals) == 0 {
		g.admirals = append([]byte(nil), g.res.Admiral...)
		g.res.Admiral = g.admirals
	}
	return g.admirals
}

// staffOps 回某個陣營的軍事專任參謀的營運（提督記錄 `+14`）——船艦建造的
// 單價由它決定（`docs/spec/037`、`docs/re/265`）。
//
// **電腦那一方也要**：它的建造計劃走同一條算式，用的是它自己那位參謀。
func (g *Game) staffOps(faction int) int {
	return rules.StaffOps(g.workRecord(), g.res.Admiral, rules.PostMilitary, faction)
}

// staffAdmiral 回坐某個幕僚職的提督姓名與軍階的字碼（`docs/internals/30`）。
func (g *Game) staffAdmiral(post int) (name, rank []uint16) {
	i := rules.StaffAdmiral(g.workRecord(), post, g.faction)
	if i < 0 || i >= len(g.res.AdmiralNames) {
		return nil, nil
	}
	name = g.res.AdmiralNames[i]
	all := rules.ParseAdmirals(g.res.Admiral)
	if i < len(all) {
		rank = g.rankCodes(rules.RankOf(all[i].Merit))
	}
	return name, rank
}

// rankCodes 回某一階的字碼。
//
// 十個軍階的字碼**只量過一份**，在提督資料那張畫面的規格裡
// （`screen-admiral-list-view.json` 的 `list.ranks`）。再抄一份到別的
// 資料檔就會有兩個真相，所以這裡直接跟那一份要。
func (g *Game) rankCodes(rank int) []uint16 {
	s := g.res.Screens["screen-admiral-list-view"]
	if s == nil || rank < 0 {
		return nil
	}
	for _, w := range s.Windows {
		if w.List != nil && rank < len(w.List.Ranks) {
			return w.List.Ranks[rank]
		}
	}
	return nil
}

// listState 回這個狀態的清單要哪一群提督。
//
// 「自己人」是玩家的陣營、「敵將」是另一邊——不是寫死的帝國／同盟
// （`docs/re/64`）。
func (g *Game) listState(s State) (faction int, standby, ok bool) {
	switch s {
	case StateAdmiralListView: // 提督資料：全列，幕僚職前面標 ○
		return g.faction, false, true
	case StateStandbyAdmirals, StateAdmiralList: // 待命中提督資料
		return g.faction, true, true
	case StateEnemyList: // 敵提督資料
		return 1 - g.faction, false, true
	}
	return 0, false, false
}

// listAdmiralAt 回清單上畫面 y 那一列的提督索引；不在清單上回 -1。
func (g *Game) listAdmiralAt(y int) int {
	spec := g.commandSpec(g.state)
	if spec == nil {
		return -1
	}
	// 清單是畫的時候才算的；點擊可能發生在第一次畫之前（測試就是這樣），
	// 所以這裡自己先算一次。
	g.setListRows(spec)
	var l *screen.ListSpec
	for _, w := range spec.Windows {
		if w != nil && w.List != nil {
			l = w.List
		}
	}
	if l == nil || l.Pitch <= 0 || y < l.Y {
		return -1
	}
	row := (y - l.Y) / l.Pitch
	if row < 0 || row >= l.Rows || row >= len(g.listIndex) {
		return -1
	}
	return g.listIndex[row]
}

// starListOrder 回這一方的星系清單列順序。**兩個陣營各一份**，
// 都從自己的首都起頭（`docs/re/64`）。
func (g *Game) starListOrder() []int {
	if g.faction < 0 || g.faction >= len(g.res.StarListOrder) {
		return nil
	}
	return g.res.StarListOrder[g.faction]
}

// setStarRows 算「星系資料」那張清單的列：星域名（用行星算出來的擁有者
// 的陣營色）與兩個陣營的艦隊數。
func (g *Game) setStarRows(spec *screen.MainGameSpec) {
	var l *screen.ListSpec
	for _, w := range spec.Windows {
		if w.List != nil {
			l = w.List
		}
	}
	if l == nil || g.res.Planets == nil || len(g.res.Stellar) == 0 ||
		len(g.starListOrder()) == 0 {
		return
	}
	order := g.starListOrder()
	for _, i := range order {
		if (i+1)*rules.RegionRecord > len(g.res.Stellar) {
			continue
		}
		rec := g.res.Stellar[i*rules.RegionRecord : (i+1)*rules.RegionRecord]
		owner := g.res.Planets.RegionOwner(i)
		spec.ListRows = append(spec.ListRows, screen.ListRow{Cells: []screen.ListCell{
			{X: l.NameX, Ink: rules.FactionColour(owner), Codes: rules.RegionName(rec)},
			{X: l.OwnX, Ink: rules.FactionColour(0), Text: " 0"},
			{X: l.SepX, Ink: l.Ink, Codes: []uint16{l.SepCode}},
			{X: l.EnemyX, Ink: rules.FactionColour(1), Text: " 0"},
		}})
	}
}

// drawBattle 畫戰術畫面。
//
// 視野是 g.battleView（單位是格）；缺原版檔案時回一張空畫布。
func (g *Game) drawBattle() *gfx.Framebuffer {
	if g.res.Battle == nil || g.res.BattlePanel == nil {
		return gfx.NewFramebuffer(0)
	}
	var cells []uint8
	if r := g.battleRegion; r >= 0 && r < len(g.res.BattleMaps) {
		cells = g.res.BattleMaps[r]
	}
	fb := screen.Battle(g.res.Battle, g.res.BattlePanel, g.res.BattleTiles, cells, g.battleView)
	if len(g.battleUnits) > 0 && len(g.res.Sprites) > 0 {
		screen.DrawBattleUnits(g.res.Battle, fb, g.res.Sprites, g.battleUnits,
			func(u rules.BattleUnit) int { return battleSprite(u, g.faction) }, g.battleView)
	}
	// 艦隊選點時畫兩個範圍圖形（`docs/re/120`、`docs/re/135`）。
	// 兩個都是 XOR 上去的，所以要畫在單位與格線之後。
	if g.fleetOrder != FleetOrderNone {
		if f := g.orderedFleetForRange(); f != nil {
			screen.DrawManoeuvreRange(g.res.Battle, fb, f.cx, f.cy,
				f.turn, f.dist, f.facing, g.battleView)
		}
	}
	// 「補給」選對象時右側狀態欄換成補給面板（`docs/re/143`）。
	if p, ok := g.supplyPanel(); ok {
		screen.DrawSupplyPanel(fb, g.res.SupplyPanel, g.res.Font, g.res.FrameTiles, p)
		// 「結束」鈕是另一個視窗，畫在面板上面（`docs/re/143`）。
		screen.DrawSupplyEndButton(fb, g.res.SupplyPanel, g.res.Font, g.res.FrameTiles)
	}
	if p, ok := g.repairPanel(); ok {
		screen.DrawRepairPanel(fb, g.res.SupplyPanel, g.res.Font, g.res.FrameTiles, p)
		screen.DrawSupplyEndButton(fb, g.res.SupplyPanel, g.res.Font, g.res.FrameTiles)
	}
	// 「補給」／「修理」選對象時畫查詢範圍的圈（`docs/re/140`）。
	// **候選本身沒有高亮**——原版那個模態迴圈裡只有這個圈與右側面板，
	// 沒有逐個候選的繪製。
	if cx, cy, r, ok := g.supplyRangeCircle(); ok {
		screen.DrawSupplyRange(g.res.Battle, fb, cx, cy, r, g.battleView)
	}
	// 要塞砲瞄準時畫射程橢圓（`docs/re/145`）。焦點是行星佔位框的
	// 中心（記錄座標各 ＋7），傾角跟著游標。
	if o, i, ok := g.FortressTargetRange(); ok {
		p := g.battlePlanets[g.battleGun]
		screen.DrawFortressRange(g.res.Battle, fb,
			p.X+fortressCentreBias, p.Y+fortressCentreBias,
			o, i, g.battleGunAim, g.battleView)
	}
	// 選抵達朝向時，目的地上多一個指向性的十字（`docs/re/122`）。
	if g.fleetOrder == FleetOrderFacing {
		screen.DrawFacingCross(g.res.Battle, fb,
			g.fleetOrderX, g.fleetOrderY, g.fleetOrderFacing, g.battleView)
	}
	// 「編隊」選格子時，空格子各畫一個標記（`docs/internals/46`）。
	// 原版是把標記掛進顯示串列再整張重畫，分隊在標記之後畫上去
	// ——收據 `s3form` 上壓在分隊底下的那三格看得到。
	if marks := g.formationMarkSlots(); len(marks) > 0 && len(g.res.Sprites) > 0 {
		screen.DrawFormationMarks(g.res.Battle, fb, g.res.Sprites, marks, g.battleView)
	}
	screen.DrawMinimap(g.res.Battle, fb, cells, g.battleView)
	screen.DrawMinimapPlanets(g.res.Battle, fb,
		rules.PlanetSites(cells, g.res.Battle.Field.TileW, g.res.Battle.Field.TileH))
	screen.DrawMinimapUnits(g.res.Battle, fb, g.battleUnits)
	screen.DrawBattleStatusFrame(g.res.Battle, fb, g.res.FrameTiles)
	// 點到自己的分隊時右側換成分隊面板（`docs/re/118`）。
	if v, ok := g.unitPanelView(); ok && g.res.Font != nil {
		screen.DrawUnitPanel(g.res.UnitPanel, fb, g.res.Font, v)
	} else if g.res.Font != nil {
		y, mo, d := g.Date()
		screen.DrawBattleStatus(g.res.Battle, fb, g.res.Font, screen.BattleStatus{
			Faction: g.faction, Year: y, Month: mo, Day: d,
			Turn:     g.battleTurn,
			Victory:  rules.BattleVictory(g.battleScore),
			Empire:   g.battleScore[0],
			Alliance: g.battleScore[1],
			Region:   g.regionName(g.battleRegion),
		})
	}
	if g.battleMenu && g.res.BattleMenu != nil && g.res.Font != nil {
		screen.DrawBattleMenu(g.res.BattleMenu, fb, g.res.Font, g.res.FrameTiles,
			g.battleMenuX, g.battleMenuY, 0)
	}
	// 命令選單開在游標處，反白帶跟著游標走（`docs/spec/023`）。
	if g.battleCmd && g.res.CommandMenus != nil && g.res.Font != nil {
		cm := g.res.CommandMenus
		n := cm.ItemCount(g.battleCmdKind)
		var enabled []bool
		switch {
		case g.battleCmdUnit >= 0 && g.battleCmdUnit < len(g.battleUnits):
			enabled = g.commandMenuEnabled(g.battleCmdKind, g.battleUnits[g.battleCmdUnit])
		case g.battleCmdPlanet >= 0 && g.battleCmdPlanet < len(g.battlePlanets):
			enabled = g.planetMenuEnabled(g.battleCmdPlanet)
		default:
			n = 0
		}
		if n > 0 {
			sel := cm.CommandMenuRowAt(g.battleCmdKind, g.battleCmdX, g.battleCmdY, g.mouseX, g.mouseY, n)
			screen.DrawCommandMenu(cm, fb, g.res.Font, g.res.FrameTiles, g.battleCmdKind,
				enabled, g.battleCmdX, g.battleCmdY, sel)
		}
	}
	// 「資料」是不透明的一層，蓋掉戰場那一塊（`docs/spec/021`）。
	if g.battleData && g.res.BattleData != nil && g.res.Font != nil {
		if g.battleDetail && g.res.FleetDetail != nil {
			// 詳細資料與「資料」共用外框，換掉的是裡面那一層。
			w := g.res.BattleData.Window
			g.res.BattleData.Window.Frame.Frame().Draw(fb, w.X, w.Y, w.W, w.H)
			if v, ok := g.fleetDetailView(); ok {
				screen.DrawFleetDetail(g.res.FleetDetail, fb, g.res.Font, v)
			}
		} else {
			screen.DrawBattleData(g.res.BattleData, fb, g.res.Font,
				g.battleDataTab, g.battleDataRows(g.battleDataTab))
		}
	}
	// 行星情報視窗（`docs/re/144`）。
	if v, ok := g.planetInfoView(); ok {
		screen.DrawPlanetInfo(fb, g.res.PlanetInfo, g.res.Font, g.res.FrameTiles, v)
	}
	// 訊息框畫在最上面（模態）。
	g.drawMessage(fb)
	return fb
}

// regionName 回第 i 個星域的名字；沒有資料就回 nil。
func (g *Game) regionName(i int) []uint16 {
	if i < 0 || (i+1)*rules.RegionRecord > len(g.res.Stellar) {
		return nil
	}
	return rules.RegionName(g.res.Stellar[i*rules.RegionRecord : (i+1)*rules.RegionRecord])
}

// EnterBattle 開戰術畫面，看第 region 個星域、視野從格 (col,row) 起。
//
// **這是診斷入口**：原版是戰略地圖上遭遇才進戰場，那條路徑還沒逆向
// （`docs/re/82`）。走這裡進來的畫面不能當玩家路徑收據。
func (g *Game) EnterBattle(region, col, row int) {
	g.battleRegion = region
	g.battleView = screen.BattleView{Col: col, Row: row}
	g.battleTurn = 1
	g.state = StateBattle
}

// capitalPlanet 是某個陣營首都星域裡收船的那一顆：帝國是切片的第 0 顆
// （奧丁），同盟是第 4 顆（海尼森）。原版 `2B5F:001F` 把這兩組
// `(星域, 切片內序號)` 寫死成 `(0, 0)` 與 `(31, 4)`（`docs/re/264`）。
var capitalPlanetSlot = [2]int{0, 4}

// deliverShips 是原版月底本體的第一件事（`2B5F:001F`，`docs/spec/036`）：
// 把上個月「建造中」的船艦加進自己陣營首都行星的補充存量，同時記進
// 首都星域記錄的艦艇彙總。
//
// 兩邊都寫是照原版做的——星域那一份每回合會被重算（`docs/internals/22`），
// 但同一個回合內新編艦隊讀的是它。
func (g *Game) deliverShips() {
	if g.res.Planets == nil || len(g.res.Stellar) == 0 {
		return
	}
	for f := 0; f < 2; f++ {
		region := ai.Capitals[f]
		start, n := g.res.Planets.Range(region)
		slot := capitalPlanetSlot[f]
		if slot >= n {
			continue
		}
		rec := g.res.Planets.Record(start + slot)
		st, ok := rules.ReadPlanetStock(rec)
		if !ok {
			continue
		}
		o := region * rules.RegionRecord
		if o+rules.RegionRecord > len(g.res.Stellar) {
			continue
		}
		plan := g.building[f]
		for row := 0; row < rules.BuildPlanRows; row++ {
			ships := plan[row]
			if ships == 0 {
				continue
			}
			st.Ships[row] += ships
			rules.ReturnRegionShips(g.res.Stellar[o:o+rules.RegionRecord], f, row, ships)
		}
		rules.WritePlanetStock(rec, st)
	}
}
