package game

import (
	"fmt"

	"github.com/wicanr2/logh3/internal/ai"
	"github.com/wicanr2/logh3/internal/rules"
)

// enhancedBrain 只替換決策；建隊與整補仍用原版規則（規格 068）。
type enhancedBrain struct {
	*ai.Focused
	builder ai.Builder
}

func (b *enhancedBrain) Build(w ai.World, faction, turn int) []ai.NewFleet {
	if b.builder == nil {
		return nil
	}
	return b.builder.Build(w, faction, turn)
}

func (*enhancedBrain) OriginalRefit() {}

// SetAILevel 設定執行期偏好：-1 原版，0–3 增強。偏好不進原版存檔。
func (g *Game) SetAILevel(level int) error {
	if level < -1 || level > 3 {
		return fmt.Errorf("電腦 AI 難度必須為 -1 或 0–3：%d", level)
	}
	g.aiMode = level + 1 // 零值是原版，不影響未選強化版的既有入口。
	g.reconnectAI()
	if level < 0 {
		g.SetTactic(ai.OriginalTactic{}, false)
	}
	return nil
}

// AILevel 回執行期偏好，供入口與正常玩家路徑驗收。
func (g *Game) AILevel() int { return g.aiMode - 1 }

// reconnectAI 在新局／載入後重建資料閉包與暫態，不丟失玩家選項。
func (g *Game) reconnectAI() {
	original := g.OriginalBrain()
	if g.aiMode == 0 {
		g.brain = original
		return
	}
	builder, _ := original.(ai.Builder)
	g.brain = &enhancedBrain{Focused: &ai.Focused{Difficulty: g.AILevel()}, builder: builder}
	g.SetTactic(&ai.Encircle{Difficulty: g.AILevel()}, false)
}

// liveFleetStrength 套原版權重於現況編制；保存底稿與舊快取不能取代戰損。
func (g *Game) liveFleetStrength(f rules.Fleet) int {
	if len(f.Composition) > 0 && len(g.res.Weights) > 0 {
		n := 0
		for _, gp := range f.Composition {
			if gp.Kind >= 0 && gp.Kind < len(g.res.Weights) && gp.Full > 0 {
				n += g.res.Weights[gp.Kind] * gp.Ships * 1000 / gp.Full
			}
		}
		return n
	}
	if f.Strength > 0 {
		return f.Strength
	}
	return 1 // 沿用欠缺資料的診斷退路，不製造正式編制。
}
