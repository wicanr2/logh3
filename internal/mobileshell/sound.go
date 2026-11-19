package mobileshell

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	eaudio "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/wicanr2/logh3/internal/audio"
	"github.com/wicanr2/logh3/internal/game"
)

// 行動版的配樂與音效（規格161 M1）。
//
// 規則全在 internal/：選哪一首走 audio.TuneFor，要播哪一段音效由 Game 的佇列
// （DrainSoundEffects）決定，播完才放行下一段（SetSoundEffectPlaying，規格105）。
// 這裡只有播放這一層：解 OGG、循環播配樂、播音效。桌面的同一層在 cmd/gin3
// （music.go、effects.go），那兩個檔被四十多份文件以路徑引用，不搬，行動版另寫。
//
// 只用原版曲庫（音樂目錄根層的 OGG）；HD 音色庫不在 M1 的完整版裡。

type sound struct {
	ctx    *eaudio.Context
	music  *audio.Library
	sfx    *audio.Library
	cur    *eaudio.Player
	tune   string
	effect *eaudio.Player
	last   game.State

	mu  sync.Mutex
	pcm map[int][]byte // 音效解好的 PCM，第一次播時解
	bad bool           // 音效壞了就整個關掉，只印一次
}

// newSound 建播放器；目錄沒有曲子回 nil（安靜，但印出原因）。
func newSound(dir string) *sound {
	if dir == "" {
		return nil
	}
	if _, err := os.Stat(dir); err != nil {
		log.Printf("logh3: 沒有音樂目錄 %s，這一趟沒有聲音", dir)
		return nil
	}
	music, sfx := audio.NewLibrary(dir), audio.NewSoundLibrary(dir)
	if music.Have() == 0 && sfx.Have() == 0 {
		log.Printf("logh3: %s 裡沒有配樂也沒有音效", dir)
		return nil
	}
	log.Printf("logh3: 配樂 %d 首、音效 %d/28 段", music.Have(), sfx.Have())
	return &sound{ctx: eaudio.NewContext(audio.SampleRate), music: music, sfx: sfx, pcm: map[int][]byte{}}
}

func decode(lib *audio.Library, track string) (*vorbis.Stream, bool, error) {
	path, ok := lib.Path(track)
	if !ok {
		return nil, false, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, false, err
	}
	st, err := vorbis.DecodeWithoutResampling(bytes.NewReader(raw))
	if err != nil {
		return nil, false, fmt.Errorf("解 %s：%w", track, err)
	}
	if st.SampleRate() != audio.SampleRate {
		return nil, false, fmt.Errorf("%s 取樣率 %d，預期 %d", track, st.SampleRate(), audio.SampleRate)
	}
	return st, true, nil
}

// beforeTick 在 Game.Tick 之前：告訴遊戲音效還在不在播（規格105）。
func (s *sound) beforeTick(g *game.Game) {
	if s == nil {
		return
	}
	if s.effect != nil && !s.effect.IsPlaying() {
		_ = s.effect.Close()
		s.effect = nil
	}
	g.SetSoundEffectPlaying(s.effect != nil)
}

// afterTick 在 Game.Tick 之後：同步配樂，播佇列裡的音效。
func (s *sound) afterTick(g *game.Game) {
	if s == nil {
		return
	}
	st := g.State()
	// 原版標題逾時會重播開頭動畫，同一首也要從頭播（桌面 syncMusic 同一條，規格112）。
	if s.last == game.StateTitle && st == game.StateIntro {
		s.stopMusic()
	}
	s.last = st
	m := g.MusicSettings()
	tune := audio.TuneFor(st, g.Faction(), audio.Settings{Device: m.Device, BGM: m.BGM, Battle: m.Battle})
	if tune == "" {
		s.stopMusic()
	} else if tune != s.tune || s.cur == nil || !s.cur.IsPlaying() {
		s.playMusic(tune)
	}
	for _, id := range g.DrainSoundEffects() {
		s.playEffect(g, id)
	}
}

func (s *sound) stopMusic() {
	if s.cur != nil {
		_ = s.cur.Close()
	}
	s.cur, s.tune = nil, ""
}

func (s *sound) playMusic(tune string) {
	st, ok, err := decode(s.music, tune)
	if err != nil || !ok {
		if err != nil {
			log.Printf("logh3: 配樂：%v", err)
		}
		s.stopMusic()
		s.tune = tune // 缺曲或壞檔不每格重試
		return
	}
	s.stopMusic()
	p, err := s.ctx.NewPlayer(eaudio.NewInfiniteLoop(st, st.Length()))
	if err != nil {
		log.Printf("logh3: 配樂播放器：%v", err)
		return
	}
	p.Play()
	s.cur, s.tune = p, tune
}

func (s *sound) playEffect(g *game.Game, id int) {
	if s.bad || id < 0 || id >= 28 {
		return
	}
	s.mu.Lock()
	pcm := s.pcm[id]
	s.mu.Unlock()
	if pcm == nil {
		st, ok, err := decode(s.sfx, audio.SoundTrack(id))
		if err == nil && ok {
			pcm, err = io.ReadAll(st)
			if err == nil && int64(len(pcm)) != st.Length() {
				err = fmt.Errorf("音效%d串流截斷", id)
			}
		}
		if err != nil || !ok {
			// 與桌面同一條規則：音效壞了就整個關掉並說出原因（updateEffects）。
			log.Printf("logh3: 音效%d 無法播放，關掉音效：%v", id, err)
			s.bad = true
			_ = g.EnableSoundEffects(false)
			return
		}
		s.mu.Lock()
		s.pcm[id] = pcm
		s.mu.Unlock()
	}
	if s.effect != nil {
		_ = s.effect.Close()
	}
	p := s.ctx.NewPlayerFromBytes(pcm)
	p.Play()
	s.effect = p
}
