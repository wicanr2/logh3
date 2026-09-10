package rules_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 收據 `staffpick`（`docs/re/268`）：新遊戲按一次「實行」，在同盟 AI 指派
// 內政專任參謀的前後各倒一份提督表，並用 `-watch` 錄下寫入本身。
//
//	[watch] #258991706 5DA80: 3C → 3F  ← 2D10:02F1     60 → 63
//
// 原版寫進去的是 63，所以規則算出來的也必須是 63。
func loadStaffReceipt(t *testing.T, name string) []byte {
	t.Helper()
	p := filepath.Join("..", "..", "workplace", "oracle", "staffpick", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Skipf("缺收據 %s，跳過（不用代用品）", p)
	}
	return b
}

func TestDomesticStaffPickMatchesOracle(t *testing.T) {
	pre := loadStaffReceipt(t, "staffpre.bin")
	work := loadStaffReceipt(t, "staffwork.bin")

	// 指派當下的現任（`WORK +0x3B`／`+0x3C`）。
	if got := rules.StaffAdmiral(work, rules.PostDomestic, 1); got != 60 {
		t.Fatalf("收據裡同盟的內政參謀是提督 %d，應該是 60", got)
	}
	if got := rules.StaffAdmiral(work, rules.PostDomestic, 0); got != 16 {
		t.Fatalf("收據裡帝國的內政參謀是提督 %d，應該是 16", got)
	}

	// 同盟（電腦座位）：AI 跑過，換成 63。
	if got := rules.PickDomesticStaff(pre, 1, 31, 60); got != 63 {
		t.Errorf("同盟的內政參謀算出提督 %d，收據是 63", got)
	}
	// 帝國（玩家座位）：同一條規則算出 39，但原版那一格從頭到尾沒動
	// ——玩家的回合不跑 AI，`2D10:0038` 不會執行。這一行釘的是
	// 「規則本身不分陣營」，不是「原版會這樣做」。
	if got := rules.PickDomesticStaff(pre, 0, 0, 16); got != 39 {
		t.Errorf("帝國那一側算出提督 %d，同一份傾印應該是 39", got)
	}
}

// TestIntelStaffStaysMatchesOracle 是同一份收據的第二個獨立觀測：指派前後
// 整張提督表**只有兩個位元組不同**（60 卸任、63 就任），所以先跑的
// `2D10:0143`（情報專任參謀）沒有換人——現任必須就是規則挑出來的那個。
func TestIntelStaffStaysMatchesOracle(t *testing.T) {
	pre := loadStaffReceipt(t, "staffpre.bin")
	work := loadStaffReceipt(t, "staffwork.bin")
	// 只有同盟（電腦座位）跑得到這一支。帝國那一側算出提督 32 而現任是
	// 13，那不是矛盾——玩家的回合根本不會執行 `2D10:0143`。
	// 收據裡同盟的情報參謀是 **75**，而存檔範本（`docs/internals/30`）記的是
	// 58——那一格同樣被 AI 換過，`docs/re/248` 量到的執行期值也是 75。
	cur := rules.StaffAdmiral(work, rules.PostIntel, 1)
	if cur != 75 {
		t.Fatalf("收據裡同盟的情報參謀是提督 %d，應該是 75", cur)
	}
	if got := rules.PickIntelStaff(pre, 1, 31, cur); got != cur {
		t.Errorf("同盟的情報參謀算出提督 %d，收據裡沒有換人（現任 %d）", got, cur)
	}
}

// TestStaffAssignmentSideEffects 釘住就任／卸任只動 `+0x09` bit7 這件事
// ——`docs/internals/08` 原本把那個位元記成「沒帶艦隊」，那與「載入時被
// 夾成 0x7F 就沒有人合格」互相矛盾；收據顯示它是「沒有專任參謀職位」。
func TestStaffAssignmentSideEffects(t *testing.T) {
	pre := loadStaffReceipt(t, "staffpre.bin")
	post := loadStaffReceipt(t, "staffpost.bin")
	if len(pre) != len(post) {
		t.Fatalf("兩份提督表長度不同：%d／%d", len(pre), len(post))
	}
	changed := map[int][]int{}
	for i := range pre {
		if pre[i] != post[i] {
			changed[i/rules.AdmiralRecord] = append(
				changed[i/rules.AdmiralRecord], i%rules.AdmiralRecord)
		}
	}
	if len(changed) != 2 {
		t.Fatalf("指派前後有 %d 位提督被改，收據是 2 位（60 卸任、63 就任）", len(changed))
	}
	for _, idx := range []int{60, 63} {
		offs, ok := changed[idx]
		if !ok || len(offs) != 1 || offs[0] != 0x09 {
			t.Errorf("提督 %d 被改的欄位是 %v，收據只有 +0x09", idx, offs)
		}
	}
	if pre[60*rules.AdmiralRecord+9] != 0x7F || post[60*rules.AdmiralRecord+9] != 0x80 {
		t.Errorf("提督 60 的 +0x09 是 %02X→%02X，收據是 7F→80（卸任）",
			pre[60*rules.AdmiralRecord+9], post[60*rules.AdmiralRecord+9])
	}
	if pre[63*rules.AdmiralRecord+9] != 0x80 || post[63*rules.AdmiralRecord+9] != 0x7F {
		t.Errorf("提督 63 的 +0x09 是 %02X→%02X，收據是 80→7F（就任）",
			pre[63*rules.AdmiralRecord+9], post[63*rules.AdmiralRecord+9])
	}
}
