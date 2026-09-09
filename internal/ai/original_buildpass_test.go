package ai

import (
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

func adm(i, level, command int) AdmiralView {
	return AdmiralView{Index: i, Level: level, Command: command,
		Ability: [4]int{100, 100, 100, 100}}
}

func buildBase() BuildInput {
	return BuildInput{
		Faction: 0, CapitalOwner: 0,
		Tax: 100000, Upkeep: 0,
		Candidates: []AdmiralView{adm(1, 20, 100), adm(2, 10, 100), adm(3, 5, 100)},
	}
}

// TestBuildPassCapitalGate 首都不是自己的就整趟不做（`2D76:0002`）。
func TestBuildPassCapitalGate(t *testing.T) {
	in := buildBase()
	in.CapitalOwner = 1
	if got := BuildPass(in); got != nil {
		t.Errorf("首都不是自己的還建了 %v", got)
	}
}

// TestBuildPassMoneyGate 維持費 × 20 追上稅收就不建（`2D76:0060`）。
// **要三支以上才看錢**，所以先給三支。
func TestBuildPassMoneyGate(t *testing.T) {
	in := buildBase()
	for i := 0; i < 3; i++ {
		in.Fleets = append(in.Fleets, rules.Fleet{Faction: 0})
	}
	in.Tax, in.Upkeep = 10000, 500 // 500 × 20 ＝ 10000，追上了
	if got := BuildPass(in); got != nil {
		t.Errorf("錢不夠還建了 %v", got)
	}
	// 錢夠的話 `2D76:0002` 會一直繞——候選只有三位，繞到沒人可用為止。
	in.Upkeep = 499
	if got := BuildPass(in); len(got) != 3 {
		t.Errorf("錢夠卻建出 %d 支，預期 3（候選用完）", len(got))
	}
}

// TestBuildPassLoopsUntilBlocked `2D76:0002` 是迴圈——建到擋住為止
// （`docs/re/210`）。這裡的候選只有三位，所以停在三支，而且**不會有人
// 帶兩支**（原版靠清掉 `+0x09` bit7 達成）。
func TestBuildPassLoopsUntilBlocked(t *testing.T) {
	got := BuildPass(buildBase())
	if len(got) != 3 {
		t.Fatalf("建出 %d 支，預期 3（候選三位用完）", len(got))
	}
	seen := map[int]bool{}
	for _, nf := range got {
		if seen[nf.Admiral] {
			t.Fatalf("提督 %d 帶了兩支", nf.Admiral)
		}
		seen[nf.Admiral] = true
	}
	// **一支都沒有時四類的「現有 − 應有」全是 0**（總數 0 → 應有全 0），
	// 平手就照 0 1 2 3 的順序，所以試的是類別 0，鍵取自它那張表。
	if rules.FleetClass(got[0].ClassKey) != 0 {
		t.Errorf("類別鍵 ＝ %X（類別 %d），預期類別 0",
			got[0].ClassKey, rules.FleetClass(got[0].ClassKey))
	}
	// 亂數是 nil → 每一輪都挑當時分數最高的（等級 20 那位先）。
	if got[0].Admiral != 1 {
		t.Errorf("第一支挑到提督 %d，預期 1（分數最高）", got[0].Admiral)
	}
	if got[0].Region != Capitals[0] {
		t.Errorf("編在星域 %d，預期首都 %d", got[0].Region, Capitals[0])
	}
}

// TestBuildPassFallsThroughToNextClass 這一類編不成就換下一類
// ——而且亂數已經照原版的次序用掉了。
func TestBuildPassFallsThroughToNextClass(t *testing.T) {
	in := buildBase()
	in.CanForm = func(AdmiralView) bool { return false }
	if got := BuildPass(in); got != nil {
		t.Errorf("四類都編不成卻建了 %v", got)
	}
	// 四類各用掉兩個亂數 ＝ 八次；四類都失敗，外圈就結束，不會再繞。
	r := rules.NewRand(1)
	in.Rand = r
	BuildPass(in)
	want := rules.NewRand(1)
	for i := 0; i < 8; i++ {
		want.Next()
	}
	if r.Seed() != want.Seed() {
		t.Errorf("亂數用掉的次數不是 8（種子 %08X，預期 %08X）", r.Seed(), want.Seed())
	}
}

// TestBuildPassAdjutant 預設副官照抄。
func TestBuildPassAdjutant(t *testing.T) {
	in := buildBase()
	in.Adjutant = func(i int) int {
		if i == 1 {
			return 75
		}
		return -1
	}
	got := BuildPass(in)
	if len(got) == 0 || got[0].Adjutant != 75 {
		t.Errorf("第一支的副官 ＝ %v，預期 75", got)
	}
}

// TestOriginalStrategyBuildNeedsSupport 沒有 `Support` 就不建艦隊。
func TestOriginalStrategyBuildNeedsSupport(t *testing.T) {
	w := World{Owner: func(int) int { return 0 }}
	if got := (OriginalStrategy{}).Build(w, 0, 0); got != nil {
		t.Errorf("沒有 Support 卻建了 %v", got)
	}
	o := OriginalStrategy{Support: &BuildSupport{
		Money:    func(int) (int, int, bool) { return 100000, 0, false },
		Admirals: func(int) []AdmiralView { return []AdmiralView{adm(7, 20, 100)} },
	}}
	got := o.Build(w, 0, 0)
	if len(got) != 1 || got[0].Admiral != 7 {
		t.Errorf("接了 Support 卻得到 %v", got)
	}
}

// TestBuildPassPicksMostLackingClass 有艦隊之後才看得出配額：
// 十支全是類別 0 → 應有 4／4／0／1，最缺的是類別 1。
func TestBuildPassPicksMostLackingClass(t *testing.T) {
	in := buildBase()
	for i := 0; i < 10; i++ {
		in.Fleets = append(in.Fleets, rules.Fleet{Faction: 0, ClassKey: 0})
	}
	got := BuildPass(in)
	if len(got) == 0 {
		t.Fatal("一支都沒建")
	}
	if c := rules.FleetClass(got[0].ClassKey); c != 1 {
		t.Errorf("第一支補的是類別 %d，預期 1（配額 45％ 缺最多）", c)
	}
}
