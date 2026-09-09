package rules

import (
	"encoding/binary"
	"path/filepath"
	"testing"
)

// TestPlanetAssaultMatchesOracle 對拍打行星那一步。
//
// 收據 `s3assault`（`docs/re/150`、`docs/re/151`）：S3 不必 poke——
// 雷奧米爾在第八個回合被同盟打下來。傾印是**同一次執行**在那一步
// （`181B:03D0`）的前後各一份。
//
// 驗的是整條 production path：兩個距離帶的標記、艦隊的處理順序、
// 亂數的消耗順序、傷害、換手條件、行星反擊怎麼分攤到各分隊。
func TestPlanetAssaultMatchesOracle(t *testing.T) {
	pre := loadGunDump(t, "aspre.bin")
	post := loadGunDump(t, "aspost.bin")
	kinds, err := LoadShipKindTable(filepath.Join("../..",
		"docs/internals/data/ship-kind-table.json"))
	if err != nil {
		t.Fatal(err)
	}

	// 戰鬥中 `ds:CFF8`／`ds:CFFC` 指的是**這個星域的行星切片**
	// （`docs/re/150`），所以照它建 planets。
	base := pre.u16(oracleDS+0xCFFA)*16 + pre.u16(oracleDS+0xCFF8)
	n := int(pre[oracleDS+0xCFFC])
	var planets []BattlePlanet
	for i := 0; i < n; i++ {
		rec := pre[base+i*82 : base+(i+1)*82]
		st, _ := ReadPlanetStock(rec)
		planets = append(planets, BattlePlanet{
			Index:      i,
			Owner:      int(rec[0x11]),
			X:          int(binary.LittleEndian.Uint16(rec[0x46:])),
			Y:          int(binary.LittleEndian.Uint16(rec[0x48:])),
			Defence:    int(binary.LittleEndian.Uint16(rec[0x14:])),
			DefenceMax: int(binary.LittleEndian.Uint16(rec[0x16:])),
			Fortress:   rec[0x12] != 0,
			GunState:   int(rec[0x45]),
			GunAim:     int(binary.LittleEndian.Uint16(rec[0x50:])),
			Stock:      st,
		})
	}

	// 在場艦隊的清單在 `ds:9898`、筆數在 `ds:9894`（`docs/re/149`）。
	cnt := int(pre[oracleDS+0x9894])
	var fleets []BattleFleet
	unitOf := map[int]int{} // 艦艇池索引 → units 的索引
	var units []BattleUnit
	for i := 0; i < cnt; i++ {
		fn := int(pre[oracleDS+0x9898+i])
		f := pre.fleet(fn)
		fleets = append(fleets, BattleFleet{
			Index: fn, Faction: int(f[1]), Facing: int(f[0x4A]),
			Ability0: int(f[0x3D]), Ability3: int(f[0x40]),
			Defence: int(f[0x40]) * int(f[0x41]) / 100,
			Integrity: int(f[0x3C]), Flag0: f[0]&1 != 0,
			AdmiralLossFactor: int(pre[oracleDS+0xB77C+int(f[4])*AdmiralRecord+0x0B]),
			Action: int(f[0x48]) & 0x0F,
			X:      int(binary.LittleEndian.Uint16(f[0x5C:])),
			Y:      int(binary.LittleEndian.Uint16(f[0x5E:])),
		})
		for k := 0; k < int(f[0x0F]); k++ {
			si := int(binary.LittleEndian.Uint16(f[0x10+k*4:]))
			s := pre.ship(si)
			unitOf[si] = len(units)
			units = append(units, BattleUnit{
				Fleet: fn, Squadron: int(s[3]), Faction: int(s[0]),
				Kind: int(s[4]),
				X:    int(binary.LittleEndian.Uint16(s[0x34:])),
				Y:    int(binary.LittleEndian.Uint16(s[0x36:])),
				Ships:     int(binary.LittleEndian.Uint16(s[8:])),
				FullShips: int(binary.LittleEndian.Uint16(s[0x0A:])),
				RepairCap: int(binary.LittleEndian.Uint16(s[0x0C:])),
				Missiles:  int(binary.LittleEndian.Uint32(s[0x0E:])),
				Fighters:  int(binary.LittleEndian.Uint32(s[0x12:])),
			})
		}
	}

	// 順帶驗編制完整度（`+0x3C`）算得出來：傾印是在打行星那一步的進場
	// 拍的，那時原版的值是最新的（`docs/re/153`）。
	for i := range fleets {
		if got := FleetIntegrity(&fleets[i], units); got != fleets[i].Integrity {
			t.Errorf("艦隊 %d 的編制完整度算出 %d，收據是 %d",
				fleets[i].Index, got, fleets[i].Integrity)
		}
	}

	rng := NewRand(uint32(pre.u32(oracleDS + 0x94DC)))
	hits := PlanetAssault(planets, fleets, units, kinds, rng)
	if len(hits) == 0 {
		t.Fatal("一支都沒打，這份收據驗不到東西")
	}

	// 行星：擁有者、防衛力與換手的三格。
	for i := range planets {
		rec := post[base+i*82 : base+(i+1)*82]
		if got, want := planets[i].Owner, int(rec[0x11]); got != want {
			t.Errorf("行星 %d 擁有者算出 %d，收據是 %d", i, got, want)
		}
		if got, want := planets[i].Defence,
			int(binary.LittleEndian.Uint16(rec[0x14:])); got != want {
			t.Errorf("行星 %d 防衛力算出 %d，收據是 %d", i, got, want)
		}
		st, _ := ReadPlanetStock(rec)
		if planets[i].Stock.Missiles != st.Missiles ||
			planets[i].Stock.Fighters != st.Fighters {
			t.Errorf("行星 %d 存量算出 %d／%d，收據是 %d／%d", i,
				planets[i].Stock.Missiles, planets[i].Stock.Fighters,
				st.Missiles, st.Fighters)
		}
	}

	// 逐個分隊比艘數與副作用。
	bad := 0
	for si, ui := range unitOf {
		u, s1 := units[ui], post.ship(si)
		if u.Ships != int(binary.LittleEndian.Uint16(s1[8:])) {
			t.Errorf("艦艇 %d 算出剩 %d 艘，收據是 %d",
				si, u.Ships, int(binary.LittleEndian.Uint16(s1[8:])))
			bad++
		}
		if u.RepairCap != int(binary.LittleEndian.Uint16(s1[0x0C:])) {
			t.Errorf("艦艇 %d 的 +0x0C 算出 %d，收據是 %d",
				si, u.RepairCap, int(binary.LittleEndian.Uint16(s1[0x0C:])))
			bad++
		}
		if bad > 8 {
			t.Fatal("差太多，先停")
		}
	}
	var dmg, counter int
	for _, h := range hits {
		dmg += h.Damage
		counter += h.Counter
	}
	t.Logf("%d 支艦隊出手：打掉防衛力 %d、挨反擊 %d 艘，逐一相同",
		len(hits), dmg, counter)
}
