package game

import "github.com/wicanr2/logh3/internal/rules"

// WORK+E0／+E1按陣營索引。0654:0034同時檢查旗標及回合餘數；
// 共同尾巴清該陣營旗標，0374:03DF／03E2在進位後重設兩格（規格088）。
const workMonthSeatFlag = 0xE0

func (g *Game) monthSeatWorkAvailable(f int) bool {
	return g.res != nil && f >= 0 && f < 2 && len(g.workRecord()) >= rules.WorkRecord
}

func (g *Game) monthSeatDue(f int) bool {
	return g.monthSeatWorkAvailable(f) && g.turn%rules.TurnsPerMonth == 0 &&
		g.workRecord()[workMonthSeatFlag+f] != 0
}

func (g *Game) clearMonthSeat(f int) {
	if !g.versus && g.monthSeatWorkAvailable(f) {
		g.mutableWork()[workMonthSeatFlag+f] = 0
	}
}

func (g *Game) armMonthSeats() {
	if !g.versus && g.monthSeatWorkAvailable(0) {
		w := g.mutableWork()
		w[workMonthSeatFlag], w[workMonthSeatFlag+1] = 1, 1
	}
}
