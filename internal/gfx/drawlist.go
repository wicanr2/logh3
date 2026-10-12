package gfx

import "image"

// DrawKind 是正式顯示層可替換的操作；Fallback 保留原畫序及原色號。
type DrawKind uint8

const (
	DrawBarrier DrawKind = iota
	DrawPixels
	DrawFill
	DrawImage
	DrawPanel
	DrawSurface
	DrawGlyph
	DrawIcon
	DrawTerrain
	DrawMinimap
	DrawStarMap
	DrawUnitSymbol
	DrawDirection
	DrawCursor
	DrawAnimation
	DrawProjection
)

type DrawCommand struct {
	Kind                   DrawKind
	X, Y, W, H             int
	Colour                 uint8
	Key                    string
	SourceW, SourceH       int
	Source                 []uint8
	Writes                 []uint32
	Fallback               []DrawCommand
	Code                   uint16
	Half, Bold             bool
	Shadow                 uint8
	ShadowDY               int
	Clip                   image.Rectangle
	WorldX, WorldY         int
	OpenEdges              uint8 // 左1、右2、上4、下8：相鄰世界格不是星雲。
	Base, Regions, Colours []uint8
}

// BeginGroup 隔開前後原像素寫入，避免後畫像素被併入較早的命令。
func (f *Framebuffer) BeginGroup() int {
	i := len(f.Commands)
	f.Commands = append(f.Commands, DrawCommand{Kind: DrawBarrier})
	return i
}

// EndGroup 用語意操作包住原操作。HD 缺素材時仍能完整重播。
func (f *Framebuffer) EndGroup(start int, command DrawCommand) {
	if start < 0 || start >= len(f.Commands) {
		return
	}
	command.Fallback = append([]DrawCommand(nil), f.Commands[start+1:]...)
	f.Commands = append(f.Commands[:start], command)
}

func (f *Framebuffer) recordPixel(x, y int, c uint8) {
	if f.mute != 0 {
		return
	}
	n := len(f.Commands)
	if n == 0 || f.Commands[n-1].Kind != DrawPixels {
		f.Commands = append(f.Commands, DrawCommand{Kind: DrawPixels})
		n++
	}
	f.Commands[n-1].Writes = append(f.Commands[n-1].Writes, uint32(y*Width+x)<<8|uint32(c))
}

// SurfaceRect 只替換介面底色；原色號寫入與 FillRect 相同。
func (f *Framebuffer) SurfaceRect(x, y, w, h int, c uint8) {
	start := f.BeginGroup()
	f.FillRect(x, y, w, h, c)
	f.EndGroup(start, DrawCommand{Kind: DrawSurface, X: x, Y: y, W: w, H: h, Colour: c})
}

// DrawSource 記錄一次不透明原圖貼入，保留來源尺寸與本次裁切。
func (f *Framebuffer) DrawSource(key string, x, y, w, h, sw, sh int, source []uint8) {
	if sw <= 0 || sh <= 0 || len(source) < sw*sh {
		return
	}
	w, h = min(w, sw), min(h, sh)
	if w <= 0 || h <= 0 {
		return
	}
	f.mute++
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			f.Set(x+i, y+j, source[j*sw+i])
		}
	}
	f.mute--
	if f.mute == 0 {
		f.Commands = append(f.Commands, DrawCommand{Kind: DrawImage, Key: key, X: x, Y: y, W: w, H: h,
			SourceW: sw, SourceH: sh, Source: append([]uint8(nil), source...)})
	}
}

// ReplayIndexed 是顯示紀錄的原色重播，供完整性驗證使用。
func ReplayIndexed(commands []DrawCommand) []uint8 {
	result := make([]uint8, Width*Height)
	var replay func([]DrawCommand)
	replay = func(commands []DrawCommand) {
		for _, c := range commands {
			switch c.Kind {
			case DrawPixels:
				for _, p := range c.Writes {
					if int(p>>8) < len(result) {
						result[p>>8] = uint8(p)
					}
				}
			case DrawFill:
				for y := max(0, c.Y); y < min(Height, c.Y+c.H); y++ {
					for x := max(0, c.X); x < min(Width, c.X+c.W); x++ {
						result[y*Width+x] = c.Colour
					}
				}
			case DrawImage:
				for y := max(0, -c.Y); y < min(c.H, Height-c.Y); y++ {
					for x := max(0, -c.X); x < min(c.W, Width-c.X); x++ {
						result[(c.Y+y)*Width+c.X+x] = c.Source[y*c.SourceW+x]
					}
				}
			default:
				replay(c.Fallback)
			}
		}
	}
	replay(commands)
	return result
}
