package audio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHDLibraryRequiresCompleteMIDICatalog(t *testing.T) {
	dir := t.TempDir()
	for _, name := range Tracks {
		if strings.HasSuffix(name, "M") {
			if err := os.WriteFile(filepath.Join(dir, name+".ogg"), []byte{1}, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	library := NewHDLibrary(dir)
	if missing := library.Missing(); len(missing) != 0 {
		t.Fatal(missing)
	}
	fm, ok := library.Path("BGM_6F")
	if !ok || filepath.Base(fm) != "BGM_6M.ogg" {
		t.Fatalf("同曲映射錯誤：%s", fm)
	}
	if err := os.Remove(filepath.Join(dir, "BGM_6M.ogg")); err != nil {
		t.Fatal(err)
	}
	if missing := NewHDLibrary(dir).Missing(); len(missing) != 2 {
		t.Fatalf("缺少一首時應同時缺FM鍵與MIDI鍵：%v", missing)
	}
}
