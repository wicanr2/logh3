package game

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// #59：正常Game入口，僅在開清單前套用明示原版-poke的相同欄位。
// 這是受控same-state靜態畫面，不把局面注入當正常玩法完成。
func TestListTabEdgesMatchOracle(t *testing.T) {
	for _, c := range []struct {
		name, hash string
		ally       bool
		counts     []byte
		flags      byte
		morale     int
	}{
		{"empire-counts", "15030cbdb4a4a0fb36b8ca478dee48e65e5a4426a2a8b076bdb62c3c14f73d1b", false, []byte{9, 7, 6, 8}, 0, 0},
		{"empire-hidden", "89b8fe811c7b864e567ef4672ad6be3084cd9a3521a98a2b9a9aa6100e3120a8", false, []byte{255, 7, 6, 255}, 0, 0},
		{"ally-visible-counts", "357d39fa360c8556397cb6b3388ee360445c42b30685cb3d72fa0cb72e0d53af", true, []byte{9, 7, 6, 8}, 0, 0},
		{"ally-visible-hidden", "64ac0924c3985cf88709e392a39ec6f301902670fa6919d9534e2b96a88665cd", true, []byte{9, 255, 255, 8}, 0, 0},
		{"star-battle", "0413a8f541f8ab0966c8529ea6e3530ce17cb5ac21ec4822dc9eaf17b3c68c94", false, nil, 4, 0},
		{"morale49", "27c30c1b2c467aa5fc013dc420a49457c1e64fcd74e4db8e31e0b84dcc0fd11a", false, nil, 0, 49},
		{"morale50", "7ec2786405184f6e0e1b09653e9879a894f9b39f1d8753902f0c451818b68ae4", false, nil, 0, 50},
		{"battle", "9aea2a31eaa023f4775372ce172893d4a8a5dd7e113c50a1bee2a260aa38d4e3", false, nil, 0x14, 0},
		{"occupied", "9262128cb8f647338ad5f3946490aed483fd6c5570b5a3162d9cf648345fc3b2", false, nil, 0x10, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			want, err := os.ReadFile(filepath.Join(root, "workplace/oracle/list59-r59-"+c.name+"-b/screen.bin"))
			if os.IsNotExist(err) {
				t.Skip("缺第59輪原版收據；tools/list_tab_edge_oracle.sh重播")
			}
			if err != nil || len(want) != 640*480 || fmt.Sprintf("%x", sha256.Sum256(want)) != c.hash {
				t.Fatalf("原版收據漂移或無效：%v", err)
			}
			var g *Game
			region := 0
			if c.ally {
				g = reachCommandPhaseOnMapAlly(t)
				region = 31
			} else {
				g = reachFleetDecided(t)
			}
			if c.counts != nil {
				copy(g.regionRecord(region)[0x28:0x2C], c.counts)
			}
			if c.flags != 0 {
				g.regionRecord(region)[0x13] = c.flags
			}
			if c.morale != 0 {
				g.fleets[0].Morale = c.morale
			}
			g.Click(448, 125)
			if c.counts != nil || c.name == "star-battle" {
				g.Click(460, 125)
			} else {
				g.Click(460, 145)
				if c.flags != 0 {
					g.Click(584, 70)
				}
			}
			fb := g.Frame()
			if out := os.Getenv("LOGH3_LIST_TAB_EDGE_OUT"); out != "" {
				out = filepath.Join(out, c.name)
				if err := os.Mkdir(out, 0755); err != nil {
					t.Fatal(err)
				}
				if err := fb.WriteIndex(filepath.Join(out, "screen.bin")); err != nil {
					t.Fatal(err)
				}
				if err := fb.WritePNG(filepath.Join(out, "screen.png"), g.res.Menu.Palette()); err != nil {
					t.Fatal(err)
				}
				if err := g.SaveSlot(out, 0); err != nil {
					t.Fatal(err)
				}
			}
			if !bytes.Equal(fb.Pix, want) {
				n := 0
				for i := range want {
					if fb.Pix[i] != want[i] {
						n++
					}
				}
				t.Fatalf("完整畫面差%d／307200像素", n)
			}
			if c.counts != nil {
				if err := g.SaveSlot(t.TempDir(), 0); err != nil {
					t.Fatal(err)
				}
				wantCounts := append([]byte(nil), c.counts[2:]...)
				got := g.slot.Stellar[region*rules.RegionRecord+0x2A : region*rules.RegionRecord+0x2C]
				if !bytes.Equal(got, wantCounts) {
					t.Fatalf("保存改寫情報兩格：%x，原有%x", got, wantCounts)
				}
				g.LoadSlot(g.slot)
				// 091補證：正式讀檔重算可見數。本輪艦艇Detail都是0。
				if !bytes.Equal(g.regionRecord(region)[0x2A:0x2C], []byte{0, 0}) {
					t.Fatal("讀檔沒有重算過期情報兩格")
				}
			}
		})
	}
}

func TestListTabNormalPlayer(t *testing.T) {
	g := reachCommandPhaseOnMapAlly(t)
	g.Click(448, 125)
	g.Click(460, 125)
	wantState(t, g, StateStarList, "正常一覽表")
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle/list59-r59-ally-normal-b/screen.bin"))
	if os.IsNotExist(err) {
		t.Skip("缺第59輪正常同盟收據；tools/list_tab_edge_oracle.sh ally-normal-b")
	}
	if err != nil || len(want) != 307200 || fmt.Sprintf("%x", sha256.Sum256(want)) != "ebf406cea30ff5c9f8bc8f470386d39b76b7e2e83d81e59df83dc927923c5326" {
		t.Fatalf("正常入口收據無效或漂移：%v", err)
	}
	if !bytes.Equal(g.Frame().Pix, want) {
		t.Fatal("正常同盟入口與原版全圖不符")
	}
	if out := os.Getenv("LOGH3_LIST_TAB_EDGE_OUT"); out != "" {
		out = filepath.Join(out, "normal-ally")
		if err := os.Mkdir(out, 0755); err != nil {
			t.Fatal(err)
		}
		fb := g.Frame()
		if err := fb.WriteIndex(filepath.Join(out, "screen.bin")); err != nil {
			t.Fatal(err)
		}
		if err := fb.WritePNG(filepath.Join(out, "screen.png"), g.res.Menu.Palette()); err != nil {
			t.Fatal(err)
		}
	}
	g.RightClick(300, 240)
	wantState(t, g, StateCommandPhase, "正常清單返回")
}
