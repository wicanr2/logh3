package rules

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fleetStepReceipt struct {
	Receipt struct {
		Plan struct {
			Common []stepInjection `json:"common"`
			Cases  []struct {
				Injections []stepInjection `json:"injections"`
			} `json:"cases"`
		} `json:"plan"`
		Rows []struct {
			Name    string            `json:"name"`
			Before  map[string]string `json:"before"`
			After   map[string]string `json:"after"`
			Returns []uint32          `json:"returns"`
			Calls   []struct {
				Entry string `json:"entry"`
				Args  []int  `json:"args"`
			} `json:"calls"`
			Seed string `json:"seed_after_hex"`
		} `json:"rows"`
	} `json:"receipt"`
}

type stepInjection struct {
	Linear int    `json:"linear"`
	Hex    string `json:"hex"`
}

func loadFleetStepReceipt(t *testing.T) fleetStepReceipt {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "docs", "internals", "data", "battle-step.json"))
	if err != nil {
		t.Fatal(err)
	}
	var result fleetStepReceipt
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Receipt.Rows) != 102 {
		t.Fatal("原版102案收據不完整")
	}
	return result
}

func stepBytes(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func stepWord(b []byte, off int) int   { return int(binary.LittleEndian.Uint16(b[off:])) }
func stepBudget(b []byte, off int) int { return int(int32(binary.LittleEndian.Uint32(b[off:]))) }

// 規格081：先保存原版返回值與額度邊界，舊AdvanceFleet應失敗。
func TestFleetSingleStepMatchesOriginal(t *testing.T) {
	fixture := loadFleetStepReceipt(t)
	terrain, err := LoadBattleTerrain(filepath.Join("..", "..", "docs/internals/data/battle-terrain.json"))
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for index, row := range fixture.Receipt.Rows {
		if !strings.HasPrefix(row.Name, "fleet-") {
			continue
		}
		count++
		t.Run(row.Name, func(t *testing.T) {
			before, after := stepBytes(t, row.Before["fleet"]), stepBytes(t, row.After["fleet"])
			f := FleetMover{X: stepWord(before, 0x5C), Y: stepWord(before, 0x5E), DestX: stepWord(before, 0x64), DestY: stepWord(before, 0x66), Facing: int(before[0x4A]), TargetFacing: int(before[0x4C]), TurnSpeed: int(before[0x4D]), Budget: stepBudget(before, 0x56), Action: int(before[0x48]), MoveFlags: int(before[0x49])}
			units := []*BattleUnit{stepUnit(t, row.Before["flagship"]), stepUnit(t, row.Before["unit"])}
			cells, enemyCount := stepMap(t, fixture, index)
			var occupants []MovementOccupant
			if enemyCount == 1 {
				occupants = []MovementOccupant{{Pool: 142, Unit: stepUnit(t, row.Before["enemy"])}}
			}
			var points [][2]int
			check := func(x, y int) int {
				points = append(points, [2]int{x, y})
				return BattleMovementCheck(terrain, cells, occupants, int(before[1]), x, y)
			}
			processed := AdvanceFleetStep(&f, units, check)
			if processed != (row.Returns[0]&255 != 0) || f.X != stepWord(after, 0x5C) || f.Y != stepWord(after, 0x5E) || f.Facing != int(after[0x4A]) || f.Budget != stepBudget(after, 0x56) {
				t.Fatalf("原版：AL=%d、(%d,%d)、朝向%d、額度%d；remake：處理%v、(%d,%d)、朝向%d、額度%d", row.Returns[0]&255, stepWord(after, 0x5C), stepWord(after, 0x5E), after[0x4A], stepBudget(after, 0x56), processed, f.X, f.Y, f.Facing, f.Budget)
			}
			actual := bytes.Clone(before)
			actual[0x49], actual[0x4A] = byte(f.MoveFlags), byte(f.Facing)
			binary.LittleEndian.PutUint32(actual[0x56:], uint32(f.Budget))
			binary.LittleEndian.PutUint16(actual[0x5C:], uint16(f.X))
			binary.LittleEndian.PutUint16(actual[0x5E:], uint16(f.Y))
			if !bytes.Equal(actual, after) {
				t.Fatalf("完整122-byte艦隊不同：\n原版%x\nremake%x", after, actual)
			}
			for i, label := range []string{"flagship", "unit"} {
				want := stepBytes(t, row.After[label])
				if units[i].Attack != stepWord(want, 0x16) {
					t.Fatalf("%s追隊旗標不同", label)
				}
			}
			var expected [][2]int
			for _, call := range row.Calls {
				if call.Entry == "2583:040C" {
					expected = append(expected, [2]int{call.Args[0], call.Args[1]})
				}
			}
			if len(points) != len(expected) {
				t.Fatalf("通行候選數%d，要%d", len(points), len(expected))
			}
			for i := range points {
				if points[i] != expected[i] {
					t.Fatalf("候選%d=%v，要%v", i, points[i], expected[i])
				}
			}
			if row.Seed != "78563412" {
				t.Fatal("原版種子不符")
			}
		})
	}
	if count != 38 {
		t.Fatalf("艦隊單步案例%d，要38", count)
	}
}

func stepUnit(t *testing.T, raw string) *BattleUnit {
	b := stepBytes(t, raw)
	return &BattleUnit{Faction: int(b[0]), Kind: int(b[4]), Attack: stepWord(b, 0x16), Action: int(b[0x1E]), X: stepWord(b, 0x34), Y: stepWord(b, 0x36), TargetX: stepWord(b, 0x28), TargetY: stepWord(b, 0x2A)}
}

func stepMap(t *testing.T, fixture fleetStepReceipt, index int) ([]uint8, int) {
	t.Helper()
	cells := make([]uint8, 128*102)
	enemyCount := 0
	all := append(append([]stepInjection(nil), fixture.Receipt.Plan.Common...), fixture.Receipt.Plan.Cases[index].Injections...)
	for _, v := range all {
		data := stepBytes(t, v.Hex)
		if v.Linear >= 0x6D3F4 && v.Linear+len(data) <= 0x6D3F4+len(cells) {
			copy(cells[v.Linear-0x6D3F4:], data)
		}
		if v.Linear == 0x54B30+0x9896 {
			enemyCount = stepWord(data, 0)
		}
	}
	if enemyCount < 0 || enemyCount > 1 {
		t.Fatal("此受控局面只定義0或1筆敵佔位")
	}
	return cells, enemyCount
}

func TestBattleMovementCollisionMatchesOriginal(t *testing.T) {
	fixture := loadFleetStepReceipt(t)
	count := 0
	for _, row := range fixture.Receipt.Rows {
		if !strings.HasPrefix(row.Name, "collision-") {
			continue
		}
		count++
		t.Run(row.Name, func(t *testing.T) {
			got := BattleMovementCheck(nil, nil, []MovementOccupant{{Pool: 142, Unit: stepUnit(t, row.Before["enemy"])}}, 1, 1000, 1000)
			if got != int(row.Returns[0]&65535) {
				t.Fatalf("碰撞%d，要原版AX=%d", got, row.Returns[0]&65535)
			}
		})
	}
	if count != 9 {
		t.Fatalf("碰撞案例%d，要9", count)
	}
}
