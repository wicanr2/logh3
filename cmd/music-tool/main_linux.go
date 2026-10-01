// Linux 隨包第三方音樂工具的原生啟動器（規格 072）。
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

type dependencies struct {
	Platform  string   `json:"platform"`
	Libraries []string `json:"libraries"`
}

func regular(path string) error {
	s, err := os.Lstat(path)
	if err != nil || !s.Mode().IsRegular() {
		return fmt.Errorf("缺隨包普通檔案：%s", path)
	}
	return nil
}

func launchConfig(base, name string, args, env []string) (string, []string, []string, error) {
	if name != "ffmpeg" && name != "fluidsynth" && name != "opl2wav" {
		return "", nil, nil, fmt.Errorf("未知隨包工具：%s", name)
	}
	p := filepath.Join(base, "DEPENDENCIES.json")
	if err := regular(p); err != nil {
		return "", nil, nil, err
	}
	st, err := os.Stat(p)
	if err != nil || st.Size() > 1<<20 {
		return "", nil, nil, fmt.Errorf("依賴清單超過 1 MiB 或不可讀")
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		return "", nil, nil, err
	}
	var d dependencies
	if err := json.Unmarshal(raw, &d); err != nil {
		return "", nil, nil, fmt.Errorf("依賴清單格式錯誤：%w", err)
	}
	if d.Platform != "linux/amd64" || len(d.Libraries) == 0 || len(d.Libraries) > 512 {
		return "", nil, nil, fmt.Errorf("依賴清單平台或項數錯誤")
	}
	seen := map[string]bool{}
	for _, lib := range d.Libraries {
		if lib == "" || lib == "." || lib == ".." || strings.ContainsAny(lib, `/\`) || seen[lib] {
			return "", nil, nil, fmt.Errorf("依賴清單檔名錯誤：%q", lib)
		}
		seen[lib] = true
		if err := regular(filepath.Join(base, "lib", lib)); err != nil {
			return "", nil, nil, err
		}
	}
	loader := filepath.Join(base, "lib", "ld-linux-x86-64.so.2")
	program := filepath.Join(base, "real", name)
	for _, path := range []string{loader, program} {
		if err := regular(path); err != nil {
			return "", nil, nil, err
		}
	}
	var clean []string
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, "LD_") || key == "GLIBC_TUNABLES" || key == "GCONV_PATH" || key == "LOCPATH" {
			continue
		}
		clean = append(clean, entry)
	}
	// $ORIGIN 由載入器解析，避免把安裝目錄中的空白／分隔符當成搜尋列表。
	argv := []string{loader, "--inhibit-cache", "--library-path", "$ORIGIN/../lib", program}
	return loader, append(argv, args...), clean, nil
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "隨包音樂工具失敗：", err)
		os.Exit(1)
	}
}

func run() error {
	if runtime.GOARCH != "amd64" {
		return fmt.Errorf("此工具包只支援 Linux amd64")
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return err
	}
	loader, args, env, err := launchConfig(filepath.Dir(exe), filepath.Base(os.Args[0]), os.Args[1:], os.Environ())
	if err != nil {
		return err
	}
	return syscall.Exec(loader, args, env) // 原生替換程序；保留退出碼與逾時信號。
}
