package game

import "testing"

// TestRallyPointsMatchOriginal 對原版量到的七個集結點。
//
// `docs/re/100`：盯著同盟 16 的目標欄位看誰寫它，抓到 `31DB:01B6` 從
// `ds:0D70C` 抽點，倒出來剛好七筆，全部落在 `MAP26.CMP` 圖塊 1 的格子上。
// 這條測試把那七個座標釘住——remake 掃地圖掃出來的要一模一樣。
func TestRallyPointsMatchOriginal(t *testing.T) {
	g := newTestGame(t)
	if len(g.res.BattleMaps) <= 26 || len(g.res.BattleMaps[26]) == 0 {
		t.Skip("缺 MAP26.CMP，跳過（不用代用品）")
	}
	want := [][2]int{
		{816, 660}, {368, 800}, {672, 920}, {1168, 1160},
		{864, 1220}, {1328, 1540}, {1040, 1660},
	}
	got := g.battleRegionRally(26)
	if len(got) != len(want) {
		t.Fatalf("集結點 %d 個，原版量到 %d 個", len(got), len(want))
	}
	for i, w := range want {
		if got[i].X != w[0] || got[i].Y != w[1] {
			t.Errorf("第 %d 個 (%d,%d)，原版是 (%d,%d)",
				i, got[i].X, got[i].Y, w[0], w[1])
		}
	}
}

// TestPlanetIsPassable 釘住「行星那一格走得過去」。
//
// `docs/re/80` 的同盟 31 旗艦就站在雷奧米爾的座標 (1320,812) 上，而
// `ds:9BB4` 裡圖塊 6 的遮罩是 00（`docs/re/100`）。要是哪天把行星改成
// 障礙，戰術層就再也帶不到行星上，這條會先叫。
func TestPlanetIsPassable(t *testing.T) {
	g := newTestGame(t)
	cells := g.res.BattleMaps
	if len(cells) <= 26 || len(cells[26]) == 0 || g.res.Terrain == nil {
		t.Skip("缺 MAP26.CMP 或地形表，跳過")
	}
	if !g.res.Terrain.Passable(cells[26], 1320, 812) {
		t.Error("雷奧米爾 (1320,812) 被判成障礙")
	}
}
