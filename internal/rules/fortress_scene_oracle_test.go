package rules

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"path/filepath"
	"testing"
)

// 原版完整保存狀態續跑，同次 2264:0248 入口／返回。收據缺失就 skip。
// 直接搬全部有效 SHIP，不以原版的命中標記替 remake 篩選目標。
// 證據與範圍：docs/playtest/122-fortress-gun-trigger.md。
func TestFortressGunSceneOracle(t *testing.T) {
	base := "oracle/fortress48-r133-r3-hit-a"
	pre := loadGunDump(t, filepath.Join(base, "gun-before-mem.bin"))
	post := loadGunDump(t, filepath.Join(base, "gun-return-181b-03bf-01-mem.bin"))
	for _, v := range []struct {
		dump gunDump
		hash string
	}{
		{pre, "25f0bc4a42b43b168a96b526022819212de367eda48317f0f4c93186f8c54c5f"},
		{post, "a7bf187dd706e05bfd5e26ebde603da4b7e522638ae12de40200f910bff4e6b0"},
	} {
		if got := fmt.Sprintf("%x", sha256.Sum256(v.dump)); got != v.hash {
			t.Fatalf("原版收據遭替換：%s != %s", got, v.hash)
		}
	}
	u16 := func(b []byte, off int) int { return int(binary.LittleEndian.Uint16(b[off:])) }
	u32 := func(b []byte, off int) int { return int(binary.LittleEndian.Uint32(b[off:])) }
	paddr := pre.u16(oracleDS+0xCFF8) + 16*pre.u16(oracleDS+0xCFFA)
	p := pre[paddr : paddr+82]
	planets := []BattlePlanet{{Index: 63, Owner: int(p[0x11]), Fortress: p[0x12] != 0,
		DefenceMax: u16(p, 0x16), GunState: int(p[0x45]), GunAim: u16(p, 0x50),
		X: u16(p, 0x46), Y: u16(p, 0x48)}}
	var units []BattleUnit
	var poolIndices []int
	for i := 0; i < 336; i++ {
		s := pre.ship(i)
		if s[0] > 1 || s[1] >= 32 || u16(s, 8) == 0 {
			continue
		}
		poolIndices = append(poolIndices, i)
		units = append(units, BattleUnit{Fleet: int(s[1]), Squadron: int(s[3]),
			Faction: int(s[0]), Kind: int(s[4]), X: u16(s, 0x34), Y: u16(s, 0x36),
			Ships: u16(s, 8), FullShips: u16(s, 10), RepairCap: u16(s, 12),
			Missiles: u32(s, 14), Fighters: u32(s, 18)})
	}
	var fleets []BattleFleet
	for i := 0; i < 32; i++ {
		f := pre.fleet(i)
		fleets = append(fleets, BattleFleet{Index: i, Faction: int(f[1]),
			Facing: int(f[0x4A]), Ability0: int(f[0x3D])})
	}
	kinds, err := LoadShipKindTable("../../docs/internals/data/ship-kind-table.json")
	if err != nil {
		t.Fatal(err)
	}
	rng := NewRand(uint32(pre.u32(oracleDS + 0x94DC)))
	hits := FortressGunFire(planets, fleets, units, kinds, rng)
	var total int
	var scores [2]int
	for _, h := range hits {
		total += h.Lost
		scores[h.Scorer] += h.Score
	}
	if len(hits) != 10 || total != 10637 {
		t.Fatalf("命中 %d 支、擊沉 %d 艘，原版是 10 支、10637 艘", len(hits), total)
	}
	for j, i := range poolIndices {
		u, s := units[j], post.ship(i)
		if u.Ships != u16(s, 8) || u.RepairCap != u16(s, 12) ||
			u.Missiles != u32(s, 14) || u.Fighters != u32(s, 18) {
			t.Errorf("SHIP[%d] 戰損／庫存副作用不符原版", i)
		}
	}
	if got, want := rng.Seed(), uint32(post.u32(oracleDS+0x94DC)); got != want {
		t.Errorf("末端 seed %08X != 原版 %08X", got, want)
	}
	for _, off := range []int{0x8FE4, 0x8FEC} {
		for side, score := range scores {
			if want := post.u32(oracleDS+off+4*side) - pre.u32(oracleDS+off+4*side); int64(score) != want {
				t.Errorf("ds:%04X[%d] 分數 %d != 原版 %d", off, side, score, want)
			}
		}
	}
	t.Logf("%d 支有效分隊：命中 %d 支、擊沉 %d 艘；逐支副作用、雙計分器及末端 seed 相同", len(units), len(hits), total)
}
