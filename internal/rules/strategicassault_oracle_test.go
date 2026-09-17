package rules

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// 戰略打行星的同狀態對拍（規格 058、`docs/re/330`）：`tools/fleetcmd_receipts.sh` 第 22 段
// 在原版進入 `20B4:1497` 與回到 `20B4:1507` 那兩步各倒一份 1 MB（`mem-in.bin`／`mem-out.bin`）。
// 拿進入時的行星表、艦隊表、艦艇池、提督表與亂數種子驅動 `StrategicAssault`，
// 逐欄比離開時的記憶體。
//
// 線性位址（`ds` ＝ 54B3）：星域表 `ds:A79C`、行星表 `53132`（`ds:5336` far）、
// 每格行星顆數 `ds:CFB8`、艦隊表 `ds:C074`、艦艇池 `668B:0002`（這條鏈固定）、
// 提督表 `ds:B77C`、亂數 `ds:94DC`、計分器 `ds:8FE4`／`ds:8FEC`。
const (
	memDS         = 0x54B30
	memPlanets    = 0x53132
	memShipPool   = 0x668B2
	memRegionSize = 126
)

func TestStrategicAssaultMatchesOracle(t *testing.T) {
	for _, c := range []struct {
		batch   string
		region  int
		defence int // 打的那一顆離開時的防衛力
		morale  int
	}{
		{"fcmd-assault-ast", 16, 3768, 67}, // 亞斯提：艦隊 0 打巴溫
		{"fcmd-assault-phz", 14, 8792, 47}, // 費沙（無主）
	} {
		t.Run(c.batch, func(t *testing.T) {
			in, out := readMem(t, c.batch, "in"), readMem(t, c.batch, "out")

			// 行星切片：前面每格的顆數加起來。
			start := 0
			for r := 0; r < c.region; r++ {
				start += int(binary.LittleEndian.Uint16(in[memDS+0xCFB8+2*r:]))
			}
			n := int(binary.LittleEndian.Uint16(in[memDS+0xCFB8+2*c.region:]))
			lo, hi := memPlanets+start*PlanetRecord, memPlanets+(start+n)*PlanetRecord
			planets := append([]byte(nil), in[lo:hi]...)
			// 行星座標由 `2AF5:00CB` 在 1497 開頭算；進入那一刻還是上一次的，取離開時的。
			sites := make([]PlanetSite, n)
			for i := range sites {
				p := out[lo+i*PlanetRecord:]
				sites[i] = PlanetSite{X: int(binary.LittleEndian.Uint16(p[0x46:])), Y: int(binary.LittleEndian.Uint16(p[0x48:]))}
			}

			slotIn := memSlot(in)
			fleets := slotIn.Fleets(nil)
			admIn := in[memDS+0xB77C : memDS+0xB77C+82*AdmiralRecord]
			// 攻方 ＝ 這一格有沒有帝國艦隊（`297B:14F8`，士氣低落的不算）。
			attacker := 1
			for _, f := range fleets {
				if f.Region == c.region && f.Faction == 0 && f.Flags&fleetFlagLowMorale == 0 {
					attacker = 0
				}
			}
			var sfs []StrategicFleet
			for i := range fleets {
				f := &fleets[i]
				if f.Region != c.region || f.Faction != attacker {
					continue
				}
				rec := slotIn.fleet(f.Index)
				adm := admIn[f.Admiral*AdmiralRecord : (f.Admiral+1)*AdmiralRecord]
				var adj []byte
				if f.Adjutant >= 0 {
					adj = admIn[f.Adjutant*AdmiralRecord : (f.Adjutant+1)*AdmiralRecord]
				}
				ab := FleetAbilities(adm, adj)
				if int(ab[3]) != int(rec[0x40]) || int(ab[4]) != int(rec[0x41]) {
					t.Fatalf("艦隊 %d 能力 +40/+41 由提督算出 %d/%d，記錄 %d/%d", f.Index, ab[3], ab[4], rec[0x40], rec[0x41])
				}
				sfs = append(sfs, StrategicFleet{Fleet: f, Ability40: int(ab[3]), Ability41: int(ab[4]), Command: int(adm[0x0B])})
			}
			if len(sfs) == 0 {
				t.Fatal("這一格沒有攻方艦隊")
			}
			r := NewRand(binary.LittleEndian.Uint32(in[memDS+0x94DC:]))
			res := StrategicAssault(planets, sites, attacker, sfs, r)

			if got, want := r.Seed(), binary.LittleEndian.Uint32(out[memDS+0x94DC:]); got != want {
				t.Errorf("亂數種子 %08X，原版 %08X（消耗次數對不上）", got, want)
			}
			// 行星：擁有者、防衛力、存量、飛彈、艦載機逐筆逐欄。
			for i := 0; i < n; i++ {
				g, w := planets[i*PlanetRecord:], out[lo+i*PlanetRecord:]
				for _, off := range []int{planetOwner, planetDefence, planetDefence + 1} {
					if g[off] != w[off] {
						t.Errorf("行星 %d +%02X ＝ %02X，原版 %02X", start+i, off, g[off], w[off])
					}
				}
				for off := planetKindStock; off < planetFighters+4; off++ {
					if g[off] != w[off] {
						t.Errorf("行星 %d 存量 +%02X ＝ %02X，原版 %02X", start+i, off, g[off], w[off])
						break
					}
				}
			}
			slotOut := memSlot(out)
			for _, sf := range sfs {
				f := sf.Fleet
				wantF := slotOut.fleet(f.Index)
				if f.Morale != int(wantF[0x3C]) || f.Flags != int(wantF[0x00]) {
					t.Errorf("艦隊 %d 士氣 %d／+00 %02X，原版 %d／%02X", f.Index, f.Morale, f.Flags, wantF[0x3C], wantF[0x00])
				}
				wantC := slotOut.FleetComposition(f.Index)
				for gi, g := range f.Composition {
					if gi < len(wantC) && g.Ships != wantC[gi].Ships {
						t.Errorf("艦隊 %d 第 %d 組 %d 艘，原版 %d", f.Index, gi, g.Ships, wantC[gi].Ships)
					}
				}
			}
			admOut := out[memDS+0xB77C:]
			for a, v := range res.Merit {
				before := binary.LittleEndian.Uint32(admIn[a*AdmiralRecord+2:])
				if got, want := before+uint32(v), binary.LittleEndian.Uint32(admOut[a*AdmiralRecord+2:]); got != want {
					t.Errorf("提督 %d 功績 %d，原版 %d", a, got, want)
				}
			}
			for fac := 0; fac < 2; fac++ {
				for _, base := range []int{0x8FE4, 0x8FEC} {
					before := binary.LittleEndian.Uint32(in[memDS+base+4*fac:])
					if got, want := before+uint32(res.Score[fac]), binary.LittleEndian.Uint32(out[memDS+base+4*fac:]); got != want {
						t.Errorf("ds:%04X[%d] ＝ %d，原版 %d", base, fac, got, want)
					}
				}
			}
			// 收據上的兩個數（`docs/re/322` §四）。
			hit := sfs[0].Fleet
			found := false
			for i := 0; i < n; i++ {
				if int(binary.LittleEndian.Uint16(planets[i*PlanetRecord+planetDefence:])) == c.defence {
					found = true
				}
			}
			if !found || hit.Morale != c.morale {
				t.Errorf("收據的防衛力 %d／士氣 %d 沒有出現（士氣 %d）", c.defence, c.morale, hit.Morale)
			}
		})
	}
}

func readMem(t *testing.T, batch, which string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "workplace/oracle", batch, "mem-"+which+".bin"))
	if err != nil || len(b) != 1<<20 {
		t.Skipf("缺收據 %s 的 mem-%s.bin，跳過（tools/fleetcmd_receipts.sh 第 22 段）", batch, which)
	}
	return b
}

// memSlot 把 1 MB 傾印裡的艦隊表與艦艇池包成 Slot（只為了沿用它的解析）。
func memSlot(m []byte) *Slot {
	return &Slot{
		Fleet: m[memDS+0xC074 : memDS+0xC074+FleetCount*FleetRecord],
		Ship:  m[memShipPool : memShipPool+ShipCount*ShipRecord],
	}
}
