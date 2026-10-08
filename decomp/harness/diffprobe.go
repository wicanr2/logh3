// diffprobe：描述檔驅動的原版函式差異探針。
//
// 在 dosgolem 的 module 內執行（decomp/tools/diff_run.sh），不改動執行器。
// 載入已核對雜湊的保存狀態，依描述檔產生參數，直接呼叫原版函式，記錄：
//   - 回傳值（AX 或 DX:AX）
//   - 函式實際讀取的位元組（讀取集），作為下一輪突變的候選
//   - 堆疊以外的全部記憶體寫入（全 1 MB 前後比對）
//   - 執行過的指令位址，回饋覆蓋率
//
// 突變導向讀取集：只改函式真的讀過的位元組，所以不必手寫資料結構的亂數產生器。
// 覆蓋到新指令的輸入留進語料庫，語料庫再被突變，直到迭代用完。
// 輸出向量檔由 decomp/test/check_diff.c 在 C 參考實作上重跑並比對。
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/dos"
	"github.com/wicanr2/dosgolem/internal/machine"
	"github.com/wicanr2/dosgolem/internal/state"
)

const (
	dgroupSeg    = 0x54B3
	imageLoad    = 0x3490  // 映像在 dosgolem 的線性位址
	codeEnd      = 0x3D200 // 程式碼區結束（映像 0x39D11 之後略有餘裕）
	sentinelSeg  = 0xFFF0
	sentinelIP   = 0x0000
	nearSentinel = 0xFFF0
	stackBand    = 0x400
	memSize      = 1 << 20
	scratchBase  = 0x9D000
	stepBudget   = 200000
)

type TableSpec struct {
	Base   string `json:"base"` // "ds:C074" 或 "ptr:CFB4:CFB6"（位移在前、段在後的兩個 DS 變數）
	Stride int    `json:"stride"`
	Count  int    `json:"count"`
}

type GenSpec struct {
	Min     *int   `json:"min"`
	Max     *int   `json:"max"`
	Choices []int  `json:"choices"`
	Table   string `json:"table"`
	Offset  int    `json:"offset"`
	DS      string `json:"ds"`      // 固定的 DS 位移（hex）
	Scratch int    `json:"scratch"` // 指向新配置的暫存緩衝區，長度 bytes
	Null    bool   `json:"null"`
	// AlsoRandom 讓 choices 有一半機率改取均勻隨機值
	AlsoRandom bool `json:"also_random"`
}

type ArgSpec struct {
	Name string  `json:"name"`
	Kind string  `json:"kind"` // word byte dword farptr nearptr reg
	Reg  string  `json:"reg"`  // kind=reg 時的暫存器：ax cx dx bx si di bp
	Gen  GenSpec `json:"gen"`
}

type Spec struct {
	ID          string               `json:"id"`
	Entry       string               `json:"entry"` // "203A:0424"
	Conv        string               `json:"conv"`  // far | near
	Ret         string               `json:"ret"`   // ax | dxax | none | pack:ax,bx,...（依序由高位到低位串成一個值，最多 4 個暫存器）
	CodeSHA16   string               `json:"code_sha256_16"`
	Size        int                  `json:"size"`
	BaseState   string               `json:"base_state"`
	BaseSHA     string               `json:"base_state_sha256"`
	Tables      map[string]TableSpec `json:"tables"`
	Args        []ArgSpec            `json:"args"`
	Iters       int                  `json:"iters"`
	MaxCases    int                  `json:"max_cases"`
	Seed        int64                `json:"seed"`
	MinCoverage float64              `json:"min_coverage"`
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func digest(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

type mut struct {
	addr uint32
	val  uint8
}

type input struct {
	args []uint16
	muts map[uint32]uint8 // 相對於基底傾印的差異
}

func (in *input) clone() *input {
	c := &input{args: append([]uint16(nil), in.args...), muts: map[uint32]uint8{}}
	for k, v := range in.muts {
		c.muts[k] = v
	}
	return c
}

type result struct {
	ok     bool
	reason string
	ret    uint64
	steps  uint64
	reads  []uint32
	writes []mut
	cover  map[uint32]struct{}
}

type probe struct {
	spec      Spec
	m         *machine.Machine
	d         *dos.DOS
	rng       *rand.Rand
	base      []byte
	baseCPU   cpu.CPU
	seg, off  uint16
	funcLin   uint32
	insns     map[uint32]struct{}
	tables    map[string]resolvedTable
	spLin     uint32 // 堆疊頂的線性位址（初始 SP 之處）
	stackLo   uint32
	stackHi   uint32
	intCount  int
	scratchAt uint32
	readPool  []uint32
	readSeen  map[uint32]struct{}
	covered   map[uint32]struct{}
	curReads  map[uint32]struct{}
	recording bool
}

type resolvedTable struct {
	seg, off uint16
	stride   int
	count    int
}

func parseAddr(s string) (uint16, uint16) {
	p := strings.SplitN(s, ":", 2)
	seg, err := strconv.ParseUint(p[0], 16, 16)
	must(err)
	off, err := strconv.ParseUint(p[1], 16, 16)
	must(err)
	return uint16(seg), uint16(off)
}

func (p *probe) rd16(seg, off uint16) uint16 {
	return uint16(p.base[cpu.Addr(seg, off)&0xFFFFF]) | uint16(p.base[cpu.Addr(seg, off+1)&0xFFFFF])<<8
}

func (p *probe) resolveTables() {
	p.tables = map[string]resolvedTable{}
	for name, t := range p.spec.Tables {
		var rt resolvedTable
		switch {
		case strings.HasPrefix(t.Base, "ds:"):
			off, err := strconv.ParseUint(t.Base[3:], 16, 16)
			must(err)
			rt.seg, rt.off = dgroupSeg, uint16(off)
		case strings.HasPrefix(t.Base, "ptr:"):
			f := strings.Split(t.Base, ":")
			o, err := strconv.ParseUint(f[1], 16, 16)
			must(err)
			s, err := strconv.ParseUint(f[2], 16, 16)
			must(err)
			rt.off = p.rd16(dgroupSeg, uint16(o))
			rt.seg = p.rd16(dgroupSeg, uint16(s))
		default:
			panic("不支援的表基底：" + t.Base)
		}
		rt.stride, rt.count = t.Stride, t.Count
		p.tables[name] = rt
	}
}

func edge(r *rand.Rand, lo, hi int) int {
	switch r.Intn(6) {
	case 0:
		return lo
	case 1:
		return hi
	case 2:
		if lo+1 <= hi {
			return lo + 1
		}
	case 3:
		if hi-1 >= lo {
			return hi - 1
		}
	}
	return lo + r.Intn(hi-lo+1)
}

func (p *probe) genValue(g GenSpec, bits int) uint16 {
	if len(g.Choices) > 0 {
		if !(g.AlsoRandom && p.rng.Intn(2) == 0) {
			return uint16(g.Choices[p.rng.Intn(len(g.Choices))])
		}
		return uint16(p.rng.Intn(1 << bits))
	}
	lo, hi := 0, (1<<bits)-1
	if g.Min != nil {
		lo = *g.Min
	}
	if g.Max != nil {
		hi = *g.Max
	}
	return uint16(edge(p.rng, lo, hi))
}

// genArgs 依描述檔產生一組新的參數與需要寫進記憶體的暫存緩衝區內容。
func (p *probe) genArgs() *input {
	in := &input{muts: map[uint32]uint8{}}
	next := uint32(scratchBase)
	for _, a := range p.spec.Args {
		switch a.Kind {
		case "byte":
			in.args = append(in.args, p.genValue(a.Gen, 8))
		case "reg":
			in.args = append(in.args, p.genValue(a.Gen, 16))
		case "word", "nearptr":
			if a.Kind == "nearptr" && a.Gen.DS != "" {
				v, err := strconv.ParseUint(a.Gen.DS, 16, 16)
				must(err)
				in.args = append(in.args, uint16(v))
			} else {
				in.args = append(in.args, p.genValue(a.Gen, 16))
			}
		case "dword":
			v := uint32(p.genValue(a.Gen, 16)) | uint32(p.genValue(a.Gen, 16))<<16
			in.args = append(in.args, uint16(v), uint16(v>>16))
		case "farptr":
			switch {
			case a.Gen.Null:
				in.args = append(in.args, 0, 0)
			case a.Gen.Table != "":
				t, ok := p.tables[a.Gen.Table]
				if !ok {
					panic("沒有這張表：" + a.Gen.Table)
				}
				idx := p.rng.Intn(t.count)
				in.args = append(in.args, t.off+uint16(idx*t.stride+a.Gen.Offset), t.seg)
			case a.Gen.DS != "":
				v, err := strconv.ParseUint(a.Gen.DS, 16, 16)
				must(err)
				in.args = append(in.args, uint16(v), dgroupSeg)
			case a.Gen.Scratch > 0:
				base := next
				next += uint32((a.Gen.Scratch + 15) &^ 15)
				for i := 0; i < a.Gen.Scratch; i++ {
					var v uint8
					switch p.rng.Intn(4) {
					case 0:
						v = 0
					case 1:
						v = 0xFF
					default:
						v = uint8(p.rng.Intn(256))
					}
					if p.base[base+uint32(i)] != v {
						in.muts[base+uint32(i)] = v
					}
				}
				in.args = append(in.args, 0, uint16(base>>4))
			default:
				panic("farptr 需要 table、ds、scratch 或 null")
			}
		default:
			panic("不支援的參數種類：" + a.Kind)
		}
	}
	return in
}

// run 以輸入呼叫原版函式一次。
func (p *probe) run(in *input) result {
	m := p.m
	copy(m.Mem[:memSize], p.base)
	*m.CPU = p.baseCPU
	for a, v := range in.muts {
		m.Mem[a] = v
	}
	pre := make([]byte, memSize)
	copy(pre, m.Mem[:memSize])

	c := m.CPU
	ss, sp := c.Seg[cpu.SS], c.R[cpu.SP]
	push := func(v uint16) {
		sp -= 2
		m.Write16(cpu.Addr(ss, sp), v)
	}
	// 展開參數：堆疊上的字由右向左推入；kind=reg 的值放進暫存器
	var stackWords []uint16
	type regVal struct {
		idx int
		v   uint16
	}
	var regVals []regVal
	k := 0
	for _, a := range p.spec.Args {
		n := argSize(a)
		if a.Kind == "reg" {
			regVals = append(regVals, regVal{regIndex(a.Reg), in.args[k]})
		} else {
			stackWords = append(stackWords, in.args[k:k+n]...)
		}
		k += n
	}
	for i := len(stackWords) - 1; i >= 0; i-- {
		push(stackWords[i])
	}
	if p.spec.Conv == "far" {
		push(sentinelSeg)
		push(sentinelIP)
	} else {
		push(nearSentinel)
	}
	c.R[cpu.SP] = sp
	c.Seg[cpu.CS], c.IP = p.seg, p.off
	c.Seg[cpu.DS] = dgroupSeg
	c.Flags &^= 0x200
	for _, r := range regVals {
		c.R[r.idx] = r.v
	}

	p.intCount = 0
	logLen := len(m.PortLog)
	var planar [4]uint64
	planar = m.WriteModeUse
	p.curReads = map[uint32]struct{}{}
	p.recording = true
	cover := map[uint32]struct{}{}
	start := m.Steps
	res := result{ok: true}
	for {
		var done bool
		if p.spec.Conv == "far" {
			done = c.Seg[cpu.CS] == sentinelSeg && c.IP == sentinelIP
		} else {
			done = c.Seg[cpu.CS] == p.seg && c.IP == nearSentinel
		}
		if done {
			break
		}
		if m.Steps-start > stepBudget {
			res.ok, res.reason = false, "步數超過上限"
			break
		}
		cover[cpu.Addr(c.Seg[cpu.CS], c.IP)] = struct{}{}
		if err := m.Step(); err != nil {
			res.ok, res.reason = false, "Step 失敗："+err.Error()
			break
		}
	}
	p.recording = false
	res.steps = m.Steps - start
	res.cover = cover
	if res.ok {
		switch {
		case p.intCount > 0:
			res.ok, res.reason = false, "執行了 INT"
		case len(m.PortLog) != logLen:
			res.ok, res.reason = false, "寫了 I/O 埠"
		case m.WriteModeUse != planar:
			res.ok, res.reason = false, "寫了平面視訊記憶體"
		}
	}
	res.ret = p.packRet(c)
	for a := range p.curReads {
		res.reads = append(res.reads, a)
	}
	sort.Slice(res.reads, func(i, j int) bool { return res.reads[i] < res.reads[j] }) // map 走訪順序是隨機的，排序才能決定性
	if res.ok {
		for a := 0; a < memSize; a++ {
			if m.Mem[a] != pre[a] && (uint32(a) < p.stackLo || uint32(a) >= p.stackHi) {
				res.writes = append(res.writes, mut{uint32(a), m.Mem[a]})
			}
		}
	}
	return res
}

func (p *probe) onRead(addr uint32, _ uint8) {
	if p.recording {
		p.curReads[addr] = struct{}{}
	}
}

func (p *probe) learn(res result) {
	for _, a := range res.reads {
		if a >= imageLoad && a < codeEnd {
			continue
		}
		if a >= p.stackLo && a < p.stackHi {
			continue
		}
		if a >= 0xA0000 {
			continue
		}
		if _, ok := p.readSeen[a]; !ok {
			p.readSeen[a] = struct{}{}
			p.readPool = append(p.readPool, a)
		}
	}
}

var edgeBytes = []uint8{0, 1, 2, 0x7F, 0x80, 0xFE, 0xFF}
var edgeWords = []uint16{0, 1, 0xFF, 0x100, 0x7FFF, 0x8000, 0xFFFF, 100, 150, 255, 256, 3000}

func (p *probe) mutate(in *input) *input {
	out := in.clone()
	if p.rng.Intn(5) == 0 {
		out.args = p.genArgs().args
	}
	if len(p.readPool) == 0 {
		return out
	}
	n := 1 + p.rng.Intn(4)
	for i := 0; i < n; i++ {
		a := p.readPool[p.rng.Intn(len(p.readPool))]
		cur := p.base[a]
		if v, ok := out.muts[a]; ok {
			cur = v
		}
		set := func(addr uint32, v uint8) {
			if p.base[addr] == v {
				delete(out.muts, addr)
			} else {
				out.muts[addr] = v
			}
		}
		switch p.rng.Intn(5) {
		case 0:
			set(a, edgeBytes[p.rng.Intn(len(edgeBytes))])
		case 1:
			set(a, uint8(p.rng.Intn(256)))
		case 2:
			if p.rng.Intn(2) == 0 {
				set(a, cur+1)
			} else {
				set(a, cur-1)
			}
		default:
			w := edgeWords[p.rng.Intn(len(edgeWords))]
			if p.rng.Intn(3) == 0 {
				w = uint16(p.rng.Intn(65536))
			}
			set(a, uint8(w))
			set(a+1, uint8(w>>8))
		}
	}
	return out
}

func sigOf(res result) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%016x|", res.ret)
	for _, w := range res.writes {
		fmt.Fprintf(&b, "%x:%x,", w.addr, w.val)
	}
	return b.String()
}

func main() {
	specPath := flag.String("spec", "", "描述檔 JSON")
	outDir := flag.String("out", "", "輸出目錄（vectors.txt 與 base.bin；必須是新的）")
	insnPath := flag.String("insns", "", "指令位址表 insns.tsv（台帳產生），用來算覆蓋率")
	head := flag.String("dosgolem-head", "", "執行器 Git HEAD（由呼叫端核對後傳入）")
	flag.Parse()
	if *specPath == "" || *outDir == "" || *head == "" {
		panic("缺 -spec、-out 或 -dosgolem-head")
	}
	var spec Spec
	raw, err := os.ReadFile(*specPath)
	must(err)
	must(json.Unmarshal(raw, &spec))
	if spec.Iters == 0 {
		spec.Iters = 20000
	}
	if spec.MaxCases == 0 {
		spec.MaxCases = 400
	}
	if spec.Conv == "" {
		spec.Conv = "far"
	}
	game, err := os.ReadFile("/orig/GIN3.COM")
	must(err)
	if digest(game) != "770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918" {
		panic("GIN3.COM 雜湊不符")
	}
	stateRaw, err := os.ReadFile(spec.BaseState)
	must(err)
	if digest(stateRaw) != spec.BaseSHA {
		panic("保存狀態雜湊不符：" + digest(stateRaw))
	}

	p := &probe{spec: spec, rng: rand.New(rand.NewSource(spec.Seed)), readSeen: map[uint32]struct{}{}, covered: map[uint32]struct{}{}}
	p.m = machine.New()
	p.d = dos.New(p.m, "/orig")
	p.d.Install()
	must(state.Load(spec.BaseState, p.m, p.d))
	p.seg, p.off = parseAddr(spec.Entry)
	p.funcLin = cpu.Addr(p.seg, p.off)
	code := append([]byte(nil), p.m.Mem[p.funcLin:p.funcLin+uint32(spec.Size)]...)
	if spec.CodeSHA16 != "" && digest(code)[:16] != spec.CodeSHA16 {
		panic(fmt.Sprintf("函式位元組雜湊不符：%s，描述檔 %s", digest(code)[:16], spec.CodeSHA16))
	}
	p.m.CPU.Flags &^= 0x200
	prev := p.m.CPU.IntHook
	p.m.CPU.IntHook = func(c *cpu.CPU, n uint8) bool {
		p.intCount++
		if prev != nil {
			return prev(c, n)
		}
		return false
	}
	p.baseCPU = *p.m.CPU // 要在 WatchReads 之前擷取：保留取指令的快速路徑，取指令就不會算進讀取集
	p.base = make([]byte, memSize)
	copy(p.base, p.m.Mem[:memSize])
	p.spLin = cpu.Addr(p.baseCPU.Seg[cpu.SS], p.baseCPU.R[cpu.SP])
	p.stackLo, p.stackHi = p.spLin-stackBand, p.spLin
	p.resolveTables()

	p.m.WatchReads(0, 0xFFFFF, p.onRead)

	if *insnPath != "" {
		p.insns = map[uint32]struct{}{}
		f, err := os.Open(*insnPath)
		must(err)
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		want := fmt.Sprintf("%05x", p.funcLin)
		for sc.Scan() {
			cols := strings.Split(sc.Text(), "\t")
			if len(cols) == 2 && cols[0] == want {
				for _, o := range strings.Split(cols[1], ",") {
					v, err := strconv.ParseUint(o, 16, 32)
					must(err)
					p.insns[p.funcLin+uint32(v)] = struct{}{}
				}
			}
		}
		f.Close()
		if len(p.insns) == 0 {
			panic("insns.tsv 沒有這個函式：" + want)
		}
	}

	var recorded []*input
	var results []result
	seenSig := map[string]struct{}{}
	corpus := []*input{}
	discarded := map[string]int{}
	covOf := func() int {
		if p.insns == nil {
			return len(p.covered)
		}
		n := 0
		for a := range p.insns {
			if _, ok := p.covered[a]; ok {
				n++
			}
		}
		return n
	}

	plain := 0 // 不新穎（沒有新覆蓋、輸出簽章重複）但仍記錄的案數，保留對突變的鑑別力
	for it := 0; it < spec.Iters && len(recorded) < 2*spec.MaxCases && plain < spec.MaxCases; it++ {
		var in *input
		if len(corpus) == 0 || p.rng.Intn(6) == 0 {
			in = p.genArgs()
			if len(corpus) > 0 && p.rng.Intn(2) == 0 {
				in = p.mutate(in)
			}
		} else {
			in = p.mutate(corpus[p.rng.Intn(len(corpus))])
		}
		res := p.run(in)
		if !res.ok {
			discarded[res.reason]++
			continue
		}
		p.learn(res)
		newCover := false
		for a := range res.cover {
			if a >= p.funcLin && a < p.funcLin+uint32(spec.Size) {
				if _, ok := p.covered[a]; !ok {
					p.covered[a] = struct{}{}
					newCover = true
				}
			}
		}
		sig := sigOf(res)
		_, seen := seenSig[sig]
		if newCover || !seen {
			seenSig[sig] = struct{}{}
			recorded = append(recorded, in)
			results = append(results, res)
			corpus = append(corpus, in)
		} else if plain < spec.MaxCases {
			plain++
			recorded = append(recorded, in)
			results = append(results, res)
		}
	}

	cov := covOf()
	total := len(p.insns)
	if p.insns == nil {
		total = 0
	}
	pct := -1.0
	if total > 0 {
		pct = float64(cov) * 100 / float64(total)
	}
	if spec.MinCoverage > 0 && pct < spec.MinCoverage {
		fmt.Printf("警告：指令覆蓋率 %.1f%% 低於描述檔要求的 %.1f%%\n", pct, spec.MinCoverage)
	}

	must(os.MkdirAll(*outDir, 0o755))
	basePath := *outDir + "/base.bin"
	must(os.WriteFile(basePath, p.base, 0o644))
	vec, err := os.OpenFile(*outDir+"/vectors.txt", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	must(err)
	w := bufio.NewWriter(vec)
	fmt.Fprintf(w, "H func=%s nargs=%d ret=%s base=base.bin base_sha256=%s stack_lo=%x stack_hi=%x\n",
		spec.ID, argWords(spec), spec.Ret, digest(p.base), p.stackLo, p.stackHi)
	fmt.Fprintf(w, "# schema decomp.diff-vectors.v2\n# dosgolem_head %s\n# state_sha256 %s\n# game_sha256 770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918\n",
		*head, spec.BaseSHA)
	fmt.Fprintf(w, "# entry %s conv %s size %d code_sha256_16 %s\n", spec.Entry, spec.Conv, spec.Size, digest(code)[:16])
	fmt.Fprintf(w, "# seed %d iters %d recorded %d\n", spec.Seed, spec.Iters, len(recorded))
	fmt.Fprintf(w, "# coverage %d/%d instructions (%.1f%%)\n", cov, total, pct)
	reasons := make([]string, 0, len(discarded))
	for k, v := range discarded {
		reasons = append(reasons, fmt.Sprintf("%s=%d", k, v))
	}
	sort.Strings(reasons)
	fmt.Fprintf(w, "# discarded %s\n", strings.Join(reasons, " "))
	for i, in := range recorded {
		res := results[i]
		args := make([]string, len(in.args))
		for k, a := range in.args {
			args[k] = fmt.Sprintf("%x", a)
		}
		argStr := strings.Join(args, ",")
		if argStr == "" {
			argStr = "-"
		}
		fmt.Fprintf(w, "C %d %s %d %s %016x %d %s %d\n", i, argStr, len(in.muts), pairs(in.muts), res.ret,
			len(res.writes), wpairs(res.writes), res.steps)
	}
	must(w.Flush())
	must(vec.Close())
	fmt.Printf("%s：記錄 %d 案（迭代上限 %d），指令覆蓋 %d/%d (%.1f%%)，丟棄 %s\n",
		spec.ID, len(recorded), spec.Iters, cov, total, pct, strings.Join(reasons, " "))
}

func argSize(a ArgSpec) int {
	switch a.Kind {
	case "farptr", "dword":
		return 2
	}
	return 1
}

func argWords(s Spec) int {
	n := 0
	for _, a := range s.Args {
		n += argSize(a)
	}
	return n
}

func regIndex(name string) int {
	switch name {
	case "ax":
		return cpu.AX
	case "cx":
		return cpu.CX
	case "dx":
		return cpu.DX
	case "bx":
		return cpu.BX
	case "si":
		return cpu.SI
	case "di":
		return cpu.DI
	case "bp":
		return cpu.BP
	}
	panic("不支援的暫存器：" + name)
}

// packRet 依描述檔的 ret 取回傳值：ax、dxax，或 pack:ax,bx,…（由高位到低位串成一個值）。
func (p *probe) packRet(c *cpu.CPU) uint64 {
	switch {
	case p.spec.Ret == "ax":
		return uint64(c.R[cpu.AX])
	case p.spec.Ret == "dxax":
		return uint64(c.R[cpu.DX])<<16 | uint64(c.R[cpu.AX])
	case strings.HasPrefix(p.spec.Ret, "pack:"):
		var v uint64
		for _, n := range strings.Split(p.spec.Ret[5:], ",") {
			if n == "flags" {
				// 只取跳躍條件用得到的四個旗標：OF SF ZF CF
				v = v<<16 | uint64(c.Flags&0x08C1)
				continue
			}
			v = v<<16 | uint64(c.R[regIndex(n)])
		}
		return v
	}
	return 0
}

func pairs(m map[uint32]uint8) string {
	if len(m) == 0 {
		return "-"
	}
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%x:%x", k, m[uint32(k)])
	}
	return strings.Join(parts, ",")
}

func wpairs(w []mut) string {
	if len(w) == 0 {
		return "-"
	}
	parts := make([]string, len(w))
	for i, x := range w {
		parts[i] = fmt.Sprintf("%x:%x", x.addr, x.val)
	}
	return strings.Join(parts, ",")
}
