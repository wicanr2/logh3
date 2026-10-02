package game

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// 非零邊界由原版 cmp byte,0 決定。正常S3另由正式bootstrap探測驗證。
func TestNormalAIGateConsumesOnceAndPreservesSave(t *testing.T) {
	for _, level := range []int{-1, 1} {
		for _, flag := range []byte{1, 2, 255} {
			t.Run(fmt.Sprintf("level%d/flag%d", level, flag), func(t *testing.T) {
				g := reachConfiguredAI(t, level, 0)
				g.mutableWork()[0x32] = flag // 明示的邊界測試，非自然S3。
				beforeWork := bytes.Clone(g.workRecord())
				beforeWork[0x32] = 0
				seed, commands := g.rand94DC().Seed(), aiFleetCommands(g, 1)
				stellar, military := bytes.Clone(g.res.Stellar), g.military
				g.runBrain()
				if !bytes.Equal(g.workRecord(), beforeWork) || g.rand94DC().Seed() != seed ||
					!reflect.DeepEqual(commands, aiFleetCommands(g, 1)) ||
					!bytes.Equal(stellar, g.res.Stellar) || g.military != military {
					t.Fatal("一次性跳過改變AI資料或未只清除旗標")
				}
				dir := t.TempDir()
				if err := g.SaveSlot(dir, 0); err != nil {
					t.Fatal(err)
				}
				work, err := os.ReadFile(filepath.Join(dir, "WORK.DT0"))
				if err != nil || work[0x32] != 0 {
					t.Fatalf("消費狀態未保存：%v", err)
				}
				g.res.SaveDir = dir
				g.Click(544, 125)
				g.Click(560, 145)
				g.Click(300, 122)
				wantState(t, g, StateLoadFaction, "存檔列後選陣營")
				g.Click(318, 208)
				wantState(t, g, StateCommandPhase, "消費旗標正式存讀")
				if g.workRecord()[0x32] != 0 || g.AILevel() != level {
					t.Fatal("載入恢復跳過旗標或丟失AI模式")
				}
				g.runBrain()
				if reflect.DeepEqual(commands, aiFleetCommands(g, 1)) {
					t.Fatal("第二次AI呼叫未恢復建隊／命令")
				}
			})
		}
	}
}

func TestNormalAIGatePreservedWithoutComputer(t *testing.T) {
	for _, reason := range []string{"versus", "no-brain", "no-graph", "no-planets"} {
		t.Run(reason, func(t *testing.T) {
			g := reachConfiguredAI(t, -1, 0)
			g.mutableWork()[0x32] = 1
			switch reason {
			case "versus":
				g.versus = true
			case "no-brain":
				g.brain = nil
			case "no-graph":
				g.res.Graph = nil
			case "no-planets":
				g.res.Planets = nil
			}
			g.runBrain()
			if g.workRecord()[0x32] != 1 {
				t.Fatal("沒有可執行AI卻消費旗標")
			}
		})
	}
}
