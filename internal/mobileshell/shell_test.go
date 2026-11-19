package mobileshell

import (
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
)

// 邏輯寬：高 480，寬依比例、不小於 640。
func TestLogicalWidth(t *testing.T) {
	for _, c := range []struct{ cw, ch, w int }{
		{2400, 1080, 1066}, {1280, 576, 1066}, {1920, 1080, 853}, {1024, 768, 640}, {800, 800, 640}, {0, 0, 640},
	} {
		if w := LogicalWidth(c.cw, c.ch); w != c.w {
			t.Errorf("LogicalWidth(%d,%d) = %d，預期 %d", c.cw, c.ch, w, c.w)
		}
	}
}

// 整張置中：中央換回原版座標，兩側不屬於畫面。
func TestFullLayoutToGame(t *testing.T) {
	l := computeLayout(1066, false)
	if x, y, ok := l.toGame(533, 240); !ok || x != 320 || y != 240 {
		t.Errorf("中央換成 %d,%d %v", x, y, ok)
	}
	for _, x := range []float64{100, 1000} {
		if _, _, ok := l.toGame(x, 240); ok {
			t.Errorf("x=%v 在側邊，不該對到原版畫面", x)
		}
	}
}

// 戰略版面：地圖靠左滿高，卡片槽在右；三塊互不重疊、都在畫布內；點回原版座標正確。
func TestStrategicLayout(t *testing.T) {
	for _, lw := range []int{1066, 960} {
		l := computeLayout(lw, true)
		if !l.strategic || len(l.regions) != 3 {
			t.Fatalf("寬 %d 沒有排成戰略版面：%+v", lw, l)
		}
		m := l.regions[0]
		if m.src != srcMap || m.x != sidebarW || m.y != 0 || float64(m.src.Dy())*m.scale != gfx.Height {
			t.Errorf("寬 %d 地圖區塊 %+v", lw, m)
		}
		for i, r := range l.regions {
			w, h := float64(r.src.Dx())*r.scale, float64(r.src.Dy())*r.scale
			if r.x < 0 || r.y < 0 || r.x+w > float64(lw)+0.01 || r.y+h > gfx.Height+0.01 {
				t.Errorf("寬 %d 區塊 %d 超出畫布 %+v", lw, i, r)
			}
			// 每塊左上角那一點要換回該塊的原版左上角。
			if x, y, ok := l.toGame(r.x+0.5, r.y+0.5); !ok || x != r.src.Min.X || y != r.src.Min.Y {
				t.Errorf("寬 %d 區塊 %d 左上角換成 %d,%d %v", lw, i, x, y, ok)
			}
		}
		if l.regions[1].x < m.x+float64(m.src.Dx())*m.scale {
			t.Errorf("寬 %d 卡片槽壓到地圖", lw)
		}
	}
	// 16:9 的卡片槽只有 237，卡片會縮到 0.74 倍、點陣字太小，退回整張置中。
	if l := computeLayout(853, true); l.strategic {
		t.Error("寬 853 卡片槽不足 280，應退回整張置中")
	}
}
