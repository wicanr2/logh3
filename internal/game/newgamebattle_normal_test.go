package game_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
)

// 用正式 `bootstrap.Load` 的資源走新遊戲：測試夾具缺的演出資料這裡都有。
// 帝國星域 15 與同盟星域 13 都是電腦開局會去的地方，第一場在第 4 回合。
func TestNewGameContinuesAfterFirstBattle(t *testing.T) {
	root := filepath.Join("..", "..")
	orig := filepath.Join(root, "org_game/extracted")
	if _, err := os.Stat(filepath.Join(orig, "GIN3.COM")); os.IsNotExist(err) {
		t.Skip("缺本機原版")
	}
	for _, c := range []struct {
		name                      string
		menuRow, target, loadRowY int
	}{
		{"帝國", 0, 15, 208},
		{"同盟", 1, 13, 228},
	} {
		t.Run(c.name, func(t *testing.T) {
			r, err := bootstrap.Load(orig, filepath.Join(root, "docs/internals/data"))
			if err != nil {
				t.Fatal(err)
			}
			game.RunNewGameContinuesForTest(t, r, c.menuRow, c.target, c.loadRowY)
		})
	}
}
