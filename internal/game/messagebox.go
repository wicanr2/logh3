package game

import (
	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/screen"
)

// 訊息框（原版 `3826:0005` ＋ `3826:023B`，`docs/re/137`、`138`）。
//
// **它是模態的**：`3826:023B` 一直等，只認左鍵，而且只有點在子元件
// （按鈕）上才會離開迴圈；點在別處什麼都不做，框留著。離開之後回傳
// 那顆按鈕的編號，呼叫端據此決定下一步。三句訊息的呼叫端都是「印完就
// 回 `0xFFFF`」，也就是整個命令作廢（`docs/re/121`）。

// showMessage 開一個訊息框。name 是 `message-box.json` 的 `messages`。
func (g *Game) showMessage(name string) bool {
	if g.res.MessageBox == nil {
		return false
	}
	kind, codes, ok := g.res.MessageBox.Message(name)
	if !ok {
		return false
	}
	g.msgOn, g.msgKind, g.msgCodes = true, kind, codes
	return true
}

// messageClick 處理訊息框開著時的左鍵。回有沒有吃掉這一下——**永遠是
// true**，因為原版在那個迴圈裡不理會別的東西。
func (g *Game) messageClick(x, y int) bool {
	if !g.msgOn {
		return false
	}
	if _, hit := g.res.MessageBox.MessageBoxButtonAt(g.msgKind, x, y); hit {
		g.msgOn, g.msgCodes = false, nil
	}
	return true
}

// drawMessage 把訊息框畫在最上面。
func (g *Game) drawMessage(fb *gfx.Framebuffer) {
	if !g.msgOn || g.res.MessageBox == nil || g.res.FrameTiles == nil || g.res.Font == nil {
		return
	}
	s := g.res.MessageBox
	screen.DrawMessageBox(fb, s, g.msgKind, g.res.FrameTiles)
	screen.DrawMessageBoxText(fb, s, g.msgKind, g.res.Font, 0, g.msgCodes)
	screen.DrawMessageBoxButtons(fb, s, g.msgKind, g.res.Font)
}
