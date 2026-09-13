package rules

import "encoding/binary"

// `0374:0556` 在沒有戰鬥或單方占據的星域依序做三件事
// （`docs/re/292`）：無主星域行星脫離、防衛力回復、要塞砲重置。
const (
	planetRecoveryRate = 0x18 // 16 位元百分比
)

// DepartNeutralPlanet 是 `2703:1B48`：星域本身無主且仍有非無主行星時，
// 固定抽一次全域亂數，從那些行星中選一顆套用 RebelPlanet。
// 回被選中的全域行星序號；沒有可處理的行星時回 -1。
func DepartNeutralPlanet(record []byte, planets *Planets, region int, rnd *Rand) int {
	if planets == nil || rnd == nil || RegionOwnerOf(record) != RegionOwnerNone ||
		len(record) <= regionCount+2 || record[regionCount+2] == 0 {
		return -1
	}
	start, n := planets.Range(region)
	eligible := make([]int, 0, n)
	for i := start; i < start+n; i++ {
		if planets.Owner(i) != RegionOwnerNone {
			eligible = append(eligible, i)
		}
	}
	if len(eligible) == 0 {
		return -1
	}
	picked := eligible[rnd.Next()%len(eligible)]
	RebelPlanet(planets.Record(picked))
	return picked
}

// RecoverRegionDefences 是 `2703:1CD7`：每顆未滿上限的行星最多做九次
// `目前值 += floor(上限 × 回復率 / 100)`，每次都夾到上限。
func RecoverRegionDefences(planets *Planets, region int) {
	if planets == nil {
		return
	}
	start, n := planets.Range(region)
	for i := start; i < start+n; i++ {
		r := planets.Record(i)
		limit := int(binary.LittleEndian.Uint16(r[planetGrowthDef:]))
		step := limit * int(binary.LittleEndian.Uint16(r[planetRecoveryRate:])) / 100
		for turn := 0; turn < 9; turn++ {
			defence := int(binary.LittleEndian.Uint16(r[planetDefence:]))
			if defence >= limit {
				break
			}
			defence += step
			if defence > limit {
				defence = limit
			}
			binary.LittleEndian.PutUint16(r[planetDefence:], uint16(defence))
		}
	}
}

// ResetRegionFortressGuns 是 `2703:1D85`：有要塞的行星每回合把砲設成
// 待命並清除瞄準方位；沒有要塞的行星完全不動。
func ResetRegionFortressGuns(planets *Planets, region int) {
	if planets == nil {
		return
	}
	start, n := planets.Range(region)
	for i := start; i < start+n; i++ {
		r := planets.Record(i)
		if r[planetFortress] == 0 {
			continue
		}
		r[planetGunState] = FortressGunIdle
		binary.LittleEndian.PutUint16(r[planetGunAim:], FortressGunNoAim)
	}
}
