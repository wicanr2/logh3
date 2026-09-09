package rules

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// 收據 `s3gunpre`／`s3gunpost` 是**同一次執行**在要塞砲那一步的前後各倒
// 一份 1 MB（`-dump-mem-at`）。分兩次執行去比會混進整場戰鬥的差異——
// 已經踩過一次：`+0x45 ＝ 3` 會讓戰術 AI 走不同的路，兩次執行從第二個
// 階段就分岔了（`docs/re/148`）。
const (
	oracleDS       = 0x54B30 // 資料段的線性起點
	oraclePlanet63 = 0x54560 // 被 poke 的那顆行星（全域索引 63）
	shipRecLen     = 78
	fleetRecLen    = 122
	kindRowLen     = 26
	kindRowsPerFac = 208
)

type gunDump []byte

func (d gunDump) u16(a int) int { return int(binary.LittleEndian.Uint16(d[a:])) }
func (d gunDump) u32(a int) int64 {
	return int64(binary.LittleEndian.Uint32(d[a:]))
}
func (d gunDump) poolBase() int {
	off := d.u16(oracleDS + 0xCFB4)
	seg := d.u16(oracleDS + 0xCFB6)
	return seg*16 + off
}
func (d gunDump) ship(i int) []byte {
	b := d.poolBase() + i*shipRecLen
	return d[b : b+shipRecLen]
}
func (d gunDump) fleet(i int) []byte {
	b := oracleDS + 0xC074 + i*fleetRecLen
	return d[b : b+fleetRecLen]
}
func (d gunDump) kindRow(faction, kind int) []byte {
	b := oracleDS + 0x4E3A + faction*kindRowsPerFac + kind*kindRowLen
	return d[b : b+kindRowLen]
}

func loadGunDump(t *testing.T, name string) gunDump {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("../..", "workplace", name))
	if err != nil {
		t.Skipf("缺收據 %s，跳過：%v", name, err)
	}
	if len(b) != 1<<20 {
		t.Fatalf("%s 是 %d bytes，該是 1 MB", name, len(b))
	}
	return b
}

// TestFortressGunMatchesOracle 對拍要塞砲開火。
//
// 驗三件事，都是**不含亂數**的部分：
//
//  1. 哪些分隊被判進射程（艦艇 `+0x3C`）——完全由幾何決定；
//  2. 每一支被扣的艘數落在算式給的區間裡；
//  3. `203A:0593` 的副作用（`+0x0C` 掉 實扣² ÷ 滿編、`+0x0E`／`+0x12`
//     按剩餘比例縮減、兩個計分器各 ＋ 實扣 ÷ 20）——給定實扣之後全是
//     決定性的。
func TestFortressGunMatchesOracle(t *testing.T) {
	pre := loadGunDump(t, "gunpre.bin")
	post := loadGunDump(t, "gunpost.bin")

	p := pre[oraclePlanet63 : oraclePlanet63+82]
	defMax := int(binary.LittleEndian.Uint16(p[0x16:]))
	aim := int(binary.LittleEndian.Uint16(p[0x50:]))
	px := int(binary.LittleEndian.Uint16(p[0x46:]))
	py := int(binary.LittleEndian.Uint16(p[0x48:]))
	a, b := FortressRange(defMax)
	t.Logf("行星 (%d,%d) 防衛力上限 %d → 長半軸 %d 短半軸 %d 焦距 %d 瞄準 %d",
		px, py, defMax, a, b, FortressFocalDistance(a, b), aim)

	// **只比候選集合。** 原版是先用範圍查詢（`17A0:00AC`）把附近的
	// 分隊列出來，才逐一做橢圓判定；沒被列到的艦艇 `+0x3C` 留在 0xFF。
	// 那支查詢還沒實作，而它只要是橢圓的超集就不影響判定，所以這裡驗
	// 的是「**候選集合裡每一支的判定逐一相同**」。
	hitWant, hitGot := map[int]bool{}, map[int]bool{}
	cand := 0
	for i := 0; i < 352; i++ {
		mark := post.ship(i)[0x3C]
		if mark != 0 && mark != 1 {
			continue // 0xFF：沒被列進候選
		}
		cand++
		s := pre.ship(i)
		sx := int(binary.LittleEndian.Uint16(s[0x34:]))
		sy := int(binary.LittleEndian.Uint16(s[0x36:]))
		if FortressGunInRange(px, py, aim, a, b, sx, sy) {
			hitWant[i] = true
		}
		if mark == 1 {
			hitGot[i] = true
		}
	}
	for i := range hitWant {
		if !hitGot[i] {
			s := pre.ship(i)
			t.Errorf("艦艇 %d (%d,%d) 算出來在射程裡，原版判在外", i,
				int(binary.LittleEndian.Uint16(s[0x34:])),
				int(binary.LittleEndian.Uint16(s[0x36:])))
		}
	}
	for i := range hitGot {
		if !hitWant[i] {
			s := pre.ship(i)
			t.Errorf("艦艇 %d (%d,%d) 原版判在射程裡，算出來在外", i,
				int(binary.LittleEndian.Uint16(s[0x34:])),
				int(binary.LittleEndian.Uint16(s[0x36:])))
		}
	}
	if len(hitWant) == 0 {
		t.Fatal("射程裡一支都沒有，這份收據驗不到東西")
	}
	t.Logf("候選 %d 支，其中射程內 %d 支，逐一相同", cand, len(hitWant))

	// 逐支對損失。先按原版的比較函式排序（帝國的 `2264:052F`：
	// 艦隊編號遞減、`+0x03` 遞增、索引遞增），再從收據的種子接上亂數
	// ——那兩個 `rand()` 是進了 `2264:0EE3` 才取的，順序完全由這個排序
	// 決定。
	order := make([]int, 0, len(hitGot))
	for i := range hitGot {
		order = append(order, i)
	}
	sort.Slice(order, func(x, y int) bool {
		u, v := pre.ship(order[x]), pre.ship(order[y])
		if u[1] != v[1] {
			return u[1] > v[1]
		}
		if u[3] != v[3] {
			return u[3] < v[3]
		}
		return order[x] < order[y]
	})
	rng := NewRand(uint32(pre.u32(oracleDS + 0x94DC)))

	var total, scoreWant int
	for _, i := range order {
		s0, s1 := pre.ship(i), post.ship(i)
		n0 := int(binary.LittleEndian.Uint16(s0[8:]))
		n1 := int(binary.LittleEndian.Uint16(s1[8:]))
		lost := n0 - n1
		total += lost
		scoreWant += lost / 20
		full := int(binary.LittleEndian.Uint16(s0[0x0A:]))
		fl := pre.fleet(int(s0[1]))
		row := pre.kindRow(int(s0[0]), int(s0[4]))
		am := int(binary.LittleEndian.Uint16(row[0x16:]))
		an := int(binary.LittleEndian.Uint16(row[0x18:]))
		sx := int(binary.LittleEndian.Uint16(s0[0x34:]))
		sy := int(binary.LittleEndian.Uint16(s0[0x36:]))
		d := FortressGunDefence(px, py, sx, sy, int(fl[0x4A]), am, an, int(fl[0x3D]))
		base := FortressGunBase(n0, defMax, d)
		var r1 int
		if base != 0 {
			r1 = rng.Next()
		}
		r2 := rng.Next()
		want := FortressGunLoss(base, r1, r2)
		if want < 0 {
			want = 0
		}
		if want > n0 {
			want = n0
		}
		if lost != want {
			t.Errorf("艦艇 %d 掉 %d 艘，算出來 %d（基準 %d、防禦係數 %d）",
				i, lost, want, base, d)
		}
		// 副作用（給定實扣之後是決定性的）
		if want := int(binary.LittleEndian.Uint16(s0[0x0C:])) - lost*lost/full; want !=
			int(binary.LittleEndian.Uint16(s1[0x0C:])) {
			t.Errorf("艦艇 %d 的 +0x0C 是 %d，該是 %d（%d − %d²÷%d）", i,
				int(binary.LittleEndian.Uint16(s1[0x0C:])), want,
				int(binary.LittleEndian.Uint16(s0[0x0C:])), lost, full)
		}
		if s0[4] != 5 {
			for _, off := range []int{0x0E, 0x12} {
				if row[off] == 0xFF {
					continue
				}
				w := int64(binary.LittleEndian.Uint32(s0[off:])) / int64(n0) * int64(n0-lost)
				if g := int64(binary.LittleEndian.Uint32(s1[off:])); g != w {
					t.Errorf("艦艇 %d 的 +0x%02X 是 %d，該是 %d", i, off, g, w)
				}
			}
		}
	}
	t.Logf("總共擊沉 %d 艘（%d 支）", total, len(hitGot))

	// 同一份收據再走一次 production path：FortressGunFire 應該算出
	// 一模一樣的結果。
	checkFortressGunFire(t, pre, post, order, hitGot)

	// 計分器：每一支各加 ⌊實扣 ÷ 20⌋（**逐支捨去**，不是總和再除）。
	if want := scoreWant; want > 0 {
		for _, base := range []int{0x8FE4, 0x8FEC} {
			g0 := pre.u32(oracleDS + base)
			g1 := post.u32(oracleDS + base)
			if g1-g0 != int64(want) {
				t.Errorf("計分器 ds:%04X[0] 加了 %d，該加 %d", base, g1-g0, want)
			}
		}
	}

	// 計分器。
	for _, base := range []int{0x8FE4, 0x8FEC} {
		for f := 0; f < 2; f++ {
			g0 := pre.u32(oracleDS + base + f*4)
			g1 := post.u32(oracleDS + base + f*4)
			if g1 == g0 {
				continue
			}
			t.Logf("計分器 ds:%04X[%d] %d → %d（差 %d）", base, f, g0, g1, g1-g0)
		}
	}
}

// checkFortressGunFire 把收據的狀態搬進 remake 的資料結構，跑一次
// `FortressGunFire`，再與傾印逐支比。前面那一段驗的是算式，這一段驗的是
// **接進去的那條路**：候選、排序、亂數順序、副作用一次全走。
func checkFortressGunFire(t *testing.T, pre, post gunDump, order []int, hit map[int]bool) {
	t.Helper()
	kinds, err := LoadShipKindTable(filepath.Join("../..",
		"docs/internals/data/ship-kind-table.json"))
	if err != nil {
		t.Fatal(err)
	}
	p := pre[oraclePlanet63 : oraclePlanet63+82]
	planets := []BattlePlanet{{
		Index:      63,
		Owner:      int(p[0x11]),
		X:          int(binary.LittleEndian.Uint16(p[0x46:])),
		Y:          int(binary.LittleEndian.Uint16(p[0x48:])),
		Fortress:   p[0x12] != 0,
		DefenceMax: int(binary.LittleEndian.Uint16(p[0x16:])),
		GunState:   int(p[0x45]),
		GunAim:     int(binary.LittleEndian.Uint16(p[0x50:])),
	}}
	// 只搬**候選集合裡**的分隊：範圍查詢還沒實作，池裡沒上場的那些
	// 記錄座標是垃圾，全搬進去會多打到不存在的東西。
	var units []BattleUnit
	idxOf := map[int]int{}
	for i := 0; i < 352; i++ {
		mark := post.ship(i)[0x3C]
		if mark != 0 && mark != 1 {
			continue
		}
		s := pre.ship(i)
		idxOf[i] = len(units)
		units = append(units, BattleUnit{
			Fleet:     int(s[1]),
			Squadron:  int(s[3]),
			Faction:   int(s[0]),
			Kind:      int(s[4]),
			X:         int(binary.LittleEndian.Uint16(s[0x34:])),
			Y:         int(binary.LittleEndian.Uint16(s[0x36:])),
			Ships:     int(binary.LittleEndian.Uint16(s[8:])),
			FullShips: int(binary.LittleEndian.Uint16(s[0x0A:])),
			RepairCap: int(binary.LittleEndian.Uint16(s[0x0C:])),
			Missiles:  int(binary.LittleEndian.Uint32(s[0x0E:])),
			Fighters:  int(binary.LittleEndian.Uint32(s[0x12:])),
		})
	}
	var fleets []BattleFleet
	for n := 0; n < 32; n++ {
		f := pre.fleet(n)
		fleets = append(fleets, BattleFleet{
			Index: n, Faction: int(f[1]), Facing: int(f[0x4A]),
			Ability0: int(f[0x3D]),
		})
	}
	rng := NewRand(uint32(pre.u32(oracleDS + 0x94DC)))
	hits := FortressGunFire(planets, fleets, units, kinds, rng)
	if len(hits) != len(hit) {
		t.Fatalf("FortressGunFire 打中 %d 支，收據是 %d 支", len(hits), len(hit))
	}
	var score int
	for _, h := range hits {
		score += h.Score
	}
	for i := range idxOf {
		u := units[idxOf[i]]
		s1 := post.ship(i)
		if u.Ships != int(binary.LittleEndian.Uint16(s1[8:])) {
			t.Errorf("艦艇 %d 算出剩 %d 艘，收據是 %d",
				i, u.Ships, int(binary.LittleEndian.Uint16(s1[8:])))
		}
		if u.RepairCap != int(binary.LittleEndian.Uint16(s1[0x0C:])) {
			t.Errorf("艦艇 %d 的 +0x0C 算出 %d，收據是 %d",
				i, u.RepairCap, int(binary.LittleEndian.Uint16(s1[0x0C:])))
		}
		if int64(u.Missiles) != int64(binary.LittleEndian.Uint32(s1[0x0E:])) {
			t.Errorf("艦艇 %d 的飛彈算出 %d，收據是 %d",
				i, u.Missiles, binary.LittleEndian.Uint32(s1[0x0E:]))
		}
		if int64(u.Fighters) != int64(binary.LittleEndian.Uint32(s1[0x12:])) {
			t.Errorf("艦艇 %d 的艦載機算出 %d，收據是 %d",
				i, u.Fighters, binary.LittleEndian.Uint32(s1[0x12:]))
		}
	}
	want := int(post.u32(oracleDS+0x8FE4) - pre.u32(oracleDS+0x8FE4))
	if score != want {
		t.Errorf("FortressGunFire 加了 %d 分，收據是 %d", score, want)
	}
	FortressGunTickAll(planets)
	if planets[0].GunState != 2 {
		t.Errorf("開火之後狀態 %d，該是 2", planets[0].GunState)
	}
	t.Logf("FortressGunFire 打中 %d 支、加 %d 分，與收據逐一相同", len(hits), score)
}
