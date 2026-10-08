package rules

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"testing"
)

// TestStrategicAttackAgainstOriginalVectors 用原版 203A:0424 產生的向量（decomp/）驗證
// strategicAttack。向量檔由 decomp/tools/fleet_attack_run.sh 產生，沒有環境變數就 skip，
// 不以自製資料代替（CLAUDE.md「缺檔就 skip」）。
//
// 原版有 16 位元回繞（艘數合計）與截斷（第一次除法的商只留低 16 位元），
// remake 用 int，兩者只可能在 溢位區 不同：
//
//	艘數合計 >= 65536，或 艘數合計 × 能力 ÷ 100 >= 65536。
//
// 測試要求溢位區以外一筆都不能不符，並把溢位區內的差異數量記進 log。
// 溢位區是否可達（艘數是否真能累到這麼大）另案處理，見 docs/spec/124。
func TestStrategicAttackAgainstOriginalVectors(t *testing.T) {
	path := os.Getenv("LOGH3_DECOMP_FLEET_ATTACK_VECTORS")
	if path == "" {
		t.Skip("未設 LOGH3_DECOMP_FLEET_ATTACK_VECTORS；向量檔由 decomp/tools/fleet_attack_run.sh 產生")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	var total, match, inZone, inZoneDiff int
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "C ") {
			continue
		}
		fs := strings.Fields(line)
		id := fs[1]
		fleet, err := hex.DecodeString(fs[2])
		if err != nil || len(fleet) != 122 {
			t.Fatalf("case %s 艦隊記錄格式錯誤", id)
		}
		npool, _ := strconv.Atoi(fs[3])
		pool := map[int][]byte{}
		for _, tok := range fs[4 : 4+npool] {
			kv := strings.SplitN(tok, ":", 2)
			idx, _ := strconv.Atoi(kv[0])
			rec, err := hex.DecodeString(kv[1])
			if err != nil || len(rec) != 78 {
				t.Fatalf("case %s 艦艇記錄格式錯誤", id)
			}
			pool[idx] = rec
		}
		ax64, _ := strconv.ParseUint(fs[4+npool], 16, 16)

		sf := StrategicFleet{Fleet: &Fleet{}, Ability40: int(fleet[0x40])}
		trueMain := 0
		for i := 0; i < int(fleet[0x0F]); i++ {
			idx := int(binary.LittleEndian.Uint16(fleet[0x10+i*4:]))
			rec := pool[idx]
			g := FleetGroup{Kind: int(rec[4]), Ships: int(binary.LittleEndian.Uint16(rec[8:]))}
			sf.Fleet.Composition = append(sf.Fleet.Composition, g)
			if g.Kind == 0 || g.Kind == 1 {
				trueMain += g.Ships
			}
		}

		total++
		zone := trueMain >= 65536 || trueMain*sf.Ability40/100 >= 65536
		if zone {
			inZone++
		}
		got := strategicAttack(sf)
		if got == int(ax64) {
			match++
			continue
		}
		if !zone {
			t.Errorf("case %s 溢位區以外不符：remake=%d 原版=%d 艘數合計=%d 能力=%d", id, got, ax64, trueMain, sf.Ability40)
			continue
		}
		inZoneDiff++
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if total == 0 {
		t.Fatal("向量檔沒有任何案")
	}
	t.Logf("案數 %d，相符 %d，溢位區案數 %d（其中不符 %d）", total, match, inZone, inZoneDiff)
}
