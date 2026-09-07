package rules

// RegionSummary 是戰略地圖介面左下角那塊要印的東西（規格 `docs/spec/001`）。
//
// 這裡只算「印什麼」，不管印在哪——版面在
// `docs/internals/data/region-panel.json`，畫的人是 `internal/screen`。
type RegionSummary struct {
	// Owner 是星域的擁有者：0 帝國、1 同盟、2 無主。
	Owner int

	// Name 是星域名，Big5 字碼。
	Name []uint16

	// Own ＝ 這個星域屬於玩家陣營。治安只有自己這一方的才印數字。
	Own bool

	// Unrest 是未安定度，治安與治安色都從它算出來。
	Unrest int

	// Planets 是玩家陣營在這個星域的行星數，Total 是總數（畫面上的 n/m）。
	Planets, Total int

	// OwnFleets／EnemyFleets 是「艦隊 a：b」那兩個數字。
	OwnFleets, EnemyFleets int
}

// Security 回治安百分比。
func (s RegionSummary) Security() int { return Security(s.Unrest) }

// SecurityColour 回治安那個數字的色號。
func (s RegionSummary) SecurityColour() uint8 { return SecurityColour(s.Unrest) }

// RegionName 從星域記錄開頭讀出 Big5 字碼。名稱是 `\0` 結束的全形字串
// （`docs/internals/22`）。
func RegionName(record []byte) []uint16 {
	var out []uint16
	for i := 0; i+1 < len(record); i += 2 {
		if record[i] == 0 {
			break
		}
		out = append(out, uint16(record[i])<<8|uint16(record[i+1]))
	}
	return out
}

// Summarise 把一個星域的執行期記錄整理成面板要印的東西。
//
// 擁有者與行星數不從記錄讀，**照 `PLADAT.DAT` 現算**——原版載入時就是
// 這樣重算的（`docs/re/17`、`docs/re/23`）。未安定度、艦隊數只有記錄裡有。
func Summarise(record []byte, planets *Planets, region, faction int) RegionSummary {
	s := RegionSummary{Name: RegionName(record)}
	if planets != nil {
		s.Owner = planets.RegionOwner(region)
		e, a, o := planets.RegionPlanets(region)
		s.Total = e + a + o
		switch faction {
		case 0:
			s.Planets = e
		case 1:
			s.Planets = a
		}
	}
	s.Own = s.Owner == faction
	if len(record) > regionUnrest {
		s.Unrest = int(record[regionUnrest])
	}
	s.OwnFleets, s.EnemyFleets = Fleets(record, faction)
	return s
}
