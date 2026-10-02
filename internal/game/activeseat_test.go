package game

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/rules"
)

func activeSeatOriginal(t *testing.T, kind, stage string) []byte {
	t.Helper()
	m, err := os.ReadFile(filepath.Join(root, "workplace/oracle/normal89-r48-"+kind+"-a", stage+".bin"))
	if err != nil {
		t.Skipf("缺第48輪原版活動座位收據：%v", err)
	}
	if len(m) != 1048576 {
		t.Fatal("原版完整RAM長度錯誤")
	}
	return m[0x5DA4D:0x5DA4F]
}

func requireActiveSeat(t *testing.T, g *Game, original []byte) {
	t.Helper()
	if !bytes.Equal(g.workRecord()[0x09:0x0B], original) {
		t.Fatalf("活動陣營／座位=%v，原版=%v", g.workRecord()[0x09:0x0B], original)
	}
}

func TestActiveSeatNormalLifecycleMatchesOriginal(t *testing.T) {
	g := reachFleetDecided(t)
	requireActiveSeat(t, g, activeSeatOriginal(t, "clock", "initial"))
	g.Click(544, 125)
	g.Click(560, 125)
	wantState(t, g, StateTurnThinking, "正常實行思考")
	requireActiveSeat(t, g, activeSeatOriginal(t, "clock", "thinking"))
	for i := 0; i < 1000 && g.state == StateTurnThinking; i++ {
		g.Tick(time.Second / 60)
	}
	wantState(t, g, StateCommandPhase, "完整回合返回命令")
	requireActiveSeat(t, g, activeSeatOriginal(t, "clock", "command"))
}

// 正常移動保留處理畫面；只核對活動控制邊界，不宣稱與clock局面同態。
func TestActiveSeatNormalProcessingBoundary(t *testing.T) {
	g := reachFleetTargetPick(t)
	g.Click(85, 128)
	g.Click(432, 263)
	pressExecute(g)
	wantState(t, g, StateTurnWindow, "正常艦隊處理")
	requireActiveSeat(t, g, activeSeatOriginal(t, "clock", "processing"))
	if g.faction != 0 || g.turn != 0 {
		t.Fatal("控制上下文改了玩家視角或提早進位")
	}
	g.Click(560, 125)
	g.Click(560, 125)
	wantState(t, g, StateCommandPhase, "移動處理後命令")
	requireActiveSeat(t, g, activeSeatOriginal(t, "clock", "command"))
}

func TestActiveSeatVersusHandoffMatchesOriginal(t *testing.T) {
	g := reachLoadFaction(t)
	g.Click(318, 248)
	requireActiveSeat(t, g, activeSeatOriginal(t, "versus", "loaded"))
	g.Click(432, 270)
	requireActiveSeat(t, g, activeSeatOriginal(t, "versus", "empire"))
	g.Click(544, 125)
	g.Click(560, 125)
	wantState(t, g, StateVersusInput, "同盟交接")
	requireActiveSeat(t, g, activeSeatOriginal(t, "versus", "handoff"))
	g.Click(432, 270)
	requireActiveSeat(t, g, activeSeatOriginal(t, "versus", "alliance"))
	dir := t.TempDir()
	if err := g.SaveSlot(dir, 0); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "WORK.DT0"))
	if err != nil || !bytes.Equal(raw[0x09:0x0B], activeSeatOriginal(t, "versus", "alliance")) {
		t.Fatalf("同盟正式保存活動控制不同：%v", err)
	}
}

func TestActiveSeatBothAIModesPreservePlayer(t *testing.T) {
	for _, level := range []int{-1, 1} {
		for faction := 0; faction < 2; faction++ {
			t.Run(fmt.Sprintf("level%d/faction%d", level, faction), func(t *testing.T) {
				g := reachConfiguredAI(t, level, faction)
				g.turn = 3 // 明示控制上下文邊界，不當作原版局面。
				g.mutableWork()[0x32] = 0
				commands := aiFleetCommands(g, 1-faction)
				g.beginTurn()
				requireActiveSeat(t, g, []byte{byte(1 - faction), 1})
				g.runBrain()
				requireActiveSeat(t, g, []byte{byte(1 - faction), 1})
				if g.faction != faction || g.turn != 3 || g.AILevel() != level {
					t.Fatal("活動控制改變玩家、日期或AI選擇")
				}
				if reflect.DeepEqual(commands, aiFleetCommands(g, 1-faction)) {
					t.Fatal("實際AI未產生建隊／命令")
				}
			})
		}
	}
}

func TestActiveSeatPreservesUnknownAndResourceTemplate(t *testing.T) {
	w := bytes.Repeat([]byte{0xA5}, rules.WorkRecord)
	g := &Game{res: &Resources{WorkTemplate: w}, faction: 1}
	want := bytes.Clone(w)
	want[0x09], want[0x0A] = 1, 0
	g.setHumanSeat()
	if !bytes.Equal(g.workRecord(), want) || !bytes.Equal(w, bytes.Repeat([]byte{0xA5}, rules.WorkRecord)) {
		t.Fatal("未知欄位或原始資源被覆寫")
	}
	for _, seat := range []int{-1, 2} {
		g.setActiveSeat(seat)
		if !bytes.Equal(g.workRecord(), want) {
			t.Fatal("非法座位改了控制")
		}
	}
	g.faction = 2
	g.setActiveSeat(0)
	g.setProcessingFaction()
	if !bytes.Equal(g.workRecord(), want) {
		t.Fatal("非法玩家陣營改了控制")
	}
	for _, broken := range []*Game{{}, {res: &Resources{WorkTemplate: []byte{1}}}} {
		broken.setHumanSeat()
		broken.setProcessingFaction()
		if len(broken.work) != 0 {
			t.Fatal("缺WORK仍創造替代控制資料")
		}
	}
}
