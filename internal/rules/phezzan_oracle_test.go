package rules

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestPhezzanEventMatchesPZX296Receipt 直接重播 `pzx296` 的本體＋懲罰。
// 原版素材與收據不進版控；本機缺檔時照專案契約 skip。
func TestPhezzanEventMatchesPZX296Receipt(t *testing.T) {
	prePath := filepath.Join("..", "..", "workplace", "oracle", "pzx296", "pzx296_228.bin")
	postPath := filepath.Join("..", "..", "workplace", "oracle", "pzx296", "pzx296_232.bin")
	plaPath := filepath.Join("..", "..", "org_game", "extracted", "PLADAT.DAT")
	pre, err := os.ReadFile(prePath)
	if err != nil {
		t.Skipf("沒有 pzx296 收據：%v", err)
	}
	post, err := os.ReadFile(postPath)
	if err != nil {
		t.Skipf("沒有 pzx296 收據：%v", err)
	}
	pla, err := os.ReadFile(plaPath)
	if err != nil {
		t.Skipf("沒有原版 PLADAT.DAT：%v", err)
	}
	planets, err := LoadPlanets(pla)
	if err != nil {
		t.Fatal(err)
	}
	const planetBytes = 81 * PlanetRecord
	const stellarAt = 0x5F2CC - 0x53132
	copy(planets.Data[PlanetHeader:], pre[:planetBytes])
	records := make([]byte, RegionCount*RegionRecord)
	for i := 0; i < RegionCount; i++ {
		records[i*RegionRecord+regionOwner] = RegionOwnerNone
	}
	copy(records, pre[stellarAt:])
	RunPhezzanEvent(records, planets, 0, true, 0, nil)
	if !bytes.Equal(planets.Data[PlanetHeader:], post[:planetBytes]) {
		n := 0
		for i, got := range planets.Data[PlanetHeader:] {
			if got != post[i] {
				t.Errorf("行星表 +%04X：remake %02X、原版 %02X", i, got, post[i])
				n++
				if n == 12 {
					break
				}
			}
		}
		t.FailNow()
	}
	// 原版包裝最後還重算 +0x7D；那支已有自己的收據，這裡只略過該欄。
	wantRecords := post[stellarAt:]
	for i := 0; i+RegionRecord <= len(wantRecords); i += RegionRecord {
		for j := 0; j < RegionRecord; j++ {
			if j == RegionUnrest {
				continue
			}
			if records[i+j] != wantRecords[i+j] {
				t.Fatalf("星域 %d +0x%02X：remake %02X，收據 %02X",
					i/RegionRecord, j, records[i+j], wantRecords[i+j])
			}
		}
	}
}
