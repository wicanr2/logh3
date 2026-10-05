package asset

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestTitle350Original(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(origDir(t), "SPTIT.350"))
	if err != nil {
		t.Skipf("缺原版SPTIT.350：%v", err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != "c38203aa19c9cab462b9e4e7806225e668c7593ecbdbf30d03b883bdea0ba747" {
		t.Fatal("來源不是規格111固定版本")
	}
	img, err := DecodeTitle350(b)
	if err != nil {
		t.Fatal(err)
	}
	if img.W != 640 || img.H != 350 || img.Used != len(b) || fmt.Sprintf("%x", sha256.Sum256(img.Pix)) != "addde5e0929eb5f99a5430a714b19041a2cbe5530a4720a3641abebbfa3bf8ae" {
		t.Fatal("完整來源尺寸、消耗或色號不符原版雙跑")
	}
	for _, name := range []string{"title68-r96-a", "title68-r96-b"} {
		want, e := os.ReadFile(filepath.Join("../..", "workplace/oracle", name, "screen.bin"))
		if e != nil {
			t.Logf("缺本機原版收據%s，保留固定來源及色號SHA核對", name)
			continue
		}
		if !bytes.Equal(img.Pix, want) {
			t.Fatal("色號不符", name)
		}
	}
	for cut := 0; cut < len(b); cut++ {
		if partial, e := DecodeTitle350(b[:cut]); e == nil || partial != nil {
			t.Fatalf("截斷%d仍回部分圖或無錯誤", cut)
		}
	}
	for _, off := range []int{9684, 18023, 26614, 33345} {
		bad := bytes.Clone(b)
		bad[off] = 0
		if _, e := DecodeTitle350(bad); e == nil {
			t.Fatal("接受錯分隔byte", off)
		}
	}
	for _, bad := range [][]byte{append(bytes.Clone(b), 0), {0, 0, 0, 0}, {0, 0, 0x7f, 0xff}} {
		if partial, e := DecodeTitle350(bad); e == nil || partial != nil {
			t.Fatal("接受尾端多資料、零run或超長run")
		}
	}
}
