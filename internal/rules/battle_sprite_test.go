package rules

import "testing"

// 原版兩份收據裡逐塊比對出來的十六個樣本（`docs/re/85`）。
// 陣營全是同盟（1），所以基底是 50。
func TestBattleSpriteMatchesOriginal(t *testing.T) {
	cases := []struct {
		name    string
		faction int
		detail  int
		kind    int
		want    int
	}{
		{"s3fight 艦隊28 組0", 1, 1, 0, 74},
		{"s3fight 艦隊31 組0-2", 1, 2, 1, 67},
		{"s3fight 艦隊31 組3-5", 1, 1, 2, 74},
		{"s3fight 艦隊31 組6", 1, 1, 6, 74},
		{"s3kinds2 艦隊31 組0", 1, 2, 0, 66},
		{"s3kinds2 艦隊31 組1", 1, 2, 1, 67},
		{"s3kinds2 艦隊31 組2", 1, 2, 2, 68},
		{"s3kinds2 艦隊31 組3", 1, 1, 3, 74},
	}
	for _, c := range cases {
		if got := BattleSprite(c.faction, c.detail, c.kind); got != c.want {
			t.Errorf("%s：BattleSprite(%d,%d,%d) ＝ %d，原版是 %d",
				c.name, c.faction, c.detail, c.kind, got, c.want)
		}
	}
}

// 算式本身（`25DB:03E8`）：陣營基底 20／50，艦種 7 與層級 0 都畫第 9 塊。
func TestBattleSpriteFormula(t *testing.T) {
	cases := []struct {
		faction, detail, kind, want int
	}{
		{0, 2, 0, 36},
		{0, 2, 5, 41},
		{0, 1, 0, 44},
		{1, 2, 0, 66},
		{1, 1, 0, 74},
		{0, 2, 7, BlankSprite},
		{1, 2, 7, BlankSprite},
		{1, 0, 1, BlankSprite},
		{0, 0, 0, BlankSprite},
	}
	for _, c := range cases {
		if got := BattleSprite(c.faction, c.detail, c.kind); got != c.want {
			t.Errorf("BattleSprite(%d,%d,%d) ＝ %d，要 %d",
				c.faction, c.detail, c.kind, got, c.want)
		}
	}
}
