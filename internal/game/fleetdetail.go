package game

import (
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 「資料」點一列開出來的艦隊詳細資料（`docs/re/112`、規格 `docs/spec/022`）。
//
// 版面在 `internal/screen`；這一層負責把值湊齊。每一格的來源都反組譯過，
// 沒有猜的欄位。

// OpenFleetDetail 打開第 row 列的詳細資料；那一列不是艦隊就不開。
func (g *Game) OpenFleetDetail(row int) bool {
	if g.res.FleetDetail == nil || row < 0 || row >= len(g.battleDataUnits(g.battleDataTab)) {
		return false
	}
	g.battleDetail, g.battleDetailRow = true, row
	return true
}

// CloseFleetDetail 回到「資料」那張表。
func (g *Game) CloseFleetDetail() { g.battleDetail = false }

// FleetDetailOpen 回詳細資料開著沒有。
func (g *Game) FleetDetailOpen() bool { return g.battleDetail }

// battleDataUnits 回某個分頁列出來的分隊（行星分頁沒有分隊）。
func (g *Game) battleDataUnits(tab int) []rules.BattleUnit {
	if tab == 1 {
		return nil
	}
	want := g.faction
	if tab == 2 {
		want = 1 - g.faction
	}
	var out []rules.BattleUnit
	for _, u := range g.battleUnits {
		if u.Faction == want {
			out = append(out, u)
		}
	}
	return out
}

// admiralRec 回提督表第 i 筆；越界回 nil。
func (g *Game) admiralRec(i int) []byte {
	if i < 0 || (i+1)*AdmiralRecord > len(g.res.Admiral) {
		return nil
	}
	return g.res.Admiral[i*AdmiralRecord : (i+1)*AdmiralRecord]
}

// admiralLine 回「名字 ＋ 軍階」的字碼。
func (g *Game) admiralLine(i int) []uint16 {
	rec := g.admiralRec(i)
	if rec == nil || i >= len(g.res.AdmiralNames) {
		return nil
	}
	out := append([]uint16(nil), g.res.AdmiralNames[i]...)
	return append(out, g.rankCodes(rules.RankOf(int(rec[admiralMeritOff])))...)
}

// 提督記錄上這一層讀的兩個欄位（`docs/re/112`）。
const (
	admiralMeritOff   = 0x06 // 畫面上的「等級」，也是軍階的來源
	admiralExpOff     = 0x02 // 經驗值（32 位元）
	admiralCommandOff = 0x0B // 畫面上的「統率力」
)

func u32At(rec []byte, off int) int {
	if rec == nil || off+4 > len(rec) {
		return 0
	}
	return int(rec[off]) | int(rec[off+1])<<8 | int(rec[off+2])<<16 | int(rec[off+3])<<24
}

func byteAt(rec []byte, off int) int {
	if rec == nil || off >= len(rec) {
		return 0
	}
	return int(rec[off])
}

// fleetDetailView 組出詳細資料要顯示的內容。
func (g *Game) fleetDetailView() (screen.FleetDetailView, bool) {
	var v screen.FleetDetailView
	units := g.battleDataUnits(g.battleDataTab)
	if g.battleDetailRow >= len(units) || g.res.FleetDetail == nil {
		return v, false
	}
	u := units[g.battleDetailRow]
	f := g.fleetByIndex(u.Fleet)
	if f == nil {
		return v, false
	}
	adm := g.admiralRec(f.Admiral)
	adj := g.admiralRec(f.Adjutant)

	v.FactionInk = battleDataInk(u.Faction)
	if g.res.Battle != nil && len(g.res.Battle.Status.Labels.Factions) > u.Faction%2 {
		v.Faction = g.res.Battle.Status.Labels.Factions[u.Faction%2].Codes
	}
	v.FleetNo = f.Number
	v.Commander = g.admiralLine(f.Admiral)
	v.CmdLevel, v.CmdExp = byteAt(adm, admiralMeritOff), u32At(adm, admiralExpOff)
	if adj != nil {
		v.Adjutant = g.admiralLine(f.Adjutant)
		v.AdjLevel, v.AdjExp = byteAt(adj, admiralMeritOff), u32At(adj, admiralExpOff)
	}

	// 十二個數值。順序與規格的 values 相同。
	ab := rules.FleetAbilities(adm, adj)
	turn, cruise, fight := fleetSpeeds(g.battleUnits, u.Fleet)
	v.Values = [12]int{
		f.Morale, turn, cruise, fight,
		byteAt(adm, admiralCommandOff), int(ab[9]), int(ab[10]), int(ab[3]),
		int(ab[0]), int(ab[1]), int(ab[2]), int(ab[4]),
	}
	for i, val := range g.res.FleetDetail.Values {
		if val.Source == nil || adj == nil {
			continue
		}
		if byteAt(adj, *val.Source) > byteAt(adm, *val.Source) {
			v.ValueInks[i] = adjutantInk
		}
	}

	v.UnitNo = u.Squadron + 1
	v.Ships, v.Full, v.UnitMorale = u.Ships, u.FullShips, u.Morale
	if s := g.res.FleetDetail; s != nil && u.Kind >= 0 && u.Kind < len(s.ShipKinds) {
		v.KindName = s.ShipKinds[u.Kind].Codes
	}
	if k, ok := g.res.ShipKinds.Kind(u.Faction, u.Kind); ok {
		v.TopSpeed, v.BattleSpeed, v.Durability = k.TopSpeed(), k.BattleSpeed(), k.Durability()
		v.Beam = g.weapon(k.BeamWeapon())
		v.Missile = g.weapon(k.MissileWeapon())
		v.Launcher = g.weapon(k.LauncherPer())
		v.Launcher.Name = nil // 那一格印的是固定的「發射裝置」
		v.Fighter = g.weapon(k.FighterWeapon())
		for _, w := range []*screen.FleetDetailWeapon{&v.Beam, &v.Missile, &v.Launcher, &v.Fighter} {
			w.Cur, w.Max = w.Cur*u.Ships, w.Max*u.FullShips
		}
	}
	if g.res.Faces != nil {
		if img, err := g.res.Faces.Decode(f.Admiral); err == nil {
			v.Face = img
		}
	}
	return v, true
}

// adjutantInk 是「這一項副官比提督高」的顏色（`docs/re/112`）。
//
// 「艦隊士氣值」讀的是艦隊記錄 `+0x3C`（`rules.Fleet.Morale`）：建立艦隊
// 時寫 100（`0x2A7F8`：`mov byte es:[bx+3Ch],64h`），戰鬥中會變——收據
// `s3cmdfl3` 的第 13 艦隊是 99（`docs/re/118`）。
const adjutantInk = 5

// weapon 把「名字索引、每艘幾件」變成一組裝備；索引 −1 ＝ 沒搭載。
func (g *Game) weapon(name, per int) screen.FleetDetailWeapon {
	if name < 0 {
		return screen.FleetDetailWeapon{}
	}
	var codes []uint16
	if s := g.res.FleetDetail; s != nil && name < len(s.Weapons) {
		codes = s.Weapons[name].Codes
	}
	return screen.FleetDetailWeapon{Name: codes, Cur: per, Max: per}
}

// fleetSpeeds 回一支艦隊的回轉／巡航／戰鬥速度：**各分隊取最小值**
// （`0x2A7AE`，`docs/re/112`）。
func fleetSpeeds(units []rules.BattleUnit, fleet int) (turn, cruise, fight int) {
	first := true
	for _, u := range units {
		if u.Fleet != fleet {
			continue
		}
		if first {
			turn, cruise, fight = u.TurnSpeed, u.CruiseSpeed, u.FightSpeed
			first = false
			continue
		}
		turn = min(turn, u.TurnSpeed)
		cruise = min(cruise, u.CruiseSpeed)
		fight = min(fight, u.FightSpeed)
	}
	return
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
