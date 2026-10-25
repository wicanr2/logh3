package main

import (
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/presentation"
	"image"
	"testing"
)

func TestToolbarRevealHideAndInputBarrier(t *testing.T) {
	tb := toolbar{}
	tb.move(50, 20)
	if tb.visible || tb.blocks(50, 20) {
		t.Fatal("非頂緣操作意外攔截遊戲")
	}
	tb.move(50, 2)
	tb.move(50, 20)
	if !tb.visible || !tb.blocks(50, 20) || tb.actionAt(50, 20) != 1 {
		t.Fatal("功能列進入後無法保持或命中")
	}
	tb.move(100, 33)
	if tb.visible || tb.blocks(100, 33) {
		t.Fatal("移開後未恢復遊戲輸入")
	}
	tb.help = true
	tb.move(300, 300)
	if !tb.blocks(0, 479) || tb.actionAt(300, 348) != 4 {
		t.Fatal("說明沒有攔截底層或關閉鈕失效")
	}
	for i, b := range toolbarButtons {
		if !b.In(image.Rect(0, 0, 640, toolbarHeight)) || tb.actionAt(b.Min.X+2, b.Min.Y+2) != toolbarActions[i] {
			t.Fatal("繪製與命中未共用矩形")
		}
	}
}

func TestToolbarFontAndAIShareSettings(t *testing.T) {
	a := &app{g: &game.Game{}, hd: &presentation.Renderer{}}
	a.toolbarAction(3)
	a.toolbar.move(320, 40)
	if !a.toolbar.fontMenu || !a.toolbar.blocks(600, 400) || a.toolbar.actionAt(320, 40) != 100 {
		t.Fatal("字型下拉未攔截／命中")
	}
	a.toolbarAction(100)
	if a.hd.FontStyle() != "original" || a.toolbar.fontMenu {
		t.Fatal("字型選擇沒有套用既有renderer")
	}
	a.toolbarAction(5)
	if !a.toolbar.aiMenu || a.toolbar.actionAt(500, 40) != 200 {
		t.Fatal("AI下拉未開啟")
	}
	a.toolbarAction(204)
	if a.g.AILevel() != 3 || a.toolbar.aiMenu {
		t.Fatal("工具列AI與Game不同步")
	}
	a.toolbarAction(5)
	a.toolbarAction(0)
	if a.toolbar.aiMenu || a.toolbar.fontMenu {
		t.Fatal("點選單外未關閉")
	}
}

func TestToolbarGraphicsActionKeepsGame(t *testing.T) {
	a := &app{hd: &presentation.Renderer{}}
	a.toolbarAction(2)
	if !a.original {
		t.Fatal("功能列沒有切換原版")
	}
	a.toolbarAction(2)
	if a.original {
		t.Fatal("功能列沒有切回HD")
	}
	a.toolbarAction(1)
	a.toolbarAction(4)
	if a.toolbar.help {
		t.Fatal("說明關閉失效")
	}
}
