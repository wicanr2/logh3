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

func TestCommandMenuHoverSharesClickGeometry(t *testing.T) {
	g, e := New(load(t))
	if e != nil {
		t.Fatal(e)
	}
	g.Click(300, 230)
	for _, c := range []byte("138\r") {
		g.Key(c)
	}
	g.RightClick(300, 240)
	g.RightClick(300, 240)
	g.MoveMouse(548, 125)
	if g.Frame().HoverControl != image.Rect(544, 120, 608, 140) {
		t.Fatal("主指令仍沿JSON的舊滑鼠y偏移")
	}
	g.Click(544, 125)
	g.MoveMouse(560, 165)
	f := g.Frame()
	if !f.HoverMenu || f.HoverControl != image.Rect(544, 160, 592, 180) {
		t.Fatal("存進度hover未沿commandBar正式列", f.HoverControl)
	}
	g.MoveMouse(532, 165)
	if !g.Frame().HoverControl.Empty() {
		t.Fatal("選單邊框不能點卻出現hover")
	}
}
