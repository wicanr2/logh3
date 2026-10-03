package game

import (
	"encoding/binary"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 088 READY：WORK保存原版控制值；讀到這些欄位不代表戰場續讀已完成。
const (
	battleWorkActive = 0x02
	battleWorkRegion = 0x03
	battleWorkTurn   = 0x06
	battleWorkView   = 0x26
)

func (g *Game) setScannedRegion(region int) {
	if region >= 0 && region <= rules.RegionCount && len(g.workRecord()) > battleWorkRegion {
		g.mutableWork()[battleWorkRegion] = byte(region)
	}
}

func (g *Game) battleViewAt(x, y int) screen.BattleView {
	tw, th, cols, rows := asset.BattleTileW, asset.BattleTileH, 28, 18
	if g.res.Battle != nil {
		field := g.res.Battle.Field
		if field.TileW > 0 && field.TileH > 0 && field.Cols > 0 && field.Rows > 0 {
			tw, th, cols, rows = field.TileW, field.TileH, field.Cols, field.Rows
		}
	}
	return screen.BattleView{
		Col: clampCol((x-cols*tw/2)/tw, cols),
		Row: clampRow((y-rows*th/2)/th, rows),
	}
}

func (g *Game) initBattleWork() {
	if g.battleRegion < 0 || g.battleRegion >= rules.RegionCount || len(g.workRecord()) < rules.WorkRecord {
		return
	}
	w := g.mutableWork()
	w[battleWorkActive], w[battleWorkRegion] = 1, byte(g.battleRegion)
	binary.LittleEndian.PutUint16(w[battleWorkTurn:], uint16(g.battleTurn-1))
	for seat := 0; seat < 2; seat++ {
		faction := int(w[0x11+seat])
		if faction > 1 {
			continue
		}
		for _, f := range g.battleFleets {
			if f.Faction != faction {
				continue
			}
			view := g.battleViewAt(f.X, f.Y)
			g.writeBattleView(w, seat, view)
			if !g.versus && seat == 0 {
				g.battleView = view
			}
			break
		}
	}
}

func (g *Game) writeBattleView(w []byte, seat int, view screen.BattleView) {
	o := battleWorkView + seat*4
	binary.LittleEndian.PutUint16(w[o:], uint16(view.Col*asset.BattleTileW))
	binary.LittleEndian.PutUint16(w[o+2:], uint16(view.Row*asset.BattleTileH))
}

// 原版12AA:034C..0373在Input返回後保存鏡頭；純捲動只改目前視窗。
func (g *Game) saveBattleWorkView() {
	w := g.workRecord()
	if len(w) < rules.WorkRecord || w[battleWorkActive] == 0 {
		return
	}
	seat := int(w[workActiveSeat])
	if seat > 1 || int(w[0x11+seat]) != g.faction {
		return
	}
	g.writeBattleView(g.mutableWork(), seat, g.battleView)
}

func (g *Game) updateBattleWorkTurn() {
	w := g.workRecord()
	if len(w) >= rules.WorkRecord && w[battleWorkActive] != 0 {
		binary.LittleEndian.PutUint16(g.mutableWork()[battleWorkTurn:], uint16(g.battleTurn-1))
	}
}

func (g *Game) clearBattleWork() {
	if len(g.workRecord()) > battleWorkActive {
		g.mutableWork()[battleWorkActive] = 0
	}
}

// 1261:0067及090A:0E6D的經費快取。Save只保存快取，不重跑戰場規則。
func (g *Game) refreshFleetExpenses(faction int) {
	if g.res.ShipKinds == nil || faction < 0 || faction > 1 || len(g.workRecord()) < rules.WorkRecord {
		return
	}
	rules.WriteWorkFleetExpenses(g.mutableWork(), faction, g.fleetUpkeep(faction), g.fleetMoveCost(faction))
}
