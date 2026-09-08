package rules

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRankOfMatchesOracle 把軍階換算釘在收據上量到的對應。
//
// 27 個資料點來自三張清單畫面（`lst2e`／`sortie1e`／`s3sortie`）。
// 12 與 15 同階、29 以上都是上將，所以是分段不是查表。
func TestRankOfMatchesOracle(t *testing.T) {
	for _, c := range []struct {
		merit int
		want  string
	}{
		{4, "上尉"}, {6, "少校"}, {12, "上校"}, {15, "上校"},
		{16, "准將"}, {20, "少將"}, {24, "中將"}, {29, "上將"},
		{0, "少尉"}, {2, "中尉"}, {8, "中校"}, {41, "上將"},
	} {
		if got := RankNames[RankOf(c.merit)]; got != c.want {
			t.Errorf("+0x06 ＝ %d 換出 %s，預期 %s", c.merit, got, c.want)
		}
	}
}

// TestAdmiralListNewGame 用原版的新遊戲樣板驗清單：帝國軍前 16 名的軍階
// 與收據 `lst2e` 相同，而且待命中那張會少掉兩名幕僚。
func TestAdmiralListNewGame(t *testing.T) {
	path := filepath.Join("../..", "org_game/extracted/ADMIRAL.DTX")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("缺 %s，跳過（不用代用品）", path)
	}
	all := ParseAdmirals(data)
	if len(all) != 82 {
		t.Fatalf("解出 %d 筆，預期 82", len(all))
	}
	want := []string{
		"上將", "少將", "少將", "少將", "准將", "准將", "准將", "准將",
		"准將", "准將", "准將", "少將", "准將", "准將", "少將", "准將",
	}
	work, err := os.ReadFile(filepath.Join("../..", "org_game/extracted/WORK.DTX"))
	if err != nil {
		work, err = os.ReadFile(filepath.Join("../..", "org_game/extracted/WORK.DT9"))
		if err != nil {
			t.Skipf("缺 WORK 樣板，跳過")
		}
	}
	staff := StaffAdmirals(work, 0)
	list := AdmiralList(all, 0, false, staff)
	for i, w := range want {
		if got := list[i].Rank(); got != w {
			t.Errorf("第 %d 名的軍階 %s，預期 %s", i, got, w)
		}
	}
	// 收據 lst2e 的第 6、13 名前面標 ○；奧貝斯坦（16）是第三個幕僚職。
	wantStaff := map[int]bool{6: true, 13: true, 16: true}
	gotStaff := map[int]bool{}
	for _, i := range staff {
		gotStaff[i] = true
	}
	if len(gotStaff) != len(wantStaff) {
		t.Fatalf("幕僚職 %v，預期 %v", staff, wantStaff)
	}
	for i := range wantStaff {
		if !gotStaff[i] {
			t.Errorf("幕僚職 %v 裡少了 %d", staff, i)
		}
	}
	standby := AdmiralList(all, 0, true, staff)
	if len(standby) != len(list)-len(staff) {
		t.Errorf("待命中 %d 名，總數 %d 名、幕僚 %d 名", len(standby), len(list), len(staff))
	}
	// 收據 sortie1e 的前 16 列：跳過 6、13、16。
	wantOrder := []int{0, 1, 2, 3, 4, 5, 7, 8, 9, 10, 11, 12, 14, 15, 17, 18}
	for i, w := range wantOrder {
		if standby[i].Index != w {
			t.Fatalf("待命中第 %d 列是提督 %d，預期 %d", i, standby[i].Index, w)
		}
	}
}

// TestStaffAdmiralsByFaction 釘住「一職兩個位元組、帝國在前同盟在後」。
//
// 判準是**兩邊各自的提督清單缺了誰**（`docs/re/61`）：新遊戲載入之後
// `ds:95DA` 那份清單，帝國缺 6／13／16、同盟缺 0x3A／0x3C／0x4F，
// 而那六個就是 `WORK +0x37` 的六個位元組。同盟的月初報告署名
// 「內政專任參謀 邱吾權上校」＝ 提督 60 ＝ 0x3C，對得上。
func TestStaffAdmiralsByFaction(t *testing.T) {
	work, err := os.ReadFile(filepath.Join("../..", "org_game/extracted/WORK.DT9"))
	if err != nil {
		t.Skipf("缺 WORK，跳過")
	}
	for _, c := range []struct {
		faction int
		want    []int
	}{
		{0, []int{6, 13, 16}},  // 梅克林格／克斯拉／奧貝斯坦
		{1, []int{79, 58, 60}}, // 派特里契夫／巴格達胥／邱吾權
	} {
		got := StaffAdmirals(work, c.faction)
		if len(got) != len(c.want) {
			t.Fatalf("陣營 %d 拿到 %v，預期 %v", c.faction, got, c.want)
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Fatalf("陣營 %d 的第 %d 個是 %d，預期 %d（全部 %v）",
					c.faction, i, got[i], c.want[i], got)
			}
		}
	}
	if got := StaffAdmiral(work, PostDomestic, 1); got != 60 {
		t.Fatalf("同盟的內政專任參謀是 %d，預期 60（邱吾權）", got)
	}
}
