package rules

// AdmiralFate 是失蹤提督（記錄 +0x00 == 0x80）在一回合內被判定出的結果。
// 原版入口與狀態寫入見 `056A:0219`、`28E0:057F`（`docs/re/307`）。
type AdmiralFate struct {
	Admiral int
	Faction int
	Dead    bool
}

const (
	admiralStatus = 0x00
	admiralFate   = 0x0B
)

// ResolveMissingAdmirals 依提督表順序判定失蹤者是否生還或死亡，並就地寫回
// 原版的狀態轉移。human 表示該陣營是否由人類控制；它只改死亡門檻，
// 不負責決定畫面是否顯示。
func ResolveMissingAdmirals(data []byte, human [2]bool, rnd *Rand) []AdmiralFate {
	if rnd == nil {
		return nil
	}
	var out []AdmiralFate
	for i := 0; i < len(data)/AdmiralRecord; i++ {
		rec := data[i*AdmiralRecord : (i+1)*AdmiralRecord]
		if rec[admiralStatus] != 0x80 {
			continue
		}
		faction := int(rec[admiralFaction])
		resolved, dead := false, false
		if i == 0 || i == 52 {
			resolved = rnd.Next() < 0x2000
		} else if rnd.Next() < 0x0444 {
			resolved = true
			divisor := int(rec[admiralFate])
			// 原版直接 DIV；正式資料的欄位非零。對破損／人工資料採安全
			// fallback，不把無效記錄變成 Go panic。
			if divisor > 0 {
				limit := 30
				if faction >= 0 && faction < len(human) && human[faction] {
					limit = 45
				}
				dead = rnd.Next()%divisor <= limit
			}
		}
		if !resolved {
			continue
		}
		if dead {
			rec[0x09] |= 0x80
			rec[admiralStatus] = 0
			for o := 0x02; o < 0x06; o++ {
				rec[o] = 0
			}
		} else {
			rec[admiralStatus] = 0xFF
			rec[admiralRegion] = byte(CapitalRegion(faction))
		}
		out = append(out, AdmiralFate{Admiral: i, Faction: faction, Dead: dead})
	}
	return out
}
