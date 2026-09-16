package screen

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

// fullDigit0 是 Big5 的全形「０」。全形數字十個連號（`docs/internals/04`）。
const fullDigit0 = 0xA2AF

// DeliveryPanel 是月底「船艦接收報告」視窗裡算出來的那幾塊。
//
// 視窗其餘的字（標題的固定部分、七種船艦的名稱、欄位標題、按鈕）是
// 靜態的，留在畫面規格裡；這裡只放**跟狀態走**的東西：標題的年月、
// 參謀的姓名與軍階、七列的造艦數與部隊數。
type DeliveryPanel struct {
	Ink    uint8 `json:"ink"`
	Header struct {
		Y          int `json:"y"`
		YearX      int `json:"year_x"`
		MonthRight int `json:"month_right"`
	} `json:"header"`
	Staff struct {
		Y int `json:"y"`
		X int `json:"x"`
	} `json:"staff"`
	RowsY   []int `json:"rows_y"`
	Columns struct {
		Ships int `json:"ships"`
		Units int `json:"units"`
	} `json:"columns"`
}

// LoadDeliveryPanel 讀版面。
func LoadDeliveryPanel(path string) (*DeliveryPanel, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p DeliveryPanel
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Draw 把算出來的部分畫上去。
//
// year／month 是報告的年月（＝剛結束的那個月），delivered 是這個月交付的
// 七個艦種艘數（原版 `ds:8F9C`），name／rank 是簽這份報告的參謀。
func (p *DeliveryPanel) Draw(fb *gfx.Framebuffer, fnt *font.Font,
	year, month int, delivered rules.Building, name, rank []uint16) {
	// 標題的年月是**全形數字**（Big5 A2AF 起），不是半形。
	drawFullDigits(fb, fnt, p.Header.YearX, p.Header.Y, year, p.Ink)
	m := fullDigits(month)
	drawFull(fb, fnt, p.Header.MonthRight-len(m)*font.GlyphW, p.Header.Y, m, p.Ink)

	if len(name) > 0 {
		codes := append(append([]uint16(nil), name...), rank...)
		drawFull(fb, fnt, p.Staff.X, p.Staff.Y, codes, p.Ink)
	}
	for i, y := range p.RowsY {
		if i >= rules.BuildPlanRows {
			break
		}
		drawRight(fb, fnt, p.Columns.Ships, y, strconv.Itoa(delivered.Ships(i)), p.Ink)
		drawRight(fb, fnt, p.Columns.Units, y, strconv.Itoa(delivered.Units(i)), p.Ink)
	}
}

// fullDigits 把一個非負整數轉成全形數字的字碼。
func fullDigits(n int) []uint16 {
	s := strconv.Itoa(n)
	out := make([]uint16, 0, len(s))
	for i := 0; i < len(s); i++ {
		out = append(out, uint16(fullDigit0+int(s[i]-'0')))
	}
	return out
}

func drawFullDigits(fb *gfx.Framebuffer, fnt *font.Font, x, y, n int, ink uint8) {
	drawFull(fb, fnt, x, y, fullDigits(n), ink)
}

// FullDigits 是 fullDigits 給外面用的版本（提督詳細資料的「第Ｎ艦隊」）。
func FullDigits(n int) []uint16 { return fullDigits(n) }
