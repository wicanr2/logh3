package asset

import "fmt"

// FleetSceneSprites 是095下方場景的原版圖塊與指令標籤。
// 原始檔由玩家匯入，不內嵌素材。
type FleetSceneSprites struct {
	Blocks [4]*Image
	Labels [5]*Image
}

// DecodeFleetSceneSprite 解析整列plane格式；不能用word交錯的DecodeTiles。
func DecodeFleetSceneSprite(raw []byte, w, h int) (*Image, error) {
	if !((w == 48 && h == 16) || (w == 64 && h == 9)) {
		return nil, fmt.Errorf("交戰圖塊尺寸不符：%d×%d", w, h)
	}
	wb := w / 8
	if len(raw) != wb*h*4 {
		return nil, fmt.Errorf("交戰圖塊長度不符：%d，需%d", len(raw), wb*h*4)
	}
	im := &Image{W: w, H: h, Used: len(raw), Pix: make([]uint8, w*h)}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			for p := 0; p < 4; p++ {
				im.Pix[y*w+x] |= ((raw[y*wb*4+p*wb+x/8] >> (7 - x%8)) & 1) << p
			}
		}
	}
	return im, nil
}

// DecodeFleetSceneSprites 驗實際檔案形狀，錯誤回傳給匯入端。
func DecodeFleetSceneSprites(blocks, flags []byte) (*FleetSceneSprites, error) {
	if len(blocks) != 1536 || len(flags) != 1440 {
		return nil, fmt.Errorf("交戰素材長度不符：3DBLOCK=%d、COM_FLAG=%d", len(blocks), len(flags))
	}
	s := &FleetSceneSprites{}
	for i := range s.Blocks {
		im, err := DecodeFleetSceneSprite(blocks[i*384:(i+1)*384], 48, 16)
		if err != nil {
			return nil, err
		}
		s.Blocks[i] = im
	}
	for i := range s.Labels {
		im, err := DecodeFleetSceneSprite(flags[i*288:(i+1)*288], 64, 9)
		if err != nil {
			return nil, err
		}
		s.Labels[i] = im
	}
	return s, nil
}
