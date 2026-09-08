package rules

// 艦隊能力（`docs/internals/33`）。
//
// 艦隊記錄的 `+0x3D`–`+0x47` 是 11 個 byte，載入時由提督與副官逐項算出來，
// **前五項與後六項的合成方式不同**：
//
//	第 0–4 項：提督 ≥ 副官 ? 提督 : (提督 ＋ 副官) ÷ 2   ; 副官只補一半
//	第 5–10 項：max(提督, 副官)                          ; 逐項取大
//
// 抄過來的來源是提督記錄的這 11 個位移，順序固定：

// AbilitySources 是那 11 項在 `ADMIRAL.DT*` 裡的位移，依艦隊記錄的順序排。
var AbilitySources = [FleetAbilityCount]int{15, 16, 17, 19, 18, 12, 13, 20, 22, 23, 14}

const (
	// FleetAbilityCount 是艦隊能力的欄位數。
	FleetAbilityCount = 11
	// FleetAbilityOff 是艦隊記錄裡第一項的位移。
	FleetAbilityOff = 0x3D
	// abilityHalfItems 是「副官只補一半」的那幾項（前五項）。
	abilityHalfItems = 5
)

// FleetAbilities 由提督與副官的記錄算出艦隊能力那 11 個 byte。
//
// 副官給 nil 就只看提督。長度不足的記錄那一項當 0。
func FleetAbilities(admiral, adjutant []byte) [FleetAbilityCount]byte {
	var out [FleetAbilityCount]byte
	for i, off := range AbilitySources {
		a, j := int(byteAt(admiral, off)), int(byteAt(adjutant, off))
		v := a
		if j > a {
			if i < abilityHalfItems {
				v = (a + j) / 2
			} else {
				v = j
			}
		}
		out[i] = byte(v)
	}
	return out
}

func byteAt(rec []byte, off int) byte {
	if off < 0 || off >= len(rec) {
		return 0
	}
	return rec[off]
}
