package game

import (
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/asset"
)

func fleetSoundAssets(t *testing.T, g *Game) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "org_game/extracted/GIN3SP.VOC"))
	if os.IsNotExist(err) {
		t.Skip("缺原版音效")
	}
	if err != nil {
		t.Fatal(err)
	}
	g.res.SoundSamples, err = asset.DecodeSVF(b)
	if err != nil {
		t.Fatal(err)
	}
	g.res.FleetSound, err = asset.LoadFleetSoundSpec(filepath.Join(root, "docs/internals/data/fleet-sound-spec.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = g.EnableSoundEffects(true); err != nil {
		t.Fatal(err)
	}
}

func TestFleetSoundFirstFrameWait(t *testing.T) {
	m := fleetSceneReceipt(t)
	g := fleetChainRAM(t, m, m.read(t, "workplace/oracle/scene48-r69-v2-laser-a/start-mem.bin"))
	fleetSoundAssets(t, g)
	g.display.AnimSpeed = 2
	previousDuration := 2 * time.Second
	g.soundEffects.remaining = previousDuration
	g.fleetCombat = fleetCombatState{active: true, fleets: []int{12}, current: -1}
	g.beginFleetCombatScenes()
	s := g.fleetCombat.scene
	if s == nil || len(g.DrainSoundEffects()) != 0 {
		t.Fatal("前一樣本未完就請求音效")
	}
	seed, ships := g.battleRand.Seed(), g.battleUnits[s.target].Ships
	elapsed := time.Duration(0)
	for elapsed+g.TimerTick() < previousDuration {
		g.Tick(g.TimerTick())
		elapsed += g.TimerTick()
		if len(g.DrainSoundEffects()) != 0 || s.damaged || g.battleUnits[s.target].Ships != ships || g.battleRand.Seed() != seed {
			t.Fatal("等待音效期間提前進入攻擊／戰損")
		}
		v, err := g.fleetFrameView(s, false)
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range v.Pictures {
			if len(p.Segment) > 2 && p.Segment[2] == 'A' {
				t.Fatal("音效尚未開始就畫攻擊首格")
			}
		}
		g.Frame()
		g.Frame()
		if len(g.DrainSoundEffects()) != 0 {
			t.Fatal("繪圖重播音效")
		}
	}
	var ids []int
	for i := 0; i < 4 && len(ids) == 0; i++ {
		g.Tick(g.TimerTick())
		elapsed += g.TimerTick()
		ids = g.DrainSoundEffects()
	}
	if len(ids) != 1 || ids[0] != 8 || elapsed < previousDuration {
		t.Fatal("音效完成後未續接原版攻擊編號", ids, elapsed)
	}
}

func TestFleetSoundZeroLoss(t *testing.T) {
	m := fleetSceneReceipt(t)
	g := fleetChainRAM(t, m, m.read(t, "workplace/oracle/scene48-r69-v2-laser-a/start-mem.bin"))
	fleetSoundAssets(t, g)
	g.fleetCombat = fleetCombatState{active: true, fleets: []int{12}, current: -1}
	g.beginFleetCombatScenes()
	s := g.fleetCombat.scene
	if s == nil {
		t.Fatal("缺交戰")
	}
	// 局部零損失分支診斷；不宣稱此注入是正常原版局面。
	g.battleUnits[s.target].Ships = 0
	if err := g.applyFleetSceneDamage(s); err != nil {
		t.Fatal(err)
	}
	if s.hit.Lost != 0 {
		t.Fatal("未到零損失分支")
	}
	for _, c := range s.sounds {
		if c.defence {
			t.Fatal("零損失仍請求守方音效")
		}
	}
}

func TestFleetSoundNative(t *testing.T) {
	m := fleetSceneReceipt(t)
	var native struct {
		Pairs []struct {
			Case    string `json:"case"`
			Records []struct {
				Record  []byte `json:"record"`
				Static  int    `json:"static"`
				Animate int    `json:"animate"`
			} `json:"comani_records"`
		} `json:"pairs"`
	}
	b, err := os.ReadFile(filepath.Join(root, "docs/internals/data/fleet-scene.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &native); err != nil {
		t.Fatal(err)
	}
	for _, pair := range native.Pairs {
		for _, speed := range []int{0, 1, 2, 3} {
			t.Run(pair.Case+string(rune('0'+speed)), func(t *testing.T) {
				dir := "workplace/oracle/scene48-r69-v2-" + pair.Case + "-a/"
				g := fleetChainRAM(t, m, m.read(t, dir+"start-mem.bin"))
				fleetSoundAssets(t, g)
				g.display.AnimSpeed = speed
				if speed == 3 {
					g.display.Anim = 0
				}
				g.fleetCombat = fleetCombatState{active: true, fleets: []int{12}, current: -1}
				g.beginFleetCombatScenes()
				var events []int
				elapsed, previous, previousID := int64(0), int64(-1), -1
				collect := func() {
					for _, id := range g.DrainSoundEffects() {
						if previous >= 0 && elapsed-previous < int64(g.res.SoundSamples.Samples[previousID].Duration()) {
							t.Fatal("樣本尚未完成就播下一段", previousID, id, elapsed-previous)
						}
						previous, previousID = elapsed, id
						events = append(events, id)
					}
					if len(g.DrainSoundEffects()) != 0 {
						t.Fatal("事件重播")
					}
				}
				collect()
				for ticks := 0; ticks < 4000 && g.fleetCombat.active; ticks++ {
					g.Tick(g.TimerTick())
					elapsed += int64(g.TimerTick())
					collect()
					if g.fleetCombat.err != nil {
						t.Fatal(g.fleetCombat.err)
					}
					g.Frame()
					g.Frame()
					if len(g.DrainSoundEffects()) != 0 {
						t.Fatal("Frame產生音效")
					}
				}
				if g.fleetCombat.active {
					t.Fatal("音效等待未結束")
				}
				var want []int
				if speed != 3 {
					for _, r := range pair.Records {
						if r.Static == 0 && r.Animate != 0 && r.Record[6] != 255 {
							want = append(want, int(r.Record[6])%28)
						}
					}
				}
				if !reflect.DeepEqual(events, want) {
					t.Fatalf("原版音效%v，Game%v", want, events)
				}
				end := m.read(t, dir+"end-mem.bin")
				for _, pool := range []int{80, 82, 102} {
					p := pool * 78
					if !reflect.DeepEqual(g.slot.Ship[p+8:p+22], end[0x668b2+p+8:0x668b2+p+22]) {
						t.Fatal("音效改變戰損／庫存", pool)
					}
				}
				if g.battleRand.Seed() != binary.LittleEndian.Uint32(end[0x5e00c:]) {
					t.Fatal("音效改變seed")
				}
				t.Log("原版音效序列", events, "硬體規格近似；非原版wall-clock")
			})
		}
	}
}
