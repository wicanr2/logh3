package rules

import (
	"os"
	"path/filepath"
	"testing"
)

const origDir = "../../org_game/extracted"

func readOrig(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(origDir, name))
	if err != nil {
		t.Skipf("缺 %s，跳過（不用代用品）", name)
	}
	return b
}

// TestParseSaveIndex 對 `SAVEDATA.IDX`：11 格，前八格空的，後三格是那三個進度。
func TestParseSaveIndex(t *testing.T) {
	got, err := ParseSaveIndex(readOrig(t, "SAVEDATA.IDX"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != SaveSlots {
		t.Fatalf("格數 %d，要 %d", len(got), SaveSlots)
	}
	for i := 0; i < 8; i++ {
		if got[i].Used {
			t.Errorf("第 %d 格應該是空的", i)
		}
	}
	for _, c := range []struct {
		index, num, turn int
		name             string
		ext              string
	}{
		{8, 1, 93, "亞姆力札會戰", "8"},
		{9, 2, 346, "諸神的黃昏", "9"},
		{10, 3, 380, "巴米利恩會戰", "A"},
	} {
		e := got[c.index]
		if !e.Used || e.Num != c.num || e.Turn != c.turn {
			t.Errorf("第 %d 格：used=%v num=%d turn=%d", c.index, e.Used, e.Num, e.Turn)
		}
		if n := len(e.Name); n != len([]rune(c.name)) {
			t.Errorf("第 %d 格名稱 %d 個字，要 %d", c.index, n, len([]rune(c.name)))
		}
		if SlotExt(c.index) != c.ext {
			t.Errorf("第 %d 格副檔名末字 %q，要 %q", c.index, SlotExt(c.index), c.ext)
		}
	}
}

// loadSlot 讀一格存檔的六個檔。
func loadSlot(t *testing.T, ext string) *Slot {
	t.Helper()
	s, err := LoadSlot(SlotFiles{
		Stellar:      readOrig(t, "STELLAR.DT"+ext),
		Planet:       readOrig(t, "PLANET.DT"+ext),
		Admiral:      readOrig(t, "ADMIRAL.DT"+ext),
		Fleet:        readOrig(t, "FLEET.DT"+ext),
		Ship:         readOrig(t, "SHIP.DT"+ext),
		Work:         readOrig(t, "WORK.DT"+ext),
		PlanetHeader: readOrig(t, "PLADAT.DAT"),
	})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// TestLoadSlot 對兩格中局：回合數與兩筆錢與畫面上印的相同（`docs/spec/002`），
// 行星擁有者是那個局面的、不是開局的。
func TestLoadSlot(t *testing.T) {
	for _, c := range []struct {
		ext                    string
		turn, military, tax    int
		region, owner, planets int
		fleets                 int
	}{
		// S2「諸神的黃昏」798/11/19；S3「巴米利恩會戰」799/3/1。
		{"9", 346, 150000, 7280, 0, 0, 4, 18},
		// 干達巴爾開局是三顆中立行星，S3 已經全是帝國的。
		{"A", 380, 160000, 13131, 25, 0, 3, 18},
	} {
		t.Run(c.ext, func(t *testing.T) {
			s := loadSlot(t, c.ext)
			if s.Turn != c.turn || s.Military != c.military || s.TaxTotal != c.tax {
				t.Errorf("回合 %d 軍事費 %d 總稅收 %d，要 %d／%d／%d",
					s.Turn, s.Military, s.TaxTotal, c.turn, c.military, c.tax)
			}
			if o := s.Planets.RegionOwner(c.region); o != c.owner {
				t.Errorf("星域 %d 擁有者 %d，要 %d", c.region, o, c.owner)
			}
			e, a, _ := s.Planets.RegionPlanets(c.region)
			if n := e + a; n != c.planets {
				t.Errorf("星域 %d 有 %d 顆行星，要 %d", c.region, n, c.planets)
			}
			n := 0
			for i := 0; i < FleetCount; i++ {
				if s.FleetUsed(i) {
					n++
				}
			}
			if n != c.fleets {
				t.Errorf("在用的艦隊 %d 支，要 %d", n, c.fleets)
			}
		})
	}
}

// TestSlotShipPool 對艦艇池的帳（`docs/re/42`）：十個槽引用的，加上每支艦隊
// 多出來的那一筆（該隊最小索引 ＋1），剛好就是池裡有內容的那些記錄。
//
// 第 8 格（亞姆力札會戰）不在裡面——那一格有四個索引被重複引用，
// 還多一筆對不到，原因未解。
func TestSlotShipPool(t *testing.T) {
	for _, c := range []struct {
		ext   string
		total int
	}{{"9", 160}, {"A", 148}} {
		t.Run(c.ext, func(t *testing.T) {
			s := loadSlot(t, c.ext)
			seen := map[int]bool{}
			dup := 0
			for i := 0; i < FleetCount; i++ {
				if !s.FleetUsed(i) {
					continue
				}
				ships := s.FleetShips(i)
				if len(ships) == 0 {
					t.Fatalf("艦隊 %d 在用卻沒有艦艇", i)
				}
				lo := ships[0]
				for _, k := range ships {
					if seen[k] {
						dup++
					}
					seen[k] = true
					if k < lo {
						lo = k
					}
				}
				seen[lo+1] = true
			}
			if dup != 0 {
				t.Errorf("有 %d 個艦艇索引被重複引用", dup)
			}
			used := 0
			for i := 0; i < ShipCount; i++ {
				if shipFilled(s.ShipAt(i)) {
					used++
					if !seen[i] {
						t.Errorf("艦艇 %d 有內容卻沒有艦隊指到它", i)
					}
				} else if seen[i] {
					t.Errorf("艦艇 %d 被指到卻是空的", i)
				}
			}
			if used != c.total {
				t.Errorf("有內容的艦艇 %d 筆，要 %d", used, c.total)
			}
		})
	}
}

// shipFilled ＝ 這一筆艦艇記錄不是整片 FF。`+0x38` 之後三格存檔完全一樣，
// 所以只看前面那段會變的（`docs/re/42`）。
func shipFilled(r []byte) bool {
	n := 0
	for _, b := range r[:0x38] {
		if b != 0xFF {
			n++
		}
	}
	return n >= 10
}
