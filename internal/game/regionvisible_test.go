package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

func TestRegionVisibleFleetNormalCommand(t *testing.T) {
	g := reachFleetDecided(t)
	if g.rand94DC().Seed() != 0xD3DC167E {
		t.Fatal("正常回合的初始亂數種子不同")
	}
	executeTurn(t, g)
	wantState(t, g, StateCommandPhase, "正常一回合")
	if g.turn != 1 {
		t.Fatal("正常回合計數不同")
	}
	for r := 0; r < rules.RegionCount; r++ {
		if !bytes.Equal(g.regionRecord(r)[0x2A:0x2C], []byte{0, 0}) {
			t.Fatalf("正常星域%d可見數不同", r)
		}
	}
	g.Click(448, 125)
	g.Click(460, 125)
	wantState(t, g, StateStarList, "正常一回合後一覽表")
	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle/intel59-r60-normal-b/screen.bin"))
	if os.IsNotExist(err) {
		t.Skip("缺第60輪正常原版收據；tools/region_visible_fleets_oracle.sh normal-b")
	}
	if err != nil || len(want) != 307200 || fmt.Sprintf("%x", sha256.Sum256(want)) != "257b2e6ffdd365d35055888838cb14ebbccb4f461eca6421ab53f139b7fee265" {
		t.Fatalf("正常原版畫面漂移或無效：%v", err)
	}
	if out := os.Getenv("LOGH3_REGION_VISIBLE_OUT"); out != "" {
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
		if err := g.SaveSlot(out, 0); err != nil {
			t.Fatal(err)
		}
	}
	if !bytes.Equal(g.Frame().Pix, want) {
		n := 0
		for i, b := range g.Frame().Pix {
			if b != want[i] {
				n++
			}
		}
		t.Fatalf("正常清單全圖差%d／307200", n)
	}
}

// 正常Game建隊後明示修改索敵／陣營，以驗證正式保存與載入資料鏈。
// 非正常局面或全圖對拍聲明；原版導出值另由rules的固定收據核對。
func TestRegionVisibleFleetSaveLoad(t *testing.T) {
	for _, faction := range []int{0, 1} {
		t.Run(fmt.Sprintf("faction%d", faction), func(t *testing.T) {
			g := reachFleetDecided(t)
			g.fleets[0].Faction = faction
			g.fleets[0].Composition[0].Detail = 255
			copy(g.regionRecord(0)[0x2A:0x2C], []byte{254, 255})
			// 開關清單和實數重算不能刷新這份快照。
			g.recountRegionFleets()
			g.Click(448, 125)
			g.Click(460, 125)
			g.RightClick(300, 240)
			if !bytes.Equal(g.regionRecord(0)[0x2A:0x2C], []byte{254, 255}) {
				t.Fatal("在玩家入口前提前刷新可見數")
			}
			if err := g.SaveSlot(t.TempDir(), 0); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(g.slot.Stellar[0x2A:0x2C], []byte{254, 255}) {
				t.Fatal("保存改寫快照")
			}
			g.LoadSlot(g.slot)
			want := []byte{0, 0}
			want[faction] = 1
			if !bytes.Equal(g.regionRecord(0)[0x2A:0x2C], want) || g.fleets[0].Composition[0].Detail != 255 {
				t.Fatal("正式讀檔沒有依保存的索敵值重算")
			}
			copy(g.regionRecord(0)[0x2A:0x2C], []byte{254, 255})
			g.versus = true
			g.enterSide(1)
			if !bytes.Equal(g.regionRecord(0)[0x2A:0x2C], want) {
				t.Fatal("雙人換邊沒有重算")
			}
		})
	}
}

// 受控舊值在無戰鬥星域清理後，經正常命令入口導出；不宣稱全RAM相同。
func TestRegionVisibleFleetPeacefulClear(t *testing.T) {
	g := reachFleetDecided(t)
	g.fleets[0].Composition[0].Detail = 1
	for r := 0; r < rules.RegionCount; r++ {
		copy(g.regionRecord(r)[0x2A:0x2C], []byte{254, 255})
	}
	executeTurn(t, g)
	wantState(t, g, StateCommandPhase, "受控索敵清零後返回命令階段")
	if g.fleets[0].Composition[0].Detail != 0 {
		t.Fatal("無戰鬥星域沒有清零索敵值")
	}
	for r := 0; r < rules.RegionCount; r++ {
		if !bytes.Equal(g.regionRecord(r)[0x2A:0x2C], []byte{0, 0}) {
			t.Fatalf("星域%d保留過期可見數", r)
		}
	}
}

func TestRegionVisibleFleetClearMatchesOracle(t *testing.T) {
	read := func(name, hash string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(root, "workplace/oracle/intel59-r60-peace-b", name+".bin"))
		if os.IsNotExist(err) {
			t.Skip("缺第60輪清理收據；tools/region_visible_fleets_oracle.sh peace-b")
		}
		if err != nil || len(b) != 1048576 || fmt.Sprintf("%x", sha256.Sum256(b)) != hash {
			t.Fatalf("原版清理RAM無效或漂移：%v", err)
		}
		return b
	}
	before := read("clear-input", "8efb45626b8679599575e5a036be5b37288cf25c6ddf4d1b0f923a9eed4df677")
	after := read("clear-output", "7fe188f7724812b41ad0c1155a350b7b1b84946462572e5e159db99ca0a7a74c")
	pool := int(binary.LittleEndian.Uint16(before[0x61AE4:])) + 16*int(binary.LittleEndian.Uint16(before[0x61AE6:]))
	s := &rules.Slot{Fleet: bytes.Clone(before[0x60BA4 : 0x60BA4+rules.FleetCount*rules.FleetRecord]),
		Ship: bytes.Clone(before[pool : pool+rules.ShipCount*rules.ShipRecord])}
	g := &Game{slot: s, fleets: s.Fleets(nil)}
	g.clearRegionFleetVisibility(0)
	if !bytes.Equal(s.Fleet, after[0x60BA4:0x60BA4+rules.FleetCount*rules.FleetRecord]) ||
		!bytes.Equal(s.Ship, after[pool:pool+rules.ShipCount*rules.ShipRecord]) {
		t.Fatal("清理後完整艦隊／艦艇表不符原版")
	}
	if g.fleets[0].Action != 4 || g.fleets[0].Composition[0].Detail != 0 {
		t.Fatal("typed資料與原始表不同步")
	}
}
