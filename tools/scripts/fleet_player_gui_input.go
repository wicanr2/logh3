//go:build ignore

// 常駐XTEST連線，避免每個按下／放開都啟動xdotool的延遲。
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgb/xtest"
)

func main() {
	c, err := xgb.NewConn()
	if err != nil {
		panic(err)
	}
	defer c.Close()
	if err = xtest.Init(c); err != nil {
		panic(err)
	}
	s := xproto.Setup(c)
	m, err := xproto.GetKeyboardMapping(c, s.MinKeycode, byte(int(s.MaxKeycode)-int(s.MinKeycode)+1)).Reply()
	if err != nil {
		panic(err)
	}
	keys := map[xproto.Keysym]byte{}
	for i, k := range m.Keysyms {
		if k != 0 {
			if _, ok := keys[k]; !ok {
				keys[k] = byte(int(s.MinKeycode) + i/int(m.KeysymsPerKeycode))
			}
		}
	}
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		var p struct{ Type, Button, Key int }
		if err = json.Unmarshal(input.Bytes(), &p); err != nil {
			panic(err)
		}
		detail := byte(p.Button)
		if p.Type == xproto.KeyPress || p.Type == xproto.KeyRelease {
			symbol := xproto.Keysym(p.Key)
			if p.Key == 13 {
				symbol = 0xff0d
			}
			var ok bool
			detail, ok = keys[symbol]
			if !ok {
				panic("X11鍵盤缺測試按鍵")
			}
		} else if (p.Type != xproto.ButtonPress && p.Type != xproto.ButtonRelease) || (p.Button != 1 && p.Button != 3) {
			panic("非法測試輸入")
		}
		// 不同步等X server回覆；正式app的觀測器另核送達及放開。
		xtest.FakeInput(c, byte(p.Type), detail, 0, s.DefaultScreen(c).Root, 0, 0, 0)
		fmt.Println("{\"queued\":true}")
	}
	if err = input.Err(); err != nil {
		panic(err)
	}
}
