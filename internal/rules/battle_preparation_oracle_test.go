package rules

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 只驗共用行星補充；完整自動第一階段仍 DRAFT，不把孤立規則升格。
func TestPlanetReplenishMatchesPreparationOracle(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "docs/internals/data/battle-preparation.json"))
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(raw)
	if hex.EncodeToString(h[:]) != "39ee36b91de0cce822293b255fb83ad8fa13a6c3f37175c366f317ac5fa0c749" {
		t.Fatal("原版解出欄位身份不符，須重新審查證據")
	}
	type planet struct {
		Ships    [7]int `json:"ships"`
		Missiles int    `json:"missiles"`
		Fighters int    `json:"fighters"`
	}
	var receipt struct {
		Seed string `json:"seed"`
		Rows []struct {
			Name         string       `json:"name"`
			Before       []BattleUnit `json:"before"`
			After        []BattleUnit `json:"after"`
			PlanetBefore planet       `json:"planet_before"`
			PlanetAfter  planet       `json:"planet_after"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Seed != "0x12345678" {
		t.Fatal("原版事前種子不符")
	}
	n := 0
	for _, c := range receipt.Rows {
		if c.Name == "supply-radius-192" {
			t.Run("戰後只有彈藥變更", func(t *testing.T) {
				s := &Slot{Fleet: make([]byte, FleetCount*FleetRecord), Ship: make([]byte, 3*ShipRecord)}
				fleet := s.Fleet[16*FleetRecord : 17*FleetRecord]
				fleet[15] = 3
				for i, u := range c.Before {
					binary.LittleEndian.PutUint16(fleet[16+4*u.Squadron:], uint16(i))
					r := s.ShipAt(i)
					binary.LittleEndian.PutUint16(r[8:], uint16(u.Ships))
					binary.LittleEndian.PutUint16(r[12:], uint16(u.RepairCap))
					binary.LittleEndian.PutUint32(r[14:], uint32(u.Missiles))
					binary.LittleEndian.PutUint32(r[18:], uint32(u.Fighters))
				}
				if n := s.ApplyBattleResult(c.After); n != 0 {
					t.Fatalf("原版只變彈藥，艘數變更筆數應為0，得到%d", n)
				}
				for i, u := range c.After {
					r := s.ShipAt(i)
					if int(binary.LittleEndian.Uint16(r[8:])) != u.Ships || int(binary.LittleEndian.Uint16(r[12:])) != u.RepairCap ||
						int(binary.LittleEndian.Uint32(r[14:])) != u.Missiles || int(binary.LittleEndian.Uint32(r[18:])) != u.Fighters {
						t.Fatalf("單位%d未保存原版欄位：%x", i, r[8:22])
					}
				}
			})
		}
		if !strings.HasPrefix(c.Name, "planet-") {
			continue
		}
		n++
		t.Run(c.Name, func(t *testing.T) {
			if len(c.Before) != 3 || len(c.After) != 3 {
				t.Fatal("單位數不符")
			}
			units := append([]BattleUnit(nil), c.Before...)
			stock := PlanetStock{Ships: c.PlanetBefore.Ships, Missiles: c.PlanetBefore.Missiles, Fighters: c.PlanetBefore.Fighters}
			PlanetReplenish(&stock, &units[0])
			for i := range units {
				if units[i] != c.After[i] {
					t.Errorf("單位 %d：得到 %+v，原版 %+v", i, units[i], c.After[i])
				}
			}
			want := PlanetStock{Ships: c.PlanetAfter.Ships, Missiles: c.PlanetAfter.Missiles, Fighters: c.PlanetAfter.Fighters}
			if stock != want {
				t.Errorf("行星庫存 %+v，原版 %+v", stock, want)
			}
		})
	}
	if n != 5 {
		t.Fatalf("行星案例 %d，應有5案", n)
	}
}
