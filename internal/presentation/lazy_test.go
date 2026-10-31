package presentation

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
)

func TestPortraitsDecodeOnDemandAndReleasePixels(t *testing.T) {
	faces, _ := asset.OpenAMG(makeAMG())
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "face"), 0755); err != nil {
		t.Fatal(err)
	}
	im := image.NewRGBA(image.Rect(0, 0, 64, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 64; x++ {
			im.SetRGBA(x, y, color.RGBA{uint8(x * 3), uint8(y * 11), 70, 255})
		}
	}
	path := filepath.Join(dir, "face/face-00.png")
	writePNG(t, path, im)
	r, err := Load(dir, faces)
	if err != nil {
		t.Fatal(err)
	}
	if _, n, b := r.ImageCacheStats(); n != 0 || b != 0 || len(r.images) != 0 {
		t.Fatal("啟動解碼了尚未使用的人物")
	}
	got := r.artworkImage("face/00")
	want := resize(im, 8*Scale, 2*Scale)
	if got == nil || !bytes.Equal(got.Pix, want.Pix) {
		t.Fatal("首用圖片像素改變")
	}
	if r.artworkImage("face/00") != got || r.artDecodes != 1 {
		t.Fatal("重複解碼")
	}
	if editor := r.artworkImage("editor-face/00"); !bytes.Equal(editor.Pix, resize(want, 64*Scale, 80*Scale).Pix) {
		t.Fatal("編輯器人物改變")
	}
	output := filepath.Join(t.TempDir(), "face")
	if err := WriteReleasePortraits(dir, output, faces); err != nil {
		t.Fatal(err)
	}
	packedRoot := filepath.Dir(output)
	packed, err := Load(packedRoot, faces)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(packed.artworkImage("face/00").Pix, want.Pix) {
		t.Fatal("預縮發行圖與原runtime像素不同")
	}
	if err := WriteReleasePortraits(dir, output, faces); err == nil {
		t.Fatal("允許覆寫既有發行圖")
	}
	// 登記後破損的圖片不使renderer崩潰，也不每幀重試。
	if err := os.WriteFile(path, []byte("broken after indexing"), 0644); err != nil {
		t.Fatal(err)
	}
	fresh, err := Load(packedRoot, faces)
	if err != nil {
		t.Fatal(err)
	}
	fresh.artSources["face/00"] = r.artSources["face/00"]
	if fresh.artworkImage("face/00") != nil || !fresh.failedArt["face/00"] {
		t.Fatal("破損未安全退回")
	}
	writePNG(t, path, im)
	if fresh.artworkImage("face/00") != nil {
		t.Fatal("錯誤圖片每幀重試")
	}
}

func TestArtworkCacheIsBoundedAndReloadsEvictedImages(t *testing.T) {
	path := filepath.Join(t.TempDir(), "image.png")
	writePNG(t, path, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	r := &Renderer{artSources: map[string]artworkSource{}}
	for i := 0; i < 9; i++ {
		key := string(rune('a' + i))
		r.artSources[key] = artworkSource{path: path, width: 1, height: 1, w: 2048, h: 1024}
		if r.artworkImage(key) == nil {
			t.Fatal("無法首次載入", key)
		}
		if r.artBytes > artworkCacheLimit {
			t.Fatal("快取超出64MiB")
		}
	}
	if r.images["a"] != nil || r.images["i"] == nil {
		t.Fatal("不是淘汰最久未用圖片")
	}
	if r.artworkImage("a") == nil || r.artDecodes != 10 || r.artBytes > artworkCacheLimit {
		t.Fatal("淘汰後不能重載")
	}
}
