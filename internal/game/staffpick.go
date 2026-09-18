package game

import "github.com/wicanr2/logh3/internal/rules"

// 參謀任命（#62，`docs/re/342` §5、`docs/re/348`、規格 064）。
//
// 情報收集畫面的「參謀」鈕（`0A02:0AC4`）開的是**與出擊命令同一套**的
// 「待命中提督資料」清單與提督詳細資料視窗（`0C0A:00F0(1)` → `0D92:000D`），
// 只是詳細視窗底下第三顆鈕在這條路上是「決定」——按下去寫
// `ds:8F4D[座位] ← 提督編號`，回到情報收集就換成新的姓名與頭像。
//
// 按「中止」回清單、不是關窗（收據 `fcmd-staff-cancel`：`ds:8F4D` 留在原值）。

// staffPick 是參謀任命這條路上的狀態。零值 ＝ 沒在任命（走的是出擊那條）。
type staffPick struct {
	// on ＝ 現在開著的清單／詳細視窗屬於參謀任命，不是出擊命令。
	on bool
	// post 是要換哪一個幕僚職（`rules.PostIntel` 等）。
	post int
	// back 是關掉之後回哪一張畫面。
	back State
}

// beginStaffPick 從情報收集（或徵稅計劃）的「參謀」鈕進來。
func (g *Game) beginStaffPick(post int, back State) {
	if g.commandSpec(StateStandbyAdmirals) == nil {
		return
	}
	g.staff = staffPick{on: true, post: post, back: back}
	g.detailAdmiral = -1
	g.state = StateStandbyAdmirals
}

// staffPickDecide 是詳細資料視窗的「決定」：把選中的提督寫進專任參謀表，
// 回到進來時那一張畫面。
func (g *Game) staffPickDecide() {
	defer func() {
		back := g.staff.back
		g.staff = staffPick{}
		g.state = back
	}()
	if g.detailAdmiral < 0 {
		return
	}
	rules.SetStaffAdmiral(g.mutableWork(), g.staff.post, g.faction, g.detailAdmiral)
}

// staffPickCancel 是「中止」：回清單，什麼都不寫。
func (g *Game) staffPickCancel() {
	g.state = StateStandbyAdmirals
}

// staffPickClose 是在清單上按右鍵：整條收掉，回進來時那一張。
func (g *Game) staffPickClose() {
	back := g.staff.back
	g.staff = staffPick{}
	g.state = back
}
