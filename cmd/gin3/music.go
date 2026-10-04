package main

// 配樂的播放端。**放在 cmd/gin3 而不是 internal/audio**，因為它要
// `ebiten/audio`，而那條路上的 `oto` 建置時要 ALSA 的 pkg-config——
// 標準建置 image（`tools/go.sh`）沒有，放進 internal 會讓
// `go test ./...` 從此建不起來。GUI 這一支本來就只在
// `tools/gui.sh` 的 image 裡建。
//
// 曲庫（哪一首在哪裡、缺哪幾首）在 `internal/audio`，那半是純的、測得到。

import (
	"bytes"
	"fmt"
	"io"
	"os"

	eaudio "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"

	"github.com/wicanr2/logh3/internal/audio"
)

// SampleRate 是輸出的取樣率。轉檔工具照這個值產 OGG，執行期就不必重取樣
// ——重取樣要在每一個 buffer 上做，而它會把 CPU 花在聽不出差別的地方。
const sampleRate = audio.SampleRate

// Player 播一首、循環，換曲就把上一首停掉。
//
// ⚠ **一個行程只能有一個 `audio.Context`**（Ebitengine 的限制），
// 所以 Context 由呼叫端建好傳進來，這裡不自己建。
type musicPlayer struct {
	ctx  *eaudio.Context
	lib  *audio.Library
	cur  *eaudio.Player
	name string
}

// NewPlayer 綁定一個已經建好的 Context 與一份曲庫。
func newMusicPlayer(ctx *eaudio.Context, lib *audio.Library) *musicPlayer {
	return &musicPlayer{ctx: ctx, lib: lib}
}

// Play 播 track 並無限循環。已經在播同一首就什麼都不做——
// **不要每一幀都重播**，那會把曲子卡在開頭 0.016 秒。
//
// 找不到檔案回 nil 不回 error：缺配樂不該讓遊戲跑不起來。
// 缺了哪幾首由 `Library.Missing` 交代，那份清單呼叫端要印出來。
func (p *musicPlayer) Play(track string) error {
	if p.name == track && p.cur != nil && p.cur.IsPlaying() {
		return nil
	}
	st, ok, err := p.decode(track)
	if err != nil || !ok {
		return err
	}
	p.Stop()
	loop := eaudio.NewInfiniteLoop(st, st.Length())
	pl, err := p.ctx.NewPlayer(loop)
	if err != nil {
		return fmt.Errorf("建 %s 的播放器：%w", track, err)
	}
	pl.Play()
	p.cur, p.name = pl, track
	return nil
}

// decode 驗曲庫中的一首，但不碰音效裝置。正式 Play 與無裝置整合測試共用
// 同一支，避免「測過的解碼條件」與真正播放器分岔。
func (p *musicPlayer) decode(track string) (*vorbis.Stream, bool, error) {
	path, ok := p.lib.Path(track)
	if !ok {
		return nil, false, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, false, fmt.Errorf("讀配樂 %s：%w", track, err)
	}
	st, err := vorbis.DecodeWithoutResampling(bytes.NewReader(raw))
	if err != nil {
		return nil, false, fmt.Errorf("解 %s 的 OGG：%w", track, err)
	}
	if got := st.SampleRate(); got != sampleRate {
		// **取樣率不對要說出來，不要默默重取樣。** 轉檔工具與這裡用
		// 同一個常數，對不上就是轉檔那一端跑錯設定了——播出來只是
		// 速度與音高不對，聽起來像「音源不一樣」而不是像設定錯。
		return nil, false, fmt.Errorf(
			"配樂 %s 的取樣率是 %d，預期 %d（轉檔時的 -ar 設定不對）",
			track, got, sampleRate)
	}
	return st, true, nil
}

// Stop 停掉目前這首。沒在播就什麼都不做。
func (p *musicPlayer) Stop() {
	if p.cur == nil {
		return
	}
	_ = p.cur.Close()
	p.cur, p.name = nil, ""
}

// Playing 是目前在播的曲名；沒有就是空字串。
func (p *musicPlayer) Playing() string { return p.name }

// 讓 io 這個 import 有用途：DecodeWithoutResampling 吃 io.Reader，
// 這裡宣告一次型別確保簽章不會在升版時悄悄改掉。
var _ func(io.Reader) (*vorbis.Stream, error) = vorbis.DecodeWithoutResampling
