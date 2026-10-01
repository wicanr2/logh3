package asset

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestMDSMIDIOriginalExportParity(t *testing.T) {
	files, err := filepath.Glob("/orig/*.MDS")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Skip("缺原版 MDS，不製造替代品")
	}
	if len(files) != 22 {
		t.Fatalf("原版曲數 %d，預期 22", len(files))
	}
	for _, p := range files {
		t.Run(filepath.Base(p), func(t *testing.T) {
			d, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			got, err := MDSMIDI(d)
			if err != nil {
				t.Fatal(err)
			}
			name := filepath.Base(p)
			want, err := os.ReadFile(filepath.Join("../../assets/original/music/mid", name[:len(name)-4]+".mid"))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("MIDI 移植差異：原生 %d bytes，既有 %d bytes", len(got), len(want))
			}
		})
	}
}

func TestMDSMIDIRejectsCorruption(t *testing.T) {
	d := make([]byte, 260)
	copy(d, "M1")
	copy(d[256:], []byte{60, 48, 24, 100})
	if _, err := MDSMIDI(d); err != nil {
		t.Fatal(err)
	}
	for _, bad := range [][]byte{nil, d[:258], append([]byte("ZZ"), d[2:]...), append(d, 0xFE, 0, 0, 0), append(d[:256:256], 0xFC, 1, 0, 0), make([]byte, 1<<20+4)} {
		if _, err := MDSMIDI(bad); err == nil {
			t.Fatal("損壞或無音符的 MDS 未拒絕")
		}
	}
}
