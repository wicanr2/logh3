package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/audio"
	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/rules"
)

func TestSoundEffectPlayerDecode(t *testing.T) {
	dir := os.Getenv("LOGH3_SFX_DIR")
	if dir == "" {
		dir = "../../workplace/round78-sfx-final-a"
	}
	b, err := os.ReadFile(filepath.Join(dir, "SFX-RECEIPT.json"))
	if os.IsNotExist(err) {
		t.Skip("缺原版28段OGG批次")
	}
	if err != nil {
		t.Fatal(err)
	}
	var r struct {
		Tracks []struct {
			ID       int     `json:"id"`
			SHA      string  `json:"ogg_sha256"`
			Duration float64 `json:"duration_seconds"`
		} `json:"tracks"`
	}
	if err = json.Unmarshal(b, &r); err != nil || len(r.Tracks) != 28 {
		t.Fatal("轉檔批次不完整", err)
	}
	p := newEffectPlayer(nil, audio.NewSoundLibrary(dir))
	if err = p.validate(); err != nil {
		t.Fatal(err)
	}
	var records []map[string]any
	for _, track := range r.Tracks {
		name := audio.SoundTrack(track.ID)
		b, err := os.ReadFile(filepath.Join(dir, name+".ogg"))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(b)) != track.SHA {
			t.Fatal("OGG產物不符", name)
		}
		st, ok, err := p.decoder.decode(name)
		if err != nil || !ok {
			t.Fatal(name, err)
		}
		length := st.Length()
		duration := float64(length) / 4 / 44100
		if math.Abs(duration-track.Duration) > 1.0/44100 {
			t.Fatal("正式解碼有效時長不符", name, duration, track.Duration)
		}
		h := sha256.New()
		n, err := io.Copy(h, st)
		if err != nil || n != length {
			t.Fatal("正式解碼尾端長度不符", name, n, length, err)
		}
		records = append(records, map[string]any{"id": track.ID, "bytes": n, "duration_seconds": duration, "pcm_sha256": fmt.Sprintf("%x", h.Sum(nil))})
	}
	if path := os.Getenv("LOGH3_SFX_DECODE_OUT"); path != "" {
		b, err := json.MarshalIndent(records, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(path, append(b, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("28段由正式一次性播放端完整解碼，裝置及聆聽另驗")
}

type effectRecorder struct {
	elapsed int64
	events  []map[string]any
}

func (p *effectRecorder) PlayEffect(id int) error {
	p.events = append(p.events, map[string]any{"id": id, "at_ns": p.elapsed})
	return nil
}
func (p *effectRecorder) Stop()      {}
func (p *effectRecorder) Busy() bool { return false }

// 正常bootstrap、公開輸入與正式app播放邊界。overlay僅提供固定seed。
func TestNormalSoundApp(t *testing.T) {
	if os.Getenv("LOGH3_SOUND_APP") != "1" {
		t.Skip("正常收據重播需明示overlay")
	}
	for side := 0; side < 2; side++ {
		t.Run(fmt.Sprint(side), func(t *testing.T) {
			dir := fmt.Sprintf("../../workplace/round78-chain-a/normal-%d", side)
			var plan struct {
				Inputs []struct {
					Kind string
					X, Y int
					NS   int64
					Seed uint32
				}
			}
			b, err := os.ReadFile(filepath.Join(dir, "inputs.json"))
			if err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(b, &plan); err != nil {
				t.Fatal(err)
			}
			r, err := bootstrap.Load("../../org_game/extracted", "../../docs/internals/data")
			if err != nil {
				t.Fatal(err)
			}
			g, err := game.New(r)
			if err != nil {
				t.Fatal(err)
			}
			probe, ok := any(g).(interface{ FleetChainProbeSeed(uint32) })
			if !ok {
				t.Fatal("未載入固定seed的test-only overlay")
			}
			if err = g.SetAILevel(-1); err != nil {
				t.Fatal(err)
			}
			if err = g.EnableSoundEffects(true); err != nil {
				t.Fatal(err)
			}
			if g.State() == game.StateIntro {
				g.Click(300, 210)
			}
			g.Click(300, 210+20*side)
			for _, k := range []byte("138\r") {
				g.Key(k)
			}
			g.RightClick(300, 240)
			g.RightClick(300, 240)
			if g.State() != game.StateCommandPhase {
				t.Fatal("正常開局未到命令畫面")
			}
			p := &effectRecorder{}
			a := &app{g: g, effects: p} // 沒有配樂端，音效仍獨立同步。
			for _, in := range plan.Inputs {
				switch in.Kind {
				case "click":
					g.MoveMouse(in.X, in.Y)
					g.Click(in.X, in.Y)
					g.Frame()
				case "right":
					g.RightClick(in.X, in.Y)
					g.RightRelease()
				case "tick":
					g.Tick(time.Duration(in.NS))
					p.elapsed += in.NS
				case "seed":
					probe.FleetChainProbeSeed(in.Seed)
				default:
					t.Fatal("未知公開輸入")
				}
				if err = a.syncEffects(); err != nil {
					t.Fatal(err)
				}
			}
			if g.State() != game.StateBattle {
				t.Fatal("交戰未返回正常戰場")
			}
			b, err = os.ReadFile(filepath.Join(dir, "sounds.json"))
			if err != nil {
				t.Fatal(err)
			}
			var want []map[string]any
			if err = json.Unmarshal(b, &want); err != nil {
				t.Fatal(err)
			}
			got, _ := json.Marshal(p.events)
			var normalized []map[string]any
			if err = json.Unmarshal(got, &normalized); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(normalized, want) {
				t.Fatal("正式app音效序列／時間與正常Game不同", normalized, want)
			}
			save := t.TempDir()
			if err = g.SaveSlot(save, 0); err != nil {
				t.Fatal(err)
			}
			for _, n := range rules.SlotFileNames(0) {
				b, err := os.ReadFile(filepath.Join(save, n))
				if err != nil {
					t.Fatal(err)
				}
				w, err := os.ReadFile(filepath.Join(dir, n))
				if err != nil || string(b) != string(w) {
					t.Fatal("正式app保存不同", n, err)
				}
			}
			if err = os.WriteFile(fmt.Sprintf("../../workplace/round78-app-%d.json", side), append(got, '\n'), 0644); err != nil {
				t.Fatal(err)
			}
			t.Log("正式app事件", len(p.events), "六份保存相同；無裝置錄製器")
		})
	}
}
