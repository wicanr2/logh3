package screen

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// 戰略地圖右下角「敵艦隊」分頁點一列開的視窗（原版 `0B20:0336` →
// `0B20:0411`，`docs/re/325`、規格 052）：24×6 格的內容 ＋ 一格外框、
// 提督頭像、一顆「結束」鈕。內容是把原版那六次 `3617:003A` 逐一重播
// （格式字串逐位元組抄自資料段，`docs/internals/data/enemy-fleet-window.json`）。
// 索敵等級到 3 才開得出來；不到的那一列改跳訊息框「情報資料來源不足」。

// EnemyWindowSpec 是版面與格式字串。
type EnemyWindowSpec struct {
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
	KindLetters          []string          `json:"kind_letters"`
	Formats              map[string]string `json:"formats"`
	NoIntelMessage       string            `json:"no_intel_message"`
}

// LoadEnemyWindowSpec 讀 `docs/internals/data/enemy-fleet-window.json`。
func LoadEnemyWindowSpec(path string) (*EnemyWindowSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s EnemyWindowSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	if s.Cols == 0 || s.Rows == 0 || len(s.Formats) == 0 {
		return nil, fmt.Errorf("enemy-fleet-window：規格不完整（%d×%d、%d 條格式字串）", s.Cols, s.Rows, len(s.Formats))
	}
	return &s, nil
}

// EnemyWindowView 是一張視窗要印的東西。
type EnemyWindowView struct {
	Faction int    // 那一支的陣營（0 帝國、1 同盟）
	Number  int    // 第Ｎ艦隊
	Name    []byte // 提督姓名（Big5）
	Rank    []byte // 軍階（Big5）
	// Groups 是各組編制的（艦種, 索敵等級）：等級 ≥ 2 印艦種字母、1 印
	// 「？」、0 不印（`0B20:0411` 的迴圈）。
	Groups []EnemyGroup
}

// EnemyGroup 是一組編制在這張視窗要用到的兩個數。
type EnemyGroup struct {
	Kind, Detail int
}

// Buffer 把視窗排進內容緩衝區——就是 `0B20:0411` 那一串列印。
func (s *EnemyWindowSpec) Buffer(v *EnemyWindowView) *ui.TextBuffer {
	f := func(name string) []byte { return decodeHex(s.Formats[name]) }
	b := ui.NewTextBuffer(s.Cols, s.Rows, uint8(s.Colour))
	var faction []byte
	if v.Faction >= 0 && v.Faction < len(s.FactionNamesColoured) {
		faction = decodeHex(s.FactionNamesColoured[v.Faction])
	}
	b.Printf(f("head"), faction)
	writeFleetNumber(b, v.Number) // `3982:0038`
	b.Write(f("head_tail"))
	b.Printf(f("admiral"), v.Name, v.Rank)
	b.Write(f("kinds_label"))
	for _, gp := range v.Groups {
		switch {
		case gp.Detail >= 2:
			if gp.Kind >= 0 && gp.Kind < len(s.KindLetters) {
				b.Printf(f("kind"), decodeHex(s.KindLetters[gp.Kind]))
			}
		case gp.Detail != 0:
			b.Write(f("kind_unknown"))
		}
	}
	// 「結束」（`37A9:000F` 的子元件，與艦隊視窗那三顆同一種畫法）。
	for _, bt := range s.Buttons {
		putTextButton(b, bt.CellX, bt.CellY, bt.Cols, uint8(s.ButtonColour), bt.Codes)
	}
	return b
}

// Draw 把視窗畫在畫面上：外框、內容、頭像。回沒有字模的字碼。
func (s *EnemyWindowSpec) Draw(fb *gfx.Framebuffer, fnt *font.Font, tiles *ui.FrameTiles,
	v *EnemyWindowView, face *asset.Image) []uint16 {
	if s == nil || fb == nil || v == nil {
		return nil
	}
	if tiles != nil {
		(ui.TileFrame{Tiles: tiles, Style: s.Style, Fill: uint8(s.Colour >> 4)}).
			Draw(fb, s.CellX, s.CellY, s.Cols+2, s.Rows+2)
	}
	missing := ui.DrawTextBuffer(fb, fnt, tiles, s.Buffer(v), s.CellX+1, s.CellY+1)
	if face != nil {
		face.Draw(fb, s.Face.X, s.Face.Y, s.Face.W, s.Face.H)
	}
	return missing
}

// ButtonAt 回畫面座標落在哪一顆鈕上（"close"），沒有就回空字串。
// 熱區是 `1AA8` 那一筆：內容格 (21,5) 2×1。
func (s *EnemyWindowSpec) ButtonAt(x, y int) string {
	for _, bt := range s.Buttons {
		bx := (s.CellX + 1 + bt.CellX) * ui.FrameTileW
		by := (s.CellY+1+bt.CellY)*ui.FrameTileH + ui.FrameCellOriginY
		if x >= bx && x < bx+bt.Cols*ui.FrameTileW && y >= by && y < by+ui.FrameTileH {
			return bt.Name
		}
	}
	return ""
}
