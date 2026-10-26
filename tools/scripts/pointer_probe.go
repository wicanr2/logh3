//go:build ignore

// 152：正常輸入所得畫面之熱區／按鈕與重繪診斷，不取代正常GUI驗收。
package main

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"time"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/presentation"
)

func main() {
	res, err := bootstrap.Load("/orig", "docs/internals/data")
	must(err)
	g, err := game.New(res)
	must(err)
	r, err := presentation.Load("workplace/poster150/package-linux/LOGH3-v.1.6.0-20261024-linux-amd64/art", res.Faces)
	must(err)
	res.Faces.ArtworkPrefix = "face"
	g.Click(300, 230)
	for _, k := range []byte("138\r") {
		g.Key(k)
	}
	g.RightClick(300, 240)
	g.RightClick(300, 240)
	out := "workplace/pointer152"
	must(os.MkdirAll(out, 0755))
	probe := func(name string, x, y int) {
		g.MoveMouse(x, y)
		cursor := res.Cursor
		res.Cursor = nil
		start := time.Now()
		f := g.Frame()
		frameMS := float64(time.Since(start)) / 1e6
		res.Cursor = cursor
		var boxes []map[string]any
		var visit func([]gfx.DrawCommand)
		visit = func(cs []gfx.DrawCommand) {
			for _, c := range cs {
				v := image.Rect(c.X, c.Y, c.X+c.W, c.Y+c.H)
				if c.Kind == gfx.DrawPanel && (image.Pt(x, y).In(v) || !f.HoverControl.Empty() && !v.Intersect(f.HoverControl).Empty()) {
					boxes = append(boxes, map[string]any{"role": c.Key, "box": v, "contains_hover": f.HoverControl.In(v)})
				}
				visit(c.Fallback)
			}
		}
		visit(f.Commands)
		if name == "command-system" {
			var box image.Rectangle
			var trace func([]gfx.DrawCommand)
			trace = func(cs []gfx.DrawCommand) { for _,c := range cs {
				v:=image.Rect(c.X,c.Y,c.X+c.W,c.Y+c.H)
				if c.Kind==gfx.DrawPanel && (c.Key=="ui/button" || c.Key=="ui/primary" || c.Key=="" && v.Min.Y==f.HoverControl.Min.Y && v.Max.Y==f.HoverControl.Max.Y) && f.HoverControl.In(v) { box=v; fmt.Println("HOVER candidate",c.Kind,c.Key,v); continue }
				if !box.Empty() && !v.Intersect(box).Empty() && (c.Kind==gfx.DrawPanel || c.Kind==gfx.DrawImage || c.Kind==gfx.DrawSurface && !v.In(box) || c.Kind==gfx.DrawFill && !v.In(box)) { fmt.Println("HOVER covered",c.Kind,c.Key,v,"control",f.HoverControl); box=image.Rectangle{} }
				trace(c.Fallback)
			} }
			trace(f.Commands)
		}
		start = time.Now()
		im := r.Render(f, g.Palette())
		renderMS := float64(time.Since(start)) / 1e6
		before := append([]byte(nil), im.Pix...)
		f.HoverControl = image.Rectangle{}
		after := r.Render(f, g.Palette())
		changed := 0
		for i := 0; i < len(before); i += 4 {
			if string(before[i:i+4]) != string(after.Pix[i:i+4]) {
				changed++
			}
		}
		b, _ := json.Marshal(map[string]any{"name": name, "state": g.State().String(), "boxes": boxes, "hover_pixels": changed, "frame_ms": frameMS, "render_ms": renderMS})
		fmt.Println(string(b))
		g.MoveMouse(x, y)
		res.Cursor = nil
		f = g.Frame()
		res.Cursor = cursor
		must(r.WritePNG(filepath.Join(out, name+".png"), f, g.Palette()))
	}
	probe("command-system", 544, 125)
	probe("command-list", 445, 125)
	g.Click(544, 125)
	probe("system-save", 560, 165)
	probe("system-settings", 560, 185)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
