package asset

import "fmt"

// DecodeCMP 解一張戰場地圖（`MAP<nn>.CMP`，一個星域一張）。
//
// 編碼很短（`docs/re/81`、`docs/internals/36`）：
//
//	byte >= 0x80  一格，圖塊編號 ＝ byte & 0x7F
//	byte <  0x80  這個 byte 是圖塊編號，**下一個 byte 是重複次數**
//
// 重複次數本身可以到 255，所以不能假設它小於 0x80——那樣解會少算，
// 32 張裡有 16 張會短。
//
// 展開的結果是 cols×rows 格；原版是 128×102，配 16×20 的圖塊就是
// 2048×2040，與艦隊在戰場上的座標範圍對得上。
func DecodeCMP(d []byte, cols, rows int) ([]uint8, error) {
	want := cols * rows
	out := make([]uint8, 0, want)
	for i := 0; i < len(d); {
		if d[i] >= 0x80 {
			out = append(out, d[i]&0x7F)
			i++
			continue
		}
		if i+1 >= len(d) {
			out = append(out, d[i])
			i++
			continue
		}
		for n := int(d[i+1]); n > 0; n-- {
			out = append(out, d[i])
		}
		i += 2
	}
	if len(out) != want {
		return nil, fmt.Errorf("解出 %d 格，要 %d 格（%d×%d）", len(out), want, cols, rows)
	}
	return out, nil
}

// 戰場地圖的尺寸（`docs/internals/36`）。
const (
	BattleMapCols = 128
	BattleMapRows = 102
	BattleTileW   = 16
	BattleTileH   = 20
)
