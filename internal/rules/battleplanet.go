package rules

// 戰場上的行星（`docs/re/99`）。
//
// 一輪的總損失是**以行星為中心**算的，不是兩支艦隊互打：行星記錄的
// `+0x11` 是擁有者、`+0x46`／`+0x48` 是它在星域地圖上的位置、
// `+0x14` 是防衛力（16 位元，2200–5400）、`+0x12` 決定除數。

// BattlePlanet 是戰鬥期間一顆行星要用到的欄位。
type BattlePlanet struct {
	Index   int `json:"index"` // 行星表的索引
	Owner   int `json:"owner"` // 0 帝國、1 同盟、2 無主
	X       int `json:"x"`     // 星域地圖上的位置
	Y       int `json:"y"`
	Defence int `json:"defence"` // `+0x14`：防衛力
	// Fortress ＝ 行星記錄 `+0x12` 非 0：**這顆有要塞**。81 顆裡有 8 顆
	// （伊謝爾倫、連典貝爾克、禿鷹之城、格魯米遜、葛羅斯達、帕謬爾、
	// 梅爾格利德、魯德米拉）。兩個地方讀它：戰損的除數從 10 變 20，
	// 以及戰場點行星時開的是四項的選單（多「要塞砲」，`docs/re/133`）。
	Fortress bool `json:"fortress"`

	// Stock 是這顆行星身上的補給／補充存量（`docs/re/134`）。
	Stock PlanetStock `json:"stock"`

	// GunState／GunTarget 是要塞砲的狀態與鎖定的目標（記錄 `+0x45`／
	// `+0x50`）。只有狀態 0（沒鎖定）或 3（已鎖定）能重新指定。
	GunState  int `json:"gun_state"`
	GunTarget int `json:"gun_target"`
}

// 查詢半徑的夾限（`2AF5:01F7`）。
const (
	planetRangeLo  = 1500
	planetRangeHi  = 6000
	planetRangeDiv = 25
)

// PlanetQueryRadius 回「數哪些分隊」的半徑：防衛力夾在 1500–6000 再除 25。
func PlanetQueryRadius(defence int) int {
	d := defence
	if d < planetRangeLo {
		d = planetRangeLo
	}
	if d > planetRangeHi {
		d = planetRangeHi
	}
	return d / planetRangeDiv
}

// Fortress 回這顆行星有沒有要塞（記錄 `+0x12` 非 0）。
func (p *Planets) Fortress(i int) bool { return p.rec(i)[planetFortress] != 0 }

const planetFortress = 0x12

// BattlePlanetsFor 把一個星域的行星配上它們在戰場上的位置。
//
// 位置不在檔案裡，是載入星域地圖時算出來的：地圖上圖塊 `PlanetTile` 的
// 格子**依列序**對到該星域的行星（`docs/re/89`）。所以 `sites` 要是
// `PlanetSites` 掃同一張地圖的結果，兩邊按順序配。
func (p *Planets) BattlePlanetsFor(region int, sites []PlanetSite) []BattlePlanet {
	start, n := p.Range(region)
	out := make([]BattlePlanet, 0, n)
	for k := 0; k < n && k < len(sites); k++ {
		i := start + k
		out = append(out, BattlePlanet{
			Index: i, Owner: p.Owner(i),
			X: sites[k].X, Y: sites[k].Y,
			Defence: p.Defence(i), Fortress: p.Fortress(i),
			GunState:  int(p.rec(i)[planetGunState]),
			GunTarget: int(le32(p.rec(i)[planetGunTarget:]) & 0xFFFF),
			Stock:     planetStockOf(p.rec(i)),
		})
	}
	return out
}

// planetStockOf 讀一筆行星記錄的存量；讀不出來就回空的。
func planetStockOf(rec []byte) PlanetStock {
	s, _ := ReadPlanetStock(rec)
	return s
}
