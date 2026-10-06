package game

import (
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
)

// 星域易主的戰略報告（原版 `056A:0855`，`docs/spec/034`、`docs/re/246`、
// `docs/re/247`）。
//
// **它是模態的，而且擋住後面所有處理。** 原版在回合結束逐格重算擁有者
// 那一圈裡，每有一格易主就跳一則；路徑表重建與 32 格態勢重算要等它按掉
// 才會跑（收據 `chain3`：卡住時 `2703:08DE` 零次執行）。
// 滑鼠任一鍵可關，鍵盤不行（收據 `chain4`／`chain5`／`chain3`）。

// RegionFlip 是一則還沒給玩家看的易主報告。
type RegionFlip struct {
	Region   int // 哪一格
	OldOwner int // 易主前的擁有者
	NewOwner int // 易主後的擁有者（0 帝國、1 同盟、2 無主）
}

// FlipMessage 是報告第二行要印的那句話（`docs/re/247` 的訊息表）。
type FlipMessage int

// 四句話。四份收據各釘一句（`chain7`／`chain3`／`chain6`／`chain8`），
// 選法也從 `056A:089D` 讀出來了（`docs/re/249`、`250`）。
const (
	FlipNone           FlipMessage = iota
	FlipTakenByUs                  // ds:08DA　已被我方占領。
	FlipLostToEnemy                // ds:08EE　已被敵軍占領。
	FlipOurGripLost                // ds:088B　…部分地區已經無法掌控。
	FlipWePartlyLanded             // ds:08BC　部分行星我軍已成功進駐。
)

// Message 回這一則對 `seat` 這個座位要印哪一句。
//
// 四條分支照原版 `056A:089D`（`docs/re/250`）：新擁有者是無主時看
// 「**舊**擁有者是不是自己」，否則看「**新**擁有者是不是自己」。
func (f RegionFlip) Message(seat int) FlipMessage {
	switch {
	case f.NewOwner == seat:
		return FlipTakenByUs
	case f.NewOwner != rules.RegionOwnerNone:
		return FlipLostToEnemy
	case f.OldOwner == seat:
		return FlipOurGripLost
	default:
		return FlipWePartlyLanded
	}
}

// PendingFlips 回還沒看完的易主報告。
func (g *Game) PendingFlips() []RegionFlip { return g.pendingFlips }

// CurrentFlip 回現在正在顯示的那一則；沒有就回 false。
func (g *Game) CurrentFlip() (RegionFlip, bool) {
	if len(g.pendingFlips) == 0 {
		return RegionFlip{}, false
	}
	return g.pendingFlips[0], true
}

// FlipRegionName 回這一則報告要印的星域名（Big5 字碼）。
func (g *Game) FlipRegionName(f RegionFlip) []uint16 {
	rec := g.regionRecord(f.Region)
	if rec == nil {
		return nil
	}
	return rules.RegionName(rec)
}

// dismissFlip 收掉最前面那一則。全部看完之後才把後半段跑完
// ——重建路徑表、32 格態勢，然後接回原本的流程。
func (g *Game) dismissFlip() {
	g.keepMapMark()
	if f, ok := g.CurrentFlip(); ok && f.NewOwner != rules.RegionOwnerNone {
		// 原版在組報告時就把 `+0x13` bit4 清掉（`056A:089D` 的
		// `and 0EFh`），而且只在「新擁有者不是無主」那一路
		// （`docs/re/250`）。之後 `refreshRegionPosture` 還會重推一次。
		if rec := g.regionRecord(f.Region); rec != nil {
			rec[regionFlagsOff] &^= 0x10
		}
	}
	if len(g.pendingFlips) > 0 {
		g.pendingFlips = g.pendingFlips[1:]
	}
	if len(g.pendingFlips) > 0 {
		return
	}
	g.refreshRegionPosture()
	g.resumeAfterFlips()
}

// resumeAfterFlips 接回被報告打斷的那條路。
func (g *Game) resumeAfterFlips() {
	tail := g.afterFlips
	g.afterFlips = nil
	if tail != nil {
		tail()
		return
	}
	g.state = StateCommandPhase
}

// drawFlipReport 把現在這一則報告畫上去：兩行字 ＋ 頭像
// （版面 `docs/internals/data/flip-report-panel.json`，`docs/spec/034`）。
func (g *Game) drawFlipReport(fb *gfx.Framebuffer) {
	p := g.res.FlipReport
	f, ok := g.CurrentFlip()
	if p == nil || !ok {
		return
	}
	p.Draw(fb, g.res.Font, g.FlipRegionName(f), f.NewOwner,
		flipMessageKey(f.Message(g.faction)))
	g.drawReportFace(fb, g.faction)
}

// drawReportFace 畫戰略報告共用的情報參謀頭像。易主報告的索引是座位
// 陣營；叛亂報告則是發生叛亂前的星域擁有者（`docs/spec/034`、`046`）。
func (g *Game) drawReportFace(fb *gfx.Framebuffer, owner int) {
	p := g.res.FlipReport
	if p == nil {
		return
	}
	if g.res.Faces == nil {
		return
	}
	// 頭像是 `FACE.AMG` 第 `ds:8F4D[座位陣營]` 張，**只畫左上 72×93**
	// ——整張是 80×100，畫上去的是裁切過的（`docs/re/248`）。
	n := 0
	if owner >= 0 && owner < len(g.res.Reporter) {
		n = g.res.Reporter[owner]
	}
	img, err := g.res.Faces.Decode(n)
	if err != nil {
		return
	}
	img.Draw(fb, p.Face.X, p.Face.Y, p.Face.W, p.Face.H)
}

// flipMessageKey 把四句對應到版面 JSON 裡的鍵。
func flipMessageKey(m FlipMessage) string {
	switch m {
	case FlipTakenByUs:
		return "taken_by_us"
	case FlipLostToEnemy:
		return "lost_to_enemy"
	case FlipOurGripLost:
		return "our_grip_lost"
	case FlipWePartlyLanded:
		return "we_partly_landed"
	}
	return ""
}
