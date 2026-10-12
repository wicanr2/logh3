package presentation

import (
	"bytes"
	"crypto/sha256"
	"image"
	"image/color"
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
)

func themeTestRenderer(t *testing.T) *Renderer {
	t.Helper()
	themes, err := loadFactionThemes("../../assets/hd")
	if err != nil {
		t.Fatal(err)
	}
	if themes == nil {
		t.Skip("缺HD主題，不用替代資料")
	}
	return &Renderer{theme: true, classic: true, factionThemes: themes, frame: image.NewRGBA(image.Rect(0, 0, 1920, 1440)), material: image.NewRGBA(image.Rect(0, 0, 256, 256))}
}

func TestFactionUIOrderedOcclusionAndOriginalFallback(t *testing.T) {
	r := themeTestRenderer(t)
	// 此案例隔離色彩角色與覆蓋；逐列漸層由正常概念／完整視窗核對。
	r.factionThemes["alliance"].Gradient = 0
	var p gfx.Palette
	p[1], p[2], p[3], p[7] = [3]byte{0, 20, 24}, [3]byte{16, 20, 24}, [3]byte{44, 44, 48}, [3]byte{60, 60, 60}
	f := gfx.NewFramebuffer(1)
	f.InterfaceTheme = "alliance"
	f.SurfaceRect(0, 0, 80, 80, 1)
	s := f.BeginGroup()
	f.Set(10, 10, 7)
	f.EndGroup(s, gfx.DrawCommand{Kind: gfx.DrawGlyph, X: 10, Y: 10, W: 1, H: 1})
	f.DrawSource("", 20, 20, 1, 1, 1, 1, []byte{1})
	f.DrawSource("", 30, 30, 1, 1, 1, 1, []byte{7})
	s = f.BeginGroup()
	f.FillRect(40, 40, 20, 20, 2)
	f.EndGroup(s, gfx.DrawCommand{Kind: gfx.DrawPanel, Key: "ui/primary", X: 40, Y: 40, W: 20, H: 20})
	before := bytes.Clone(f.Pix)
	im := r.Render(f, p)
	skin := r.factionThemes["alliance"]
	ink := skin.White
	if skin.Ink.A != 0 {
		ink = skin.Ink
	}
	if im.RGBAAt(4, 4) != skin.Surface || im.RGBAAt(31, 31) != ink {
		t.Fatal("陣營面板或原字形配色未接線")
	}
	if im.RGBAAt(61, 61) != paletteColour(p, 1) || im.RGBAAt(91, 91) != paletteColour(p, 7) {
		t.Fatal("原圖被UI換色")
	}
	if im.RGBAAt(130, 130) != skin.Primary {
		t.Fatal("主要指令沒有沿既有矩形換色")
	}
	f.SurfaceRect(20, 20, 1, 1, 1)
	if r.Render(f, p).RGBAAt(61, 61) != skin.Surface {
		t.Fatal("後畫面板沒有遮住原圖")
	}
	if !bytes.Equal(before, f.Pix) || !bytes.Equal(gfx.ReplayIndexed(f.Commands), f.Pix) {
		t.Fatal("主題修改原色號或畫序")
	}
	f.InterfaceTheme = "unknown"
	if r.Render(f, p).RGBAAt(4, 4) != paletteColour(p, 1) {
		t.Fatal("未知主題未回退或殘留上一陣營色")
	}
}

func TestFactionStarMapCacheAndInformationStates(t *testing.T) {
	r := themeTestRenderer(t)
	c := gfx.DrawCommand{Kind: gfx.DrawStarMap, X: 4, Y: 40, W: 320, H: 300, SourceW: 320, SourceH: 300, Base: make([]byte, 320*300), Regions: bytes.Repeat([]byte{1}, 320*300), Colours: make([]byte, 32)}
	// 核准第三稿保留所有原功能色，陣營色只套在操作介面。
	c.Base[10*320+10], c.Base[100*320+100] = 8, 8
	var p gfx.Palette
	for i := range p {
		p[i] = [3]byte{byte(i * 4), byte(i * 3), byte(i * 2)}
	}
	f := gfx.NewFramebuffer(0)
	f.Commands = append(f.Commands, c)
	f.InterfaceTheme = "empire"
	first := sha256.Sum256(r.Render(f, p).Pix)
	emblem := r.frame.RGBAAt((4+10)*3+1, (40+10)*3+1)
	f.InterfaceTheme = "alliance"
	second := sha256.Sum256(r.Render(f, p).Pix)
	if first != second {
		t.Fatal("星圖功能色被陣營主題改動")
	}
	if r.frame.RGBAAt((4+10)*3+1, (40+10)*3+1) != emblem {
		t.Fatal("徽章識別色被主題改動")
	}
	f.InterfaceTheme = "empire"
	if sha256.Sum256(r.Render(f, p).Pix) != first {
		t.Fatal("切回帝國沒有恢復原主題")
	}
	for _, id := range []string{"empire", "alliance"} {
		seen := map[color.RGBA]bool{}
		f.InterfaceTheme = id
		for state := 0; state < 16; state++ {
			c.Colours[0] = byte(state)
			f.Commands[1] = c
			r.Render(f, p)
			seen[r.frame.RGBAAt(30, 150)] = true
		}
		if len(seen) != 16 {
			t.Fatalf("%s合併了情報色號：%d", id, len(seen))
		}
	}
	if len(r.starCache) > 8 {
		t.Fatal("星圖快取沒有上限")
	}
}
