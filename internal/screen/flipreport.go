package screen

import (
	"encoding/json"
	"os"

	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
)

// FlipReportPanel 是星域易主報告那一頁跟局面走的東西：星系名、那一句話，
// 以及頭像的落點（規格 `docs/spec/034`）。
//
// 版面量自收據 `chain3`（`docs/re/246`），文字取自原版訊息表
// （`docs/re/247`），頭像是 `FACE.AMG` 第 `ds:8F4D[座位陣營]` 張、
// 取左上 72×93（`docs/re/248`）。
type FlipReportPanel struct {
	LineStep int `json:"line_step"`
	Name     struct {
		X, Y int
		// InkByOwner 以**新的擁有者**索引（0 帝國 6、1 同盟 5、2 無主 7）。
		// 三份收據的玩家座位都是帝國，所以「依擁有者」與「依敵我」
		// 這一份證據分不開（`docs/spec/034`）。
		InkByOwner []uint8 `json:"ink_by_owner"`
	} `json:"name"`
	Suffix struct {
		Ink   uint8
		Codes []uint16
	} `json:"suffix"`
	Message struct {
		X, Y int
		Ink  uint8
	} `json:"message"`
	Face struct {
		X, Y, W, H int
	} `json:"face"`
	Messages map[string]struct {
		Lines [][]uint16 `json:"lines"`
	} `json:"messages"`
}

// LoadFlipReportPanel 讀版面。
func LoadFlipReportPanel(path string) (*FlipReportPanel, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p FlipReportPanel
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Draw 畫兩行字：第 1 行是「〈星系名〉星系」（名字與 `星系` 兩種顏色，
// 對應原版格式字串 `ds:085D` 的兩個顏色控制碼），第 2 行是 `key` 那一句。
//
// 頭像另外畫（要 `FACE.AMG`，見 `Face`）。
func (p *FlipReportPanel) Draw(fb *gfx.Framebuffer, fnt *font.Font,
	name []uint16, owner int, key string) {
	ink := uint8(7)
	if owner >= 0 && owner < len(p.Name.InkByOwner) {
		ink = p.Name.InkByOwner[owner]
	}
	x := p.Name.X
	for _, c := range name {
		drawFull(fb, fnt, x, p.Name.Y, []uint16{c}, ink)
		x += font.GlyphW
	}
	drawFull(fb, fnt, x, p.Name.Y, p.Suffix.Codes, p.Suffix.Ink)

	m, ok := p.Messages[key]
	if !ok {
		return
	}
	step := p.LineStep
	if step == 0 {
		step = 20
	}
	for i, line := range m.Lines {
		drawFull(fb, fnt, p.Message.X, p.Message.Y+i*step, line, p.Message.Ink)
	}
}
