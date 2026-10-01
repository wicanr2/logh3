package rules

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type battleRuleFields struct {
	Ships     int `json:"ships"`
	FullShips int `json:"full_ships"`
	RepairCap int `json:"repair_cap"`
	Missiles  int `json:"missiles"`
	Fighters  int `json:"fighters"`
}

func (f battleRuleFields) unit(kind, squadron int) BattleUnit {
	return BattleUnit{Kind: kind, Squadron: squadron, Ships: f.Ships, FullShips: f.FullShips,
		RepairCap: f.RepairCap, Missiles: f.Missiles, Fighters: f.Fighters}
}
func battleFields(u BattleUnit) battleRuleFields {
	return battleRuleFields{u.Ships, u.FullShips, u.RepairCap, u.Missiles, u.Fighters}
}

// TestBattleRulesMatchesOracle 比對 dosgolem 直接執行原版的解出欄位，
// 包含能區分舊寫法的樣本；不宣稱完整戰場八步與副作用一致。
func TestBattleRulesMatchesOracle(t *testing.T) {
	var receipt struct {
		GolemHead  string `json:"golem_head"`
		ExeSHA     string `json:"exe_sha256"`
		Seed       string `json:"seed"`
		RemakeSeed string `json:"remake_seed"`
		Runs       []struct {
			Path string `json:"path"`
			SHA  string `json:"sha256"`
		} `json:"runs"`
		Table []int `json:"atan_table"`
		Rows  []struct {
			DX    int `json:"dx"`
			DY    int `json:"dy"`
			Angle int `json:"angle"`
		} `json:"rows"`
		Losses []struct {
			Kind   int              `json:"kind"`
			Flags  int              `json:"kind_flags"`
			Amount int              `json:"amount"`
			Return int              `json:"return"`
			Before battleRuleFields `json:"before"`
			After  battleRuleFields `json:"after"`
		} `json:"losses"`
		Spreads []struct {
			Amount int                `json:"amount"`
			Return int                `json:"return"`
			Before []battleRuleFields `json:"before"`
			After  []battleRuleFields `json:"after"`
		} `json:"spreads"`
	}
	b, err := os.ReadFile("../../docs/internals/data/battle-rule-corrections.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.GolemHead != "f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c" {
		t.Fatal("原版執行器身份不符")
	}
	if receipt.Seed != "0x12345678" || receipt.RemakeSeed != receipt.Seed {
		t.Fatal("受控種子不符")
	}
	// 這三支常式不接收 RNG；只核對事前條件，不能推論整段骰序。
	exe, err := os.ReadFile("../../org_game/extracted/GIN3.COM")
	if os.IsNotExist(err) {
		t.Skip("缺原版 GIN3.COM，跳過原版對照")
	}
	if err != nil {
		t.Fatal(err)
	}
	exeSum := sha256.Sum256(exe)
	if hex.EncodeToString(exeSum[:]) != receipt.ExeSHA {
		t.Fatal("原版執行檔雜湊不符")
	}
	if len(receipt.Runs) != 2 {
		t.Fatal("缺兩輪原版身份")
	}
	for _, run := range receipt.Runs {
		data, err := os.ReadFile(filepath.Join("../..", run.Path))
		if os.IsNotExist(err) {
			t.Skipf("缺固定原版收據 %s", run.Path)
		}
		if err != nil {
			t.Fatal(err)
		}
		s := sha256.Sum256(data)
		if hex.EncodeToString(s[:]) != run.SHA {
			t.Fatal("原版收據雜湊不符")
		}
	}
	if !reflect.DeepEqual(receipt.Table, atanTable[:]) {
		t.Error("64 個反正切原始表值不符")
	}
	if len(receipt.Rows) != 12 || len(receipt.Losses) != 5 || len(receipt.Spreads) != 2 {
		t.Fatal("原版對照樣本缺項")
	}
	for _, row := range receipt.Rows {
		t.Run(fmt.Sprintf("angle_%d_%d", row.DX, row.DY), func(t *testing.T) {
			for name, fn := range map[string]func(int, int) int{"共用": atan256, "公開": Atan256, "戰略": AtanTable256} {
				if got := fn(row.DX, row.DY); got != row.Angle {
					t.Errorf("%s=%d，原版=%d", name, got, row.Angle)
				}
			}
		})
	}
	for _, c := range receipt.Losses {
		t.Run(fmt.Sprintf("loss_%d_%d", c.Kind, c.Flags), func(t *testing.T) {
			u := c.Before.unit(c.Kind, 0)
			k := ShipKind{Kind: c.Kind, Words: make([]int, 13)}
			k.Words[kindMissileFlag/2] = c.Flags
			k.Words[kindFighterFlag/2] = c.Flags
			if got := ApplyShipLoss(&u, k, c.Amount); got != c.Return {
				t.Errorf("回值 %d，原版 %d", got, c.Return)
			}
			if got := battleFields(u); got != c.After {
				t.Errorf("欄位 %+v，原版 %+v", got, c.After)
			}
		})
	}
	kinds, err := LoadShipKindTable("../../docs/internals/data/ship-kind-table.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range receipt.Spreads {
		t.Run(fmt.Sprintf("spread_%d", c.Amount), func(t *testing.T) {
			units := make([]BattleUnit, len(c.Before))
			for i, f := range c.Before {
				units[i] = f.unit(0, i)
			}
			if got := spreadOverFleet(&BattleFleet{Index: 0}, units, kinds, c.Amount); got != c.Return {
				t.Errorf("實扣 %d，原版 %d", got, c.Return)
			}
			for i, u := range units {
				if got := battleFields(u); got != c.After[i] {
					t.Errorf("分隊 %d 欄位 %+v，原版 %+v", i, got, c.After[i])
				}
			}
		})
	}
}
