// fleet_attack_probe：把原版 203A:0424（戰略艦隊攻擊力）當函式庫呼叫，產生差異測試向量。
//
// 在 dosgolem 的 module 內執行（decomp/tools/golem_probe_run.sh），不改動執行器。
// 流程：載入已核對雜湊的保存狀態 → 對每一案把艦隊記錄與艦艇記錄寫進記憶體 →
// 直接呼叫原版常式 → 記錄回傳值與「是否只動到堆疊」。
// 向量檔由 decomp/test/check_fleet_attack.c 讀回，與 C 重寫版逐筆比對。
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/dos"
	"github.com/wicanr2/dosgolem/internal/machine"
	"github.com/wicanr2/dosgolem/internal/state"
)

const (
	wantStateSHA = "c55de581a6393e0e39e304dd0f191e9d1cfa8457b86b93f2a9e3368b5751ba26"
	wantGameSHA  = "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918"

	dgroupSeg  = 0x54B3
	fleetTable = 0xC074
	fleetSize  = 122
	shipSize   = 78
	poolShips  = 352

	funcSeg = 0x203A
	funcOff = 0x0424
	funcLen = 0xAD // 0424..04D0，最後一個 byte 是 retf

	sentinelSeg = 0xFFF0
	stackSlack  = 0x100
)

var funcPrefix = []byte{0x55, 0x56, 0x57, 0x83, 0xEC, 0x16, 0x89, 0xE5}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func digest(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

type testCase struct {
	fleet []byte
	pool  map[uint16][]byte
}

func newCase(r *rand.Rand) *testCase {
	c := &testCase{fleet: make([]byte, fleetSize), pool: map[uint16][]byte{}}
	r.Read(c.fleet)
	return c
}

func (c *testCase) setSlots(idx []uint16) {
	c.fleet[0x0F] = byte(len(idx))
	for i, v := range idx {
		binary.LittleEndian.PutUint16(c.fleet[0x10+i*4:], v)
	}
}

func (c *testCase) setShip(r *rand.Rand, idx uint16, kind byte, count uint16) {
	rec := make([]byte, shipSize)
	r.Read(rec)
	rec[4] = kind
	binary.LittleEndian.PutUint16(rec[8:], count)
	c.pool[idx] = rec
}

func pick(r *rand.Rand, weights []int) int {
	total := 0
	for _, w := range weights {
		total += w
	}
	n := r.Intn(total)
	for i, w := range weights {
		if n < w {
			return i
		}
		n -= w
	}
	return len(weights) - 1
}

func randomCase(r *rand.Rand) *testCase {
	c := newCase(r)
	var count int
	switch pick(r, []int{1, 1, 1, 1, 10}) {
	case 0:
		count = 0
	case 1:
		count = 1
	case 2:
		count = 12
	case 3:
		count = 11
	default:
		count = r.Intn(13)
	}
	idx := make([]uint16, count)
	for i := range idx {
		if i > 0 && r.Intn(5) == 0 {
			idx[i] = idx[r.Intn(i)]
		} else {
			idx[i] = uint16(r.Intn(poolShips))
		}
	}
	c.setSlots(idx)
	switch pick(r, []int{5, 2, 3}) {
	case 0:
		c.fleet[0x40] = byte(78 + r.Intn(43))
	case 1:
		c.fleet[0x40] = byte(r.Intn(256))
	default:
		c.fleet[0x40] = []byte{0, 1, 100, 127, 128, 255}[r.Intn(6)]
	}
	for _, id := range idx {
		if _, done := c.pool[id]; done {
			continue
		}
		var kind byte
		switch pick(r, []int{2, 1, 1}) {
		case 0:
			kind = byte(r.Intn(2))
		case 1:
			kind = byte(2 + r.Intn(6))
		default:
			kind = byte(r.Intn(256))
		}
		var ships uint16
		switch pick(r, []int{4, 3, 3}) {
		case 0:
			ships = uint16(r.Intn(301))
		case 1:
			ships = uint16(r.Intn(65536))
		default:
			ships = []uint16{0, 1, 65535, 32768, 32767, 4369}[r.Intn(6)]
		}
		c.setShip(r, id, kind, ships)
	}
	return c
}

// fixedCases 是會被亂數漏掉或值得明寫的邊界。
func fixedCases(r *rand.Rand) []*testCase {
	var out []*testCase
	add := func(ability byte, idx []uint16, ships map[uint16][2]uint16) {
		c := newCase(r)
		c.setSlots(idx)
		c.fleet[0x40] = ability
		for id, v := range ships {
			c.setShip(r, id, byte(v[0]), v[1])
		}
		out = append(out, c)
	}
	add(100, nil, nil)                                                         // 空艦隊
	add(100, []uint16{0}, map[uint16][2]uint16{0: {0, 1500}})                  // 一般
	add(255, []uint16{0, 0}, map[uint16][2]uint16{0: {1, 40000}})              // 艘數合計回繞
	add(255, []uint16{5}, map[uint16][2]uint16{5: {1, 65535}})                 // 第一次除法商超過 16 位元
	add(0, []uint16{3}, map[uint16][2]uint16{3: {0, 50000}})                   // 能力 0
	add(120, []uint16{351, 0}, map[uint16][2]uint16{351: {2, 900}, 0: {0, 7}}) // 池尾與非主力艦種
	return out
}

type probe struct {
	m                *machine.Machine
	d                *dos.DOS
	poolSeg, poolOff uint16
}

func (p *probe) word(seg, off uint16) uint16 { return p.m.Read16(cpu.Addr(seg, off)) }

func (p *probe) load(path string) {
	raw, err := os.ReadFile(path)
	must(err)
	if digest(raw) != wantStateSHA {
		panic("保存狀態雜湊不符：" + digest(raw))
	}
	p.m = machine.New()
	p.d = dos.New(p.m, "/orig")
	p.d.Install()
	must(state.Load(path, p.m, p.d))
	p.poolSeg = p.word(dgroupSeg, 0xCFB6)
	p.poolOff = p.word(dgroupSeg, 0xCFB4)
	if p.poolSeg == 0 {
		panic("艦艇池指標為 0，保存狀態不是遊戲執行中")
	}
}

func (p *probe) funcBytes() []byte {
	lin := cpu.Addr(funcSeg, funcOff)
	return append([]byte(nil), p.m.Mem[lin:lin+funcLen]...)
}

type result struct {
	ax, dx uint16
	pure   bool
	steps  uint64
}

// call 照 oracle.Call 的作法呼叫 far 常式，另加：清 IF、設 DS、檢查只動到堆疊。
func (p *probe) call(fleetOff uint16) (result, error) {
	m := p.m
	saved := *m.CPU
	c := m.CPU
	pre := make([]byte, 1<<20)
	copy(pre, m.Mem[:1<<20])

	ss, sp0 := c.Seg[cpu.SS], c.R[cpu.SP]
	sp := sp0
	push := func(v uint16) {
		sp -= 2
		m.Write16(cpu.Addr(ss, sp), v)
	}
	push(dgroupSeg) // 第二個 word：segment
	push(fleetOff)  // 第一個 word：offset
	push(sentinelSeg)
	push(0)
	c.R[cpu.SP] = sp
	c.Seg[cpu.CS], c.IP = funcSeg, funcOff
	c.Seg[cpu.DS] = dgroupSeg
	c.Flags &^= 0x200

	start := m.Steps
	var err error
	for {
		if c.Seg[cpu.CS] == sentinelSeg && c.IP == 0 {
			break
		}
		if m.Steps-start > 100000 {
			err = fmt.Errorf("超過 100000 道指令仍未返回 %04X:%04X", c.Seg[cpu.CS], c.IP)
			break
		}
		if e := m.Step(); e != nil {
			err = e
			break
		}
	}
	res := result{ax: c.R[cpu.AX], dx: c.R[cpu.DX], steps: m.Steps - start, pure: true}
	lo := cpu.Addr(ss, sp0) - stackSlack
	hi := cpu.Addr(ss, sp0)
	for i := 0; i < 1<<20; i++ {
		if m.Mem[i] != pre[i] && (uint32(i) < lo || uint32(i) >= hi) {
			res.pure = false
			break
		}
	}
	*m.CPU = saved
	return res, err
}

func (p *probe) install(c *testCase, fleetIdx int) uint16 {
	fleetOff := uint16(fleetTable + fleetIdx*fleetSize)
	p.m.WriteBytes(cpu.Addr(dgroupSeg, fleetOff), c.fleet)
	for idx, rec := range c.pool {
		off := p.poolOff + idx*shipSize
		p.m.WriteBytes(cpu.Addr(p.poolSeg, off), rec)
	}
	return fleetOff
}

func main() {
	statePath := flag.String("state", "/src/workplace/fcmd/fcmd-base.state", "保存狀態")
	output := flag.String("output", "", "向量檔輸出路徑（必須是新檔）")
	n := flag.Int("n", 5000, "隨機案數")
	seed := flag.Int64("seed", 20261008, "亂數種子")
	headStr := flag.String("dosgolem-head", "", "執行器 Git HEAD（由呼叫端核對後傳入）")
	flag.Parse()
	if *output == "" || *headStr == "" {
		panic("缺 -output 或 -dosgolem-head")
	}
	game, err := os.ReadFile("/orig/GIN3.COM")
	must(err)
	if digest(game) != wantGameSHA {
		panic("GIN3.COM 雜湊不符")
	}

	p := &probe{}
	p.load(*statePath)
	code := p.funcBytes()
	if !bytes.HasPrefix(code, funcPrefix) || code[len(code)-1] != 0xCB {
		panic(fmt.Sprintf("203A:0424 的位元組不符預期：%x", code[:16]))
	}

	r := rand.New(rand.NewSource(*seed))
	cases := fixedCases(r)
	for i := 0; i < *n; i++ {
		cases = append(cases, randomCase(r))
	}

	out, err := os.OpenFile(*output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	must(err)
	w := bufio.NewWriter(out)
	fmt.Fprintf(w, "# schema decomp.fleet-attack-vectors.v1\n")
	fmt.Fprintf(w, "# function GIN3PS 203A:0424 bytes=%d sha256=%s\n", funcLen, digest(code))
	fmt.Fprintf(w, "# state_sha256 %s\n# game_sha256 %s\n", wantStateSHA, wantGameSHA)
	fmt.Fprintf(w, "# dosgolem_head %s\n# seed %d fixed=%d random=%d\n", *headStr, *seed, len(cases)-*n, *n)
	fmt.Fprintf(w, "# pool_seg %04X pool_off %04X fleet_table ds:%04X stride %d\n", p.poolSeg, p.poolOff, fleetTable, fleetSize)

	var minSteps, maxSteps uint64 = ^uint64(0), 0
	impure := 0
	for id, c := range cases {
		fleetOff := p.install(c, 31)
		res, err := p.call(fleetOff)
		must(err)
		if res.steps < minSteps {
			minSteps = res.steps
		}
		if res.steps > maxSteps {
			maxSteps = res.steps
		}
		if !res.pure {
			impure++
		}
		ids := make([]int, 0, len(c.pool))
		for k := range c.pool {
			ids = append(ids, int(k))
		}
		sort.Ints(ids)
		fmt.Fprintf(w, "C %d %s %d", id, hex.EncodeToString(c.fleet), len(ids))
		for _, k := range ids {
			fmt.Fprintf(w, " %d:%s", k, hex.EncodeToString(c.pool[uint16(k)]))
		}
		pure := 0
		if res.pure {
			pure = 1
		}
		fmt.Fprintf(w, " %04X %04X %d %d\n", res.ax, res.dx, pure, res.steps)
	}
	must(w.Flush())
	must(out.Close())
	fmt.Printf("案數 %d（固定 %d、隨機 %d），步數 %d..%d，動到堆疊以外記憶體 %d 案，輸出 %s\n",
		len(cases), len(cases)-*n, *n, minSteps, maxSteps, impure, *output)
}
