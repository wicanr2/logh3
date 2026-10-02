package game

// 088 READY：WORK控制上下文與固定玩家視角分開，只更新已證實的兩個bytes。
const workActiveFaction, workActiveSeat = 0x09, 0x0A

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
