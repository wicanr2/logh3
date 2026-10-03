package game_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
)

func TestNormalSpottingInformationAndSave(t *testing.T) {
	root := filepath.Join("..", "..")
	orig := filepath.Join(root, "org_game/extracted")
	if _, err := os.Stat(filepath.Join(orig, "GIN3.COM")); os.IsNotExist(err) {
		t.Skip("缺本機原版")
	}
	r, err := bootstrap.Load(orig, filepath.Join(root, "docs/internals/data"))
	if err != nil {
		t.Fatal(err)
	}
	g, err := game.New(r)
	if err != nil {
		t.Fatal(err)
	}
	game.RunNormalSpottingForTest(t, g)
}
