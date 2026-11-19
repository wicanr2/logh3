//go:build android

package mobile

import (
	ebitenmobile "github.com/hajimehoshi/ebiten/v2/mobile"
	"github.com/wicanr2/logh3/internal/mobileshell"
)

var shell = mobileshell.NewShell()

func init() {
	shell.UseNativeTouch()
	ebitenmobile.SetGame(shell)
}

// Start 由 Java 在 Activity 建立時呼叫，傳入目錄（music、art 可空）。重複呼叫只認第一次。
func Start(orig, data, save, music, art string) {
	shell.Configure(mobileshell.Config{Orig: orig, Data: data, Save: save, Music: music, Art: art})
}

// Back 由 Java 的返回手勢呼叫：排一次右鍵，下一格消費。
func Back() { shell.Back() }

// Tap 由 Java 在手指放開時呼叫，nx、ny 是在遊戲畫面元件上的比例位置（0–1）。
func Tap(nx, ny float64) { shell.QueueTap(nx, ny) }
