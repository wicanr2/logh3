package rules

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestPruneEmptyBattleGroupsOriginal(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "workplace", "cleanup79-r88-a.json"))
	if os.IsNotExist(err) {
		t.Skip("缺原版清理收據；沿tools/battle_cleanup_oracle.sh重生")
	}
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != "80f108f6db41137c245dd52040616234779cf93349fbd7c0fae87d1aa2ef9949" {
		t.Fatal("原版清理收據不符", err)
	}
	var d struct {
		Rows []struct {
			Name          string
			Fleet         int
			Before, After map[string]string
		}
	}
	if err = json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	decode := func(s string) []byte {
		b, e := hex.DecodeString(s)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	removed := 0
	for _, r := range d.Rows {
		t.Run(r.Name, func(t *testing.T) {
			s := &Slot{Fleet: decode(r.Before["fleet"]), Ship: decode(r.Before["ship"])}
			mapping, e := s.PruneEmptyBattleGroups(r.Fleet)
			if e != nil {
				t.Fatal(e)
			}
			for _, i := range mapping {
				if i < 0 {
					removed++
				}
			}
			if !bytes.Equal(s.Fleet, decode(r.After["fleet"])) || !bytes.Equal(s.Ship, decode(r.After["ship"])) {
				t.Fatal("原池完整表不同")
			}
		})
	}
	if len(d.Rows) != 35 || removed != 77 {
		t.Fatal("案例或移除數不足")
	}
	r := d.Rows[1]
	for _, bad := range []string{"zero count", "excess count", "empty flagship", "invalid pool"} {
		t.Run(bad, func(t *testing.T) {
			s := &Slot{Fleet: decode(r.Before["fleet"]), Ship: decode(r.Before["ship"])}
			f := s.Fleet[r.Fleet*FleetRecord:]
			switch bad {
			case "zero count":
				f[15] = 0
			case "excess count":
				f[15] = 11
			case "empty flagship":
				pool := int(binary.LittleEndian.Uint16(f[16:]))
				clear(s.Ship[pool*78+8 : pool*78+10])
			case "invalid pool":
				binary.LittleEndian.PutUint16(f[20:], 65535)
			}
			fb, sb := bytes.Clone(s.Fleet), bytes.Clone(s.Ship)
			if _, e := s.PruneEmptyBattleGroups(r.Fleet); e == nil {
				t.Fatal("非法資料未回錯")
			}
			if !bytes.Equal(fb, s.Fleet) || !bytes.Equal(sb, s.Ship) {
				t.Fatal("非法資料仍部分寫入")
			}
		})
	}
}
