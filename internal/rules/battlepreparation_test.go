package rules

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func readPreparationReceipt(t *testing.T, name, hash string, out any) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("../..", "docs/internals/data", name))
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(b)
	if hex.EncodeToString(h[:]) != hash {
		t.Fatal("原版欄位收據身份不符，須重新審查")
	}
	if err := json.Unmarshal(b, out); err != nil {
		t.Fatal(err)
	}
}

func TestPreparationSortMatchesOriginal(t *testing.T) {
	var r struct {
		Cases []struct {
			Entry         string                                 `json:"entry"`
			Name          string                                 `json:"name"`
			Donor         int                                    `json:"donor_fleet"`
			Units         []struct{ Index, Fleet, Squadron int } `json:"units"`
			Before, After []int
		}
	}
	readPreparationReceipt(t, "battle-preparation-sort.json", "9df9ec35eb56af3e9445cff7be6359b4107803a2528fcfe2d922ccd0a7c8292f", &r)
	n := 0
	for _, c := range r.Cases {
		if c.Entry != "3C6A:000D" {
			continue
		}
		n++
		t.Run(c.Name, func(t *testing.T) {
			units := map[int]*BattleUnit{}
			for _, u := range c.Units {
				units[u.Index] = &BattleUnit{Fleet: u.Fleet, Squadron: u.Squadron}
			}
			list := slices.Clone(c.Before)
			SortPreparationCandidates(list, func(i int) *BattleUnit { return units[i] }, func() int { return c.Donor })
			if !slices.Equal(list, c.After) {
				t.Fatalf("得到%v，原版%v", list, c.After)
			}
		})
	}
	if n != 72 {
		t.Fatalf("排序案例%d，應有72", n)
	}
}

func TestPreparationResourcesMatchesOriginal(t *testing.T) {
	type planet struct {
		Defence                         int `json:"defence"`
		DefenceMax                      int `json:"defence_max"`
		Recover                         int `json:"recover"`
		Owner, X, Y, Missiles, Fighters int
		Ships                           [7]int
	}
	var r struct {
		Rows []struct {
			Name, Entry   string
			Before, After []BattleUnit
			PlanetBefore  planet `json:"planet_before"`
			PlanetAfter   planet `json:"planet_after"`
		}
	}
	readPreparationReceipt(t, "battle-preparation.json", "39ee36b91de0cce822293b255fb83ad8fa13a6c3f37175c366f317ac5fa0c749", &r)
	kinds, err := LoadShipKindTable(filepath.Join("../..", "docs/internals/data/ship-kind-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range r.Rows {
		t.Run(c.Name, func(t *testing.T) {
			units := slices.Clone(c.Before)
			refs := make([]PreparationUnit, len(units))
			for i := range units {
				// 歷史JSON將原版+34/+36命名target；此處明示轉為現在位置。
				units[i].X, units[i].Y = units[i].TargetX, units[i].TargetY
				refs[i] = PreparationUnit{Pool: i, Unit: &units[i]}
			}
			p := c.PlanetBefore
			planets := []BattlePlanet{{Owner: p.Owner, X: p.X, Y: p.Y, Defence: p.Defence, DefenceMax: p.DefenceMax, Recover: p.Recover,
				Stock: PlanetStock{Ships: p.Ships, Missiles: p.Missiles, Fighters: p.Fighters}}}
			if c.Entry == "181B:0083" {
				RecoverBattleDefence(planets)
			} else {
				prep := NewBattlePreparation(kinds, refs, planets, &PreparationSources{}, func(i int) int { return units[i].Fleet })
				// 原版probe的ds:9896：邊界案只列前2筆，行星案只列第0筆。
				// 歷史欄位JSON仍保存3筆池記錄，不代表3筆都在戰場清單。
				if strings.Contains(c.Name, "-radius-") {
					prep.units = refs[:2]
				}
				if strings.HasPrefix(c.Name, "planet-") {
					prep.units = refs[:1]
				}
				if c.Entry == "1AAB:006D" {
					prep.phase = 2
				} // 原版此案只執行修理常式。
				for {
					_, ok := prep.Next()
					if !ok {
						break
					}
				}
			}
			for i, u := range units {
				u.X, u.Y = 0, 0 // 原收據未保存typed X/Y；其餘全部欄位照原版比較。
				if u != c.After[i] {
					t.Errorf("單位%d得到%+v，原版%+v", i, u, c.After[i])
				}
			}
			want := c.PlanetAfter
			if planets[0].Defence != want.Defence || !reflect.DeepEqual(planets[0].Stock, PlanetStock{Ships: want.Ships, Missiles: want.Missiles, Fighters: want.Fighters}) {
				t.Errorf("行星得到%+v，原版%+v", planets[0], want)
			}
		})
	}
}
