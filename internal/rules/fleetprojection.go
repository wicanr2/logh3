package rules

import "fmt"

// FleetProjectionUnit是097的完整編制來源；各側首筆為旗艦。
// 座標取目前位置，不取目的地、艦隊中心或戰場鏡頭。
type FleetProjectionUnit struct {
	ID   int
	Side uint8
	X, Y uint16
}

type FleetProjectionPoint struct{ X, Y int }

// FleetProjectedUnit保留來源池ID，Depth是旋轉後尚未壓縮的Y。
type FleetProjectedUnit struct {
	ID      int
	Side    uint8
	X, Y    int
	Depth   int
	Visible bool
}

// FleetSceneProjection不改交戰排序鍵；Order是Units的繪製索引。
type FleetSceneProjection struct {
	Angle      uint8
	References [2]FleetProjectionPoint
	Anchors    [2]FleetProjectionPoint
	Units      []FleetProjectedUnit
	Order      []int
}

// ProjectFleetScene依321F:000A生成雙艦隊投影。完整Game接線仍由092把關。
func ProjectFleetScene(sides [2][]FleetProjectionUnit) (FleetSceneProjection, error) {
	var out FleetSceneProjection
	if len(sides[0]) == 0 || len(sides[1]) == 0 || len(sides[0])+len(sides[1]) > 22 {
		return out, fmt.Errorf("交戰投影編制為空或超過22格")
	}
	seen := make(map[int]bool)
	for side, units := range sides {
		for _, u := range units {
			if int(u.Side) != side || u.ID < 0 || seen[u.ID] || u.X > formationMax || u.Y > formationMax {
				return out, fmt.Errorf("交戰投影池ID、陣營或目前座標無效")
			}
			seen[u.ID] = true
		}
		out.References[side] = FleetProjectionPoint{int(units[0].X), int(units[0].Y)}
	}
	a, b := out.References[0], out.References[1]
	out.Angle = uint8(-AtanTable256(b.X-a.X, b.Y-a.Y))
	out.Anchors = [2]FleetProjectionPoint{{170, 298}, {470, 298}}
	if FixedDist(a.X, a.Y, b.X, b.Y) <= FixedFromInt(200) {
		// 0223先把距離小數丟掉，再乘3除2。兩邊各算自己的距離。
		mid := FleetProjectionPoint{int(uint16(a.X+b.X) >> 1), int(uint16(a.Y+b.Y) >> 1)}
		out.Anchors[0].X = 320 - FixedDist(mid.X, mid.Y, a.X, a.Y).Int()*3/2
		out.Anchors[1].X = 320 + FixedDist(mid.X, mid.Y, b.X, b.Y).Int()*3/2
	}
	for side, units := range sides {
		ref, anchor := out.References[side], out.Anchors[side]
		for _, u := range units {
			x, y := (int(u.X)-ref.X)*3/2, (int(u.Y)-ref.Y)*3/2
			x, y = Rotate256(x, y, int(out.Angle))
			depth := y
			y /= 2
			x, y = x-y+anchor.X, y+anchor.Y
			out.Units = append(out.Units, FleetProjectedUnit{u.ID, u.Side, x, y, depth,
				x >= 8 && x <= 568 && y >= 220 && y <= 376})
			out.Order = append(out.Order, len(out.Units)-1)
		}
	}
	// 0643比較投影word的無號值。不可換成穩定排序，間隔交換會影響同鍵次序。
	gap := len(out.Order)
	for {
		gap = max(1, gap*10/13)
		if gap == 9 || gap == 10 {
			gap = 11
		}
		swapped := false
		for i := 0; i+gap < len(out.Order); i++ {
			a, b := out.Units[out.Order[i]], out.Units[out.Order[i+gap]]
			if uint16(a.Y) > uint16(b.Y) || (uint16(a.Y) == uint16(b.Y) && uint16(a.X) > uint16(b.X)) {
				out.Order[i], out.Order[i+gap] = out.Order[i+gap], out.Order[i]
				swapped = true
			}
		}
		if gap == 1 && !swapped {
			return out, nil
		}
	}
}
