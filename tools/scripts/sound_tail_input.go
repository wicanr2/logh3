//go:build ignore

// 第81輪的XTEST右鍵：正式Update確認按下後，由同一程序立即送出放開。
// 只讀觀測收據，不呼叫Game、不改鏡頭，也不改更新速度。
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

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
	send := func(kind, detail byte) {
		xtest.FakeInput(c, kind, detail, 0, s.DefaultScreen(c).Root, 0, 0, 0)
		// 確認X server已處理，queued不再只代表排進工具自己的佇列。
		if _, err := xproto.GetInputFocus(c).Reply(); err != nil {
			panic(err)
		}
	}
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		var p struct {
			Type, Button, Key, Edge int
			Status                  string
		}
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
		send(byte(p.Type), detail)
		if p.Status != "" {
			if p.Type != xproto.ButtonPress || p.Button != 3 || p.Edge <= 0 {
				panic("非法觀測右鍵")
			}
			func() {
				defer send(xproto.ButtonRelease, detail)
				deadline := time.Now().Add(5 * time.Second)
				for time.Now().Before(deadline) {
					var st struct {
						Edges [3]int `json:"edges"`
					}
					b, err := os.ReadFile(p.Status)
					if err == nil && json.Unmarshal(b, &st) == nil && st.Edges[1] >= p.Edge {
						return
					}
					time.Sleep(time.Millisecond)
				}
				panic("正式Update未確認右鍵")
			}()
		}
		fmt.Println("{\"queued\":true}")
	}
	if err = input.Err(); err != nil {
		panic(err)
	}
}
