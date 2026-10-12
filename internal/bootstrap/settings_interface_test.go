package bootstrap

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/gfx"
)

// 113：正常bootstrap及公開玩家輸入；出貨BUG!不可選，畫面核固定原版。
func TestSettingsBugUnavailablePlayerPath(t *testing.T) {
	res, g := titlePlayerResources(t)
	g.Click(300, 210) // 正常跳過片頭。
	g.LeftRelease()
	g.Click(300, 210) // 帝國。
	for _, c := range []byte{'1', '3', '8', '\r'} {
		g.Key(c)
	}
	g.RightClick(300, 240)
	g.RightClick(300, 240)
	g.Click(544, 125)
	g.Click(560, 185)
	if g.State() != game.StateSettings {
		t.Fatalf("正常入口沒有到設定：%v", g.State())
	}
	g.MoveMouse(600, 340)
	out := os.Getenv("LOGH3_SETTINGS_RECEIPTS")
	if out != "" {
		if !filepath.IsAbs(out) {
			t.Fatal("收據目錄必須是絕對路徑")
		}
		if err := os.Mkdir(out, 0755); err != nil {
			t.Fatal(err)
		}
	}
	digest := func(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
	var records []map[string]any
	record := func(name, want string) {
		t.Helper()
		fb, pal := g.Frame(), g.Palette()
		// 134新增展示控制，只允許其已核矩形與原版不同。此區原為
		// 設定窗的空白底色，還原後仍核原完整SHA，不更新原版golden。
		original := bytes.Clone(fb.Pix)
		r := game.EnemyVisibilityButtonBounds()
		fill := res.Screens["screen-settings"].Windows["overlay"].Frame.Fill
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				original[y*gfx.Width+x] = fill
			}
		}
		if got := digest(original); got != want {
			t.Fatalf("%s新增控制外原色SHA=%s，原版=%s", name, got, want)
		}
		var rgb []byte
		for i := 0; i < gfx.Colors; i++ {
			r, gg, b := pal.RGB8(i)
			rgb = append(rgb, r, gg, b)
		}
		records = append(records, map[string]any{"name": name, "state": g.State().String(),
			"pixels_sha256": digest(fb.Pix), "rgb16_sha256": digest(rgb)})
		if out != "" {
			for _, err := range []error{
				fb.WritePNG(filepath.Join(out, name+".png"), pal),
				fb.WriteIndex(filepath.Join(out, name+".bin")),
				os.WriteFile(filepath.Join(out, name+"-rgb16.bin"), rgb, 0644),
			} {
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	record("settings", "4b2ee8cc8d071d58fe466c2ca95b72aad079cfaf96a04fb5edeee2cf369622d6")
	before, after := t.TempDir(), t.TempDir()
	if err := g.SaveSlot(before, 0); err != nil {
		t.Fatal(err)
	}
	g.Click(410, 65)
	if g.State() != game.StateSettings {
		t.Fatalf("不可選BUG!改了狀態：%v", g.State())
	}
	record("bug", "48fff4b4c0bcc070710e76b6d36c9d7118b2683e6ae5c3dfe2c0c43805feeb5d")
	if err := g.SaveSlot(after, 0); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(before)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		a, err := os.ReadFile(filepath.Join(before, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(filepath.Join(after, e.Name()))
		if err != nil || !bytes.Equal(a, b) {
			t.Fatalf("BUG!改了保存檔%s：%v", e.Name(), err)
		}
	}
	if out != "" {
		b, err := json.MarshalIndent(records, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, "records.json"), b, 0644); err != nil {
			t.Fatal(err)
		}
	}
}
