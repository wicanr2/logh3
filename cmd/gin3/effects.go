package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	eaudio "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/wicanr2/logh3/internal/audio"
)

type effectPlayback interface {
	PlayEffect(int) error
	Busy() bool
	Stop()
}

// 不循環；與配樂共用Context，各自持有播放器。
//
// 規格160：啟動只確認 28 段檔案都在，開檔與完整解碼由 warm 在背景做，
// 解好的 PCM 留在記憶體，之後播放不再讀檔解碼。背景發現串流截斷時記下錯誤，
// 下一次 PlayEffect 回這個錯誤，由 updateEffects 照原本規則關掉音效。
type effectPlayer struct {
	decoder *musicPlayer
	mu      sync.Mutex
	pcm     [28][]byte
	broken  error
}

func newEffectPlayer(ctx *eaudio.Context, lib *audio.Library) *effectPlayer {
	return &effectPlayer{decoder: newMusicPlayer(ctx, lib)}
}

// validate 完整解碼 28 段並留下 PCM；任何一段截斷就回錯誤。
func (p *effectPlayer) validate() error {
	for id := 0; id < 28; id++ {
		pcm, err := p.decodePCM(id)
		if err != nil {
			return err
		}
		p.mu.Lock()
		p.pcm[id] = pcm
		p.mu.Unlock()
	}
	return nil
}

// warm 在背景跑 validate；失敗時記下，讓下一次播放回報並關掉音效。
func (p *effectPlayer) warm() {
	if err := p.validate(); err != nil {
		fmt.Fprintln(os.Stderr, "音效：", err)
		p.mu.Lock()
		p.broken = err
		p.mu.Unlock()
	}
}

func (p *effectPlayer) decodePCM(id int) ([]byte, error) {
	st, ok, err := p.decoder.decode(audio.SoundTrack(id))
	if err != nil {
		return nil, err
	}
	if !ok || st.Length() <= 0 {
		return nil, fmt.Errorf("音效%d無有效串流", id)
	}
	pcm, err := io.ReadAll(st)
	if err != nil || int64(len(pcm)) != st.Length() {
		return nil, fmt.Errorf("音效%d串流截斷：%d／%d %v", id, len(pcm), st.Length(), err)
	}
	return pcm, nil
}

func (p *effectPlayer) PlayEffect(id int) error {
	if id < 0 || id >= 28 {
		return fmt.Errorf("未知音效%d", id)
	}
	p.mu.Lock()
	pcm, broken := p.pcm[id], p.broken
	p.mu.Unlock()
	if broken != nil {
		return broken
	}
	var src io.Reader
	if pcm != nil {
		src = bytes.NewReader(pcm)
	} else {
		// 背景還沒解到這一段：照舊直接串流，不等。
		st, ok, err := p.decoder.decode(audio.SoundTrack(id))
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("缺音效%d", id)
		}
		src = st
	}
	p.Stop()
	pl, err := p.decoder.ctx.NewPlayer(src)
	if err != nil {
		return err
	}
	pl.Play()
	p.decoder.cur = pl
	return nil
}

func (p *effectPlayer) Stop() { p.decoder.Stop() }

func (p *effectPlayer) Busy() bool {
	return p.decoder.cur != nil && p.decoder.cur.IsPlaying()
}

func (p *effectPlayer) releaseFinished() {
	if p.decoder.cur != nil && !p.decoder.cur.IsPlaying() {
		p.Stop()
	}
}

func (a *app) syncEffects() error {
	for _, id := range a.g.DrainSoundEffects() {
		if a.effects != nil {
			if err := a.effects.PlayEffect(id); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *app) updateEffects() {
	if p, ok := a.effects.(*effectPlayer); ok {
		p.releaseFinished()
	}
	if err := a.syncEffects(); err != nil {
		fmt.Fprintln(os.Stderr, "音效：", err)
		a.effects.Stop()
		a.effects = nil
		_ = a.g.EnableSoundEffects(false)
	}
}
