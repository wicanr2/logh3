package audio

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, dir, name string, n int) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), make([]byte, n), 0o644); err != nil {
		t.Fatal(err)
	}
}

// 目錄不存在不是錯誤——那只是「玩家還沒轉檔」，而**整份清單都要算缺**。
//
// 反面不會報錯：回一個空的 Missing 會讓「一首都沒有」看起來像「都齊了」，
// 而症狀是安靜，與「沒接播放層」一模一樣。
func TestMissingWhenDirAbsent(t *testing.T) {
	l := NewLibrary(filepath.Join(t.TempDir(), "不存在"))
	if got := len(l.Missing()); got != len(Tracks) {
		t.Errorf("缺 %d 首，預期整份 %d 首都缺", got, len(Tracks))
	}
	if l.Have() != 0 {
		t.Errorf("找得到 %d 首，預期 0", l.Have())
	}
	if _, ok := l.Path("OPEN_F"); ok {
		t.Error("目錄不存在卻回報找得到")
	}
}

// 只有部分轉好時，Missing 要精確到哪幾首。
func TestMissingListsExactlyTheAbsentOnes(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "OPEN_F.ogg", 64)
	write(t, dir, "BGM_0F.ogg", 64)
	l := NewLibrary(dir)
	if l.Have() != 2 {
		t.Fatalf("找得到 %d 首，預期 2", l.Have())
	}
	miss := l.Missing()
	if len(miss) != len(Tracks)-2 {
		t.Fatalf("缺 %d 首，預期 %d", len(miss), len(Tracks)-2)
	}
	for _, m := range miss {
		if m == "OPEN_F" || m == "BGM_0F" {
			t.Errorf("%s 明明在，卻被算成缺", m)
		}
	}
}

// **零位元組的檔案要算缺。** 轉檔中途被砍掉會留下 0 bytes 的檔，
// 當成「有」的話播放層會拿它去解碼然後報一個與現場無關的錯。
func TestEmptyFileCountsAsMissing(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "OPEN_F.ogg", 0)
	l := NewLibrary(dir)
	if _, ok := l.Path("OPEN_F"); ok {
		t.Error("0 bytes 的檔被當成有")
	}
	if l.Have() != 0 {
		t.Errorf("Have=%d，預期 0", l.Have())
	}
}

// 曲名不分大小寫——呼叫端寫 `open_f` 或 `OPEN_F` 都要找得到，
// 不然「有這首卻播不出來」會被誤判成缺檔。
func TestPathIsCaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "OPEN_F.ogg", 64)
	l := NewLibrary(dir)
	for _, name := range []string{"OPEN_F", "open_f", "Open_F"} {
		if _, ok := l.Path(name); !ok {
			t.Errorf("%q 找不到", name)
		}
	}
}

// 曲目清單要與 `.MDS` 索引對得上：22 首，`_F`／`_M` 成對。
func TestTracksMatchTheOriginalSet(t *testing.T) {
	if len(Tracks) != 22 {
		t.Fatalf("曲目 %d 首，原版是 22 首（docs/internals/data/mds/index.json）", len(Tracks))
	}
	seen := map[string]int{}
	for _, n := range Tracks {
		if len(n) < 2 {
			t.Fatalf("曲名 %q 太短", n)
		}
		seen[n[:len(n)-1]]++
	}
	for base, n := range seen {
		if n != 2 {
			t.Errorf("%s 有 %d 個版本，預期 _F 與 _M 各一", base, n)
		}
	}
}
