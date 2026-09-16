package game

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 右下角「敵艦隊」分頁有列的時候（`docs/re/325`、規格 052）：與
// `tools/enemytab_receipts.sh` 的 `enm-*` 收據對。載具是 S3 帝國第一回合的
// 戰略地圖，同盟第２艦隊（槽 17）用改記憶體的方式搬進瓦爾哈拉（0）、四組
// 艦艇的索敵等級寫成要的值（規格 014：要局面就改記憶體）。分頁鈕在滑鼠
// (490,250) ＝ 畫面 (490,290)，列 k 在 (450, 308＋20k)。

// reachEnemyTab 載 S3、把槽 17 搬進瓦爾哈拉並給各組索敵等級，再點「敵艦隊」。
func reachEnemyTab(t *testing.T, details ...[]int) *Game {
	t.Helper()
	g := newTestGame(t)
	data := filepath.Join(root, "docs/internals/data")
	var err error
	if g.res.MessageBox, err = screen.LoadMessageBoxSpec(filepath.Join(data, "message-box.json")); err != nil {
		t.Fatal(err)
	}
	if g.res.EnemyWindow, err = screen.LoadEnemyWindowSpec(filepath.Join(data, "enemy-fleet-window.json")); err != nil {
		t.Fatal(err)
	}
	// 走存檔那條路載第 10 格（S3）：載入當下的未安定度重算才會與原版一樣
	// （瓦爾哈拉治安 93%）；`newTestGame` 直接 `LoadSlot` 沒有那一步。
	g.res.OrigDir = filepath.Join(root, "org_game/extracted")
	g.loadSlotFromDisk(10)
	if g.State() != StateCommandPhase || g.faction != 0 {
		t.Fatalf("S3 載進來在 %v、座位 %d", g.State(), g.faction)
	}
	for k, slot := range []int{17, 16} {
		if k >= len(details) {
			break
		}
		f := g.fleetBySlot(t, slot)
		f.Region = 0
		for i := range f.Composition {
			if i < len(details[k]) {
				f.Composition[i].Detail = details[k][i]
			}
		}
	}
	g.Click(490, 290)
	if g.listTab() != listTabEnemy {
		t.Fatalf("點了敵艦隊分頁之後在第 %d 頁", g.listTab())
	}
	return g
}

func (g *Game) fleetBySlot(t *testing.T, slot int) *rules.Fleet {
	t.Helper()
	for i := range g.fleets {
		if g.fleets[i].Index == slot {
			return &g.fleets[i]
		}
	}
	t.Fatalf("艦隊表裡沒有槽 %d", slot)
	return nil
}

// 等級 3：列印編號、提督、士氣；點列開視窗（陣營色的軍名、第Ｎ艦隊、提督＋軍階、
// 部隊種類的字母、頭像、結束）；「結束」或右鍵關、列的反白還原。
func TestEnemyTabClearMatchesOracle(t *testing.T) {
	g := reachEnemyTab(t, []int{3, 3, 3, 3})
	if rows := g.enemyRows(g.listPage(listTabEnemy)); len(rows) != 1 || !rows[0].Known || rows[0].Number != 2 {
		t.Fatalf("列 %+v，要一列、看得清楚的第２艦隊", rows)
	}
	matches(t, g, "enm-l3")
	g.Click(450, 308)
	if g.tabs.enemy < 0 || g.tabs.pressed != 0 {
		t.Fatalf("點列之後視窗 %d、反白列 %d", g.tabs.enemy, g.tabs.pressed)
	}
	matches(t, g, "enm-l3-row")
	g.Click(300, 240) // 點在視窗別處：不動
	if g.tabs.enemy < 0 {
		t.Fatal("點在視窗別處把視窗關了")
	}
	g.Click(488, 285) // 結束
	if g.tabs.enemy >= 0 || g.tabs.pressed >= 0 {
		t.Fatalf("按了結束之後視窗 %d、反白列 %d", g.tabs.enemy, g.tabs.pressed)
	}
	matches(t, g, "enm-l3-ok")
	g.Click(450, 308)
	g.RightClick(300, 240)
	if g.tabs.enemy >= 0 || g.tabs.pressed >= 0 {
		t.Fatalf("右鍵之後視窗 %d、反白列 %d", g.tabs.enemy, g.tabs.pressed)
	}
	matches(t, g, "enm-l3-rclick")
}

// 等級 1：列是「第？艦隊 ？？？？？？？？？ ??」；點列跳「情報資料來源不足」
// （模態、只認確認、右鍵不理）；關了列的反白還原。
func TestEnemyTabDimMatchesOracle(t *testing.T) {
	g := reachEnemyTab(t, []int{1, 1, 1, 1})
	matches(t, g, "enm-l1")
	g.Click(450, 308)
	if !g.msgOn || !g.tabs.enemyMsg || g.tabs.pressed != 0 {
		t.Fatalf("點列之後訊息框 %v、反白列 %d", g.msgOn, g.tabs.pressed)
	}
	matches(t, g, "enm-l1-row")
	g.Click(300, 240) // 點在框內別處：不動
	if !g.msgOn {
		t.Fatal("點在框內別處把訊息框關了")
	}
	matches(t, g, "enm-l1-row-miss")
	g.RightClick(300, 240)
	if !g.msgOn {
		t.Fatal("右鍵把訊息框關了")
	}
	matches(t, g, "enm-l1-row-rclk")
	g.Click(432, 270) // 確認（樣板 0：格 (17,3) → (128＋16＋272, 40＋140＋20＋60)）
	if g.msgOn || g.tabs.pressed >= 0 {
		t.Fatalf("按了確認之後訊息框 %v、反白列 %d", g.msgOn, g.tabs.pressed)
	}
	matches(t, g, "enm-l1-row-ok")
}

// 混合：槽 17 四組 3／2／1／0（列照最大值；視窗的部隊種類 ≥2 印字母、1 印
// 「？」、0 不印）、士氣 30 換色；槽 16 六組全 2 也搬進來（列是「第？艦隊」，
// 排在前面——表的順序）。
func TestEnemyTabMixedMatchesOracle(t *testing.T) {
	g := reachEnemyTab(t, []int{3, 2, 1, 0}, []int{2, 2, 2, 2, 2, 2})
	g.fleetBySlot(t, 17).Morale = 30
	rows := g.enemyRows(g.listPage(listTabEnemy))
	if len(rows) != 2 || rows[0].Known || rows[0].Number != 1 || !rows[1].Known || rows[1].Number != 2 {
		t.Fatalf("列 %+v，要 第１（看不清）、第２（看得清）", rows)
	}
	matches(t, g, "enm-mix")
	g.Click(450, 308)
	if !g.msgOn {
		t.Fatal("點第１艦隊那一列該跳「情報資料來源不足」")
	}
	matches(t, g, "enm-mix-row0")
	g.Click(432, 270)
	g.Click(450, 328)
	if g.tabs.enemy < 0 {
		t.Fatal("點第２艦隊那一列該開視窗")
	}
	v, _ := g.enemyWindowView()
	if len(v.Groups) != 4 || v.Groups[0].Detail != 3 || v.Groups[3].Detail != 0 {
		t.Fatalf("視窗的組 %+v", v.Groups)
	}
	matches(t, g, "enm-mix-row1")
}

// 等級 0 的不列。
func TestEnemyTabUnseenIsEmpty(t *testing.T) {
	g := reachEnemyTab(t, []int{0, 0, 0, 0})
	if rows := g.enemyRows(g.listPage(listTabEnemy)); len(rows) != 0 {
		t.Fatalf("索敵等級 0 也列了：%+v", rows)
	}
	matches(t, g, "enm-l0")
}
