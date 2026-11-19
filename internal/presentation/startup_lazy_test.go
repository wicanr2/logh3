package presentation

import (
	"os"
	"slices"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
)

// 規格160：HD 載入不再預算投影地板、不解析黑體圖集；Warm 補算地板，
// 選黑體時才載入圖集，且仍列在可選字型裡。
func TestHDLoadDefersGridAndSansAtlas(t *testing.T) {
	b, err := os.ReadFile("../../org_game/extracted/FACE.AMG")
	if err != nil {
		t.Skip("缺原版 FACE.AMG，跳過（不用代用品）")
	}
	faces, err := asset.OpenAMG(b)
	if err != nil {
		t.Fatal(err)
	}
	// 版控的 assets/hd 沒有黑體圖集，發行包有；有發行素材就用它。
	dir := "../../assets/hd"
	if _, err := os.Stat("../../workplace/release172/extracted/squashfs-root/usr/bin/art/font-sans"); err == nil {
		dir = "../../workplace/release172/extracted/squashfs-root/usr/bin/art"
	}
	_, sansErr := os.Stat(dir + "/font-sans/MANIFEST.json")
	r, err := Load(dir, faces)
	if err != nil {
		t.Fatal(err)
	}
	if r.grid != nil {
		t.Fatal("載入時就算了投影地板")
	}
	if r.fonts["sans"] != nil {
		t.Fatal("載入時就解析了黑體圖集")
	}
	r.Warm()
	if r.grid == nil {
		t.Fatal("Warm 之後投影地板還沒算")
	}
	if sansErr != nil {
		t.Log("沒有黑體圖集，略過黑體那一段")
		return
	}
	if !slices.Contains(r.FontStyles(), "sans") {
		t.Fatal("黑體沒有列在可選字型")
	}
	if err := r.SetFontStyle("sans"); err != nil {
		t.Fatal(err)
	}
	if r.fonts["sans"] == nil || len(r.fonts["sans"].glyphs) == 0 || r.font != r.fonts["sans"] {
		t.Fatal("選黑體之後圖集沒有載入或沒有生效")
	}
	if err := r.SetFontStyle("classic"); err != nil {
		t.Fatal(err)
	}
}
