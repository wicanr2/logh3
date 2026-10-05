//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/rules"
)

func main() {
	t, err := rules.LoadBattleTerrain("docs/internals/data/battle-terrain.json")
	if err != nil {
		panic(err)
	}
	for region := 0; region < 32; region++ {
		b, err := os.ReadFile(fmt.Sprintf("/orig/MAP%02d.CMP", region))
		if err != nil {
			panic(err)
		}
		cells, err := asset.DecodeCMP(b, asset.BattleMapCols, asset.BattleMapRows)
		if err != nil {
			panic(err)
		}
		for pi, site := range rules.PlanetSites(cells, 16, 20) {
			for k := 0; k < 64; k++ {
				dirs := [8][2]int{{0, -32}, {32, -32}, {32, 0}, {32, 32}, {0, 32}, {-32, 32}, {-32, 0}, {-32, -32}}
				d := dirs[k&7]
				x, y := site.X+d[0]*(k/8+1), site.Y+d[1]*(k/8+1)
				if x >= 0 && x < 0x7ef && y >= 0 && y < 0x7e7 && !t.Passable(cells, x, y) {
					fmt.Printf("region=%d planet=%d site=%d,%d first_blocked_ring=%d point=%d,%d\n", region, pi, site.X, site.Y, k, x, y)
					break
				}
			}
		}
	}
}
