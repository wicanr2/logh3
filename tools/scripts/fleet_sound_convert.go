// 建置後由tools/audio.sh執行，原版只讀、輸出拒絕覆寫。
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/logh3/internal/asset"
	"math"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	orig := flag.String("orig", "/orig/GIN3SP.VOC", "玩家自備原版SVF")
	out := flag.String("out", "", "全新輸出目錄")
	flag.Parse()
	if *out == "" {
		return fmt.Errorf("需要-out新目錄")
	}
	if _, err := os.Stat(*out); !os.IsNotExist(err) {
		return fmt.Errorf("輸出已存在或無法核對")
	}
	b, err := os.ReadFile(*orig)
	if err != nil {
		return err
	}
	sha := fmt.Sprintf("%x", sha256.Sum256(b))
	if sha != "b36a1b43f3199e776f70302671a6bc37698dd5255113995e7a7f55f78f699a9f" {
		return fmt.Errorf("原版SVF雜湊不符")
	}
	svf, err := asset.DecodeSVF(b)
	if err != nil {
		return err
	}
	if len(svf.Samples) != 28 {
		return fmt.Errorf("需要28段音效")
	}
	version, err := exec.Command("ffmpeg", "-version").Output()
	if err != nil {
		return err
	}
	if err = os.MkdirAll(*out, 0755); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(*out, ".source-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	var records []map[string]any
	for i, s := range svf.Samples {
		if s.TimeConstant != 166 {
			return fmt.Errorf("音效%d時間常數非本版166", i)
		}
		input := filepath.Join(tmp, fmt.Sprintf("%02d.u8", i))
		if err := os.WriteFile(input, s.PCM(), 0600); err != nil {
			return err
		}
		name := fmt.Sprintf("SFX_%02d.ogg", i)
		path := filepath.Join(*out, name)
		args := []string{"-v", "error", "-f", "u8", "-ar", "11111", "-ac", "1", "-i", input, "-af", "volume=-1dB", "-ar", "44100", "-ac", "2", "-c:a", "libvorbis", "-q:a", "5", "-map_metadata", "-1", path}
		if output, err := exec.Command("ffmpeg", args...).CombinedOutput(); err != nil {
			return fmt.Errorf("轉音效%d：%w %s", i, err, output)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		raw, err = asset.FixedOggSerial(raw, 0x53465800+uint32(i))
		if err != nil {
			return err
		}
		if err = os.WriteFile(path, raw, 0644); err != nil {
			return err
		}
		probe, err := exec.Command("ffprobe", "-v", "error", "-show_entries", "stream=duration_ts,sample_rate,channels", "-of", "json", path).Output()
		if err != nil {
			return err
		}
		var metadata struct {
			Streams []struct {
				Frames   int64  `json:"duration_ts"`
				Rate     string `json:"sample_rate"`
				Channels int    `json:"channels"`
			} `json:"streams"`
		}
		if err := json.Unmarshal(probe, &metadata); err != nil {
			return err
		}
		if len(metadata.Streams) != 1 || metadata.Streams[0].Rate != "44100" || metadata.Streams[0].Channels != 2 {
			return fmt.Errorf("OGG音訊形態不符")
		}
		validFrames := metadata.Streams[0].Frames
		pcm, err := exec.Command("ffmpeg", "-v", "error", "-i", path, "-af", fmt.Sprintf("atrim=end_sample=%d", validFrames), "-f", "f32le", "-ac", "2", "-ar", "44100", "-").Output()
		if err != nil {
			return err
		}
		if len(pcm)%8 != 0 {
			return fmt.Errorf("PCM非雙聲道float")
		}
		peak, energy := 0.0, 0.0
		for p := 0; p < len(pcm); p += 4 {
			f := float64(math.Float32frombits(binary.LittleEndian.Uint32(pcm[p:])))
			peak = math.Max(peak, math.Abs(f))
			energy += f * f
		}
		frames := len(pcm) / 8
		duration := float64(validFrames) / 44100
		if int64(frames) > validFrames || int64(frames) < validFrames-256 {
			return fmt.Errorf("OGG有效解碼長度不符：%d／%d", frames, validFrames)
		}
		if peak < 0.001 || peak > 1 || math.Abs(duration-s.Duration().Seconds()) > 0.005 {
			return fmt.Errorf("音效%d技術驗收：peak%f duration%f，原樣本%f", i, peak, duration, s.Duration().Seconds())
		}
		records = append(records, map[string]any{"id": i, "file": name, "source_pcm_sha256": fmt.Sprintf("%x", sha256.Sum256(s.PCM())), "ogg_sha256": fmt.Sprintf("%x", sha256.Sum256(raw)), "decoded_pcm_sha256": fmt.Sprintf("%x", sha256.Sum256(pcm)), "source_samples": len(s.PCM()), "source_duration_ns": int64(s.Duration()), "decoded_frames": frames, "duration_seconds": duration, "peak": peak, "rms": math.Sqrt(energy / float64(len(pcm)/4)), "command": append([]string{"ffmpeg"}, args...), "rate": 44100, "channels": 2, "gain_db": -1, "serial": 0x53465800 + uint32(i)})
	}
	receipt := map[string]any{"schema": "logh3.fleet-sound-convert.v1", "source_sha256": sha, "ffmpeg_version": string(version), "tracks": records, "timing_grade": "hardware-spec approximation", "listened": false, "rights": "原版衍生素材；本機或授權私有保存，不進發行包"}
	j, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(*out, "SFX-RECEIPT.json"), append(j, '\n'), 0644); err != nil {
		return err
	}
	fmt.Println("28段OGG技術驗收通過，實際聆聽待驗：", *out)
	return nil
}
