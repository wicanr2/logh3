package rules

// 費沙事件（原版 `1213:025D`／`1213:03F5`，
// `docs/internals/57-phezzan-event.md`）。
const (
	PhezzanRegion = 14
	PhezzanPlanet = 34
	phezzanFlag   = 0x34 // WORK.DT? +0x34 ＝ ds:8F48
)

var phezzanRebelRegions = [2][]int{
	{10, 11, 12},
	{15, 22, 23, 24, 25, 28, 29, 30},
}

// PhezzanOccurred 回費沙事件是否已經在這一局發生過。
func PhezzanOccurred(work []byte) bool {
	return len(work) > phezzanFlag && work[phezzanFlag] != 0
}

// SetPhezzanOccurred 寫回原版存檔裡的同一格旗標。
func SetPhezzanOccurred(work []byte) {
	if len(work) > phezzanFlag {
		work[phezzanFlag] = 1
	}
}

// PhezzanEventResult 留下本體造成的兩項可報告結果。
type PhezzanEventResult struct {
	Transferred int
	Rebellions  []int
}

// RunPhezzanEvent 跑費沙第一次易主的本體與懲罰。
//
// owner 是拿下費沙的一方；ownerHuman 表示那一方是不是人類座位；turn 是
// 原版當下尚未進位的 `ds:8F18`。garrison 回指定星域、陣營的艦隊數。
// 呼叫端負責事件的一次性旗標與觸發時點。
func RunPhezzanEvent(records []byte, planets *Planets, owner int, ownerHuman bool,
	turn int, garrison func(region, faction int) int) PhezzanEventResult {
	var out PhezzanEventResult
	if planets == nil || owner < 0 || owner > 1 {
		return out
	}

	// 費沙的八成經濟力移到對手首都：帝國拿下就去海尼森，
	// 同盟拿下就去奧丁。
	target := 0
	if owner == 0 {
		target = 80
	}
	if planets.Record(PhezzanPlanet) == nil || planets.Record(target) == nil {
		return out
	}
	amount := planets.PlanetEconomy(PhezzanPlanet) * 80 / 100
	planets.SetPlanetEconomy(PhezzanPlanet,
		planets.PlanetEconomy(PhezzanPlanet)-amount)
	planets.SetPlanetEconomy(target, planets.PlanetEconomy(target)+amount)
	out.Transferred = amount

	// 清單方向與拿下的一方相反；只翻仍屬該方且該方沒有駐軍的整格。
	for _, region := range phezzanRebelRegions[owner] {
		rec := phezzanRecord(records, region)
		if rec == nil || RegionOwnerOf(rec) != owner {
			continue
		}
		if garrison != nil && garrison(region, owner) != 0 {
			continue
		}
		Rebel(rec, planets, region, true)
		out.Rebellions = append(out.Rebellions, region)
	}

	// 電腦拿下費沙不受這一段懲罰。回合 500–599 的 12 是原版越界讀到
	// 下一串資料的第一個 byte；600 起明確歸零。
	if ownerHuman {
		pct := phezzanPenaltyPercent(turn)
		for region := 0; region < RegionCount; region++ {
			if region == PhezzanRegion {
				continue
			}
			rec := phezzanRecord(records, region)
			if rec == nil || RegionOwnerOf(rec) != owner {
				continue
			}
			start, n := planets.Range(region)
			for i := start; i < start+n; i++ {
				e := planets.PlanetEconomy(i)
				planets.SetPlanetEconomy(i, e-e*pct/100)
			}
			rec[RegionMorale] = 100
		}
	}

	// 原版包裝最後呼叫 `090A:0E6D`，先把星域經濟力重算成行星加總。
	for region := 0; region < RegionCount; region++ {
		rec := phezzanRecord(records, region)
		if rec == nil {
			continue
		}
		start, n := planets.Range(region)
		total := 0
		for i := start; i < start+n; i++ {
			total += planets.PlanetEconomy(i)
		}
		SetRegionEconomy(rec, total)
	}
	return out
}

func phezzanRecord(records []byte, region int) []byte {
	o := region * RegionRecord
	if region < 0 || o+RegionRecord > len(records) {
		return nil
	}
	return records[o : o+RegionRecord]
}

func phezzanPenaltyPercent(turn int) int {
	if turn < 0 || turn >= 600 {
		return 0
	}
	return [...]int{50, 40, 30, 20, 10, 12}[turn/100]
}
