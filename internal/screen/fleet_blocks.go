package screen

import (
	"fmt"
	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
)

// FleetBlockKind 是095已投影下方場景的操作種類。
type FleetBlockKind uint8

const (
	FleetBlockBackground FleetBlockKind = iota
	FleetBlockSprite
	FleetBlockLine
)

// FleetBlockOp 的Y/Y1是原版虛擬座標，合成器只加40一次。
// Image由row-plane解析器提供；不借用SHIP+42保存演出順序。
type FleetBlockOp struct {
	Kind         FleetBlockKind
	X, Y, X1, Y1 int
	Count        int
	Image        *asset.Image
	Ink          uint8
}

// FleetBlockBackgroundRect 與 FleetBlockSpriteClip 分別對應背景與圖塊邊界。
func FleetBlockBackgroundRect() ClipRect { return ClipRect{8, 208, 631, 431} }
func FleetBlockSpriteClip() ClipRect     { return ClipRect{8, 208, 615, 429} }

func validateFleetBlockOp(op FleetBlockOp) error {
	switch op.Kind {
	case FleetBlockBackground:
		return nil
	case FleetBlockLine:
		if op.Ink > 15 || op.X < -32768 || op.X > 32767 || op.X1 < -32768 || op.X1 > 32767 ||
			op.Y < -32768 || op.Y > 32767 || op.Y1 < -32768 || op.Y1 > 32767 {
			return fmt.Errorf("交戰線段參數越界")
		}
	case FleetBlockSprite:
		im := op.Image
		if im == nil || !((im.W == 48 && im.H == 16) || (im.W == 64 && im.H == 9)) || len(im.Pix) != im.W*im.H ||
			op.X < 2 || op.X > 568 || op.Y < -324 || op.Y > 385 || op.Count < 1 || op.Count > 131 {
			return fmt.Errorf("交戰圖塊參數不在095合法域")
		}
		for _, p := range im.Pix {
			if p > 15 {
				return fmt.Errorf("交戰圖塊色號越界")
			}
		}
		for n := 0; n < op.Count; n++ {
			y := op.Y - 4*n
			if y+im.H == 168 {
				return fmt.Errorf("原版零高度裁切未支援：Y=%d、高%d", y, im.H)
			}
		}
	default:
		return fmt.Errorf("未知交戰場景操作%d", op.Kind)
	}
	return nil
}

// DrawFleetBlocks 合成完整095操作序列。先驗全部資料，錯誤不改畫布。
// 092尚未READY，Game不能用它猜補完整交戰生命週期。
func DrawFleetBlocks(fb *gfx.Framebuffer, ops []FleetBlockOp) error {
	if fb == nil || len(fb.Pix) != gfx.Width*gfx.Height {
		return fmt.Errorf("交戰畫布形狀不符")
	}
	for _, op := range ops {
		if err := validateFleetBlockOp(op); err != nil {
			return err
		}
	}
	mask := uint8(15)
	for _, op := range ops {
		switch op.Kind {
		case FleetBlockBackground:
			r := FleetBlockBackgroundRect()
			for y := r.Y0; y <= r.Y1; y++ {
				for x := r.X0; x <= r.X1; x++ {
					ink := uint8(0)
					if (y-r.Y0)%12 == 0 || (x-r.X0+y-r.Y0)%24 == 0 {
						ink = 14
					}
					fb.Set(x, y, ink)
				}
			}
			mask = 15
		case FleetBlockSprite:
			r, im := FleetBlockSpriteClip(), op.Image
			for n := 0; n < op.Count; n++ {
				top := op.Y + 40 - 4*n
				if top > r.Y1 || top+im.H <= r.Y0 {
					continue
				}
				for sy := 0; sy < im.H; sy++ {
					for sx := 0; sx < im.W; sx++ {
						x, y := op.X+sx, top+sy
						ink := im.Pix[sy*im.W+sx]
						if ink != 0 && r.Inside(x, y) {
							fb.Set(x, y, ink)
						}
					}
				}
				mask = 8
			}
		case FleetBlockLine:
			r := ClipRect{0, 40, 639, 439}
			x0, y0, x1, y1, ok := r.Clip(op.X, op.Y+40, op.X1, op.Y1+40)
			if !ok {
				continue
			}
			r.line(x0, y0, x1, y1, func(x, y int) { fb.Set(x, y, (fb.At(x, y)&(15^mask))|(op.Ink&mask)) })
			mask = 15
		}
	}
	return nil
}
