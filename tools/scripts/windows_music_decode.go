//go:build ignore

// 用與正式musicPlayer相同的Vorbis解碼器保存指定曲目的完整float32來源。
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

func main() {
	if len(os.Args) != 3 {
		panic("用法：曲庫目錄 新輸出目錄")
	}
	root, out := os.Args[1], os.Args[2]
	if err := os.Mkdir(out, 0755); err != nil {
		panic(err)
	}
	var rows []map[string]any
	for _, name := range []string{"OPEN_M", "BGM_0F", "BGM_0M", "BGM_1F"} {
		path := filepath.Join(root, name+".ogg")
		raw, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		stream, err := vorbis.DecodeWithoutResampling(bytes.NewReader(raw))
		if err != nil {
			panic(err)
		}
		if stream.SampleRate() != 44100 {
			panic("取樣率不同")
		}
		pcm, err := io.ReadAll(stream)
		if err != nil {
			panic(err)
		}
		if len(pcm)%4 != 0 || len(pcm) == 0 {
			panic("PCM形狀不同")
		}
		if err := os.WriteFile(filepath.Join(out, name+".s16le"), pcm, 0644); err != nil {
			panic(err)
		}
		floats := make([]byte, len(pcm)*2)
		for i := 0; i < len(pcm); i += 2 {
			value := float32(int16(binary.LittleEndian.Uint16(pcm[i:]))) / 32768
			binary.LittleEndian.PutUint32(floats[i*2:], math.Float32bits(value))
		}
		if err := os.WriteFile(filepath.Join(out, name+".f32le"), floats, 0644); err != nil {
			panic(err)
		}
		rows = append(rows, map[string]any{"track": name, "ogg": path,
			"ogg_sha256": fmt.Sprintf("%x", sha256.Sum256(raw)), "bytes": len(floats),
			"sample_rate": 44100, "channels": 2, "format": "float32le",
			"pcm_sha256": fmt.Sprintf("%x", sha256.Sum256(floats)), "frames": len(pcm) / 4,
			"formal_decode_format": "signed16 little endian", "float_conversion": "int16/32768，與NewPlayer混音端相同",
			"s16_sha256": fmt.Sprintf("%x", sha256.Sum256(pcm))})
	}
	raw, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(out, "DECODE.json"), append(raw, '\n'), 0644); err != nil {
		panic(err)
	}
	fmt.Println("PASS 4 formal Vorbis decoded sources")
}
