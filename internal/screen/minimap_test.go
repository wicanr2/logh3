package screen

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

// 原版空白Input作底圖，只驗摘要疊圖；正常Game全圖另由第52輪收據驗。
func TestInputMinimapOverlayMatchesOriginal(t *testing.T) {
	root := "../.."
	base, err := os.ReadFile(filepath.Join(root, "workplace/oracle/normal89-r51-empty-a/screen.bin"))
	if err != nil {
		t.Skipf("缺原版空白Input：%v", err)
	}
	if len(base) != gfx.Width*gfx.Height {
		t.Fatal("原版畫面尺寸不符")
	}
	s, err := LoadBattleSpec(filepath.Join(root, "docs/internals/data/screen-battle.json"))
	if err != nil {
		t.Fatal(err)
	}
	hashes := map[string]string{
		"flagship":     "e87697e47c1558d21d81118b07b1b623a633ed6112a08fc6dd611b0833d82b65",
		"squad":        "9cb2bcfe0208a1fcbaaa6dece017cd16a23340b66db7c715b4a46c93936fa8a4",
		"fleet-target": "efa72e6c10c6ac92b6899f5d203576904aeda6e2eaf5f86ef06e7f62e5d016d9",
		"squad-target": "f602b9ee4f944364cab7819b4a0296fd2a8165fe827f345eaf0c230bf76c1d42",
	}
	for _, name := range []string{"flagship", "squad", "fleet-target", "squad-target"} {
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join(root, "workplace/oracle/normal89-r52-"+name+"-a")
			m, err := os.ReadFile(filepath.Join(dir, "end.bin"))
			if err != nil {
				t.Skipf("缺第52輪原版RAM：%v", err)
			}
			want, err := os.ReadFile(filepath.Join(dir, "screen.bin"))
			if err != nil {
				t.Skipf("缺第52輪原版畫面：%v", err)
			}
			if len(m) != 1048576 || len(want) != len(base) {
				t.Fatal("原版完整RAM／畫面不符")
			}
			hash := sha256.Sum256(m)
			if hex.EncodeToString(hash[:]) != hashes[name] {
				t.Fatal("原版RAM身份不符第52輪收據")
			}
			fleet := name == "flagship" || name == "fleet-target"
			pool := 82
			if fleet {
				pool = 80
			}
			o := 0x668B2 + pool*rules.ShipRecord
			word := func(p int) int { return int(binary.LittleEndian.Uint16(m[p:])) }
			current := [2]int{word(o + 0x34), word(o + 0x36)}
			target := [2]int{word(o + 0x28), word(o + 0x2A)}
			if fleet {
				o = 0x60BA4 + int(m[o+1])*rules.FleetRecord
				current = [2]int{word(o + 0x5C), word(o + 0x5E)}
				target = [2]int{word(o + 0x64), word(o + 0x66)}
			}
			fb := gfx.NewFramebuffer(0)
			copy(fb.Pix, base)
			DrawMinimapInputHighlight(s, fb, current, target, fleet)
			minimap := s.Minimap
			for y := minimap.Y; y < minimap.Y+minimap.H; y++ {
				for x := minimap.X; x < minimap.X+minimap.W; x++ {
					if fb.At(x, y) != want[y*gfx.Width+x] {
						t.Fatalf("原版小地圖(%d,%d)：remake=%d want=%d", x, y, fb.At(x, y), want[y*gfx.Width+x])
					}
				}
			}
			t.Log("原版小地圖128×128完整零差異；非正常Game全圖聲明")
		})
	}
}

// TestMinimapMatchesOracle 是小地圖的同狀態對拍。
func TestMinimapMatchesOracle(t *testing.T) {
	root := "../.."
	specPath := filepath.Join(root, "docs/internals/data/screen-battle.json")
	orig := filepath.Join(root, "org_game/extracted")
	oracle := filepath.Join(root, "workplace/oracle/s3fight/screen.bin")
	for _, p := range []string{specPath, oracle, filepath.Join(orig, "MAP26.CMP")} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("缺 %s，跳過同狀態對拍（不用代用品）", p)
		}
	}
	s, err := LoadBattleSpec(specPath)
	if err != nil {
		t.Fatal(err)
	}
	cmp, err := os.ReadFile(filepath.Join(orig, "MAP26.CMP"))
	if err != nil {
		t.Fatal(err)
	}
	cells, err := asset.DecodeCMP(cmp, s.Field.MapCols, s.Field.MapRows)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	fb := gfx.NewFramebuffer(0)
	DrawMinimap(s, fb, cells, BattleView{Col: 66, Row: 30})
	DrawMinimapPlanets(s, fb, rules.PlanetSites(cells, s.Field.TileW, s.Field.TileH))
	DrawMinimapUnits(s, fb, minimapOriginalUnits(t, "s3fight.bin"))
	m := &s.Minimap
	bad, firstX, firstY := 0, -1, -1
	for y := m.Y; y < m.Y+m.H; y++ {
		for x := m.X; x < m.X+m.W; x++ {
			if want[y*640+x] != fb.Pix[y*640+x] {
				if bad == 0 {
					firstX, firstY = x, y
				}
				bad++
			}
		}
	}
	total := m.W * m.H
	t.Logf("小地圖 %d×%d：不同 %d／%d（%.1f%%）", m.W, m.H, bad, total,
		100*float64(bad)/float64(total))
	// 地形 ＋ 格線 ＋ 行星 ＋ 分隊 ＋ 視野框五層都齊了，應該一個都不差。
	if bad > 0 {
		t.Fatalf("差 %d／%d 個像素（超過 0.2%%），第一個在 (%d,%d)：收據 %d、remake %d",
			bad, total, firstX, firstY, want[firstY*640+firstX], fb.Pix[firstY*640+firstX])
	}
}

// 直接讀原版+19／+1A／+1C，不把現在位置當標記快照（規格082）。
func minimapOriginalUnits(t *testing.T, name string) []rules.BattleUnit {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("../..", "workplace", name))
	if os.IsNotExist(err) {
		t.Skip("缺原版標記RAM，跳過，不用代用品")
	}
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"s3fight.bin": "0f8a91acb618d0ec508e857853be1cfe249ec8fe522c8aacac658ec03cbaf060",
		"s3ai6.bin":   "427c15414ce8d265551a05ecd2d49e8157ecede82a80154b6319c7ced4a8d3e0",
	}[name]
	h := sha256.Sum256(b)
	if want == "" || len(b) != 1<<20 || hex.EncodeToString(h[:]) != want {
		t.Fatal("原版標記RAM身份不同")
	}
	const ds = 0x54B30
	pool := int(binary.LittleEndian.Uint16(b[ds+0xCFB4:])) + 16*int(binary.LittleEndian.Uint16(b[ds+0xCFB6:]))
	// +02在這兩份原版記錄是0，不能作星域篩選。依DS:9898戰場艦隊名單取單位。
	count := int(binary.LittleEndian.Uint16(b[ds+0x9894:]))
	if count > 32 || pool < 0 || pool+352*rules.ShipRecord > len(b) {
		t.Fatal("原版標記表越界")
	}
	fleets := make(map[byte]bool, count)
	for _, fleet := range b[ds+0x9898 : ds+0x9898+count] {
		fleets[fleet] = true
	}
	var units []rules.BattleUnit
	for i := 0; i < 352; i++ {
		r := b[pool+i*rules.ShipRecord : pool+(i+1)*rules.ShipRecord]
		if !fleets[r[1]] {
			continue
		}
		units = append(units, rules.BattleUnit{Faction: int(r[0]), MarkerFlags: int(r[0x19]),
			MarkerX: int(binary.LittleEndian.Uint16(r[0x1A:])), MarkerY: int(binary.LittleEndian.Uint16(r[0x1C:]))})
	}
	return units
}
