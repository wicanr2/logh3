//go:build ignore

// 只開啟／關閉正式ALSA default裝置，不生成測試音訊。
package main

// #cgo pkg-config: alsa
// #include <alsa/asoundlib.h>
// static int open_default_probe(void) {
//   snd_pcm_t *pcm = NULL;
//   int code = snd_pcm_open(&pcm, "default", SND_PCM_STREAM_PLAYBACK, SND_PCM_NONBLOCK);
//   if (code >= 0) snd_pcm_close(pcm);
//   return code;
// }
import "C"

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	_, err := os.Stat("/dev/snd")
	code := C.open_default_probe()
	b, err := json.MarshalIndent(map[string]any{"dev_snd_present": err == nil, "default_pcm_open_code": int(code), "default_pcm_available": code >= 0, "error": C.GoString(C.snd_strerror(code)), "pulse_server": os.Getenv("PULSE_SERVER")}, "", "  ")
	if err != nil {
		panic(err)
	}
	p := filepath.Join(os.Getenv("LOGH3_AUDIO_SINK_OUT"), "alsa-probe.json")
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		panic(err)
	}
	if _, err = f.Write(append(b, '\n')); err != nil {
		f.Close()
		panic(err)
	}
	if err = f.Close(); err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	if code < 0 {
		os.Exit(1)
	}
}
