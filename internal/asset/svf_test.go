package asset

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSVFOriginal(t *testing.T) {
	p := filepath.Join("..", "..", "org_game", "extracted", "GIN3SP.VOC")
	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		t.Skip("缺原版GIN3SP.VOC")
	}
	if err != nil {
		t.Fatal(err)
	}
	var source struct {
		Samples []struct {
			PlayedSamples int    `json:"played_samples"`
			DurationNS    int64  `json:"duration_ns"`
			PlayedSHA     string `json:"played_pcm_sha256"`
		}
	}
	j, err := os.ReadFile("../../docs/internals/data/fleet-sound-source.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(j, &source); err != nil {
		t.Fatal(err)
	}
	s, err := DecodeSVF(b)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Samples) != 28 {
		t.Fatal(len(s.Samples))
	}
	for i, sample := range s.Samples {
		want := source.Samples[i]
		if len(sample.PCM()) != want.PlayedSamples || int64(sample.Duration()) != want.DurationNS || fmt.Sprintf("%x", sha256.Sum256(sample.PCM())) != want.PlayedSHA {
			t.Fatalf("段%d與原始索引不同", i)
		}
	}
	for _, bad := range [][]byte{b[:17], b[:len(b)-1], append(append([]byte(nil), b...), 0)} {
		if _, err := DecodeSVF(bad); err == nil {
			t.Fatal("接受截斷或多餘bytes")
		}
	}
	bad := append([]byte(nil), b...)
	bad[21] = 0x75
	if _, err := DecodeSVF(bad); err == nil {
		t.Fatal("接受未支援ADPCM")
	}
}

func TestFleetSoundRecordedCues(t *testing.T) {
	s, err := LoadFleetSoundSpec("../../docs/internals/data/fleet-sound-spec.json")
	if err != nil {
		t.Fatal(err)
	}
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
	b, err := os.ReadFile("../../docs/internals/data/fleet-scene.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &native); err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, pair := range native.Pairs {
		for _, r := range pair.Records {
			if r.Static != 0 || r.Animate == 0 {
				continue
			}
			v := r.Record
			got, err := s.Cue(int(v[0]), int(v[1]), int(v[3]), v[4] == 0, v[2] != 0)
			if err != nil || got != int(v[6]) {
				t.Fatalf("%s記錄%x got%d err%v", pair.Case, v, got, err)
			}
			n++
		}
	}
	if n != 10 {
		t.Fatalf("動態原版記錄%d", n)
	}
}
