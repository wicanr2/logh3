//go:build linux

package main

/*
#cgo pkg-config: alsa
#include <alsa/asoundlib.h>

static int logh3_default_pcm_available(void) {
    snd_pcm_t *pcm = NULL;
    int code = snd_pcm_open(&pcm, "default", SND_PCM_STREAM_PLAYBACK, SND_PCM_NONBLOCK);
    if (code < 0) return 0;
    snd_pcm_close(pcm);
    return 1;
}
*/
import "C"

import (
	"os"
	"path/filepath"
	"strconv"
)

// ALSA default 可以連至 PulseAudio，不一定需要 /dev/snd。
// 先確認已配置端點，再試開真正的 default，避免失效端點被當成裝置。
// 初始化後被占用或服務斷線仍由正式後端回報錯誤。
func hasAudioDevice() bool {
	if _, err := os.Stat("/dev/snd"); err == nil {
		return true
	}
	configured := os.Getenv("PULSE_SERVER") != ""
	for _, dir := range []string{os.Getenv("XDG_RUNTIME_DIR"), filepath.Join("/run/user", strconv.Itoa(os.Getuid()))} {
		if dir == "" {
			continue
		}
		if info, err := os.Stat(filepath.Join(dir, "pulse/native")); err == nil && info.Mode()&os.ModeSocket != 0 {
			configured = true
		}
	}
	return configured && C.logh3_default_pcm_available() != 0
}
