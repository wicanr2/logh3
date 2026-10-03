package game

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/ai"
)

// 原版完整正常回合收據，不以remake內部自洽當作原版時點證明。
func TestTurnCounterNormalCommandMatchesOriginal(t *testing.T) {
	m, err := os.ReadFile(filepath.Join(root, "workplace/oracle/normal89-r47-clock-a/command.bin"))
	if err != nil {
		t.Skipf("缺第47輪原版完整RAM：%v", err)
	}
	if len(m) != 1048576 {
		t.Fatal("原版RAM長度錯誤")
	}
	g := reachFleetDecided(t)
	if g.turn != 0 || g.rand94DC().Seed() != 0xD3DC167E {
		t.Fatal("原版與remake實行前回合／seed不同")
	}
	executeTurn(t, g)
	want := int(binary.LittleEndian.Uint16(m[0x5DA48:]))
	if g.state != StateCommandPhase || g.turn != want || want != 1 {
		t.Fatalf("正常尾段回合不同：state=%v turn=%d want=%d", g.state, g.turn, want)
	}
	dir := t.TempDir()
	if err := g.SaveSlot(dir, 0); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "WORK.DT0"))
	if err != nil || int(binary.LittleEndian.Uint16(raw[4:])) != want {
		t.Fatalf("正式保存WORK+04不同：%v", err)
	}
	if raw[battleWorkActive] != m[0x5DA44+battleWorkActive] ||
		raw[battleWorkRegion] != m[0x5DA44+battleWorkRegion] {
		t.Fatal("正常無戰鬥尾段的旗標／掃描星域不同")
	}
}

type counterObservedBrain struct {
	ai.Brain
	turns []int
}

func (b *counterObservedBrain) Orders(w ai.World, faction, turn int) []ai.Order {
	b.turns = append(b.turns, turn)
	return b.Brain.Orders(w, faction, turn)
}

func (b *counterObservedBrain) Build(w ai.World, faction, turn int) []ai.NewFleet {
	b.turns = append(b.turns, turn)
	if builder, ok := b.Brain.(ai.Builder); ok {
		return builder.Build(w, faction, turn)
	}
	return nil
}

// 保留原版整補標記，觀測不能改變所包覆AI的能力。
type counterObservedRefitter struct{ *counterObservedBrain }

func (*counterObservedRefitter) OriginalRefit() {}

func TestTurnCounterBothAIModesUseCurrentRound(t *testing.T) {
	for _, level := range []int{-1, 1} {
		for faction := 0; faction < 2; faction++ {
			t.Run(fmt.Sprintf("level%d/faction%d", level, faction), func(t *testing.T) {
				g := reachConfiguredAI(t, level, faction)
				g.turn = 3 // 明示回合參數邊界，非原版局面對拍。
				g.mutableWork()[0x32] = 0
				observer := &counterObservedBrain{Brain: g.brain}
				g.brain = observer
				if _, ok := observer.Brain.(ai.Refitter); ok {
					g.brain = &counterObservedRefitter{observer}
				}
				g.beginTurn()
				g.runBrain()
				if g.turn != 3 || len(observer.turns) != 2 {
					t.Fatalf("AI前進位或未呼叫建隊／命令：turn=%d calls=%v", g.turn, observer.turns)
				}
				for _, turn := range observer.turns {
					if turn != 3 {
						t.Fatalf("AI讀取下一回合：%v", observer.turns)
					}
				}
			})
		}
	}
}
