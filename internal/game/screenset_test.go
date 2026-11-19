package game

import (
	"os"
	"path/filepath"
	"testing"
)

// 畫面規格改成按需解析（規格160）之後，格式錯誤不會在啟動時爆出來，
// 這支測試把每一張都解析一次、每一組星域色都建一次，補回那道檢查。
func TestCommandScreensAllParse(t *testing.T) {
	data := filepath.Join(root, "docs/internals/data")
	bit, err1 := os.ReadFile(filepath.Join(root, "org_game/extracted/STRMAP.BIT"))
	rgn, err2 := os.ReadFile(filepath.Join(root, "org_game/extracted/STRMAP.RGN"))
	if err1 != nil || err2 != nil {
		t.Skip("缺原版 STRMAP，跳過（不用代用品）")
	}
	set, maps, err := LoadCommandScreens(data, filepath.Join(data, "strmap-regions.json"), bit, rgn)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range set.Names() {
		spec := set.Get(name)
		if spec == nil {
			t.Errorf("%s 解析失敗", name)
			continue
		}
		if spec != set.Get(name) {
			t.Errorf("%s 第二次取回不同的指標", name)
		}
		if spec.Map.Colours != "" && maps.Get(spec.Map.Colours) == nil {
			t.Errorf("%s 的星域色 %s 建不出地圖", name, spec.Map.Colours)
		}
	}
	if set.Get("screen-no-such-thing") != nil {
		t.Error("不存在的規格回了非 nil")
	}
}
