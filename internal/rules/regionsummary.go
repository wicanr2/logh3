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

	// Flags 是星域記錄的 `+0x13`：左下角在星域名後面依它印「處於戰鬥狀態」
	// （bit2）／「處於占領狀態」（bit4）／「處於孤立狀態」（bit0）
	// （`docs/re/315`，規格 `docs/spec/001`）。
	Flags int
}

// 星域 `+0x13` 的三個位元（`docs/internals/22`）。
const (
	RegionFlagIsolated = 0x01
	RegionFlagBattle   = 0x04
	RegionFlagOccupied = 0x10
)

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
	if len(record) > RegionUnrest {
		s.Unrest = int(record[RegionUnrest])
	}
	s.OwnFleets, s.EnemyFleets = Fleets(record, faction)
	if len(record) > regionFlagsOff {
		s.Flags = int(record[regionFlagsOff])
	}
	return s
}

// regionFlagsOff 是星域記錄的旗標位移（`+0x13`）。
const regionFlagsOff = 0x13

// CapitalSummary 是右側面板要印的東西（規格 `docs/spec/002`）。
//
// 分兩種來源：**算得出來的**（日期、防衛力、經濟力、成長率、徵稅額、稅率）
// 與**存在遊戲狀態裡的**（軍事費、總稅收）。後兩個是月底結算留下來的全域
// （`docs/internals/27`），畫面只是把它們印出來，不是每一格重算。
type CapitalSummary struct {
	// Region 是首都那一塊摘要，與左下角同一個版面。
	Region RegionSummary

	Year, Month, Day int

	// Military ＝ 軍事費（`ds:8F52`）、TaxTotal ＝ 總稅收（`ds:8F6A`）。
	Military, TaxTotal int

	Defence, Economy, Tax int

	// GrowthPerMille／TaxRatePerMille 以千分之一為單位，畫面印成 `4.5%`。
	GrowthPerMille, TaxRatePerMille int

	// Foreign ＝ 這一格不是玩家陣營的（右側面板改看別的星域時，`073F:02EE`）：
	// 成長率、徵稅額、稅率印 `----`，防衛力與經濟力照印（收據 `fcmd-rinfo-ballat`）。
	Foreign bool
}

// SummariseCapital 把首都那一筆記錄與遊戲狀態整理成右側面板要印的東西。
func SummariseCapital(record []byte, planets *Planets, region, faction, turn, epoch,
	military, taxTotal int) CapitalSummary {
	y, m, d := Date(turn, epoch)
	c := CapitalSummary{
		Region: Summarise(record, planets, region, faction),
		Year:   y, Month: m, Day: d,
		Military: military, TaxTotal: taxTotal,
		Economy:         RegionEconomy(record),
		Tax:             RegionTax(record),
		TaxRatePerMille: RegionTaxRate(record),
	}
	if len(record) > regionGrowth {
		c.GrowthPerMille = int(record[regionGrowth])
	}
	if planets != nil {
		c.Defence = planets.RegionDefence(region)
	}
	c.Foreign = RegionOwnerOf(record) != faction
	return c
}
