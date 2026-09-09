package screen

import (
	"encoding/hex"
	"encoding/json"
	"os"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// 「補給」選對象時右側那個面板（原版 `1930:0AA2`，`docs/re/143`）。
//
// 視窗就是戰場的右側狀態欄（`ds:9890`，內容 8 × 8 格，`docs/re/142`），
// 選對象期間被整個蓋掉。格式字串裡**沒有換行**：每一段剛好佔滿 8 格，
// 寫到第 9 格自動折行（`docs/re/141`）。

// SupplyPanelSpec 是面板的版面與格式字串。
type SupplyPanelSpec struct {
	CellX  int `json:"cell_x"`
	CellY  int `json:"cell_y"`
	Cols   int `json:"cols"`
	Rows   int `json:"rows"`
	Colour int `json:"colour"`

	Lines struct {
		Head      hexStr `json:"head"`
		FleetTail hexStr `json:"fleet_suffix"`
		Transport hexStr `json:"transport"`
		Missile   hexStr `json:"missile"`
		Fighter   hexStr `json:"fighter"`
	} `json:"lines"`

	Number struct {
		LT10 hexStr `json:"lt10"`
		GE10 hexStr `json:"ge10"`
	} `json:"number"`

	// EndButton 是「結束」鈕。它是**另一個視窗**（`1629:0154`），格
	// (34,17)、4 × 1 格、旗標 0（不畫框）：三段各自換旗標與顏色，
	// 左端 `0x10`、本體 `0x08` 配色 `0x87`、右端 `0x20`。
	EndButton struct {
		CellX  int `json:"cell_x"`
		CellY  int `json:"cell_y"`
		Cols   int `json:"cols"`
		Rows   int `json:"rows"`
		Colour int `json:"colour"`

		Segments []struct {
			Flags  uint8  `json:"flags"`
			Colour int    `json:"colour"`
			String hexStr `json:"string"`
		} `json:"segments"`
	} `json:"end_button"`
}

// hexStr 是「原版資料段那一串位元組」，JSON 裡用十六進位存。
type hexStr struct {
	Offset string `json:"offset"`
	Bytes  string `json:"bytes"`
	Text   string `json:"text"`
}

// Raw 回那一串位元組。
func (h hexStr) Raw() []byte {
	b, err := hex.DecodeString(h.Bytes)
	if err != nil {
		return nil
	}
	return b
}

// LoadSupplyPanelSpec 讀面板的版面。
func LoadSupplyPanelSpec(path string) (*SupplyPanelSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s SupplyPanelSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// SupplyPanel 是面板上要顯示的數字。
type SupplyPanel struct {
	FactionName []byte // 陣營名（`ds:60F4`，字串自己帶顏色跳脫碼）
	Fleet       int    // 艦隊編號
	Transport   int    // 補給艦現有艘數（艦艇 `+0x08`）
	Missiles    int64  // 艦艇 `+0x0E` 32 位元
	Fighters    int64  // 艦艇 `+0x12` 32 位元
}

// DrawSupplyPanel 把面板畫進右側狀態欄。tiles 只用來畫框的圖塊字碼，
// 面板本身不重畫框（原版也沒有——它是印進既有視窗的內容區）。
func DrawSupplyPanel(fb *gfx.Framebuffer, s *SupplyPanelSpec, fnt *font.Font,
	tiles *ui.FrameTiles, p SupplyPanel) {
	if fb == nil || s == nil {
		return
	}
	b := ui.NewTextBuffer(s.Cols, s.Rows, uint8(s.Colour))
	b.Printf(s.Lines.Head.Raw(), p.FactionName)
	// 艦隊編號固定佔一格：小於 10 用全形數字，否則兩個半形（`3982:0038`）。
	if p.Fleet < 10 {
		b.Printf(s.Number.LT10.Raw(), p.Fleet)
	} else {
		b.Printf(s.Number.GE10.Raw(), p.Fleet)
	}
	b.Write(s.Lines.FleetTail.Raw())
	b.Printf(s.Lines.Transport.Raw(), p.Transport)
	b.Printf(s.Lines.Missile.Raw(), p.Missiles)
	b.Printf(s.Lines.Fighter.Raw(), p.Fighters)
	ui.DrawTextBuffer(fb, fnt, tiles, b, s.CellX, s.CellY)
}

// DrawSupplyEndButton 畫「結束」鈕。原版是另一個視窗，所以位置與面板
// 無關（`docs/re/143`）。
func DrawSupplyEndButton(fb *gfx.Framebuffer, s *SupplyPanelSpec, fnt *font.Font,
	tiles *ui.FrameTiles) {
	if fb == nil || s == nil || s.EndButton.Cols <= 0 {
		return
	}
	e := &s.EndButton
	b := ui.NewTextBuffer(e.Cols, e.Rows, uint8(e.Colour))
	for _, seg := range e.Segments {
		b.Colour, b.Flags = uint8(seg.Colour), seg.Flags
		b.Write(seg.String.Raw())
	}
	ui.DrawTextBuffer(fb, fnt, tiles, b, e.CellX, e.CellY)
}
