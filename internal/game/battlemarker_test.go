package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 正常S3主選單／實行入口；不注入分隊、旗標或座標。
// 比較限於原池身份與標記，沒有外推整份進場記錄或畫面same-state。
func TestBattleMarkersNormalEntryAndSave(t *testing.T) {
	g := battleTestGame(t)
	if g == nil {
		return
	}
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/battle-markers.json"))
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(b)
	if hex.EncodeToString(h[:]) != "6f67e3c5107f249b70280b8115d795492c7c823648d7ed9ccb935f1d66873424" {
		t.Fatal("原版標記收據身份不同")
	}
	var oracle struct {
		Normal struct {
			Units []struct {
				Pool   int
				Record string `json:"record_hex"`
			}
		}
	}
	if err := json.Unmarshal(b, &oracle); err != nil {
		t.Fatal(err)
	}
	if len(oracle.Normal.Units) != 55 {
		t.Fatal("原版清單不是55筆")
	}
	drawn := 0
	markers := make(map[int][]byte)
	for _, entry := range oracle.Normal.Units {
		want, err := hex.DecodeString(entry.Record)
		if err != nil || len(want) != rules.ShipRecord {
			t.Fatal("原版記錄不完整")
		}
		got := g.slot.ShipAt(entry.Pool)
		if len(got) != rules.ShipRecord || got[0x19] != want[0x19] {
			t.Fatalf("原池%d標記旗標不同", entry.Pool)
		}
		markers[entry.Pool] = bytes.Clone(got[0x19:0x1E])
		if want[0x19]&0xC0 != 0xC0 {
			continue // 未畫快照保持底稿，不猜未知初值。
		}
		drawn++
		if !bytes.Equal(got[0x19:0x1E], want[0x19:0x1E]) {
			t.Fatalf("原池%d已畫標記快照不同：%X／%X", entry.Pool, got[0x19:0x1E], want[0x19:0x1E])
		}
		found := false
		for _, u := range g.battleUnits {
			if g.battleUnitPool(u) == entry.Pool {
				found = u.Marked && u.MarkerFlags == int(got[0x19]) && u.MarkerX == u.X && u.MarkerY == u.Y
			}
		}
		if !found {
			t.Fatalf("原池%d正式單位沒有相同標記", entry.Pool)
		}
	}
	if drawn != 8 {
		t.Fatalf("已畫%d筆，原版8筆", drawn)
	}
	// 第一次戰術實行之前保存，避免由移動回寫掩蓋進場標記未保存。
	dir := t.TempDir()
	if err := g.SaveSlot(dir, 0); err != nil {
		t.Fatal(err)
	}
	for pool, want := range markers {
		if !bytes.Equal(g.slot.ShipAt(pool)[0x19:0x1E], want) {
			t.Fatalf("存檔覆蓋原池%d標記", pool)
		}
	}
	ship, fleet := bytes.Clone(g.slot.Ship), bytes.Clone(g.slot.Fleet)
	loaded := newTestGame(t)
	loaded.res.SaveDir = dir
	loaded.loadSlotFromDisk(0)
	if loaded.SaveErr() != nil || loaded.slot == nil || !bytes.Equal(ship, loaded.slot.Ship) || !bytes.Equal(fleet, loaded.slot.Fleet) {
		t.Fatal("進場標記完整SHIP／FLEET正式存讀不同")
	}
	t.Log("正常入口55筆旗標與8筆已畫快照符合原版，首次實行前完整SHIP／FLEET存讀相同")
}
