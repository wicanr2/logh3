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
	StateShipDelivery:    "screen-ship-delivery",
	StateAdmiralDetail:   "screen-admiral-detail",
	StateAdmiralBio:      "screen-admiral-bio",
	StateFleetOrg:        "screen-fleet-org",
	StateFleetNumber:     "screen-fleet-number",
	StateFleetAdjutant:   "screen-fleet-adjutant",
	StateFleetComposition: "screen-fleet-composition",
	StateFormation:        "screen-formation",
	StateVersusInput:      "screen-versus-input",
	StateVersusAccount:    "screen-versus-account",
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
	BioPanel     *screen.BioPanel
	Biographies  [][]uint16

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
	res   *Resources
	state State

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

	// delivered 是月底報告上那一份：跨月的當下交付出去的船。
	delivered              rules.BuildPlan
	reportYear, reportMonth int

	// 目前選中的星域。左下角那塊寫的就是它。**還沒選過是 −1**，
	// 那塊是空的——收據 `main1e` 就是這個狀態（關月初報告的那兩下右鍵
	// 也點在星圖上，但沒有選中任何星域）。
	selectedRegion int

	// saveErr 是上一次存檔的結果（nil ＝ 成功或還沒存過）。
	saveErr error

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
	battleFleets  []rules.BattleFleet
	// 戰術選單開在游標處：battleMenu 是「開著沒有」，battleMenuX／Y 是框的
	// 左上角（`docs/internals/34`）。
	battleMenu           bool

	// battleData ＝ 戰術選單「資料」開著沒有，battleDataTab 是看哪一個
	// 分頁（0 艦隊／1 行星／2 敵艦隊）。
	battleData    bool
	battleDataTab int

	// battleDetail ＝ 艦隊詳細資料那一層開著沒有，battleDetailRow 是
	// 點的是表上第幾列。
	battleDetail    bool
	battleDetailRow int
	battleMenuX, battleMenuY int

	// battleCmd ＝ 戰場的命令選單開著沒有（`docs/spec/023`）；
	// battleCmdUnit 是點到的分隊在 battleUnits 裡的索引，battleCmdKind
	// 是五張表的哪一張，battleCmdX／Y 是框的左上角。
	battleCmd                bool
	battleCmdUnit            int
	battleCmdKind            screen.CommandMenuKind
	battleCmdX, battleCmdY   int

	// battleMove ＝ 選了「移動」之後在等玩家點目的地（`docs/spec/024`）；
	// battleMoveUnit 是要移動的分隊。
	battleMove     bool
	battleMoveUnit int

	// battleForm ＝ 選了「編隊」之後在等玩家點隊形格（`docs/internals/46`）；
	// battleFormUnit 是要換位置的分隊。
	battleForm     bool
	battleFormUnit int

	battleRally   []rules.RallySite
	battlePlanets []rules.BattlePlanet
	battleRand   *rules.Rand
	battleLosses []rules.BattleLoss
	battleScore  [2]int
	battleOver   bool

	// brain 是電腦那一方的決策（`internal/ai`、`docs/spec/017`）。
	// **預設是 `ai.Original`——按實行時什麼都不做**，那是量出來的原版
	// 行為（`docs/re/79`）。要強化版就換成 `ai.Focused`。
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
	military, taxTotal int
	pct                [rules.BuildPlanRows]int
	treasury           rules.Treasury
	monthReport        rules.Budget
	building           rules.BuildPlan
	fleets             []rules.Fleet
}

// 星圖畫在畫面的 (0,40)，320×300，圖塊 16×20（`docs/internals/12`）。
const (
	mapX, mapY = 0, 40
	mapW, mapH = 320, 300
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
	// 電腦那一方先下命令，再讓所有艦隊走。預設的 `ai.Original`
	// 一道都不下，所以這一段對原版對拍沒有影響。
	g.runBrain()
	// 有命令的艦隊往目的地走一格（`docs/re/77`、`internal/rules.Advance`）。
	// **一回合走一格是假設待驗**——量到的三組移動都只隔一格。
	if g.res.Graph != nil {
		g.res.Graph.MoveFleets(g.fleets)
	}
	// 兩邊的艦隊走到同一個星域就直接切進戰術畫面（`docs/internals/34`
	// 「怎麼走到」：S3 按六次實行，雙方會合在達希利就開戰）。
	// **哪一個星域先打、能不能不打，都還沒解**——同時有多個星域符合時
	// 這裡取編號最小的那一個。
	if r := g.contestedRegion(); r >= 0 {
		g.StartBattle(r)
		return
	}
	// 一個月十回合；跨過去就結算（`docs/re/26`、`docs/re/33`）
	// 並跳出船艦接收報告。
	//
	// ⚠ **對戰模式的跨月沒有量過**（`docs/spec/016`）：這裡照單人模式
	// 只結算目前這一方，另一方的錢停在上個月。要往下走得先有收據。
	if g.turn%rules.TurnsPerMonth == 0 {
		g.delivered = g.building
		g.reportYear, g.reportMonth, _ = g.Date()
		// 報告掛的是**剛結束的那個月**，日期已經翻到次月 1 日。
		g.reportMonth--
		if g.reportMonth == 0 {
			g.reportMonth, g.reportYear = 12, g.reportYear-1
		}
		g.monthEnd()
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
func (g *Game) contestedRegion() int {
	best := -1
	for r := 0; r < rules.RegionCount; r++ {
		var side [2]bool
		for _, f := range g.fleets {
			if f.Region == r && f.Faction >= 0 && f.Faction < 2 {
				side[f.Faction] = true
			}
		}
		if side[0] && side[1] {
			if best < 0 || r < best {
				best = r
			}
		}
	}
	return best
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
	for _, o := range g.brain.Orders(w, foe, g.turn) {
		if o.Fleet < 0 || o.Fleet >= len(g.fleets) {
			continue
		}
		g.fleets[o.Fleet].Target = o.Target
		g.fleets[o.Fleet].Parked = false
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
	return side{military: g.military, taxTotal: g.taxTotal, pct: g.pct,
		treasury: g.treasury, monthReport: g.monthReport,
		building: g.building, fleets: g.fleets}
}

// enterSide 換到某一方：陣營、地圖顏色與那一方的錢一起換過去。
func (g *Game) enterSide(faction int) {
	g.faction = faction
	g.starMap = g.factionStarMap(faction, true)
	g.plainMap = g.factionStarMap(faction, false)
	s := g.sides[faction]
	g.military, g.taxTotal, g.pct = s.military, s.taxTotal, s.pct
	g.treasury, g.monthReport = s.treasury, s.monthReport
	g.building, g.fleets = s.building, s.fleets
	// 交棒的當下底板是乾淨的：右側面板只剩陣營那一行與按鈕，
	// 左下角是空的（收據 `vs3`／`vs5`）。
	g.selectedRegion = -1
	g.detailAdmiral = -1
	g.reportLayer = 0
}

func (g *Game) startNewGame(faction int) {
	g.restoreWorld()
	g.faction = faction
	g.starMap = g.factionStarMap(faction, true)
	g.plainMap = g.factionStarMap(faction, false)
	g.turn = 0
	g.military = g.res.StartMilitary
	g.pct = rules.OpeningPct
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
		menuRow: -1, detailAdmiral: -1, battleRegion: -1}
	g.startNewGame(res.Faction)
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
	g.res.Stellar = sl.Stellar
	g.res.Planets = sl.Planets
	g.turn = sl.Turn
	g.military = sl.Military
	g.taxTotal = sl.TaxTotal
	// 存檔裡就有兩邊的艦隊（`FLEET.DT?` ＝ 執行期那張表，`docs/re/41`）。
	// 讀進來之後 remake 才有敵方的艦隊可以畫、可以推進。
	g.fleets = sl.Fleets(g.res.Weights)
	g.selectedRegion = -1
	g.state = StateCommandPhase
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

// Typed 回防拷問答已經打進去的字串（畫面上還沒畫，見下方的未確認事項）。
func (g *Game) Typed() string { return string(g.typed) }

// mouseDY 見 `ui.MouseDY`：**Game 對外一律用畫面座標**，只有要餵給吃滑鼠
// 座標的東西時才減掉它。兩套座標曾經同時存在，症狀是「命中判定對、
// 游標畫在別的地方」——兩邊各錯一次剛好互相抵銷（`docs/re/52`）。
const mouseDY = ui.MouseDY

// MoveMouse 移動游標，座標是**畫面座標**。主選單會跟著換反白的那一列。
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
	if g.state == StateIntro {
		g.SkipIntro()
		return
	}
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
	case StateSaveGame:
		// 點下去就直接存到那一格，不問名字——原版在唯讀的原版目錄上
		// 會跳「磁碟片載入時有發生錯誤。」（`docs/spec/010`）。
		if n := g.saveRow(); n >= 0 && g.res.SaveDir != "" {
			g.saveErr = g.SaveSlot(g.res.SaveDir, n)
		}
	case StateShipDelivery:
		// 報告只有一個「確認」鈕，點哪裡都收掉——哪裡按不算數還沒量過。
		g.state = StateCommandPhase
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
						// 編成完停在自己的首都，還沒有移動命令
						// （`docs/re/70` 的收據 `dec1` 就是這個狀態）。
						Region: rules.CapitalRegion(g.faction),
						Target: rules.NoTarget,
					})
					// 左下角那塊**不會被擦掉**：艦隊那幾張畫的是自己的
					// 首都，回到地圖之後原版就讓它留在那裡（收據 `dec1`）。
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
	if g.state == StateIntro {
		g.SkipIntro()
		return
	}
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
	case StateMainGame, StateVersusAccount:
		// 月初報告是兩層。第一下關掉上面那層，第二下才到地圖介面。
		// 對戰模式那張（`vs4`）也是兩下——收據 `vs5` 用的就是兩次右鍵。
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
			// 從狀態算出來的那幾塊要畫在底板之後、疊上去的視窗之前
			// （`docs/spec/002`）。
			fb, _ := screen.MainGameLayered(spec, g.res.Font, g.starMapFor(spec),
				g.res.Faces, g.res.Cursor, g.drawBoardPanels, g.afterWindow)
			g.drawPanels(fb)
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
		fb, _ := screen.MainMenuAt(g.res.Menu, g.res.Font, g.highlight,
			g.res.Cursor, g.mouseX, g.mouseY-mouseDY)
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
	collected := rules.MonthEnd(g.res.Stellar, g.res.Planets, g.res.Growth, g.faction)
	budget := before + collected
	// 2. 照百分比切給造船，各列只買得起整數艘，零頭退回；扣掉實際花掉的。
	plan := rules.BuildPlan{Cost: budget * g.res.ShipPct / 100, Pct: g.pct, Faction: g.faction}
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
		if g.res.Planets.RegionOwner(i) != g.faction {
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
	if g.state == StateVersusInput || g.state == StateVersusAccount {
		c, _, ok := g.capitalSummary()
		if !ok {
			return
		}
		g.res.CapitalPanel.DrawHandover(fb, g.res.Font, c, g.faction)
		if g.state == StateVersusAccount {
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
		r, g.faction, g.turn, rules.Epoch, g.military, g.taxTotal), r, true
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
		StateFleetNumber, StateFleetAdjutant, StateFleetComposition,
		StateFormation,
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
	v := screen.CompositionView{
		Ships:    rules.RegionShips(g.res.Stellar[o:o+rules.RegionRecord], g.faction),
		PerUnit:  rules.ShipsPerUnit[:],
		Units:    1,
		Mobility: 100,
		Slots:    []int{0},
	}
	p := g.res.CompositionPanel
	p.Draw(fb, g.res.Font, v)
	g.drawCompositionGrid(fb, v)
}

// drawCompositionGrid 鋪格盤那 9 × 9 的圖塊（`docs/re/67`）：
// `(欄 + 列)` 為偶數的 41 格才畫，中央那一格畫部隊的圖塊。
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
	for row := 0; row < gr.Rows; row++ {
		for col := 0; col < gr.Cols; col++ {
			if (col+row)%2 != 0 {
				continue
			}
			tile := gr.Empty
			if col == gr.Cols/2 && row == gr.Rows/2 && len(v.Slots) > 0 {
				if t := g.res.CompositionPanel.TileAt(0); t >= 0 {
					tile = t
				}
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
		b := rules.Budget{Treasury: g.military, Tax: g.taxTotal, ShipPct: g.res.ShipPct}
		g.res.BudgetPanel.Draw(fb, g.res.Font, g.treasury, b)
	case StateBudgetPlan:
		// 月初報告那一張：本月份那一欄還沒有數字（全 0）。
		g.res.BudgetPanel.Draw(fb, g.res.Font, rules.Treasury{}, g.monthReport)
	}
}

// thisPlan 是**這個月**的船艦建造計劃：月底結算算出來的建造費照百分比
// 切出來的那一份。訂造報告與建造計劃這兩張印的都是它。
func (g *Game) thisPlan() rules.BuildPlan {
	return rules.BuildPlan{Cost: g.monthReport.ShipCost(), Pct: g.pct, Faction: g.faction}
}

// nextPlan 是次月份的船艦建造計劃：預估軍事費照百分比切出來的那一份。
func (g *Game) nextPlan() rules.BuildPlan {
	b := rules.Budget{Treasury: g.military, Tax: g.taxTotal, ShipPct: g.res.ShipPct}
	return rules.BuildPlan{Cost: b.ShipCost(), Pct: g.pct, Faction: g.faction}
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
	if g.state != StateShipOrder || g.res.DeliveryPanel == nil {
		return
	}
	c, _, ok := g.capitalSummary()
	if !ok {
		return
	}
	g.res.DeliveryPanel.Draw(fb, g.res.Font, c.Year, c.Month, g.thisPlan(), nil, nil)
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
		g.selectedRegion, g.faction)
	g.res.RegionPanel.Draw(fb, g.res.Font, sum, g.faction)
}

// drawLeftSummary 把一個星域的摘要畫進左下角那塊。
//
// 提督資料那幾張畫的是**自己的首都**（收據 `lst2e`／`sortie1e`／
// `admirals1` 與同盟的對應張），星系資料畫的是**滑鼠停的那一列**那個
// 星域（`star1e`／`star1eAlly`）。其餘畫面左下角是空的，或由
// `drawRegionPanel` 畫游標選中的那一格。
//
// **畫在底板那一層**（`drawBoardPanels` 裡）——待命中提督資料那張的
// 視窗會蓋掉左下角一部分，畫在視窗之後會蓋錯。
func (g *Game) drawLeftSummary(fb *gfx.Framebuffer) {
	if g.res.RegionPanel == nil || g.res.Planets == nil {
		return
	}
	region := -1
	switch g.state {
	case StateAdmiralListView, StateStandbyAdmirals, StateAdmiralList,
		StateAdmiralDetail, StateAdmiralBio, StateFleetOrg,
		StateFleetNumber, StateFleetAdjutant, StateFleetComposition,
		StateFormation:
		region = rules.CapitalRegion(g.faction)
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
	// 交棒視窗那一刻地圖不上玩家色（收據 `vs3`／`vs5`）。
	if g.state == StateVersusInput && g.plainMap != nil {
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
	if g.slot != nil && len(g.slot.Work) > 0 {
		return g.slot.Work
	}
	return g.res.WorkTemplate
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
			func(u rules.BattleUnit) int { return battleSprite(u) }, g.battleView)
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
	screen.DrawBattleStatusFrame(g.res.Battle, fb)
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
		screen.DrawBattleMenu(g.res.BattleMenu, fb, g.res.Font,
			g.battleMenuX, g.battleMenuY, 0)
	}
	// 命令選單開在游標處，反白帶跟著游標走（`docs/spec/023`）。
	if g.battleCmd && g.res.CommandMenus != nil && g.res.Font != nil &&
		g.battleCmdUnit >= 0 && g.battleCmdUnit < len(g.battleUnits) {
		cm := g.res.CommandMenus
		n := cm.ItemCount(g.battleCmdKind)
		sel := cm.CommandMenuRowAt(g.battleCmdX, g.battleCmdY, g.mouseX, g.mouseY, n)
		screen.DrawCommandMenu(cm, fb, g.res.Font, g.battleCmdKind,
			g.commandMenuEnabled(g.battleCmdKind, g.battleUnits[g.battleCmdUnit]),
			g.battleCmdX, g.battleCmdY, sel)
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
