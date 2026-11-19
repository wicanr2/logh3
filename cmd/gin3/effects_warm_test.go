package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/audio"
)

// 規格160：音效完整解碼移到背景。解好的 PCM 要與直接串流解出來的逐位元組相同；
// 有一段截斷時，背景記下錯誤，下一次 PlayEffect 回錯誤（updateEffects 據此關掉音效）。
func TestEffectWarmCachesPCMAndFlagsBroken(t *testing.T) {
	src := os.Getenv("LOGH3_SFX_DIR")
	if src == "" {
		src = "../../workplace/release172/extracted/squashfs-root/usr/bin/music"
	}
	if audio.NewSoundLibrary(src).Have() != 28 {
		t.Skip("缺28段音效OGG")
	}
	good := newEffectPlayer(nil, audio.NewSoundLibrary(src))
	good.warm()
	if good.broken != nil {
		t.Fatal(good.broken)
	}
	for id := 0; id < 28; id++ {
		st, ok, err := good.decoder.decode(audio.SoundTrack(id))
		if err != nil || !ok {
			t.Fatal(id, err)
		}
		want, _ := io.ReadAll(st)
		if !bytes.Equal(good.pcm[id], want) {
			t.Fatalf("音效%d的快取 PCM 與串流解碼不同", id)
		}
	}

	dir := t.TempDir()
	for id := 0; id < 28; id++ {
		name := audio.SoundTrack(id) + ".ogg"
		b, err := os.ReadFile(filepath.Join(src, name))
		if err != nil {
			t.Fatal(err)
		}
		if id == 5 {
			b = b[:len(b)/2] // 截掉後半
		}
		if err := os.WriteFile(filepath.Join(dir, name), b, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	bad := newEffectPlayer(nil, audio.NewSoundLibrary(dir))
	bad.warm()
	if bad.broken == nil {
		t.Fatal("截斷的音效沒有被背景驗證抓到")
	}
	if err := bad.PlayEffect(0); err == nil {
		t.Fatal("背景驗證失敗後 PlayEffect 該回錯誤")
	}
}
