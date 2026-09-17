package game

import (
	"encoding/binary"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

// 駐留命令走到目的地 → 整支解散（`113A:06A6` → `297B:025D`，`docs/re/329`）；
// 收據 `tools/fleetcmd_receipts.sh` 第 21 段。
//
//   - 亞爾提那：`fcmd-garrison-done` 按實行 → `fcmd-garrison-t1`（#665M）。
//   - 就地（瓦爾哈拉）：`fcmd-garrison-self-done` → `fcmd-garrison-self-t1`。
//
// 回合裡的「已經駐留在星系上。」那一張過場（`-report`，#570M）也整張比（規格 059）。
func TestGarrisonArrivalDisbandsMatchesOracle(t *testing.T) {
	for _, c := range []struct {
		name         string
		pickX, pickY int
		region       int
		batch        string
	}{
		{"亞爾提那", 85, 128, 2, "fcmd-garrison"},
		{"就地", 32, 72, 0, "fcmd-garrison-self"},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := reachFleetWindow(t)
			g.Click(490, 345) // 「作戰」
			g.Click(497, 352) // 「駐留」
			g.Click(c.pickX, c.pickY)
			g.Click(432, 263) // 「決定」
			if g.State() != StateCommandPhase || len(g.fleets) != 1 ||
				g.fleets[0].Move != fleetActionGarrison || g.fleets[0].Target != c.region {
				t.Fatalf("決定之後在 %v、艦隊 %+v", g.State(), g.fleets)
			}
			matches(t, g, c.batch+"-done")
			stock := rules.RegionShips(g.regionRecord(c.region), 0)[0]
			pressExecute(g)
			g.Click(560, 125) // 出發前往 → 已經駐留在星系上。
			if g.State() != StateTurnWindow || len(g.fleets) == 0 || g.fleets[0].Faction != 0 {
				t.Fatalf("「已經駐留」那一張：%v、艦隊 %d 支；原版等完之後才解散", g.State(), len(g.fleets))
			}
			matches(t, g, c.batch+"-report")
			finishTurn(t, g)
			own := 0
			for _, f := range g.fleets {
				if f.Faction == 0 {
					own++
				}
			}
			if own != 0 {
				t.Fatalf("實行之後帝國還有 %d 支艦隊，原版整支解散（+0x04 FF、ds:8F49 帝國 0）", own)
			}
			if n := receiptPeek(t, c.batch+"-t1", "lin:5DA79"); n[0] != 0 || int(n[1]) != len(g.fleets) {
				t.Fatalf("艦隊數 帝國 0／同盟 %d，原版 %d／%d", len(g.fleets), n[0], n[1])
			}
			batch := c.batch + "-t1"
			reg := receiptPeek(t, batch, "lin:5F2CC")
			o := c.region * rules.RegionRecord
			want := reg[o : o+rules.RegionRecord]
			got := g.regionRecord(c.region)
			if got[0x28] != want[0x28] || got[0x29] != want[0x29] {
				t.Fatalf("星域 %d 艦隊數 %d／%d，原版 %d／%d", c.region, got[0x28], got[0x29], want[0x28], want[0x29])
			}
			if w := int(binary.LittleEndian.Uint32(want[0x30:])); rules.RegionShips(got, 0)[0] != w {
				t.Fatalf("星域 %d 帝國戰艦存量 %d → %d，原版 %d", c.region, stock, rules.RegionShips(got, 0)[0], w)
			}
			for _, a := range []struct {
				idx  int
				addr string
			}{{0, "lin:602AC"}, {32, "lin:6062C"}} {
				w := receiptPeek(t, batch, a.addr)
				if r := g.admiralRecord(a.idx); r[9] != w[9] || r[10] != w[10] {
					t.Fatalf("提督 %d +09/+0A ＝ %02x/%02x，原版 %02x/%02x", a.idx, r[9], r[10], w[9], w[10])
				}
			}
			matches(t, g, batch)
		})
	}
}

// receiptPeek 讀收據 report.txt 裡某個 `-peek` 位址倒出來的位元組。
func receiptPeek(t *testing.T, batch, addr string) []byte {
	t.Helper()
	r, err := os.ReadFile(filepath.Join(root, "workplace/oracle", batch, "report.txt"))
	if err != nil {
		t.Skipf("沒有收據 %s，跳過", batch)
	}
	for _, line := range strings.Split(string(r), "\n") {
		if !strings.Contains(line, addr+" ") {
			continue
		}
		hexPart := line[strings.LastIndex(line, "）")+len("）"):]
		out, err := hex.DecodeString(strings.ReplaceAll(strings.TrimSpace(hexPart), " ", ""))
		if err != nil {
			t.Fatalf("收據 %s 的 %s 讀不成 hex：%v", batch, addr, err)
		}
		return out
	}
	t.Fatalf("收據 %s 的 report.txt 沒有 %s", batch, addr)
	return nil
}
