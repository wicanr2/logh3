package screen

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

// 戰略地圖上的艦隊視窗（原版 `0DEA:0040`，`docs/re/312`、規格 `docs/spec/051`）。
//
// 右下角「艦隊」分頁點一列開出來的那張：25×17 格的內容 ＋ 一格外框，
// 提督頭像、三顆鈕 作戰／變更／結束。**內容是把原版那四十幾次
// `3617:003A` 逐一重播**——格式字串逐位元組抄自資料段
// （`docs/internals/data/fleet-window.json`，`tools/scripts/fleetwindow_spec.py`），
// 顏色與底線也照原版在每一次列印之前改視窗 `[+1]`／`[+2]` 的順序。
// 這一層只會排版；值由 `internal/game` 湊齊。

// FleetWindowSpec 是版面與格式字串。
type FleetWindowSpec struct {
	CellX   int                      `json:"cell_x"`
	CellY   int                      `json:"cell_y"`
	Cols    int                      `json:"cols"`
	Rows    int                      `json:"rows"`
	Style   int                      `json:"style"`
	Colour  int                      `json:"colour"`
	Face    struct{ X, Y, W, H int } `json:"face"`
	Buttons []struct {
		Name  string   `json:"name"`
		CellX int      `json:"cell_x"`
		CellY int      `json:"cell_y"`
		Cols  int      `json:"cols"`
		Codes []string `json:"codes"`
	} `json:"buttons"`
	ButtonColour         int               `json:"button_colour"`
	FactionNamesColoured []string          `json:"faction_names_coloured"`
	FactionNames         []string          `json:"faction_names"`
	FactionColour        []int             `json:"faction_colour"`
	KindLetters          []string          `json:"kind_letters"`
	MoveNames            []string          `json:"move_names"`
	SpeedShrinkTable     []int             `json:"speed_shrink_table"`
	Formats              map[string]string `json:"formats"`
	ActionMenu           struct {
		Items     []string `json:"items"`
		ItemHex   []string `json:"item_hex"`
		Width     int      `json:"width"`
		ClampRows int      `json:"clamp_rows"`
	} `json:"action_menu"`
	Verbs           []string `json:"verbs"`
	VerbStandby     string   `json:"verb_standby"`
	ConfirmFormat   string   `json:"confirm_format"`
	Unreachable     string   `json:"unreachable"`
	UnreachableKind int      `json:"unreachable_kind"`
	ConfirmKind     int      `json:"confirm_kind"`
}

// LoadFleetWindowSpec 讀 `docs/internals/data/fleet-window.json`。
func LoadFleetWindowSpec(path string) (*FleetWindowSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s FleetWindowSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	if s.Cols == 0 || s.Rows == 0 || len(s.Formats) == 0 {
		return nil, fmt.Errorf("fleet-window：規格不完整（%d×%d、%d 條格式字串）", s.Cols, s.Rows, len(s.Formats))
	}
	return &s, nil
}

func (s *FleetWindowSpec) fmtBytes(name string) []byte {
	h, ok := s.Formats[name]
	if !ok {
		panic("fleet-window：規格沒有格式字串 " + name)
	}
	return decodeHex(h)
}

// FleetWindowGroup 是「部隊組成」那一列的一組：艦種字母與艘數。
type FleetWindowGroup struct {
	Kind  int
	Ships int
	Full  int
}

// FleetWindowView 是這張視窗要印的每一個值。欄位名跟著原版讀的記錄走
// （`docs/re/312` §二、§七）。
type FleetWindowView struct {
	Faction int
	Number  int

	CmdName, CmdRank []byte // 提督姓名、軍階（Big5）
	CmdLevel         int
	CmdExp           int

	Move   int // 艦隊 +0x08
	Target int // 艦隊 +0x07：星域（Move != 5）；-1 ＝ 0xFF
	// 現在位置與目的地的名字與擁有者（名字用擁有者的陣營色印）。
	RegionName         []byte
	RegionOwner        int
	RegionFlags        int // 星域 +0x13
	TargetName         []byte
	TargetOwner        int
	StandbyPlanetName  []byte // 待命行星（Move == 5 且 Target != -1）
	StandbyPlanetOwner int

	Groups   []FleetWindowGroup
	Expected int // 艦隊 +0x0E

	MissileCur, MissileMax int
	FighterCur, FighterMax int

	AdjPresent       bool
	AdjName, AdjRank []byte
	AdjLevel         int
	AdjExp           int

	Upkeep, MoveCost int
	CostAdjBetter    bool // 副官 +0x0E > 提督 +0x0E → 兩筆費用印色 5

	Morale  int // 艦隊 +0x3C
	Command int // 提督 +0x0B
	// 回轉／巡航／戰鬥速度（艦隊 +0x4D／+0x4E／+0x4F）與五個能力
	// （+0x3D 光線砲、+0x3E 飛彈、+0x3F 戰鬥機、+0x40 占領、+0x41 防禦）。
	Turn, Cruise, Fight int
	Abilities           [5]int
	// 每一格「副官比提督高」就印色 5：速度三格比提督記錄的
	// +0x0D／+0x0C／+0x0C，能力五格比 +0x0F／+0x10／+0x11／+0x13／+0x12。
	SpeedAdjBetter [3]bool
	AbilAdjBetter  [5]bool

	// OpsPressed ＝ 「作戰」鈕反白（選單開著或正在選目的地，收據 `fcmd-pick`）。
	OpsPressed bool
	// ChangeHidden ＝ 「變更」鈕不畫：所在星域有戰鬥（`+0x13 & 4`）或不是自己的
	// （`0DEA:0040` 開頭把熱區表 `ds:21C0` 第 1 筆的啟用旗標清 0，
	// 收據 `fcmd-battle-win`，`docs/re/315`）。
	ChangeHidden bool
	// ChangePressed ＝ 「變更」鈕反白（副官／編組小選單開著，`docs/re/316`）。
	ChangePressed bool
}

// Colour bytes the original writes into the window before each print.
const (
	fleetWindowInk        = 0x97 // 底 9、字 7
	fleetWindowInkValue   = 0x96 // 經驗值那種：字 6
	fleetWindowInkAlert   = 0x92 // 戰鬥中／占領中、編制掉到一半以下：字 2
	fleetWindowInkWarn    = 0x95 // 編制不滿但還有一半：字 5
	fleetWindowInkAdj     = 0x95 // 副官比提督高
	fleetWindowInkPressed = 0x27 // 按下去的鈕：底 2
)

// factionInk 是 `3982:0263` 印星域／行星名用的顏色：`ds:6092[擁有者] | 0x90`。
func factionInk(owner int) uint8 { return 0x90 | rules.FactionColour(owner) }

// Draw 把視窗畫在畫面上。回沒有字模的字碼。
func (s *FleetWindowSpec) Draw(fb *gfx.Framebuffer, fnt *font.Font, tiles *ui.FrameTiles,
	v *FleetWindowView, face *asset.Image) []uint16 {
	if s == nil || fb == nil || v == nil {
		return nil
	}
	if tiles != nil {
		(ui.TileFrame{Tiles: tiles, Style: s.Style, Fill: uint8(s.Colour >> 4)}).
			Draw(fb, s.CellX, s.CellY, s.Cols+2, s.Rows+2)
	}
	b := s.Buffer(v)
	missing := ui.DrawTextBuffer(fb, fnt, tiles, b, s.CellX+1, s.CellY+1)
	if face != nil {
		for y := 0; y < s.Face.H && y < face.H; y++ {
			for x := 0; x < s.Face.W && x < face.W; x++ {
				fb.Set(s.Face.X+x, s.Face.Y+y, face.Pix[y*face.W+x])
			}
		}
	}
	return missing
}

// Buffer 把整張視窗排進內容緩衝區——就是 `0DEA:0040` 那一串列印。
func (s *FleetWindowSpec) Buffer(v *FleetWindowView) *ui.TextBuffer {
	f := s.fmtBytes
	b := ui.NewTextBuffer(s.Cols, s.Rows, uint8(s.Colour))
	goto_ := func(col, row int) { b.Printf(f("goto"), col, row) }
	colouredName := func(name []byte, owner int) {
		// `3982:0263(視窗, 擁有者, 0xFFF0, 記錄)`：名字用擁有者的陣營色。
		b.Colour = factionInk(owner)
		b.Write(name)
	}

	// ── 0DEA:0040 本體 ──
	b.Write(f("clear"))
	b.Printf(f("fleet_no"), decodeHex(s.FactionNamesColoured[v.Faction]))
	// `3982:0038`：編號印成全形數字。
	b.Attr |= ui.AttrWiden
	b.Printf([]byte("%d"), v.Number)
	b.Attr &^= ui.AttrWiden
	b.Write(f("fleet_no_tail"))
	b.Printf(f("commander"), v.CmdName, v.CmdRank)
	b.Printf(f("cmd_level"), v.CmdLevel)
	b.Colour = fleetWindowInkValue
	b.Printf(f("exp"), v.CmdExp)
	b.Colour = fleetWindowInk
	b.Flags = ui.CellUnderline
	b.Write(f("rule"))
	b.Flags = 0
	b.Write(f("ops_label"))
	goto_(5, 5)
	b.Write(f("position"))
	colouredName(v.RegionName, v.RegionOwner)
	b.Colour = fleetWindowInk
	b.Flags = ui.CellUnderline
	b.Printf(f("nl_pad50"), f("space"))
	b.Flags = 0
	b.Write(f("status_head"))
	b.Flags = ui.CellUnderline
	b.Printf(f("pad50"), f("space"))
	b.Flags = 0
	b.Write(f("adjutant_label"))
	b.Flags = ui.CellUnderline
	b.Printf(f("pad50"), f("space"))
	b.Write(f("costs"))
	b.Flags = 0
	b.Write(f("abil_head"))
	b.Write(f("abil_2"))
	b.Write(f("abil_3"))
	b.Write(f("abil_4"))
	b.Write(f("abil_5"))

	// ── 0DEA:0918：副官 ──
	goto_(5, 9)
	if v.AdjPresent {
		line := append(append([]byte(nil), v.AdjName...), v.AdjRank...)
		b.Printf(f("adj_name"), line)
		goto_(5, 10)
		b.Flags = ui.CellUnderline
		b.Printf(f("adj_level"), v.AdjLevel)
		b.Colour = fleetWindowInkValue
		b.Printf(f("adj_exp"), v.AdjExp)
		b.Colour = fleetWindowInk
	} else {
		b.Printf(f("adj_name"), f("none"))
		goto_(5, 10)
		b.Flags = ui.CellUnderline
		b.Write(f("adj_none_level"))
	}
	b.Flags = 0

	// ── 0DEA:03F9：作戰那一列 ──
	goto_(5, 4)
	if v.Move == 5 {
		if v.RegionFlags&0x14 == 0 {
			b.Printf(f("ops_fight"), decodeHex(s.MoveNames[5]))
		} else {
			b.Colour = fleetWindowInkAlert
			if v.RegionFlags&4 != 0 {
				b.Write(f("ops_in_fight"))
			} else {
				b.Write(f("ops_occupying"))
			}
			b.Colour = fleetWindowInk
		}
	} else {
		name := ""
		if v.Move >= 0 && v.Move < len(s.MoveNames) {
			name = s.MoveNames[v.Move]
		}
		b.Printf(f("ops_fight"), decodeHex(name))
	}

	// ── 0DEA:04C0：現在位置底下那一列 ──
	goto_(5, 6)
	b.Flags = ui.CellUnderline
	switch {
	case v.Move != 5:
		b.Write(f("move_target"))
		colouredName(v.TargetName, v.TargetOwner)
	case v.Target >= 0:
		b.Write(f("standby_planet"))
		colouredName(v.StandbyPlanetName, v.StandbyPlanetOwner)
	default:
		b.Write(f("standby"))
	}
	b.Colour = fleetWindowInk
	b.Flags = 0

	// ── 0DEA:05DD：部隊組成 ──
	goto_(11, 7)
	n := 0
	for _, g := range v.Groups {
		ink := fleetWindowInk
		switch {
		case g.Ships == g.Full:
		case g.Full/2 > g.Ships:
			ink = fleetWindowInkAlert
		default:
			ink = fleetWindowInkWarn
		}
		letter := ""
		if g.Kind >= 0 && g.Kind < len(s.KindLetters) {
			letter = s.KindLetters[g.Kind]
		}
		b.Printf(f("colour_str"), ink, decodeHex(letter))
		n++
	}
	b.Colour = fleetWindowInk
	for ; n < v.Expected; n++ {
		b.Write(f("empty_slot"))
	}
	for ; n < 10; n++ {
		b.Write(f("two_spaces"))
	}

	// ── 0DEA:06FE：飛彈／戰鬥機 ──
	b.Flags = ui.CellUnderline
	goto_(5, 8)
	b.Write(f("missile_label"))
	if v.MissileMax != 0 {
		b.Printf(f("pct"), v.MissileCur*100/v.MissileMax)
	} else {
		b.Write(f("dashes"))
	}
	b.Write(f("fighter_label"))
	if v.FighterMax != 0 {
		b.Printf(f("pct"), v.FighterCur*100/v.FighterMax)
	} else {
		b.Write(f("dashes"))
	}
	b.Flags = 0

	// ── 0DEA:0A80：費用與十二格 ──
	b.Flags = ui.CellUnderline
	goto_(9, 11)
	costInk := fleetWindowInk
	if v.CostAdjBetter {
		costInk = fleetWindowInkAdj
	}
	b.Printf(f("colour_num6"), costInk, v.Upkeep)
	goto_(17, 11)
	b.Printf(f("num6"), v.MoveCost)
	b.Colour = fleetWindowInk
	b.Flags = 0
	goto_(10, 12)
	b.Printf(f("num3"), v.Morale)
	goto_(10, 13)
	b.Printf(f("num3"), v.Command)
	speeds := [3]int{v.Turn, v.Cruise, v.Fight}
	for i, val := range speeds {
		goto_(10, 14+i)
		b.Printf(f("colour_num3"), adjInk(v.SpeedAdjBetter[i]), val)
	}
	for i, val := range v.Abilities {
		goto_(18, 12+i)
		b.Printf(f("colour_num3"), adjInk(v.AbilAdjBetter[i]), val)
	}
	b.Colour = fleetWindowInk

	// ── 三顆鈕（`37A9:000F` 的子元件，與勝負報告的「確認」同一種畫法）──
	for _, bt := range s.Buttons {
		if bt.Name == "change" && v.ChangeHidden {
			continue
		}
		colour := uint8(s.ButtonColour)
		if (bt.Name == "ops" && v.OpsPressed) || (bt.Name == "change" && v.ChangePressed) {
			colour = fleetWindowInkPressed
		}
		putTextButton(b, bt.CellX, bt.CellY, bt.Cols, colour, bt.Codes)
	}
	return b
}

func adjInk(better bool) int {
	if better {
		return fleetWindowInkAdj
	}
	return fleetWindowInk
}

// putTextButton 把一顆鈕寫進緩衝區：左端一格、本體幾格、右端一格
// （與 `VictoryReportPanel.putButton` 同一套旗標）。codes 是每一格的
// 字碼（hex）。
func putTextButton(b *ui.TextBuffer, x, y, cols int, colour uint8, codes []string) {
	putRawButton(b, x, y, cols, colour, decodeHex(joinHex(codes)))
}

func joinHex(parts []string) string {
	out := ""
	for _, p := range parts {
		out += p
	}
	return out
}

// ButtonAt 回畫面座標 (x,y) 落在哪一顆鈕上（"ops"／"change"／"close"），
// 沒有就回空字串。熱區是 `21C0` 那三筆：內容格 (22,14)–(22,16) 各 2×1。
func (s *FleetWindowSpec) ButtonAt(x, y int) string {
	for _, bt := range s.Buttons {
		bx := (s.CellX + 1 + bt.CellX) * ui.FrameTileW
		by := (s.CellY+1+bt.CellY)*ui.FrameTileH + ui.FrameCellOriginY
		if x >= bx && x < bx+bt.Cols*ui.FrameTileW && y >= by && y < by+ui.FrameTileH {
			return bt.Name
		}
	}
	return ""
}

// Rect 回視窗（含外框）在畫面上的矩形。
func (s *FleetWindowSpec) Rect() (x, y, w, h int) {
	return s.CellX * ui.FrameTileW, s.CellY*ui.FrameTileH + ui.FrameCellOriginY,
		(s.Cols + 2) * ui.FrameTileW, (s.Rows + 2) * ui.FrameTileH
}

// ShipSpeeds 是 `297B:0E8B` 給一組艦艇算出來的三個速度（`docs/re/312` §七）：
//
//	回轉 ＝ 艦種[+8] × 艦隊[+0x43] ÷ 100
//	巡航 ＝ 艦種[+6] × 艦隊[+0x42] ÷ 80
//	戰鬥 ＝ 艦種[+7] × 艦隊[+0x42] ÷ 60
//	三個各再扣掉 自己 × f ÷ 100，f ＝ (270 − 2 × 艦隊[+0x46]) × 表[組數] ÷ 100
//
// 表就是 `speed_shrink_table`（`ds:4E2E`）。存檔 8 的 150 筆艦艇裡 111 筆
// 逐值相同，其餘 39 筆回轉差 1（同兩支艦隊，疑似能力值變過之後沒重算），
// 新編的第１艦隊三個值全中（收據 `fcmd-win`）。
func (s *FleetWindowSpec) ShipSpeeds(kindTurn, kindTop, kindBattle, abilTurn, abilMove, abilOps, groups int) (turn, cruise, fight int) {
	turn = kindTurn * abilTurn / 100
	cruise = kindTop * abilMove / 80
	fight = kindBattle * abilMove / 60
	shrink := 0
	if groups >= 0 && groups < len(s.SpeedShrinkTable) {
		shrink = (270 - 2*abilOps) * s.SpeedShrinkTable[groups] / 100
	}
	turn -= turn * shrink / 100
	cruise -= cruise * shrink / 100
	fight -= fight * shrink / 100
	return
}
