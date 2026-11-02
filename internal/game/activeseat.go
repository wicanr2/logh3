package game

// 088 READY：WORK控制上下文與固定玩家視角分開，只更新已證實的兩個bytes。
const workActiveFaction, workActiveSeat = 0x09, 0x0A

// seatControl 是WORK+0C..14的九bytes，依模式（0帝國、1同盟、2對戰），研究384；
// 模式、兩座位配置CPU、兩座位目前CPU、兩座位陣營、兩座位首都。
var seatControl = [3][9]byte{
	{0, 0, 1, 0, 1, 0, 1, 0, 31},
	{1, 0, 1, 0, 1, 1, 0, 31, 0},
	{2, 0, 0, 0, 0, 0, 1, 0, 31},
}

// writeSeatControl寫入模式的控制bytes，保留其他WORK欄位。原版新遊戲與讀檔選陣營都寫這一段。
func (g *Game) writeSeatControl(mode int) {
	if mode < 0 || mode >= len(seatControl) || len(g.workRecord()) < 0x15 {
		return
	}
	copy(g.mutableWork()[0x0C:0x15], seatControl[mode][:])
}

func (g *Game) setActiveSeat(seat int) {
	if g.res == nil || seat < 0 || seat > 1 || g.faction < 0 || g.faction > 1 || len(g.workRecord()) <= workActiveSeat {
		return
	}
	faction := g.faction
	if g.versus {
		faction = seat
	} else if seat == 1 {
		faction = 1 - faction
	}
	w := g.mutableWork()
	w[workActiveFaction], w[workActiveSeat] = byte(faction), byte(seat)
}

func (g *Game) setHumanSeat() {
	seat := 0
	if g.versus {
		seat = g.faction
	}
	g.setActiveSeat(seat)
}

func (g *Game) setProcessingFaction() {
	if g.res == nil || g.faction < 0 || g.faction > 1 || len(g.workRecord()) <= workActiveSeat {
		return
	}
	// 座位保留；原版0374主迴圈在命令完成後設活動陣營FF。
	g.mutableWork()[workActiveFaction] = 0xFF
}
