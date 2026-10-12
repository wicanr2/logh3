package game

import (
	"image"
	"testing"
)

func TestControlHoverUsesVisibleMonthlyCoordinates(t *testing.T) {
	for _, f := range []int{0, 1} {
		g, e := New(load(t))
		if e != nil {
			t.Fatal(e)
		}
		g.Click(300, 210+20*f)
		for _, c := range []byte("138\r") {
			g.Key(c)
		}
		g.MoveMouse(448, 350)
		if !image.Pt(448, 350).In(g.Frame().HoverControl) {
			t.Fatal("會計按鈕仍使用舊滑鼠偏移")
		}
		g.MoveMouse(448, 310)
		if !g.Frame().HoverControl.Empty() {
			t.Fatal("舊偏移位置誤亮")
		}
		g.Click(448, 350)
		g.MoveMouse(487, 329)
		if g.Frame().HoverControl != image.Rect(planPctDown.X0, planPctDown.Y0, planPctDown.X1+1, planPctDown.Y1+1) {
			t.Fatal("預算箭頭沒有共用實際命中來源")
		}
		g.MoveMouse(600, 420)
		if !g.Frame().HoverControl.Empty() {
			t.Fatal("移開後殘留反光")
		}
	}
}
