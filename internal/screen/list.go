package screen

import (
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// ListSpec 是一覽表那幾張清單的列版面（提督資料、待命中提督、敵提督）。
//
// 每一列是「標記 ＋ 姓名 ＋ 軍階」三段固定欄位，列的底板（底色 8 與上下
// 立體邊）已經由視窗的 boxes 畫好，這裡只補字。座標與字碼讀自字元平面
// （`docs/internals/31`）。
type ListSpec struct {
	MarkerX int `json:"marker_x"`
	NameX   int `json:"name_x"`
	RankX   int `json:"rank_x"`
	Y       int `json:"y"`
	Pitch   int `json:"pitch"`
	Rows    int `json:"rows"`

	Ink    uint8 `json:"ink"`
	Shadow uint8 `json:"shadow"`

	// StaffMark 是幕僚職那一列前面的記號（○），SortieMark 是帶著艦隊那一列的
	// （★；`+0x09` 低七位是艦隊槽，`docs/re/323`）。
	StaffMark  uint16 `json:"staff_mark"`
	SortieMark uint16 `json:"sortie_mark"`
	// Ranks 是十個軍階的字碼，由低到高，對上 rules.RankNames。
	Ranks [][]uint16 `json:"ranks"`

	// 星系清單那張的欄位：兩個艦隊數是半形數字，中間夾一個全形的「﹕」。
	OwnX    int    `json:"own_x"`
	SepX    int    `json:"sep_x"`
	EnemyX  int    `json:"enemy_x"`
	SepCode uint16 `json:"sep_code"`
}

// ListCell 是清單一列上的一段字：全形用 Codes、半形（數字）用 Text。
type ListCell struct {
	X     int
	Ink   uint8
	Codes []uint16
	Text  string
}

// ListRow 是清單上的一列。
//
// 提督清單用 Staff／Name／Rank 三段固定欄位；星系清單那種欄位不一樣的，
// 由呼叫端算好座標放進 Cells。
type ListRow struct {
	Staff  bool
	Sortie bool
	Name   []uint16
	Rank   int
	Cells  []ListCell
}

// Draw 畫清單的列。超過 Rows 的部分不畫——原版一頁就十六列。
func (l *ListSpec) Draw(fb *gfx.Framebuffer, fnt *font.Font, rows []ListRow) {
	l.drawIn(fb, fnt, rows, 0, 0, gfx.Width, gfx.Height)
}

// DrawRowsIn 只重畫與指定矩形相交的那幾列。
//
// 原版先畫游標，再把游標壓到的那一列刷回底色、把字重畫上去——
// 所以那一列會蓋掉游標的一部分，其他列不會。整份重畫的話箭頭會整個不見。
func (l *ListSpec) DrawRowsIn(fb *gfx.Framebuffer, fnt *font.Font, rows []ListRow,
	x, y, w, h int) {
	l.drawIn(fb, fnt, rows, x, y, w, h)
}

func (l *ListSpec) drawIn(fb *gfx.Framebuffer, fnt *font.Font, rows []ListRow,
	rx, ry, rw, rh int) {
	if l == nil {
		return
	}
	for i, r := range rows {
		if i >= l.Rows {
			break
		}
		// 比對用的是**列的底板**（比文字高兩列，那兩列是立體邊），
		// 因為「游標壓到哪一列」是照底板算的。
		if ry0 := l.Y - 2 + i*l.Pitch; ry0 >= ry+rh || ry >= ry0+l.Pitch {
			continue
		}
		y := l.Y + i*l.Pitch
		// 立體邊往上移那一列會落到列的白色上緣——原版的上緣是乾淨的
		// （樣式旗標畫在文字之後），所以把整列裁在內容區裡。
		draw := func(x int, c uint16) {
			ui.DrawGlyphClipped(fb, fnt, x, y, c, l.Ink, l.Shadow, -1, false,
				0, y, gfx.Width, l.Pitch)
		}
		if r.Staff && l.StaffMark != 0 {
			draw(l.MarkerX, l.StaffMark)
		} else if r.Sortie && l.SortieMark != 0 {
			draw(l.MarkerX, l.SortieMark)
		}
		for j, c := range r.Name {
			draw(l.NameX+j*font.GlyphW, c)
		}
		if len(r.Name) > 0 && r.Rank >= 0 && r.Rank < len(l.Ranks) {
			for j, c := range l.Ranks[r.Rank] {
				draw(l.RankX+j*font.GlyphW, c)
			}
		}
		for _, cell := range r.Cells {
			ink := cell.Ink
			if ink == 0 {
				ink = l.Ink
			}
			for j, c := range cell.Codes {
				ui.DrawGlyphClipped(fb, fnt, cell.X+j*font.GlyphW, y, c,
					ink, l.Shadow, -1, false, 0, y, gfx.Width, l.Pitch)
			}
			for j := 0; j < len(cell.Text); j++ {
				ui.DrawHalfGlyphClipped(fb, fnt, cell.X+j*font.HalfW, y, cell.Text[j],
					ink, l.Shadow, -1, false, 0, y, gfx.Width, l.Pitch)
			}
		}
	}
}
