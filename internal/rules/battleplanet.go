package rules

import "encoding/binary"

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

	// GunState 是要塞砲的狀態（記錄 `+0x45`）：只有 0（待命）或
	// 3（已瞄準）能重新瞄準。GunAim 是**瞄準的方位**（記錄 `+0x50`，
	// 一整圈 256，`0xFFFF` ＝ 沒瞄準），不是某一支分隊——玩家路徑
	// 由游標的方位算（`2264:0133`），電腦路徑用八格一步掃一圈挑
	// （`303F:00B0`），兩邊寫進去的都是角度（`docs/re/145`）。
	GunState int `json:"gun_state"`
	GunAim   int `json:"gun_aim"`

	// 情報視窗要的幾格（`docs/re/144`）：名字（記錄開頭那串 Big5）、
	// `+0x16` 防衛力上限、`+0x18` 回復率、`+0x1A` 經濟力（32 位元）。
	NameBytes  []byte `json:"name_bytes"`
	DefenceMax int    `json:"defence_max"`
	Recover    int    `json:"recover"`
	Economy    int    `json:"economy"`
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
			GunAim:    int(le32(p.rec(i)[planetGunAim:]) & 0xFFFF),
			Stock:     planetStockOf(p.rec(i)),

			NameBytes:  planetNameOf(p.rec(i)),
			DefenceMax: int(binary.LittleEndian.Uint16(p.rec(i)[0x16:])),
			Recover:    int(binary.LittleEndian.Uint16(p.rec(i)[0x18:])),
			Economy:    int(le32(p.rec(i)[0x1A:])),
		})
	}
	return out
}

// planetNameOf 回一筆行星記錄開頭那串名字（Big5，`0x00` 結尾）。
func planetNameOf(rec []byte) []byte {
	for i, b := range rec {
		if b == 0 {
			out := make([]byte, i)
			copy(out, rec[:i])
			return out
		}
	}
	return nil
}

// planetStockOf 讀一筆行星記錄的存量；讀不出來就回空的。
func planetStockOf(rec []byte) PlanetStock {
	s, _ := ReadPlanetStock(rec)
	return s
}
