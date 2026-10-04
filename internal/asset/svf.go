package asset

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// SVF是本版MicroWaveSVFv1.0，不是Creative VOC（規格104）。
type SVF struct{ Samples []SVFSample }

type SVFSample struct {
	Data         []byte
	TimeConstant byte
	Command      byte
}

func DecodeSVF(b []byte) (*SVF, error) {
	if len(b) < 18 || !bytes.Equal(b[:16], []byte("MicroWaveSVFv1.0")) {
		return nil, fmt.Errorf("無效MicroWave SVF標頭")
	}
	n := int(binary.LittleEndian.Uint16(b[16:18]))
	if n == 0 || n > (len(b)-18)/6 {
		return nil, fmt.Errorf("無效SVF段數%d", n)
	}
	s := &SVF{}
	at := 18
	for i := 0; i < n; i++ {
		if len(b)-at < 4 {
			return nil, fmt.Errorf("SVF段%d標頭截斷", i)
		}
		length := int(binary.LittleEndian.Uint16(b[at:]))
		tc, cmd := b[at+2], b[at+3]
		at += 4
		if length < 2 || length > len(b)-at || cmd != 0x14 {
			return nil, fmt.Errorf("SVF段%d長度或命令不支援", i)
		}
		s.Samples = append(s.Samples, SVFSample{Data: append([]byte(nil), b[at:at+length]...), TimeConstant: tc, Command: cmd})
		at += length
	}
	if at != len(b) {
		return nil, fmt.Errorf("SVF尾端多出%d bytes", len(b)-at)
	}
	return s, nil
}

// PCM保留原版N-2的DSP長度參數，播放N-1樣本。
func (s SVFSample) PCM() []byte { return s.Data[:len(s.Data)-1] }

// Duration採公開DSP契約，屬hardware-spec approximation。
func (s SVFSample) Duration() time.Duration {
	return time.Duration(len(s.PCM())) * time.Duration(256-int(s.TimeConstant)) * time.Microsecond
}

type FleetSoundSpec struct {
	AttackHex       string `json:"attack_hex"`
	DefenceHex      string `json:"defence_hex"`
	attack, defence []byte
}

func LoadFleetSoundSpec(path string) (*FleetSoundSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s FleetSoundSpec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	for _, field := range []struct {
		raw string
		out *[]byte
	}{{s.AttackHex, &s.attack}, {s.DefenceHex, &s.defence}} {
		b, err := hex.DecodeString(field.raw)
		if err != nil || len(b) != 66 {
			return nil, fmt.Errorf("艦隊音效表需66 bytes")
		}
		*field.out = b
	}
	return &s, nil
}

func (s *FleetSoundSpec) Cue(side, kind, column int, flagship, defence bool) (int, error) {
	if s == nil || side < 0 || side > 1 || kind < 0 || kind > 7 || column < 0 || column > 2 {
		return -1, fmt.Errorf("無效艦隊音效欄位")
	}
	if flagship {
		kind = 7
	}
	table := s.attack
	if defence {
		table = s.defence
	}
	if len(table) != 66 {
		return -1, fmt.Errorf("艦隊音效表未載入")
	}
	id := table[side*33+kind*3+column]
	if id == 255 {
		return -1, nil
	}
	return int(id) % 28, nil
}
