package rules

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRegionFleetsMatchesFleetTable 釘住 `docs/re/199`：星域記錄的
// `+0x28`／`+0x29` 就是帝國／同盟在那一格的艦隊數。
//
// 三個劇本各自把 `FLEET.DT*` 數一遍，32 格逐一比 `STELLAR.DT*`。
// 缺原版檔就 skip——不用代用品。
func TestRegionFleetsMatchesFleetTable(t *testing.T) {
	for _, ext := range []string{"DT8", "DT9", "DTA"} {
		t.Run(ext, func(t *testing.T) {
			dir := filepath.Join("../..", "org_game/extracted")
			fleets, err := os.ReadFile(filepath.Join(dir, "FLEET."+ext))
			if err != nil {
				t.Skipf("缺 FLEET.%s，跳過（不用代用品）", ext)
			}
			stellar, err := os.ReadFile(filepath.Join(dir, "STELLAR."+ext))
			if err != nil {
				t.Skipf("缺 STELLAR.%s，跳過（不用代用品）", ext)
			}
			var want [RegionCount][2]int
			for i := 0; i+FleetRecord <= len(fleets); i += FleetRecord {
				r := fleets[i : i+FleetRecord]
				if r[fleetUnused] == 0xFF {
					continue
				}
				f, reg := int(r[fleetFaction]), int(r[fleetRegion])
				if f < 2 && reg < RegionCount {
					want[reg][f]++
				}
			}
			for k := 0; k < RegionCount; k++ {
				o := k * RegionRecord
				if o+RegionRecord > len(stellar) {
					t.Fatalf("STELLAR.%s 只有 %d bytes，放不下 32 筆", ext, len(stellar))
				}
				rec := stellar[o : o+RegionRecord]
				for f := 0; f < 2; f++ {
					if got := RegionFleets(rec, f); got != want[k][f] {
						t.Errorf("星域 %d 陣營 %d：記錄寫 %d，艦隊表數出來 %d",
							k, f, got, want[k][f])
					}
				}
				// `+0x2A`／`+0x2B` 在三個劇本裡都是 0（`docs/re/199`）。
				if rec[0x2A] != 0 || rec[0x2B] != 0 {
					t.Errorf("星域 %d 的 +0x2A/+0x2B ＝ %d/%d，預期 0/0",
						k, rec[0x2A], rec[0x2B])
				}
			}
		})
	}
}
