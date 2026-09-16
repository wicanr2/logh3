package audio

import (
	"testing"

	"github.com/wicanr2/logh3/internal/game"
)

// TestTuneForMatchesOriginal 把 `docs/re/274` 的對照釘住。
//
// 三份出貨存檔的 `WORK.DT? +0x20` 分別是 0／1／3，而既有收據載完那三格
// 之後讀進去的正是 `BGM_0F`／`BGM_1F`／`BGM_3F`——下面前三列就是那三筆。
func TestTuneForMatchesOriginal(t *testing.T) {
	cases := []struct {
		name    string
		st      game.State
		faction int
		s       Settings
		want    string
	}{
		// 新遊戲：兩格清零、出貨 MUSIC.INI 沒有 MIDI 介面 → BGM_0F。
		{"新遊戲的遊戲本體", game.StateMainGame, 0, DefaultSettings, "BGM_0F"},
		// WORK.DT9 的 +0x20 ＝ 1（收據 ownS2 讀 bgm_1f.mds）。
		{"存檔槽 9", game.StateCommandPhase, 0, Settings{Device: 1, BGM: 1}, "BGM_1F"},
		// WORK.DTA 的 +0x20 ＝ 3（收據 s3cmd1 讀 bgm_3f.mds）。
		{"存檔槽 A", game.StateCommandPhase, 0, Settings{Device: 1, BGM: 3}, "BGM_3F"},

		// 戰術畫面沿用平時配樂：收據 s3assault 的 BGM_3F 撐過進戰場那一段。
		{"戰術畫面", game.StateBattle, 0, Settings{Device: 1, BGM: 3}, "BGM_3F"},

		// 開頭動畫一律 _M：OPEN.EXE 兩個外皮都寫死 'm'。
		{"開頭動畫", game.StateIntro, 0, DefaultSettings, "OPEN_M"},
		{"開頭動畫不看音源設定", game.StateIntro, 0, Settings{Device: 0}, "OPEN_M"},

		// 主選單與防拷問答沒有配樂。
		{"主選單", game.StateMainMenu, 0, DefaultSettings, ""},
		{"防拷問答", game.StateCopyProtection, 0, DefaultSettings, ""},

		// 結局照命令列第二個數字：9 1 帝國、9 2 同盟。
		{"帝國結局", game.StateEnding, 0, DefaultSettings, "END_1F"},
		{"同盟結局", game.StateEnding, 1, DefaultSettings, "END_2F"},

		// 音源設定換的是最後那個字元。
		{"MIDI", game.StateMainGame, 0, Settings{Device: 0, BGM: 2}, "BGM_2M"},
		{"關閉", game.StateMainGame, 0, Settings{Device: 2, BGM: 2}, ""},
		{"關閉連開頭動畫都不播", game.StateIntro, 0, Settings{Device: 2}, ""},
	}
	for _, c := range cases {
		if got := TuneFor(c.st, c.faction, c.s); got != c.want {
			t.Errorf("%s：TuneFor ＝ %q，要 %q", c.name, got, c.want)
		}
	}
}

// TestBattleTune 釘住「戰鬥曲 ＝ 8F35 ＋ 6，值 2 時沿用平時那一首」。
func TestBattleTune(t *testing.T) {
	cases := []struct {
		s    Settings
		want string
	}{
		{Settings{Device: 1, BGM: 3, Battle: 0}, "BGM_6F"}, // 全艦攻擊（收據 s3assault）
		{Settings{Device: 1, BGM: 3, Battle: 1}, "BGM_7F"}, // 決戰前夕
		{Settings{Device: 1, BGM: 3, Battle: 2}, "BGM_3F"}, // 使用原有的ＢＧＭ
		{Settings{Device: 0, BGM: 0, Battle: 0}, "BGM_6M"},
		{Settings{Device: 2}, ""},
	}
	for _, c := range cases {
		if got := c.s.BattleTune(); got != c.want {
			t.Errorf("%+v：BattleTune ＝ %q，要 %q", c.s, got, c.want)
		}
	}
}

// TestEveryTuneIsReachable 是**對照表與曲庫的對帳**：22 首裡，除了
// `OPEN_F`（原版永遠不會打開，`docs/re/274` §8）之外每一首都要有一條
// 走得到它的路。漏一首就代表對照表少接了一格。
func TestEveryTuneIsReachable(t *testing.T) {
	reached := map[string]bool{}
	for _, dev := range []int{0, 1} {
		for _, f := range []int{0, 1} {
			for bgm := 0; bgm < 6; bgm++ {
				s := Settings{Device: dev, BGM: bgm}
				reached[TuneFor(game.StateMainGame, f, s)] = true
				reached[TuneFor(game.StateEnding, f, s)] = true
				reached[TuneFor(game.StateIntro, f, s)] = true
				// 戰鬥曲從交戰演出那個狀態走到（規格 056）。
				for b := 0; b < 3; b++ {
					s.Battle = b
					reached[TuneFor(game.StateEngagement, f, s)] = true
				}
			}
		}
	}
	for _, name := range Tracks {
		if name == "OPEN_F" {
			if reached[name] {
				t.Errorf("%s 不該走得到：原版沒有任何一處送 'f' 給 open_%%c.mds", name)
			}
			continue
		}
		if !reached[name] {
			t.Errorf("%s 沒有任何一條路走得到——對照表少接了一格", name)
		}
	}
}
