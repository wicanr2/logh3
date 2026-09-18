package screen

import (
	"encoding/json"
	"os"
)

// InfoPanel 是情報收集畫面（`screen-info`）三個分頁的字串與地圖色表
// （`docs/internals/data/info-panel.json`，`docs/re/342`）。
//
// 三個分頁（治安／勢力／敵艦，順序與畫面上的鈕相同）各印一段標題與五行圖例；
// 地圖那一側三個分頁共用同一張色表（`ds:1780[等級]`，等級 0–4）。
type InfoPanel struct {
	MapColours []uint8 `json:"map_colours"`
	Tabs       []struct {
		Name   string   `json:"name"`
		Title  string   `json:"title"`
		Legend []string `json:"legend"`
	} `json:"tabs"`
}

// LoadInfoPanel 讀那份 JSON。
func LoadInfoPanel(path string) (*InfoPanel, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p InfoPanel
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// TitleCodes 回第 tab 個分頁的標題字碼；Legend 回第 k 行圖例。
func (p *InfoPanel) TitleCodes(tab int) []uint16 {
	if p == nil || tab < 0 || tab >= len(p.Tabs) {
		return nil
	}
	return big5Codes(decodeHex(p.Tabs[tab].Title))
}

// LegendCodes 回第 tab 個分頁第 k 行圖例的字碼。
func (p *InfoPanel) LegendCodes(tab, k int) []uint16 {
	if p == nil || tab < 0 || tab >= len(p.Tabs) || k < 0 || k >= len(p.Tabs[tab].Legend) {
		return nil
	}
	return big5Codes(decodeHex(p.Tabs[tab].Legend[k]))
}

// Colour 回等級 band（0–4）在地圖上的色號。
func (p *InfoPanel) Colour(band int) uint8 {
	if p == nil || band < 0 || band >= len(p.MapColours) {
		return 3
	}
	return p.MapColours[band]
}

// big5Codes 把 Big5 位元組切成一格一個字碼：高位元組 >= 0x80 的是全形（兩個
// byte 一格），其餘兩個半形併成一格（與 `ui.ContentCell` 同一套，`docs/internals/04`）。
// 原版這幾行全是全形字與全形空白，所以這裡只處理全形與成對半形。
func big5Codes(b []byte) []uint16 {
	var out []uint16
	for i := 0; i < len(b); {
		if b[i] >= 0x80 && i+1 < len(b) {
			out = append(out, uint16(b[i])<<8|uint16(b[i+1]))
			i += 2
			continue
		}
		hi := uint16(b[i]) << 8
		lo := uint16(' ')
		if i+1 < len(b) {
			lo = uint16(b[i+1])
		}
		out = append(out, hi|lo)
		i += 2
	}
	return out
}
