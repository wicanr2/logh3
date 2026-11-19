package mobileshell

import "testing"

// 邏輯畫布：高固定 480，寬依螢幕比例，不小於 640；原版畫面水平置中。
func TestLogicalSizeAndGamePoint(t *testing.T) {
	for _, c := range []struct{ ow, oh, w, h int }{
		{2400, 1080, 1066, 480}, // 20:9
		{1280, 576, 1066, 480},
		{1920, 1080, 853, 480}, // 16:9
		{1024, 768, 640, 480},  // 4:3
		{800, 800, 640, 480},   // 比 4:3 還窄：不小於 640
		{0, 0, 640, 480},
	} {
		if w, h := LogicalSize(c.ow, c.oh); w != c.w || h != c.h {
			t.Errorf("LogicalSize(%d,%d) = %d,%d，預期 %d,%d", c.ow, c.oh, w, h, c.w, c.h)
		}
	}
	s := &Shell{lastW: 1066}
	if x, y, ok := s.gamePoint(533, 240); !ok || x != 320 || y != 240 {
		t.Errorf("中央換算成 %d,%d %v", x, y, ok)
	}
	if _, _, ok := s.gamePoint(100, 240); ok {
		t.Error("左側欄被當成遊戲畫面")
	}
	if _, _, ok := s.gamePoint(1000, 240); ok {
		t.Error("右側欄被當成遊戲畫面")
	}
}
