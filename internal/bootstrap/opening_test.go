package bootstrap

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/rules"
)

// 正式載入器走正常玩家入口，核對原版開局整表（研究 352／規格 065）。
func TestOpeningRuntimeTablesMatchesOriginal(t *testing.T) {
	root := filepath.Join("..", "..")
	orig := filepath.Join(root, "org_game", "extracted")
	if _, err := os.Stat(filepath.Join(orig, "GIN3.COM")); err != nil {
		t.Skipf("缺原版輸入：%v", err)
	}
	mem, err := os.ReadFile(filepath.Join(root, "workplace", "oracle",
		"planarrow-budget-unchanged-post", "memory.bin"))
	if err != nil {
		t.Skipf("缺 dosgolem 原版完整收據：%v", err)
	}
	if len(mem) != 1048576 {
		t.Fatalf("原版記憶體長度 %d", len(mem))
	}
	r, err := Load(orig, filepath.Join(root, "docs", "internals", "data"))
	if err != nil {
		t.Fatal(err)
	}
	g, err := game.New(r)
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, 210)
	for _, c := range []byte{'1', '3', '8', '\r'} {
		g.Key(c)
	}
	g.RightClick(300, 240)
	g.RightClick(300, 240)
	if g.State() != game.StateCommandPhase {
		t.Fatalf("正常入口沒有到命令階段：%v", g.State())
	}
	for _, c := range []struct {
		name      string
		got, want []byte
	}{
		{"星域", r.Stellar, mem[0x5F2CC : 0x5F2CC+rules.RegionCount*rules.RegionRecord]},
		{"行星", r.Planets.Data[rules.PlanetHeader:], mem[0x53132 : 0x53132+81*rules.PlanetRecord]},
	} {
		if !bytes.Equal(c.got, c.want) {
			t.Errorf("正式載入器的%s整表不一致", c.name)
		}
	}
}
