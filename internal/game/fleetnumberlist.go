package game

import (
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 艦隊編號清單（`StateFleetNumber`，#65）：艦隊編成按「艦隊編號」開出來的十六列。
//
// 點一列 ＝ **挑這一支要進艦隊表的哪一格**（編號 ＝ 陣營內的槽號 ＋ 1，`docs/internals/33`）：
// 收據 `fcmd-fnum-row`／`fcmd-fnum-decide` 裡挑第 3 列之後，「決定」建出來的那一支
// 在槽 2（槽 0 整筆還是 FF），右下角清單印「第３艦隊萊因哈特」。
//
// **已經在用的列畫成暗色，點下去只關清單、不換編號**（收據 `fcmd-fnum2`／`fcmd-fnum2-used`：
// 已有第１艦隊時第 01 列是暗的，點它標題仍是「第２艦隊」）。
//
// 版面取自 `screen-fleet-number.json` 的 `number` 視窗：外框 (320,60) 96×359，
// 列 k 的字基線在 y 82＋20k，帶狀是 y 81＋20k 起 20 列。

const (
	fleetNumberRows  = 16
	fleetNumberTop   = 81 // 列 0 帶狀的第一列（畫面座標）
	fleetNumberPitch = 20
	fleetNumberX     = 320
	fleetNumberW     = 96
)

// fleetNumberRowAt 回 (x, y) 落在第幾列；不在清單上回 −1。
func fleetNumberRowAt(x, y int) int {
	if x < fleetNumberX || x >= fleetNumberX+fleetNumberW {
		return -1
	}
	k := (y - fleetNumberTop) / fleetNumberPitch
	if y < fleetNumberTop || k < 0 || k >= fleetNumberRows {
		return -1
	}
	return k
}

// fleetSlotUsed ＝ 自己陣營的第 k 格已經有艦隊（`07F9:03AC` 挑槽時跳過的那些）。
func (g *Game) fleetSlotUsed(k int) bool {
	idx := rules.FleetTableIndex(g.faction, k)
	for _, f := range g.fleets {
		if f.Index == idx {
			return true
		}
	}
	return false
}

// fleetNumberClick 是清單上的左鍵：點在列上就關清單回艦隊編成，
// 那一格沒人用才換編號。
func (g *Game) fleetNumberClick(x, y int) {
	k := fleetNumberRowAt(x, y)
	if k < 0 {
		return
	}
	if !g.fleetSlotUsed(k) {
		g.orgSlot = k
	}
	g.state = StateFleetOrg
}

// orgFleetSlot 回**玩家從清單挑的**那一格；沒挑過回 −1，由呼叫端沿用原本的預設。
//
// 預設值那條規則（remake 現在算「提督在自己陣營裡的序號 ＋ 1」）仍然是
// **假設待驗**（`docs/re/66`）：手上兩個資料點（新局的萊因哈特、已有一支時的
// 吉爾菲艾斯）在「提督序號」與「下一個空槽」兩種算法下同值，分不出來。
// 這裡不動它——挑過才覆蓋。
func (g *Game) orgFleetSlot() int {
	if g.orgSlot >= 0 && g.orgSlot < fleetNumberRows && !g.fleetSlotUsed(g.orgSlot) {
		return g.orgSlot
	}
	return -1
}

// quitConfirmKind 是「您確定要結束這個遊戲嗎？」那個框在訊息框表裡的種類：
// 1 號（外框格 (8,7) 20×4、中止在格 14、決定在格 17）——量到的鈕位置
// （中止 368–399、決定 416–447、y 260–279）與 1 號逐格相同（收據 `fcmd-quit`）。
const quitConfirmKind = 1

// Quit 回「系統指令 → 結束 → 決定」按過了。顯示層看到就收掉（原版是 `AH=4C`）。
func (g *Game) Quit() bool { return g.quit }

// fleetNumberInk 是清單列的字色：一般 7、已經在用的那一格 8（收據 `fcmd-fnum2`
// 的列 0，其餘列仍是 7）。
const (
	fleetNumberInk     = 7
	fleetNumberUsedInk = 8
)

// applyFleetNumberList 每一幀把十六列的字色重算：已經在用的那幾格畫暗色。
func (g *Game) applyFleetNumberList(spec *screen.MainGameSpec) {
	if spec == nil {
		return
	}
	w := spec.Windows["number"]
	if w == nil {
		return
	}
	for k := 0; k < fleetNumberRows; k++ {
		ink := uint8(fleetNumberInk)
		if g.fleetSlotUsed(k) {
			ink = fleetNumberUsedInk
		}
		w.SetLineInk(k, ink)
	}
}
