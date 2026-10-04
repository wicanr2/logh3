//go:build linux

package main

import (
	"os"
	"path/filepath"
	"testing"
)

// 正反案例都使用實際 ALSA。正案例由 tools/playback.sh 提供隔離 PulseAudio。
func TestAudioDeviceActualDefault(t *testing.T) {
	mode := os.Getenv("LOGH3_AUDIO_DEVICE_TEST")
	if mode == "" {
		t.Skip("需明示實際裝置案例：pulse、absent 或 broken")
	}
	if _, err := os.Stat("/dev/snd"); !os.IsNotExist(err) {
		t.Fatal("此案例需要沒有實體 /dev/snd 的容器")
	}
	want := false
	switch mode {
	case "pulse":
		if os.Getenv("PULSE_SERVER") == "" {
			t.Fatal("未提供 PulseAudio 服務")
		}
		want = true
	case "absent":
		t.Setenv("PULSE_SERVER", "")
		t.Setenv("XDG_RUNTIME_DIR", "")
	case "broken":
		t.Setenv("PULSE_SERVER", "unix:/tmp/logh3-no-such-server/native")
		t.Setenv("XDG_RUNTIME_DIR", "")
		config := filepath.Join(t.TempDir(), "alsa.conf")
		if err := os.WriteFile(config, []byte("</usr/share/alsa/alsa.conf>\npcm.!default { type pulse server \"unix:/tmp/logh3-no-such-server/native\" }\n"), 0600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("ALSA_CONFIG_PATH", config)
	default:
		t.Fatal("未知案例", mode)
	}
	if got := hasAudioDevice(); got != want {
		t.Fatalf("實際 %s 播放端可用=%t，預期%t", mode, got, want)
	}
}
