package rules

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// decomp 向量檔（decomp/harness/diffprobe.go 產生）的載入器。
//
// 向量檔記錄「原版函式在某個記憶體狀態下的行為」：基底傾印、對基底的位元組修改、參數，
// 以及原版的回傳值與堆疊以外的全部記憶體寫入。remake 的測試把同一個記憶體狀態轉成
// 自己的型別，呼叫 remake 的實作，再與原版的紀錄比對。
// 沒有設環境變數就 skip，不以自製資料代替（CLAUDE.md「缺檔就 skip」）。

type decompCase struct {
	id     int
	args   []uint16
	muts   map[uint32]byte
	ret    uint64
	writes map[uint32]byte
}

type decompVectors struct {
	funcID  string
	retKind string
	base    []byte
	cases   []decompCase
}

// loadDecompVectors 讀環境變數 envName 指的向量檔；沒設就 skip。
func loadDecompVectors(t *testing.T, envName string) *decompVectors {
	t.Helper()
	path := os.Getenv(envName)
	if path == "" {
		t.Skipf("未設 %s；向量檔由 decomp/tools/diff_run.sh 產生", envName)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	v := &decompVectors{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<26)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "H "):
			for _, tok := range strings.Fields(line)[1:] {
				kv := strings.SplitN(tok, "=", 2)
				if len(kv) != 2 {
					continue
				}
				switch kv[0] {
				case "func":
					v.funcID = kv[1]
				case "ret":
					v.retKind = kv[1]
				case "base":
					raw, err := os.ReadFile(filepath.Join(filepath.Dir(path), kv[1]))
					if err != nil {
						t.Fatal(err)
					}
					v.base = raw
				}
			}
		case strings.HasPrefix(line, "C "):
			fs := strings.Fields(line)
			c := decompCase{muts: map[uint32]byte{}, writes: map[uint32]byte{}}
			c.id, _ = strconv.Atoi(fs[1])
			if fs[2] != "-" {
				for _, a := range strings.Split(fs[2], ",") {
					n, _ := strconv.ParseUint(a, 16, 16)
					c.args = append(c.args, uint16(n))
				}
			}
			parsePairs(t, fs[4], c.muts)
			c.ret, _ = strconv.ParseUint(fs[5], 16, 64)
			parsePairs(t, fs[7], c.writes)
			v.cases = append(v.cases, c)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if v.base == nil || len(v.cases) == 0 {
		t.Fatalf("%s 沒有基底傾印或沒有案例", path)
	}
	return v
}

func parsePairs(t *testing.T, s string, out map[uint32]byte) {
	t.Helper()
	if s == "-" {
		return
	}
	for _, p := range strings.Split(s, ",") {
		kv := strings.SplitN(p, ":", 2)
		a, err1 := strconv.ParseUint(kv[0], 16, 32)
		b, err2 := strconv.ParseUint(kv[1], 16, 8)
		if err1 != nil || err2 != nil {
			t.Fatalf("向量檔格式錯誤：%q", p)
		}
		out[uint32(a)] = byte(b)
	}
}

// decompMem 是一個案例的 1 MB 記憶體：基底傾印加上該案例的位元組修改。
type decompMem struct{ b []byte }

func (v *decompVectors) mem(c decompCase) *decompMem {
	m := &decompMem{b: append([]byte(nil), v.base...)}
	for a, x := range c.muts {
		m.b[a] = x
	}
	return m
}

// lin 是 8086 的 seg:off，位移 16 位元回繞，線性位址遮到 20 位元。
func decompLin(seg, off uint16) uint32 { return (uint32(seg)<<4 + uint32(off)) & 0xFFFFF }

func (m *decompMem) u8(seg, off uint16) byte { return m.b[decompLin(seg, off)] }
func (m *decompMem) u16(seg, off uint16) uint16 {
	return uint16(m.u8(seg, off)) | uint16(m.u8(seg, off+1))<<8
}

// sortedKeys 讓失敗訊息穩定。
func sortedKeys(m map[uint32]byte) []uint32 {
	k := make([]uint32, 0, len(m))
	for a := range m {
		k = append(k, a)
	}
	sort.Slice(k, func(i, j int) bool { return k[i] < k[j] })
	return k
}
