package screen

import "github.com/wicanr2/logh3/internal/gfx"

// DrawSettlementScrollArrows 畫月底結算期間才出現在右側捲軸槽裡的上下箭頭。
// 幾何與 10×10 色號矩陣量自 `turnmonth2`（`docs/spec/045`）。
func DrawSettlementScrollArrows(fb *gfx.Framebuffer) {
	if fb == nil {
		return
	}
	draw := func(y int, rows [10]string) {
		for dy, row := range rows {
			for dx := 0; dx < len(row); dx++ {
				fb.Set(612+dx, y+dy, row[dx]-'0')
			}
		}
	}
	draw(305, [10]string{
		"8888378888",
		"8883997888",
		"8839999788",
		"8399999978",
		"7773997777",
		"8883997888",
		"8883997888",
		"8883997888",
		"8883997888",
		"8887777888",
	})
	draw(385, [10]string{
		"8883337888",
		"8883997888",
		"8883997888",
		"8883997888",
		"8883997888",
		"9333993337",
		"8399999978",
		"8839999788",
		"8883997888",
		"8888378888",
	})
}
