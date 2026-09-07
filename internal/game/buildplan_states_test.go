package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

// TestBuildPlanStatesMatchOriginal 拿改過記憶體的原版局面對拍：
// 把同一組百分比餵給 remake，整張 640×480 必須逐位元組相同。
//
// 這些局面手點到不了（例如七列各 1%、某一列吃滿 100%），是用 `-poke`
// 把百分比陣列直接寫進原版記憶體造出來的（`docs/re/11`）。點擊序列與
// `plan1` 完全一樣，所以游標位置也一樣，可以整張比。
//
// 走的是**正常玩家路徑**（主選單 → 防拷 → 訂造報告 → 建造計劃），
// 只有百分比是直接塞進去的：署名、頭像、建造費與右側面板都跟著局面走
// （`docs/spec/011`），拿規格單獨畫是畫不全的。
//
// 清單在 `workplace/explore/manifest.json`——那是一次性證據，
// 不在 Git 裡；缺檔就跳過，不做代用品。
func TestBuildPlanStatesMatchOriginal(t *testing.T) {
	dir := filepath.Join(root, "workplace/explore")
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Skip("缺 workplace/explore/manifest.json，跳過")
	}
	var mf struct {
		Cost   int                                 `json:"cost"`
		States map[string][rules.BuildPlanRows]int `json:"states"`
	}
	if err := json.Unmarshal(raw, &mf); err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(mf.States))
	for n := range mf.States {
		names = append(names, n)
	}
	sort.Strings(names)
	checked := 0
	for _, name := range names {
		want, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Skipf("缺 %s，跳過", name)
		}
		g := reachBuildPlan(t)
		if g.plan.Cost != mf.Cost {
			t.Fatalf("建造費是 %d，清單記的是 %d", g.plan.Cost, mf.Cost)
		}
		g.plan.Pct = mf.States[name]
		fb := g.Frame()
		if len(want) != len(fb.Pix) {
			t.Fatalf("%s 是 %d bytes，畫布是 %d", name, len(want), len(fb.Pix))
		}
		bad, fx, fy := 0, -1, -1
		for i := range want {
			if want[i] != fb.Pix[i] {
				if bad == 0 {
					fx, fy = i%gfx.Width, i/gfx.Width
				}
				bad++
			}
		}
		if bad != 0 {
			t.Errorf("%s：差 %d / %d 像素，第一處在 (%d,%d)", name, bad, len(want), fx, fy)
		}
		checked++
	}
	t.Logf("同狀態對拍 %d 個局面，各 %d 像素", checked, gfx.Width*gfx.Height)
}

// reachBuildPlan 走玩家路徑到次月份船艦建造計劃（帝國）。
func reachBuildPlan(t *testing.T) *Game {
	t.Helper()
	g, err := New(load(t))
	if err != nil {
		t.Fatal(err)
	}
	g.Click(300, 210)
	for _, c := range []byte{'1', '3', '8', '\r'} {
		g.Key(c)
	}
	g.Click(448, 330)
	g.Click(448, 310)
	if g.State() != StateBuildPlan || g.plan == nil {
		t.Fatalf("走不到建造計劃，停在 %v", g.State())
	}
	return g
}
