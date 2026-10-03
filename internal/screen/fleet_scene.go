package screen

import (
	"fmt"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/rules"
)

// FleetSceneUnit附上06FC消費的SHIP欄位；保留完整編制，零艘亦不提前刪除。
type FleetSceneUnit struct {
	rules.FleetProjectionUnit
	Squadron, Detail, Weapon uint8
	Ships                    uint16
}

type FleetSceneView struct {
	Sides                [2][]FleetSceneUnit
	AttackerID, TargetID int
	Visibility           FleetVisibility
	Facing               [2]uint8
}

// BuildFleetSceneOps依097投影及095操作契約重建06FC整段下方場景。
// 先完成資料與裁切驗證，錯誤時不回傳可部分繪製的計畫。
func BuildFleetSceneOps(view FleetSceneView, sprites *asset.FleetSceneSprites) ([]FleetBlockOp, error) {
	if sprites == nil || view.Visibility.Mode > 2 || view.Visibility.PlayerSide > 1 {
		return nil, fmt.Errorf("交戰場景素材或觀看設定無效")
	}
	var sources [2][]rules.FleetProjectionUnit
	units := make(map[int]FleetSceneUnit)
	for side, list := range view.Sides {
		for _, u := range list {
			sources[side] = append(sources[side], u.FleetProjectionUnit)
			units[u.ID] = u
		}
	}
	p, err := rules.ProjectFleetScene(sources)
	if err != nil {
		return nil, err
	}
	attack, aOK := units[view.AttackerID]
	target, tOK := units[view.TargetID]
	if !aOK || !tOK || attack.ID == target.ID || attack.Side == target.Side || attack.Weapon > 3 {
		return nil, fmt.Errorf("交戰場景攻方、目標或武器無效")
	}
	projected := make(map[int]rules.FleetProjectedUnit)
	for _, u := range p.Units {
		projected[u.ID] = u
	}
	ops := []FleetBlockOp{{Kind: FleetBlockBackground}}
	count := func(u FleetSceneUnit) int { return int(uint16(u.Ships+499)) / 500 }
	for _, idx := range p.Order {
		pos := p.Units[idx]
		if !pos.Visible {
			continue
		}
		u := units[pos.ID]
		control := view.Visibility.SideControl[1]
		if u.Side == view.Visibility.PlayerSide {
			control = view.Visibility.SideControl[0]
		}
		known := u.Squadron == 0 && (u.Detail == 3 || (view.Visibility.Mode != 2 && control == 0))
		sprite := int(u.Side) * 2
		if !known {
			sprite++
		}
		// 2376:083B的jcxz讓零塊不畫；仍保留選取標籤與繪製順序。
		if n := count(u); n != 0 {
			ops = append(ops, FleetBlockOp{Kind: FleetBlockSprite, X: pos.X, Y: pos.Y, Count: n, Image: sprites.Blocks[sprite]})
		}
	}
	for i, u := range []FleetSceneUnit{attack, target} {
		pos := projected[u.ID]
		if !pos.Visible {
			continue
		}
		x, y := pos.X-6, pos.Y-4*(count(u)-1)-20
		if i == 0 {
			ops = append(ops, FleetBlockOp{Kind: FleetBlockSprite, X: x, Y: y, Count: 1, Image: sprites.Labels[3]})
		}
		label := 4
		if i == 0 {
			label = int(u.Weapon)
		}
		ops = append(ops, FleetBlockOp{Kind: FleetBlockSprite, X: x, Y: y + 9, Count: 1, Image: sprites.Labels[label]})
	}
	line := func(x0, y0, x1, y1 int, ink uint8) {
		ops = append(ops, FleetBlockOp{Kind: FleetBlockLine, X: x0, Y: y0, X1: x1, Y1: y1, Ink: ink})
	}
	top := func(u FleetSceneUnit) (int, int) {
		pos := projected[u.ID]
		return pos.X + 18, pos.Y - 2 - 4*(count(u)-1)
	}
	for _, u := range []FleetSceneUnit{attack, target} {
		if projected[u.ID].Visible {
			x, y := top(u)
			line(x, y, x, y+8, 4)
		}
	}
	if projected[attack.ID].Visible && projected[target.ID].Visible {
		x0, y0 := top(attack)
		x1, y1 := top(target)
		line(x0, y0-1, x1, y1-1, 6)
		line(x0, y0, x1, y1, 4)
		line(x0, y0+1, x1, y1+1, 6)
	}
	// 原版DS:6072的四組線段，以陣營固定角落呈現朝向。
	segments := [4][4]int{{16, 0, -16, -16}, {16, 0, -16, 16}, {-16, -16, 0, 0}, {-16, 16, 0, 0}}
	for _, u := range []FleetSceneUnit{attack, target} {
		baseX, baseY := 48, 200
		if u.Side != 0 {
			baseX, baseY = 592, 368
		}
		angle := int(uint8(view.Facing[u.Side] + p.Angle))
		for _, seg := range segments {
			x0, y0 := rules.Rotate256(seg[0], seg[1], angle)
			x1, y1 := rules.Rotate256(seg[2], seg[3], angle)
			x0, y0 = x0-y0/2+baseX, y0/2+baseY
			x1, y1 = x1-y1/2+baseX, y1/2+baseY
			line(x0, y0, x1, y1, 13)
		}
	}
	for _, op := range ops {
		if err := validateFleetBlockOp(op); err != nil {
			return nil, err
		}
	}
	return ops, nil
}
