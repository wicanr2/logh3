package asset

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
)

func TestFixedOggSerialRealReference(t *testing.T) {
	want, err := os.ReadFile("../../assets/original/music/ogg/BGM_0F.ogg")
	if os.IsNotExist(err) {
		t.Skip("缺保存曲庫，不製造原版代用品")
	}
	if err != nil {
		t.Fatal(err)
	}
	serial := binary.LittleEndian.Uint32(want[14:18])
	other, err := FixedOggSerial(want, 123456789)
	if err != nil || bytes.Equal(other, want) {
		t.Fatalf("未建立不同封裝的有效串流：%v", err)
	}
	got, err := FixedOggSerial(other, serial)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("固定序號未恢復完整基準：%v", err)
	}
	native, err := os.ReadFile("../../workplace/music69-fm/render/BGM_0F-1.ogg")
	if err == nil {
		got, err := FixedOggSerial(native, serial)
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("真實重編產物未恢復完整基準：%v", err)
		}
	}
	for _, mutate := range []func([]byte) []byte{
		func(d []byte) []byte { d[22] ^= 1; return d },
		func(d []byte) []byte { d[5] = 0; return d },
		func(d []byte) []byte { d[18] = 1; return d },
		func(d []byte) []byte { return d[:26] },
		func(d []byte) []byte { return d[:len(d)-1] },
		func(d []byte) []byte { return append(d, d...) },
	} {
		d := mutate(bytes.Clone(want))
		before := bytes.Clone(d)
		if out, err := FixedOggSerial(d, serial); err == nil || out != nil || !bytes.Equal(before, d) {
			t.Fatal("損壞封裝未拒絕或改寫輸入")
		}
	}
}
