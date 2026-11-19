//go:build android

package mobile

import (
	ebitenmobile "github.com/hajimehoshi/ebiten/v2/mobile"
	"github.com/wicanr2/logh3/internal/mobileshell"
)

var shell = mobileshell.NewShell()

func init() { ebitenmobile.SetGame(shell) }

// Start 由 Java 在 Activity 建立時呼叫，傳入三個目錄。重複呼叫只認第一次。
func Start(orig, data, save string) {
	shell.Configure(mobileshell.Config{Orig: orig, Data: data, Save: save})
}

// Back 由 Java 的返回手勢呼叫：排一次右鍵，下一格消費。
func Back() { shell.Back() }
