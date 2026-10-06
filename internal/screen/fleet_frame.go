package screen

import (
	"fmt"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/ui"
)

// FleetScenePicture重播第0格至LastFrame；圖段名稱保留原始四byte。
type FleetScenePicture struct {
	Side      uint8
	Segment   string
	LastFrame int
}

// FleetFrameView是099的單幀來源；Units與Integrity按陣營索引。
// 圖段已由caller選定，本合成器不決定時間、傷害、聲音或命令。
type FleetFrameView struct {
	Scene     FleetSceneView
	Units     [2]FleetUnitView
	Weapon    FleetWeaponView
	Integrity [2]int
	Pictures  []FleetScenePicture
}

// RenderFleetFrame從空白畫布重建完整單幀，錯誤不回傳部分畫面。
func (s *FleetPresentationSpec) RenderFleetFrame(v FleetFrameView, panel *asset.Image,
	sprites *asset.FleetSceneSprites, bmg *asset.BMG, fnt *font.Font, tiles *ui.FrameTiles) (*gfx.Framebuffer, error) {
	if s == nil || fnt == nil || panel == nil || panel.W != 640 || panel.H != 400 || len(panel.Pix) != 640*400 {
		return nil, fmt.Errorf("缺交戰字型或640×400底圖")
	}
	ops, err := BuildFleetSceneOps(v.Scene, sprites)
	if err != nil {
		return nil, err
	}
	selected := map[int]FleetSceneUnit{}
	for _, list := range v.Scene.Sides {
		for _, u := range list {
			if u.Ships > 4000 {
				return nil, fmt.Errorf("交戰分隊艘數超出目前編成範圍")
			}
			selected[u.ID] = u
		}
	}
	attack, target := selected[v.Scene.AttackerID], selected[v.Scene.TargetID]
	if v.Weapon.Side != attack.Side || v.Weapon.Weapon != attack.Weapon {
		return nil, fmt.Errorf("武器標題與交戰來源不一致")
	}
	var buffers [2]*ui.TextBuffer
	for _, u := range []FleetSceneUnit{attack, target} {
		text := v.Units[u.Side]
		if text.Side != u.Side || text.Squadron != u.Squadron || text.Detail != u.Detail || text.Ships != u.Ships ||
			text.Visibility != v.Scene.Visibility || text.ShipsMax == 0 || text.ShipsMax > 4000 || text.Ships > text.ShipsMax {
			return nil, fmt.Errorf("交戰分隊文字與來源不一致")
		}
		buffers[u.Side], err = s.UnitBuffer(text)
		if err != nil {
			return nil, err
		}
		if _, ok := (EngagementGauge{FleetFaction: int(u.Side), Value: v.Integrity[u.Side]}).Height(); !ok {
			return nil, fmt.Errorf("交戰艦隊完整度超界")
		}
	}
	weapon := v.Weapon
	weapon.ShowResult = false
	middle, err := s.MiddleBuffer(weapon)
	if err != nil {
		return nil, err
	}
	if v.Weapon.ShowResult {
		if err := s.AppendResult(middle, FleetResultView{Side: target.Side, Destroyed: v.Weapon.Destroyed}); err != nil {
			return nil, err
		}
	}
	frames := make([][]asset.BMGFrame, len(v.Pictures))
	for i, p := range v.Pictures {
		if bmg == nil || p.Side > 1 || len(p.Segment) != 4 || p.LastFrame < 0 {
			return nil, fmt.Errorf("交戰圖段來源無效")
		}
		frames[i], err = bmg.Frames(p.Segment)
		if err != nil {
			return nil, err
		}
		if p.LastFrame >= len(frames[i]) {
			return nil, fmt.Errorf("交戰圖段格索引超界")
		}
		frames[i] = frames[i][:p.LastFrame+1]
	}
	fb := gfx.NewFramebuffer(0)
	for y := 0; y < panel.H; y++ {
		fb.WriteRow(0, y+40, panel.Pix[y*panel.W:(y+1)*panel.W])
	}
	for side, b := range buffers {
		w := s.Windows[side]
		ui.DrawTextBuffer(fb, fnt, tiles, b, w.X, w.Y)
		(EngagementGauge{FleetFaction: side, Value: v.Integrity[side]}).Draw(fb)
	}
	w := s.Windows[2]
	ui.DrawTextBuffer(fb, fnt, tiles, middle, w.X, w.Y)
	if len(frames) != 0 {
		planes := asset.PlanesFrom(fb.Pix)
		for i, list := range frames {
			x := 1
			if v.Pictures[i].Side == 1 {
				x = 47
			}
			for _, f := range list {
				if err := f.Draw(planes, asset.BMGWindowBase(x, 4)); err != nil {
					return nil, err
				}
			}
		}
		fb.CopyPixels(planes.Pixels())
	}
	if err := DrawFleetBlocks(fb, ops); err != nil {
		return nil, err
	}
	return fb, nil
}
