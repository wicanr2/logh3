package game

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// 098共用等待predicate診斷。刻意不把dosgolem步數映射成Game動畫tick。
func TestEngagementWaitNative(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/fleet-wait.json"))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != "8ff102acff0607b23524e0b8678fc3eef3817aa171a6ce779d5d58c56392c566" {
		t.Fatal("等待原版收據雜湊不符")
	}
	var receipt struct {
		Native struct {
			Cases []struct {
				Name     string
				Scenario struct {
					Input, Phase string
					Duration     int
				}
				Polls []struct {
					Event int `json:"event_word"`
				}
				TimerAfter int `json:"timer_after"`
				AX         int `json:"return_ax"`
			}
		}
	}
	if err := json.Unmarshal(b, &receipt); err != nil {
		t.Fatal(err)
	}
	if len(receipt.Native.Cases) != 11 {
		t.Fatal("等待案例數不符")
	}
	for _, c := range receipt.Native.Cases {
		t.Run(c.Name, func(t *testing.T) {
			g := reachEngagementFixture(t)
			if c.Scenario.Duration >= 0 {
				g.display.MsgSpeed = c.Scenario.Duration
			}
			_, result := g.engagementTimeline(g.engQueue[0])
			accept := false
			for _, p := range c.Polls {
				accept = accept || p.Event == 1 || p.Event == 2
			}
			if c.TimerAfter != 0 || c.AX != 1 {
				t.Fatal("原版返回契約不符")
			}
			send := func() {
				switch c.Scenario.Input {
				case "left":
					g.Click(300, 240)
				case "right":
					g.RightClick(300, 240)
				case "space":
					g.Key(' ')
				case "escape":
					g.Key(27)
				}
			}
			if c.Scenario.Phase == "animation" {
				g.engagementTick(result - 1)
				send()
				if g.state != StateEngagement {
					t.Fatal("演出中輸入提前退出")
				}
				g.engagementTick(1)
			} else {
				g.engagementTick(result)
				if c.Scenario.Input != "" {
					send()
				}
			}
			if accept || g.display.MsgSpeed == 0 {
				if g.state != StateBattle {
					t.Fatal("原版接受的等待未返回")
				}
				return
			}
			if g.state != StateEngagement {
				t.Fatal("原版不接受的輸入縮短等待")
			}
			g.engagementTick(g.display.MsgSpeed - 1)
			if g.state != StateEngagement {
				t.Fatal("倒數提前返回")
			}
			g.engagementTick(1)
			if g.state != StateBattle {
				t.Fatal("倒數滿未返回")
			}
		})
	}
}
