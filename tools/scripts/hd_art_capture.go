//go:build ignore

// 規格122：保存HD實際X11視窗，按HD畫面縮放核对全部可见像素。
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/bits"
	"os"
	"strconv"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

func main() {
	window := flag.Uint("window", 0, "X11 window ID")
	base := flag.String("base", "", "尚無X11快照的renderer PNG前綴")
	scale := flag.Float64("scale", 1.5, "1.5或2倍視窗；逐原生像素中心取樣")
	flag.Parse()
	c, err := xgb.NewConn()
	if err != nil {
		panic(err)
	}
	defer c.Close()
	w := xproto.Window(*window)
	g, err := xproto.GetGeometry(c, xproto.Drawable(w)).Reply()
	if err != nil {
		panic(err)
	}
	if *scale < 1 || *scale > 2 || int(g.Width) != int(640**scale) || int(g.Height) != int(480**scale) {
		panic("實際視窗形狀不符")
	}
	attr, err := xproto.GetWindowAttributes(c, w).Reply()
	if err != nil {
		panic(err)
	}
	setup := xproto.Setup(c)
	var visual *xproto.VisualInfo
	for _, s := range setup.Roots {
		for _, d := range s.AllowedDepths {
			for i := range d.Visuals {
				v := &d.Visuals[i]
				if v.VisualId == attr.Visual {
					visual = v
				}
			}
		}
	}
	if visual == nil {
		panic("未知X11 visual")
	}
	var format *xproto.Format
	for i := range setup.PixmapFormats {
		f := &setup.PixmapFormats[i]
		if f.Depth == g.Depth {
			format = f
		}
	}
	if format == nil || format.BitsPerPixel != 32 || format.ScanlinePad != 32 {
		panic("未支援的X11像素排列")
	}
	reply, err := xproto.GetImage(c, xproto.ImageFormatZPixmap, xproto.Drawable(w), 0, 0, g.Width, g.Height, ^uint32(0)).Reply()
	if err != nil {
		panic(err)
	}
	if len(reply.Data) != int(g.Width)*int(g.Height)*4 {
		panic("X11資料長度不符")
	}
	channel := func(v, mask uint32) uint8 {
		if mask == 0 || bits.OnesCount32(mask) != 8 {
			panic("未支援色彩遮罩")
		}
		return uint8((v & mask) >> bits.TrailingZeros32(mask))
	}
	capture := image.NewRGBA(image.Rect(0, 0, int(g.Width), int(g.Height)))
	for y := 0; y < int(g.Height); y++ {
		for x := 0; x < int(g.Width); x++ {
			b := reply.Data[(y*int(g.Width)+x)*4:]
			v := binary.LittleEndian.Uint32(b)
			if setup.ImageByteOrder == xproto.ImageOrderMSBFirst {
				v = binary.BigEndian.Uint32(b)
			}
			capture.SetRGBA(x, y, color.RGBA{channel(v, visual.RedMask), channel(v, visual.GreenMask), channel(v, visual.BlueMask), 255})
		}
	}
	write := func(path string, img image.Image) {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			panic(err)
		}
		if err = png.Encode(f, img); err != nil {
			f.Close()
			panic(err)
		}
		if err = f.Close(); err != nil {
			panic(err)
		}
	}
	write(*base+"-x11.png", capture)
	f, err := os.Open(*base + ".png")
	if err != nil {
		panic(err)
	}
	expected, err := png.Decode(f)
	f.Close()
	if err != nil {
		panic(err)
	}
	if expected.Bounds() != image.Rect(0, 0, 1920, 1440) {
		panic("renderer PNG形狀不符")
	}
	diff := image.NewRGBA(capture.Bounds())
	different, maxDifference := 0, 0
	bbox := [4]int{int(g.Width), int(g.Height), -1, -1}
	abs := func(v int) int {
		if v < 0 {
			return -v
		}
		return v
	}
	for y := 0; y < int(g.Height); y++ {
		for x := 0; x < int(g.Width); x++ {
			want := color.RGBAModel.Convert(expected.At(int((float64(x)+0.5)*3 / *scale), int((float64(y)+0.5)*3 / *scale))).(color.RGBA)
			got := capture.RGBAAt(x, y)
			r, gg, b := abs(int(got.R)-int(want.R)), abs(int(got.G)-int(want.G)), abs(int(got.B)-int(want.B))
			diff.SetRGBA(x, y, color.RGBA{uint8(r), uint8(gg), uint8(b), 255})
			if r != 0 || gg != 0 || b != 0 {
				different++
				bbox[0] = min(bbox[0], x)
				bbox[1] = min(bbox[1], y)
				bbox[2] = max(bbox[2], x)
				bbox[3] = max(bbox[3], y)
			}
			maxDifference = max(maxDifference, r, gg, b)
		}
	}
	write(*base+"-x11-diff.png", diff)
	var bounds any
	if different != 0 {
		bounds = bbox
	}
	b, err := json.MarshalIndent(map[string]any{"pixels": int(g.Width) * int(g.Height), "different": different, "bbox": bounds, "max_channel_difference": maxDifference,
		"window_id": strconv.FormatUint(uint64(*window), 10), "width": g.Width, "height": g.Height, "depth": g.Depth, "byte_order": setup.ImageByteOrder,
		"bits_per_pixel": format.BitsPerPixel, "masks": []uint32{visual.RedMask, visual.GreenMask, visual.BlueMask}, "normalization": "保存完整原始視窗；依最近鄰縮放核全部視窗像素，HD來源1920×1440，無排除區域"}, "", "  ")
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(*base+"-x11-comparison.json", append(b, '\n'), 0644); err != nil {
		panic(err)
	}
	fmt.Println("實際X11比較", different, "/", int(g.Width)*int(g.Height))
	if different != 0 {
		os.Exit(1)
	}
}
