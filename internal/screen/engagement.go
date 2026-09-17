package screen

import (
	"encoding/hex"
	"encoding/json"
	"os"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// EngagementSpec 是交戰演出（打行星那一種）的版面
// （`docs/internals/data/engagement-panel.json`，`docs/re/324`、規格 056）：
// `ANI_PNL1.A98` 底圖、三個樣式 0 的文字視窗與原版的格式字串（Big5 位元組，
// 含 `\x0c` 清視窗、`\x01[c,r,` 跳格、`\x03[n,` 換色那些控制碼）。
type EngagementSpec struct {
	Panel struct {
		File string `json:"file"`
		Y    int    `json:"y"`
	} `json:"panel"`
	Windows struct {
		Planet EngagementWindow `json:"planet"`
		Middle EngagementWindow `json:"middle"`
		Fleet  EngagementWindow `json:"fleet"`
	} `json:"windows"`
	Formats struct {
		Faction          []string `json:"faction"`
		PlanetTitle      string   `json:"planet_title"`
		PlanetName       string   `json:"planet_name"`
		PlanetKind       string   `json:"planet_kind"`
		PlanetDefence    string   `json:"planet_defence"`
		FleetUnknown     string   `json:"fleet_unknown"`
		FleetKnownPrefix string   `json:"fleet_known_prefix"`
		FleetKnownSuffix string   `json:"fleet_known_suffix"`
		FleetKnownName   string   `json:"fleet_known_name"`
		FleetShips       string   `json:"fleet_ships"`
		LabelDefence     string   `json:"label_defence"`
		LabelLanding     string   `json:"label_landing"`
		LabelCaptured    string   `json:"label_captured"`
		LabelFailed      string   `json:"label_failed"`
		Destroyed        string   `json:"destroyed"`
	} `json:"formats"`
	KnownDiff []EngagementRect    `json:"known_diff"`
	Animation EngagementAnimation `json:"animation"`
}

// EngagementAnimation 是演出的圖要用的常數（`COMANI.BMG`，`docs/re/331`）。
type EngagementAnimation struct {
	File string `json:"file"`
	// Intervals 是每格的間隔（計時器 tick），索引是動畫速度（`ds:8F10`）。
	Intervals []int `json:"intervals"`
	// TickNS 是一個計時器 tick 的長度（奈秒）。
	TickNS int64 `json:"tick_ns"`
	// Windows 是左（陣營 0）右（陣營 1）兩個視窗的格座標（`3A8D:003A`）。
	Windows [][2]int `json:"windows"`
	// FlagshipCodes 是 `ds:8E34`（hex）：提督索引 → 艦隊段名第二個字元。
	FlagshipCodes  string `json:"flagship_codes"`
	FortressRegion int    `json:"fortress_region"`
}

// EngagementRect 是事先列出的排除區（畫面座標）與理由。
type EngagementRect struct {
	X, Y, W, H int
	Why        string `json:"why"`
}

// EngagementWindow 是一個樣式 0 的文字視窗：格 (X,Y) 起 W×H，Colour 是視窗
// 顏色（高四位底色、低四位字色）。
type EngagementWindow struct {
	X, Y, W, H int
	Colour     int `json:"colour"`
}

// LoadEngagementSpec 讀版面。
func LoadEngagementSpec(path string) (*EngagementSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s EngagementSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// EngagementLabel 是中間那格的標題種類。
type EngagementLabel int

const (
	EngagementDefence  EngagementLabel = iota // 防衛攻擊（行星的反擊）
	EngagementLanding                         // 登陸艦攻擊
	EngagementCaptured                        // 占領成功
	EngagementFailed                          // 占領失敗
)

// EngagementView 是一場演出要印的東西。
type EngagementView struct {
	Faction    int    // 行星擁有者（0 帝國、1 同盟）
	PlanetName []byte // Big5
	Defence    int    // 防衛力 現在／上限
	DefenceMax int

	FleetKnown  bool   // 看得到艦隊編號與提督（自軍，或索敵等級夠）
	FleetNumber int    // 第Ｎ艦隊
	AdmiralName []byte // Big5
	Ships       int    // 船艦總數 現在／上限
	ShipsMax    int

	Label      EngagementLabel
	Destroyed  int // 「擊毀」的數；ShowResult 才印
	ShowResult bool
}

func fmtBytes(h string) []byte {
	b, err := hex.DecodeString(h)
	if err != nil {
		return nil
	}
	return b
}

// wideNumber 是 `3982:0038`：編號印成全形數字。
// `3982:0038` 兩位數編號應印 `%2d` 半形（`docs/re/143`、`docs/re/336` §六），這裡還是一律全形。
func wideNumber(n int) []byte {
	var out []byte
	for _, c := range fullDigits(n) {
		out = append(out, byte(c>>8), byte(c))
	}
	return out
}

// Buffers 回三個視窗的內容（照原版的格式字串排）。
func (s *EngagementSpec) Buffers(v *EngagementView) (planet, middle, fleet *ui.TextBuffer) {
	f := s.Formats
	planet = ui.NewTextBuffer(s.Windows.Planet.W, s.Windows.Planet.H, uint8(s.Windows.Planet.Colour))
	var faction []byte
	if v.Faction >= 0 && v.Faction < len(f.Faction) {
		faction = fmtBytes(f.Faction[v.Faction])
	}
	planet.Printf(fmtBytes(f.PlanetTitle), faction)
	planet.Printf(fmtBytes(f.PlanetName), fmtBytes(f.PlanetKind), v.PlanetName)
	planet.Printf(fmtBytes(f.PlanetDefence), v.Defence, v.DefenceMax)

	fleet = ui.NewTextBuffer(s.Windows.Fleet.W, s.Windows.Fleet.H, uint8(s.Windows.Fleet.Colour))
	if v.FleetKnown {
		fleet.Write(fmtBytes(f.FleetKnownPrefix))
		fleet.Write(wideNumber(v.FleetNumber))
		fleet.Write(fmtBytes(f.FleetKnownSuffix))
		fleet.Printf(fmtBytes(f.FleetKnownName), v.AdmiralName)
	} else {
		fleet.Write(fmtBytes(f.FleetUnknown))
	}
	fleet.Printf(fmtBytes(f.FleetShips), v.Ships, v.ShipsMax)

	middle = ui.NewTextBuffer(s.Windows.Middle.W, s.Windows.Middle.H, uint8(s.Windows.Middle.Colour))
	label := f.LabelDefence
	switch v.Label {
	case EngagementLanding:
		label = f.LabelLanding
	case EngagementCaptured:
		label = f.LabelCaptured
	case EngagementFailed:
		label = f.LabelFailed
	}
	middle.Write(fmtBytes(label))
	if v.ShowResult {
		middle.Printf(fmtBytes(f.Destroyed), v.Destroyed)
	}
	return planet, middle, fleet
}

// Draw 畫整張：底圖、三個視窗的字。游標在演出期間是藏起來的（`3931:0043`）。
func (s *EngagementSpec) Draw(fb *gfx.Framebuffer, fnt *font.Font, tiles *ui.FrameTiles,
	panel *asset.Image, v *EngagementView) {
	if s == nil || fb == nil {
		return
	}
	if panel != nil {
		for y := 0; y < panel.H && y+s.Panel.Y < gfx.Height; y++ {
			for x := 0; x < panel.W && x < gfx.Width; x++ {
				fb.Set(x, y+s.Panel.Y, panel.Pix[y*panel.W+x])
			}
		}
	}
	if v == nil {
		return
	}
	p, m, fl := s.Buffers(v)
	ui.DrawTextBuffer(fb, fnt, tiles, p, s.Windows.Planet.X, s.Windows.Planet.Y)
	ui.DrawTextBuffer(fb, fnt, tiles, m, s.Windows.Middle.X, s.Windows.Middle.Y)
	ui.DrawTextBuffer(fb, fnt, tiles, fl, s.Windows.Fleet.X, s.Windows.Fleet.Y)
}
