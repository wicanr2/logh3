package main

import (
	"fmt"
	"io"
	"os"

	eaudio "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/wicanr2/logh3/internal/audio"
)

type effectPlayback interface {
	PlayEffect(int) error
	Busy() bool
	Stop()
}

// 不循環；與配樂共用Context，各自持有播放器。
type effectPlayer struct{ decoder *musicPlayer }

func newEffectPlayer(ctx *eaudio.Context, lib *audio.Library) *effectPlayer {
	return &effectPlayer{decoder: newMusicPlayer(ctx, lib)}
}

func (p *effectPlayer) validate() error {
	for id := 0; id < 28; id++ {
		st, ok, err := p.decoder.decode(audio.SoundTrack(id))
		if err != nil {
			return err
		}
		if !ok || st.Length() <= 0 {
			return fmt.Errorf("音效%d無有效串流", id)
		}
		n, err := io.Copy(io.Discard, st)
		if err != nil || n != st.Length() {
			return fmt.Errorf("音效%d串流截斷：%d／%d %v", id, n, st.Length(), err)
		}
	}
	return nil
}

func (p *effectPlayer) PlayEffect(id int) error {
	if id < 0 || id >= 28 {
		return fmt.Errorf("未知音效%d", id)
	}
	st, ok, err := p.decoder.decode(audio.SoundTrack(id))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("缺音效%d", id)
	}
	p.Stop()
	pl, err := p.decoder.ctx.NewPlayer(st)
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
