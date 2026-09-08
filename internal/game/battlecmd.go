package game

import (
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 戰場的命令選單（`docs/spec/023`）。
//
// 點自己陣營的分隊會開一張命令選單：旗艦（分隊序號 0）開艦隊選單
// 「移動／戰鬥／索敵」，其餘按艦種開四張分隊選單之一。點不到分隊
// （或點到別人的）就照舊開戰術選單。

// battleUnitAt 回畫面座標 (x,y) 底下是哪一個分隊，-1 ＝ 沒有。
//
// 單位是 16×16 的圖塊、左上角對齊世界座標，與 `DrawBattleUnits` 同一組
// 幾何。重疊時取**畫在最上面的那一個**（繪製順序的最後一個）——原版的
// 命中判定還沒逆向到這一層，這是 remake 的取捨（`docs/spec/023`）。
func (g *Game) battleUnitAt(x, y int) int {
	s := g.res.Battle
	if s == nil {
		return -1
	}
	fx, fy := s.Field.X, s.Field.Y
	ox := g.battleView.Col * s.Field.TileW
	oy := g.battleView.Row * s.Field.TileH
	for i := len(g.battleUnits) - 1; i >= 0; i-- {
		u := g.battleUnits[i]
		sx, sy := fx+u.X-ox, fy+u.Y-oy
		if x >= sx && x < sx+battleUnitTile && y >= sy && y < sy+battleUnitTile {
			return i
		}
	}
	return -1
}

// battleUnitTile 是戰場上一個分隊的圖塊邊長（`SPRITE.SPR`，`docs/re/85`）。
const battleUnitTile = 16

// 分隊行動（艦艇記錄 `+0x1E`）。0x7F ＝ 沒有自己的命令。
const (
	squadronMove   = 0
	squadronForm   = 1
	squadronScout  = 2
	squadronRejoin = 3
	squadronNone   = 0x7F
)

// 艦種：5 運補艦、6 修護艦、7 索敵分遣隊（`docs/internals/40`）。
const (
	shipKindSupply = 5
	shipKindRepair = 6
	shipKindScout  = 7
)

// commandMenuKind 回點到這個分隊該開哪一張選單。
func commandMenuKind(u rules.BattleUnit) screen.CommandMenuKind {
	if u.Squadron == 0 {
		return screen.CommandMenuFleet
	}
	switch u.Kind {
	case shipKindSupply:
		return screen.CommandMenuSupply
	case shipKindRepair:
		return screen.CommandMenuRepair
	case shipKindScout:
		return screen.CommandMenuScout
	}
	return screen.CommandMenuSquadron
}

// commandMenuEnabled 回每一項可不可選（`docs/spec/023`）。
func (g *Game) commandMenuEnabled(kind screen.CommandMenuKind, u rules.BattleUnit) []bool {
	if kind == screen.CommandMenuFleet {
		// 三項先全開，再按艦隊記錄關掉。
		on := []bool{true, true, true}
		f := g.battleFleet(u.Fleet)
		if f == nil {
			return on
		}
		if f.Flag0 {
			on[1], on[2] = false, false
		}
		if f.ScoutOut {
			on[2] = false
		}
		return on
	}
	// 分隊選單四張共用同一條：已經在移動或已經在歸隊就不能再下移動。
	move := u.Action != squadronMove && u.Action != squadronRejoin
	n := g.res.CommandMenus.ItemCount(kind)
	on := make([]bool, n)
	for i := range on {
		on[i] = true
	}
	if n > 0 {
		on[0] = move
	}
	return on
}

// battleFleet 回戰場上這一支艦隊的記錄；找不到回 nil。
func (g *Game) battleFleet(index int) *rules.BattleFleet {
	for i := range g.battleFleets {
		if g.battleFleets[i].Index == index {
			return &g.battleFleets[i]
		}
	}
	return nil
}

// openCommandMenu 在游標處開命令選單。
func (g *Game) openCommandMenu(unit, x, y int) {
	s := g.res.CommandMenus
	if s == nil || unit < 0 || unit >= len(g.battleUnits) {
		return
	}
	kind := commandMenuKind(g.battleUnits[unit])
	n := s.ItemCount(kind)
	if n == 0 {
		return
	}
	g.battleCmd = true
	g.battleCmdUnit = unit
	g.battleCmdKind = kind
	g.battleCmdX, g.battleCmdY = s.CommandMenuAt(x, y, n)
}

// closeCommandMenu 關掉命令選單。
func (g *Game) closeCommandMenu() { g.battleCmd = false }

// CommandMenuOpen 回命令選單開著沒有。
func (g *Game) CommandMenuOpen() bool { return g.battleCmd }

// commandMenuClick 處理命令選單開著時的左鍵；回有沒有吃掉這一下。
func (g *Game) commandMenuClick(x, y int) bool {
	if !g.CommandMenuOpen() {
		return false
	}
	s := g.res.CommandMenus
	n := s.ItemCount(g.battleCmdKind)
	unit := g.battleCmdUnit
	if unit < 0 || unit >= len(g.battleUnits) {
		g.closeCommandMenu()
		return true
	}
	row := s.CommandMenuRowAt(g.battleCmdX, g.battleCmdY, x, y, n)
	on := g.commandMenuEnabled(g.battleCmdKind, g.battleUnits[unit])
	g.closeCommandMenu()
	if row < 0 || (row < len(on) && !on[row]) {
		return true // 點在選單外面或點到不可選的：只把選單關掉
	}
	// 「歸隊」沒有子常式，選下去就是寫一個位元組（`docs/re/117`）。
	if g.battleCmdKind == screen.CommandMenuScout && row == 1 {
		g.battleUnits[unit].Action = squadronRejoin
	}
	// 其餘各項的行為還沒逆向（`docs/spec/023`「還沒做的」）。
	return true
}
