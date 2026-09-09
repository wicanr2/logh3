package rules

import (
	"encoding/json"
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
		turn                   int
		military, tax, blend   [2]int
		region, owner, planets int
		fleets                 int
	}{
		// S2「諸神的黃昏」798/11/19；S3「巴米利恩會戰」799/3/1。
		// 帝國那一格是畫面上印的（`docs/spec/002`）；同盟那一格與 `Blend`
		// 是工作區塊裡的（`docs/internals/30`、`docs/internals/55`），
		// `Blend` 與 `docs/spec/006` 量到的執行期值相同。
		{"9", 346, [2]int{150000, 30000}, [2]int{7280, 10321}, [2]int{1, 2}, 0, 0, 4, 18},
		// 干達巴爾開局是三顆中立行星，S3 已經全是帝國的。
		{"A", 380, [2]int{160000, 30000}, [2]int{13131, 9748}, [2]int{1, 2}, 25, 0, 3, 18},
	} {
		t.Run(c.ext, func(t *testing.T) {
			s := loadSlot(t, c.ext)
			if s.Turn != c.turn || s.Military != c.military ||
				s.TaxTotal != c.tax || s.Blend != c.blend {
				t.Errorf("回合 %d 軍事費 %v 總稅收 %v Blend %v，要 %d／%v／%v／%v",
					s.Turn, s.Military, s.TaxTotal, s.Blend,
					c.turn, c.military, c.tax, c.blend)
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

// TestSlotSaveRoundTrip 是存檔寫出的驗收（`docs/spec/009`）：
// 讀一格再寫出去，六個檔要與原版逐位元組相同。
//
// 唯一的例外是 `STELLAR` 的 `+0x7D`：載入時原版會重算未安定度
// （`docs/spec/006`），寫回去的自然是重算後的值。這裡讀完不重算，
// 所以連那一欄也應該相同——真的存檔時會不一樣，那是原版的行為。
func TestSlotSaveRoundTrip(t *testing.T) {
	for _, ext := range []string{"9", "A"} {
		t.Run(ext, func(t *testing.T) {
			s := loadSlot(t, ext)
			dir := t.TempDir()
			index := 9
			if ext == "A" {
				index = 10
			}
			if err := s.Save(dir, index); err != nil {
				t.Fatal(err)
			}
			for _, name := range SlotFileNames(index) {
				got, err := os.ReadFile(filepath.Join(dir, name))
				if err != nil {
					t.Fatal(err)
				}
				want := readOrig(t, name)
				if len(got) != len(want) {
					t.Errorf("%s：寫出 %d bytes，原版 %d", name, len(got), len(want))
					continue
				}
				bad := 0
				for i := range want {
					if got[i] != want[i] {
						bad++
					}
				}
				if bad != 0 {
					t.Errorf("%s：與原版不同的位元組 %d / %d", name, bad, len(want))
				}
			}
		})
	}
}

// TestWriteSaveIndex 對存檔目錄的寫回：只動回合數與名稱，其餘照舊。
func TestWriteSaveIndex(t *testing.T) {
	orig := readOrig(t, "SAVEDATA.IDX")
	idx := append([]byte(nil), orig...)
	before, err := ParseSaveIndex(idx)
	if err != nil {
		t.Fatal(err)
	}
	// 把第 10 格改成第 9 格的回合數與名稱，再讀回來對。
	if err := WriteSaveIndex(idx, 10, before[9].Turn, before[9].Name); err != nil {
		t.Fatal(err)
	}
	after, err := ParseSaveIndex(idx)
	if err != nil {
		t.Fatal(err)
	}
	if after[10].Turn != before[9].Turn || len(after[10].Name) != len(before[9].Name) {
		t.Errorf("寫回去之後第 10 格是 %d／%d 個字，要 %d／%d 個字",
			after[10].Turn, len(after[10].Name), before[9].Turn, len(before[9].Name))
	}
	if after[10].Num != before[10].Num || after[10].Stamp != before[10].Stamp {
		t.Errorf("序號或時間戳被動到了")
	}
	// 寫成原本的值應該一個 byte 都不差。
	idx2 := append([]byte(nil), orig...)
	if err := WriteSaveIndex(idx2, 10, before[10].Turn, before[10].Name); err != nil {
		t.Fatal(err)
	}
	bad := 0
	for i := range orig {
		if orig[i] != idx2[i] {
			bad++
		}
	}
	if bad != 0 {
		t.Errorf("寫回原值之後與原版不同的位元組 %d", bad)
	}
}

// TestShipTypeUnitSize 把艦艇記錄的 `+0x08` 對上建造計劃的「一單位幾艘」
// （`docs/re/42`）：艦艇的艦種就是建造計劃的列。
func TestShipTypeUnitSize(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("../..",
		"docs/internals/data/ship-types.json"))
	if err != nil {
		t.Fatal(err)
	}
	var d struct {
		Types []struct {
			Type int `json:"type"`
			At08 int `json:"at_08"`
		} `json:"types"`
	}
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	seen := 0
	for _, ty := range d.Types {
		if ty.Type < 0 || ty.Type >= BuildPlanRows {
			t.Errorf("艦種 %d 超出建造計劃的 %d 列", ty.Type, BuildPlanRows)
			continue
		}
		if ty.At08 != ShipsPerUnit[ty.Type] {
			t.Errorf("艦種 %d 的 +0x08 是 %d，建造計劃的一單位是 %d 艘",
				ty.Type, ty.At08, ShipsPerUnit[ty.Type])
		}
		seen++
	}
	if seen < 7 {
		t.Errorf("只對到 %d 組，太少", seen)
	}
}

// TestSlotFleetsMatchDump 存檔讀出來的艦隊表要與**執行期倒出來的那一份**
// 相同（`docs/re/77` 的收據 `s2ai0`／`s3ai0`）。
//
// 兩份的來源不同：這裡讀的是 `FLEET.DT?`，收據量的是遊戲載入之後
// `ds:C074` 的內容。對得起來才表示欄位讀對了。
func TestSlotFleetsMatchDump(t *testing.T) {
	cases := []struct {
		ext  string
		want []Fleet
	}{
		// S2「諸神的黃昏」：第 9 格。
		{"9", []Fleet{
			{Number: 1, Faction: 0, Region: 2, Target: 14},
			{Number: 3, Faction: 0, Region: 7, Target: 14},
			{Number: 4, Faction: 0, Region: 8, Target: 13},
			{Number: 5, Faction: 0, Region: 0, Target: NoTarget, Parked: true},
			{Number: 6, Faction: 0, Region: 7, Target: 14},
			{Number: 8, Faction: 0, Region: 8, Target: 13},
			{Number: 9, Faction: 0, Region: 2, Target: 0, Parked: true},
			{Number: 12, Faction: 0, Region: 0, Target: NoTarget, Parked: true},
			{Number: 13, Faction: 0, Region: 2, Target: 14},
			{Number: 15, Faction: 0, Region: 8, Target: 13},
			{Number: 16, Faction: 0, Region: 2, Target: 14},
			{Number: 1, Faction: 1, Region: 13, Target: 0, Parked: true},
			{Number: 2, Faction: 1, Region: 31, Target: 4, Parked: true},
			{Number: 5, Faction: 1, Region: 31, Target: 4, Parked: true},
			{Number: 13, Faction: 1, Region: 13, Target: 0, Parked: true},
			{Number: 14, Faction: 1, Region: 31, Target: 4, Parked: true},
			{Number: 15, Faction: 1, Region: 31, Target: 4, Parked: true},
			{Number: 16, Faction: 1, Region: 13, Target: 0, Parked: true},
		}},
		// S3「巴米利恩會戰」：第 10 格。同盟六支全部在萊加爾、
		// 目的地是達希利，而且**不是 Parked**——所以它們會動。
		{"A", []Fleet{
			{Number: 1, Faction: 0, Region: 25, Target: 0, Parked: true},
			{Number: 3, Faction: 0, Region: 25, Target: 0, Parked: true},
			{Number: 4, Faction: 0, Region: 25, Target: 0, Parked: true},
			{Number: 5, Faction: 0, Region: 25, Target: 0, Parked: true},
			{Number: 6, Faction: 0, Region: 25, Target: 0, Parked: true},
			{Number: 8, Faction: 0, Region: 13, Target: 0, Parked: true},
			{Number: 9, Faction: 0, Region: 23, Target: 25},
			{Number: 10, Faction: 0, Region: 14, Target: 0, Parked: true},
			{Number: 12, Faction: 0, Region: 25, Target: 0, Parked: true},
			{Number: 13, Faction: 0, Region: 28, Target: 26},
			{Number: 15, Faction: 0, Region: 23, Target: 25},
			{Number: 16, Faction: 0, Region: 25, Target: 0, Parked: true},
			{Number: 1, Faction: 1, Region: 27, Target: 26},
			{Number: 2, Faction: 1, Region: 27, Target: 26},
			{Number: 13, Faction: 1, Region: 27, Target: 26},
			{Number: 14, Faction: 1, Region: 27, Target: 26},
			{Number: 15, Faction: 1, Region: 27, Target: 26},
			{Number: 16, Faction: 1, Region: 27, Target: 26},
		}},
	}
	for _, c := range cases {
		got := loadSlot(t, c.ext).Fleets(nil)
		if len(got) != len(c.want) {
			t.Fatalf("DT%s 讀出 %d 支，預期 %d 支", c.ext, len(got), len(c.want))
		}
		for i, w := range c.want {
			g := got[i]
			if g.Number != w.Number || g.Faction != w.Faction ||
				g.Region != w.Region || g.Target != w.Target || g.Parked != w.Parked {
				t.Errorf("DT%s 第 %d 支 = %+v，預期編號%d 陣營%d 星域%d 目的%d 停%v",
					c.ext, i, g, w.Number, w.Faction, w.Region, w.Target, w.Parked)
			}
		}
	}
}

// TestFleetStrengthUsesOriginalFormula 釘住原版的戰力算式
// （`2E54:1876`，`docs/re/79`）：權重[艦種] × 現有艘數 × 1000 ÷ 滿編艘數。
//
// S3 載入時每一筆都是滿編，所以每一艘的貢獻剛好是 `權重 × 1000`；
// 一支艦隊的戰力就是它十個槽的權重和 × 1000。
func TestFleetStrengthUsesOriginalFormula(t *testing.T) {
	w, err := LoadStrengthWeights("../../docs/internals/data/ship-strength-weights.json")
	if err != nil {
		t.Skipf("讀不到權重表：%v", err)
	}
	if len(w) != 8 {
		t.Fatalf("權重表有 %d 個，預期 8 個（艦種 0–7）", len(w))
	}
	s := loadSlot(t, "A")

	for i := 0; i < FleetCount; i++ {
		if !s.FleetUsed(i) {
			continue
		}
		want := 0
		for _, n := range s.FleetShips(i) {
			r := s.Ship[n*ShipRecord : (n+1)*ShipRecord]
			kind := int(r[shipKind])
			alive := int(r[shipAlive]) | int(r[shipAlive+1])<<8
			full := int(r[shipFull]) | int(r[shipFull+1])<<8
			if kind == shipKindEmpty || full == 0 {
				continue
			}
			if alive != full {
				t.Fatalf("第 %d 支的第 %d 筆不是滿編（%d/%d），這個測試的前提不成立",
					i, n, alive, full)
			}
			want += w[kind] * 1000
		}
		if got := s.FleetStrength(w, i); got != want {
			t.Fatalf("第 %d 支戰力 %d，照算式應該是 %d", i, got, want)
		}
		if want == 0 {
			t.Fatalf("第 %d 支算出 0 戰力——欄位大概讀錯了", i)
		}
	}
}

// TestFleetGroupsMatchesSlots 驗艦隊的 `Groups`（執行期記錄的 `+0x0F`，
// 戰略 AI 的敵方兵力圖靠它加總，`docs/re/173`）與編制槽掃到 `0xFF`
// 為止的筆數一致，而且每一支都至少有一組。
func TestFleetGroupsMatchesSlots(t *testing.T) {
	s := loadSlot(t, "8")
	fleets := s.Fleets(nil)
	if len(fleets) == 0 {
		t.Fatal("這一格沒有在用的艦隊")
	}
	for _, f := range fleets {
		// 現在編制裡的那幾組不會多於 `+0x0F`。
		if got := len(s.FleetGroupShips(f.Index)); got > f.Groups {
			t.Errorf("艦隊 %d 的編制取出 %d 組，超過 +0x0F ＝ %d", f.Index, got, f.Groups)
		}
		if f.Groups == 0 {
			t.Errorf("艦隊 %d 一組編制都沒有", f.Index)
		}
		// `+0x0E` 是應有數，不會小於現有數（`docs/re/194`）。
		if f.Expected < f.Groups {
			t.Errorf("艦隊 %d 的 +0x0E ＝ %d 小於 +0x0F ＝ %d", f.Index, f.Expected, f.Groups)
		}
		if f.ClassKey < 0 || f.ClassKey > 15 {
			t.Errorf("艦隊 %d 的類別鍵 ＝ %d，超出 0–15", f.Index, f.ClassKey)
		}
	}
}

// TestWorkShipPctAndBuildPct 釘住 `WORK.DT?` 新對上的兩個欄位
// （`docs/re/256`、`docs/re/257`）。`WORK.DTA` 是 S3「巴米利恩會戰」，
// 記憶體那一側的收據 `monthmoney` 量到 `ds:8F92` ＝ `00 14`、
// 軍事費 `160000 / 30000`、總稅收 `13131 / 9748`、回合 380——
// 存檔裡的同一組數字必須相同（`WORK.DT?` 就是 `ds:8F14` 起那 230 bytes，
// `docs/re/45`）。
func TestWorkShipPctAndBuildPct(t *testing.T) {
	work := readOrig(t, "WORK.DTA")
	if len(work) != WorkRecord {
		t.Fatalf("WORK.DTA 有 %d bytes，要 %d", len(work), WorkRecord)
	}
	if got, want := [2]int{int(work[workShipPct]), int(work[workShipPct+1])},
		[2]int{0, 20}; got != want {
		t.Errorf("ds:8F92 ＝ %v，收據是 %v", got, want)
	}
	for faction := 0; faction < 2; faction++ {
		sum := 0
		for row := 0; row < BuildPlanRows; row++ {
			sum += int(work[workBuildPct+faction*BuildPlanRows+row])
		}
		if sum != 100 {
			t.Errorf("陣營 %d 的 ds:8FD6 加起來 ＝ %d，預期 100", faction, sum)
		}
	}
	// 同一份檔案裡的三個舊欄位——基底位址錯了這三個會一起錯。
	if got := int(work[workTurn]) | int(work[workTurn+1])<<8; got != 380 {
		t.Errorf("回合 ＝ %d，預期 380", got)
	}
	if got, want := [2]int{u32(work, workMilitary), u32(work, workMilitary+4)},
		[2]int{160000, 30000}; got != want {
		t.Errorf("軍事費 ＝ %v，收據是 %v", got, want)
	}
	if got, want := [2]int{u32(work, workTaxTotal), u32(work, workTaxTotal+4)},
		[2]int{13131, 9748}; got != want {
		t.Errorf("總稅收 ＝ %v，收據是 %v", got, want)
	}
}
