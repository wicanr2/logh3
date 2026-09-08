package rules

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFleetAbilitiesMatchOriginal 拿原版存檔對：艦隊記錄 `+0x3D`–`+0x47`
// 應該等於「提督與副官逐項取大」（`docs/internals/33`）。
func TestFleetAbilitiesMatchOriginal(t *testing.T) {
	orig := filepath.Join("../..", "org_game/extracted")
	total := 0
	for _, ext := range []string{"DT8", "DT9", "DTA"} {
		total += checkAbilities(t, orig, ext)
	}
	if total == 0 {
		t.Skip("三格存檔都不在，跳過（不用代用品）")
	}
	t.Logf("三格存檔合計對過 %d 支艦隊 × %d 項", total, FleetAbilityCount)
}

// checkAbilities 對一格存檔，回對過幾支艦隊。
func checkAbilities(t *testing.T, orig, ext string) int {
	t.Helper()
	fleet, err := os.ReadFile(filepath.Join(orig, "FLEET."+ext))
	if err != nil {
		return 0
	}
	adm, err := os.ReadFile(filepath.Join(orig, "ADMIRAL."+ext))
	if err != nil {
		return 0
	}
	rec := func(i int) []byte {
		if i < 0 || (i+1)*AdmiralRecord > len(adm) {
			return nil
		}
		return adm[i*AdmiralRecord : (i+1)*AdmiralRecord]
	}
	checked := 0
	for i := 0; (i+1)*FleetRecord <= len(fleet); i++ {
		f := fleet[i*FleetRecord : (i+1)*FleetRecord]
		if f[4] == 0xFF {
			continue // 沒在用的艦隊
		}
		adj := int(f[5])
		var adjRec []byte
		if adj != 0xFF {
			adjRec = rec(adj)
		}
		got := FleetAbilities(rec(int(f[4])), adjRec)
		for k := 0; k < FleetAbilityCount; k++ {
			if got[k] != f[FleetAbilityOff+k] {
				t.Errorf("%s 艦隊 %d 第 %d 項：算出 %d，存檔是 %d",
					ext, i, k, got[k], f[FleetAbilityOff+k])
			}
		}
		checked++
	}
	return checked
}
