package presentation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMapDecorMissingCategoryKeepsOriginal(t *testing.T) {
	r := &Renderer{}
	if err := r.loadMapDecor(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	for _, im := range r.mapPlanets {
		if im != nil {
			t.Fatal("缺素材使用了替代圖")
		}
	}
	for _, im := range r.mapEmblems {
		if im != nil {
			t.Fatal("缺徽章使用了替代圖")
		}
	}
}
func TestMapDecorRejectsChangedSource(t *testing.T) {
	dir := t.TempDir()
	b, e := os.ReadFile("../../assets/hd/strategy/MANIFEST.json")
	if os.IsNotExist(e) {
		t.Skip("缺核准素材，不使用代用品")
	}
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(dir, "MANIFEST.json"), b, 0600); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(dir, "planets.png"), []byte("damaged"), 0600); e != nil {
		t.Fatal(e)
	}
	r := &Renderer{}
	if e = r.loadMapDecor(dir); e == nil || !strings.Contains(e.Error(), "雜湊") {
		t.Fatalf("未拒絕雜湊變更：%v", e)
	}
}
