//go:build ignore

// 第105輪環境診斷：以正式鎖版Oto播放既有float32 PCM，不進玩家GUI。
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ebitengine/oto/v3"
)

func main() {
	if len(os.Args) != 3 {
		panic("用法：PCM來源 新輸出JSON")
	}
	raw, err := os.ReadFile(os.Args[1])
	if err != nil || len(raw) < 8*44100*8 {
		panic("缺少8秒float32來源")
	}
	raw = raw[:8*44100*8]
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate: 44100, ChannelCount: 2, Format: oto.FormatFloat32LE,
	})
	if err != nil {
		panic(err)
	}
	select {
	case <-ready:
	case <-time.After(15 * time.Second):
		panic("音訊context未就緒")
	}
	player := ctx.NewPlayer(bytes.NewReader(raw))
	started := time.Now()
	player.Play()
	for player.IsPlaying() {
		if time.Since(started) > 30*time.Second {
			panic("8秒來源播放超過30秒")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := ctx.Err(); err != nil {
		panic(err)
	}
	if err := player.Close(); err != nil {
		panic(err)
	}
	// 讓後端已提交的尾端流到monitor；不把此等待當作玩家完成閘門。
	time.Sleep(500 * time.Millisecond)
	result, _ := json.MarshalIndent(map[string]any{
		"source": os.Args[1], "bytes": len(raw), "nominal_seconds": 8,
		"elapsed_seconds": time.Since(started).Seconds(),
		"gui_used":        false, "production_player_verified": false,
	}, "", "  ")
	file, err := os.OpenFile(os.Args[2], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if _, err := file.Write(append(result, '\n')); err != nil {
		panic(err)
	}
	fmt.Println("PASS diagnostic Oto playback, not player acceptance")
}
