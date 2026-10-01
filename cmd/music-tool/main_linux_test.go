package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestBundledToolsRefuseMissingDependency(t *testing.T) {
	base := filepath.Join(t.TempDir(), "工具 有空白")
	for _, sub := range []string{"lib", "real"} {
		if err := os.MkdirAll(filepath.Join(base, sub), 0755); err != nil {
			t.Fatal(err)
		}
	}
	write := func(name string, data []byte) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(base, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	// 只測檔案與參數組裝，這些佔位檔不會執行或作音樂驗收。
	write("lib/ld-linux-x86-64.so.2", []byte{0})
	write("lib/libc.so.6", []byte{0})
	write("real/ffmpeg", []byte{0})
	manifest := func(d dependencies) {
		raw, err := json.Marshal(d)
		if err != nil {
			t.Fatal(err)
		}
		write("DEPENDENCIES.json", raw)
	}
	manifest(dependencies{"linux/amd64", []string{"ld-linux-x86-64.so.2", "libc.so.6"}})
	loader, args, env, err := launchConfig(base, "ffmpeg", []string{"-i", "音樂 輸入.wav"}, []string{"PATH=/none", "LD_PRELOAD=/wrong.so", "LD_LIBRARY_PATH=/wrong", "GLIBC_TUNABLES=bad", "GCONV_PATH=/wrong", "HOME=/tmp"})
	if err != nil {
		t.Fatal(err)
	}
	if loader != filepath.Join(base, "lib/ld-linux-x86-64.so.2") || args[3] != "$ORIGIN/../lib" || args[4] != filepath.Join(base, "real/ffmpeg") || args[6] != "音樂 輸入.wav" || !reflect.DeepEqual(env, []string{"PATH=/none", "HOME=/tmp"}) {
		t.Fatal("相對封包定位／參數／環境清理錯誤")
	}
	if err := os.Remove(filepath.Join(base, "lib/libc.so.6")); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := launchConfig(base, "ffmpeg", nil, nil); err == nil || !strings.Contains(err.Error(), "libc.so.6") {
		t.Fatal("缺件仍可能依賴系統函式庫")
	}
	for _, d := range []dependencies{{"windows/amd64", []string{"x"}}, {"linux/amd64", nil}, {"linux/amd64", []string{"../libc.so.6"}}, {"linux/amd64", []string{"ld-linux-x86-64.so.2", "ld-linux-x86-64.so.2"}}} {
		manifest(d)
		if _, _, _, err := launchConfig(base, "ffmpeg", nil, nil); err == nil {
			t.Fatal("錯清單未拒絕")
		}
	}
	if _, _, _, err := launchConfig(base, "other", nil, nil); err == nil {
		t.Fatal("未知工具名未拒絕")
	}
	write("DEPENDENCIES.json", []byte("{壞 JSON"))
	if _, _, _, err := launchConfig(base, "ffmpeg", nil, nil); err == nil || !strings.Contains(err.Error(), "格式錯誤") {
		t.Fatal("壞 JSON 清單未明確拒絕")
	}
}
