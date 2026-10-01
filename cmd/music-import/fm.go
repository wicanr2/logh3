package main

import (
	"debug/buildinfo"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/audio"
)

type fmLock struct {
	SourceSHA256, VGMSHA256, OGGSHA256 string
	WAVSHA256, PCMSHA256               string
	Serial                             uint32
	GainDB, Route                      string
}

type fmTrackReceipt struct {
	Track        string     `json:"track"`
	Route        string     `json:"route"`
	SourceSHA256 string     `json:"source_sha256"`
	VGMSHA256    string     `json:"vgm_sha256"`
	WAVSHA256    string     `json:"wav_sha256"`
	PCMSHA256    string     `json:"pcm_sha256"`
	GainDB       string     `json:"gain_db"`
	SampleRate   int        `json:"sample_rate"`
	OGGSHA256    string     `json:"ogg_sha256"`
	Serial       uint32     `json:"ogg_serial"`
	Seconds      float64    `json:"seconds"`
	Peak         float64    `json:"peak"`
	Commands     [][]string `json:"commands"`
}

type fmReceipt struct {
	Status      string            `json:"status"`
	ProbeHead   string            `json:"probe_head"`
	ProgramHash string            `json:"program_sha256"`
	Platform    string            `json:"platform"`
	Approx      string            `json:"approximation"`
	Tools       map[string]string `json:"tools_sha256"`
	Inputs      map[string]string `json:"source_executable_sha256"`
	Tracks      []fmTrackReceipt  `json:"tracks"`
	Error       string            `json:"error,omitempty"`
}

func selectTracks(names []string, gmOnly, fmOnly bool) (gm, fm []string, err error) {
	if gmOnly && fmOnly {
		return nil, nil, fmt.Errorf("-gm-only 與 -fm-only 不可同時使用")
	}
	allowed := map[string]bool{}
	for _, n := range audio.Tracks {
		allowed[n] = true
	}
	defaultList := names == nil
	if defaultList {
		names = audio.Tracks
	}
	seen := map[string]bool{}
	for _, name := range names {
		n := strings.ToUpper(strings.TrimSpace(name))
		if !allowed[n] || seen[n] {
			return nil, nil, fmt.Errorf("不是唯一合法曲目：%q", n)
		}
		seen[n] = true
		isGM := strings.HasSuffix(n, "M")
		if (gmOnly && !isGM) || (fmOnly && isGM) {
			if defaultList {
				continue
			}
			return nil, nil, fmt.Errorf("曲目 %s 與選定音源階段不符，不會改套另一種音源", n)
		}
		if isGM {
			gm = append(gm, n)
		} else {
			fm = append(fm, n)
		}
	}
	if len(gm)+len(fm) == 0 {
		return nil, nil, fmt.Errorf("曲目清單不可為空")
	}
	return gm, fm, nil
}

func checkProbe(path string) error {
	info, err := buildinfo.ReadFile(path)
	if err != nil {
		return fmt.Errorf("讀原生 dosgolem 身份：%w", err)
	}
	settings := map[string]string{}
	for _, s := range info.Settings {
		settings[s.Key] = s.Value
	}
	if info.Path != "github.com/wicanr2/dosgolem/cmd/probe" || settings["vcs.revision"] != fmProbeHead || settings["vcs.modified"] != "false" {
		return fmt.Errorf("dosgolem 必須為 clean %s；實際 %s（modified=%s）", fmProbeHead, settings["vcs.revision"], settings["vcs.modified"])
	}
	return nil
}

func sourceFiles(root string) (map[string]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	files := map[string]string{}
	for _, e := range entries {
		if e.Type()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("原版來源含連結，不跟隨：%s", e.Name())
		}
		if e.IsDir() {
			continue // DOS 遊戲檔在根目錄，不複製任意子樹。
		}
		name := strings.ToUpper(e.Name())
		if _, ok := files[name]; ok {
			return nil, fmt.Errorf("來源有大小寫撞名：%s", name)
		}
		files[name] = filepath.Join(root, e.Name())
	}
	return files, nil
}

func copyFile(source, dest string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	st, err := in.Stat()
	if err != nil || !st.Mode().IsRegular() || st.Size() > 64<<20 {
		return fmt.Errorf("來源不是有界普通檔案：%s", source)
	}
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	closeErr := out.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func fmProbeArgs(name, game, probe, vgm string) ([]string, error) {
	exe, steps, anchor := "GIN3.COM", "1600000000", name+".MDS"
	var extra []string
	switch name {
	case "END_1F", "END_2F":
		exe = "ENDING.EXE"
		extra = []string{"-sound-blaster", "-args", " 9 " + string(name[4])}
	case "OPEN_F":
		steps, anchor = "1200000000", "OPEN_M.MDS"
	default:
		if len(name) != 6 || !strings.HasPrefix(name, "BGM_") || name[4] < '0' || name[4] > '7' || name[5] != 'F' {
			return nil, fmt.Errorf("沒有 FM 觸發路徑：%s", name)
		}
		extra = []string{"-clicks", "40000000:300:170,95000000:300:170", "-press", "1,3,8,enter", "-press-at", "105000000", "-press-every", "2000000", "-click-hold", "3000000", "-poke", "lin:5DA64@117022200=0" + string(name[4])}
	}
	args := []string{probe, "-exe", filepath.Join(game, exe), "-root", game, "-empty-env", "-tick", "20000", "-steps", steps, "-vgm-clear-on-open", anchor}
	args = append(args, extra...)
	return append(args, "-dump-vgm", vgm), nil
}

// 規格 074：未鎖定或不符的無損中間產物不得提交，即使 VGM 已符合。
func checkFMOutputHash(path, want, label string) error {
	got, err := hashFile(path)
	if err != nil || len(want) != 64 || got != want {
		return fmt.Errorf("%s SHA-256 不符或缺檔，拒絕提交", label)
	}
	return nil
}

func runFM(o options) (result error) {
	if err := validatePaths(&o); err != nil {
		return err
	}
	for _, n := range o.tracks {
		if _, ok := fmLocks[n]; !ok {
			return fmt.Errorf("沒有 FM 鎖定資料：%s", n)
		}
	}
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	probe, opl, ff := filepath.Join(o.tools, "dosgolem-probe"+suffix), filepath.Join(o.tools, "opl2wav"+suffix), filepath.Join(o.tools, "ffmpeg"+suffix)
	r := &fmReceipt{Status: "in-progress", ProbeHead: fmProbeHead, Platform: runtime.GOOS + "/" + runtime.GOARCH, Approx: "hardware-spec approximation；原版 OPL 暫存器串、固定封裝；OPEN_F substituted；未聲稱實機聆聽", Tools: map[string]string{}, Inputs: map[string]string{}}
	for _, p := range []string{probe, opl, ff} {
		h, err := hashFile(p)
		if err != nil {
			return fmt.Errorf("缺隨包工具 %s：%w", p, err)
		}
		r.Tools[p] = h
	}
	if err := checkProbe(probe); err != nil {
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if r.ProgramHash, err = hashFile(exe); err != nil {
		return err
	}
	files, err := sourceFiles(o.source)
	if err != nil {
		return err
	}
	for name, want := range fmExecutableHashes {
		got, err := hashFile(files[name])
		if err != nil || got != want {
			return fmt.Errorf("原版 %s SHA-256 不符或缺檔", name)
		}
		r.Inputs[name] = got
	}
	for _, n := range o.tracks {
		got, err := hashFile(files[n+".MDS"])
		if err != nil || got != fmLocks[n].SourceSHA256 {
			return fmt.Errorf("原版 %s.MDS SHA-256 不符或缺檔", n)
		}
	}
	if err := os.MkdirAll(o.output, 0755); err != nil {
		return err
	}
	if err := writeJSONReceipt(o.output, "FM-SOURCES.json", r); err != nil {
		return err
	}
	defer func() {
		if result != nil {
			r.Status, r.Error = "failed", result.Error()
			if err := writeJSONReceipt(o.output, "FM-SOURCES.json", r); err != nil {
				result = fmt.Errorf("%w；失敗收據亦寫入失敗：%v", result, err)
			}
		}
	}()
	tmp, err := os.MkdirTemp(o.output, ".fm-import-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	game := filepath.Join(tmp, "game")
	if err := os.Mkdir(game, 0700); err != nil {
		return err
	}
	for name, source := range files {
		if err := copyFile(source, filepath.Join(game, name)); err != nil {
			return err
		}
	}
	for _, n := range o.tracks {
		lock := fmLocks[n]
		vgm, wav, ogg := filepath.Join(tmp, n+".vgm"), filepath.Join(tmp, n+".wav"), filepath.Join(tmp, n+".ogg")
		pcm := filepath.Join(tmp, n+".f32")
		args, err := fmProbeArgs(n, game, probe, vgm)
		if err != nil {
			return err
		}
		if n == "OPEN_F" {
			if err := copyFile(files["OPEN_F.MDS"], filepath.Join(game, "OPEN_M.MDS")); err != nil {
				return err
			}
		}
		fmt.Printf("%s：原版 OPL 擷取開始（固定指令序列）\n", n)
		log, err := command(args, 15*time.Minute)
		if n == "OPEN_F" {
			if restoreErr := copyFile(files["OPEN_M.MDS"], filepath.Join(game, "OPEN_M.MDS")); restoreErr != nil {
				return restoreErr
			}
		}
		if err != nil {
			return fmt.Errorf("%s 原版擷取：%w", n, err)
		}
		vh, err := hashFile(vgm)
		if err != nil || vh != lock.VGMSHA256 {
			return fmt.Errorf("%s 缺原版 VGM／未命中錨／SHA-256 不符，拒絕合成；%s", n, log)
		}
		commands := [][]string{args, {opl, vgm, wav, fmt.Sprint(audio.SampleRate)},
			{ff, "-v", "error", "-y", "-i", wav, "-af", "volume=" + lock.GainDB + "dB", "-ar", fmt.Sprint(audio.SampleRate), "-c:a", "pcm_f32le", "-f", "f32le", pcm},
			{ff, "-v", "error", "-y", "-i", wav, "-af", "volume=" + lock.GainDB + "dB", "-ar", fmt.Sprint(audio.SampleRate), "-c:a", "libvorbis", "-q:a", "5", ogg}}
		for i, cmd := range commands[1:] {
			if _, err := command(cmd, 15*time.Minute); err != nil {
				return err
			}
			if i == 0 {
				if err := checkFMOutputHash(wav, lock.WAVSHA256, n+" 合成 WAV"); err != nil {
					return err
				}
			} else if i == 1 {
				if err := checkFMOutputHash(pcm, lock.PCMSHA256, n+" 增益後 PCM"); err != nil {
					return err
				}
			}
		}
		raw, err := os.ReadFile(ogg)
		if err != nil {
			return err
		}
		fixed, err := asset.FixedOggSerial(raw, lock.Serial)
		if err != nil {
			return err
		}
		if runtime.GOOS == "linux" && hashBytes(fixed) != lock.OGGSHA256 {
			return fmt.Errorf("%s 固定封裝後 OGG SHA-256 仍不符，拒絕提交", n)
		}
		if err := os.WriteFile(ogg, fixed, 0600); err != nil {
			return err
		}
		seconds, peak, err := oggStats(ogg)
		if err != nil {
			return err
		}
		if err := os.Rename(ogg, filepath.Join(o.output, n+".ogg")); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(o.output, n+"-probe.log"), []byte(log), 0600); err != nil {
			return err
		}
		if err := os.Rename(vgm, filepath.Join(o.output, n+".vgm")); err != nil {
			return err
		}
		r.Tracks = append(r.Tracks, fmTrackReceipt{Track: n, Route: lock.Route,
			SourceSHA256: lock.SourceSHA256, VGMSHA256: vh, WAVSHA256: lock.WAVSHA256,
			PCMSHA256: lock.PCMSHA256, GainDB: lock.GainDB, SampleRate: audio.SampleRate,
			OGGSHA256: hashBytes(fixed), Serial: lock.Serial, Seconds: seconds, Peak: peak, Commands: commands})
		if err := writeJSONReceipt(o.output, "FM-SOURCES.json", r); err != nil {
			return err
		}
		if err := os.Remove(wav); err != nil {
			return err
		}
		if err := os.Remove(pcm); err != nil {
			return err
		}
		fmt.Printf("%s：VGM／WAV／增益後 PCM 符合基準，OGG 已解碼，%.2f 秒\n", n, seconds)
	}
	r.Status = "complete-fm-stage"
	return writeJSONReceipt(o.output, "FM-SOURCES.json", r)
}
