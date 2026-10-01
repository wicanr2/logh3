// music-import 是玩家自產曲庫的原生 FM／GM 入口（規格 069／071）。
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jfreymuth/oggvorbis"
	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/audio"
)

const fontHash = "74594e8f4250680adf590507a306655a299935343583256f3b722c48a1bc1cb0"

type options struct {
	source, output, tools, font string
	tracks                      []string
}

type trackReceipt struct {
	Track        string     `json:"track"`
	Source       string     `json:"source"`
	SourceSHA256 string     `json:"source_sha256"`
	MIDISHA256   string     `json:"midi_sha256"`
	OGGSHA256    string     `json:"ogg_sha256"`
	Seconds      float64    `json:"seconds"`
	Peak         float64    `json:"peak"`
	Commands     [][]string `json:"commands"`
}

type receipt struct {
	Scope       string            `json:"scope"`
	Status      string            `json:"status"`
	Approx      string            `json:"approximation"`
	Platform    string            `json:"platform"`
	ProgramHash string            `json:"program_sha256"`
	FontHash    string            `json:"soundfont_sha256"`
	Tools       map[string]string `json:"tools_sha256"`
	Versions    map[string]string `json:"tool_versions"`
	Tracks      []trackReceipt    `json:"tracks"`
	Error       string            `json:"error,omitempty"`
}

func main() {
	exe, err := os.Executable()
	if err != nil {
		fail(err)
	}
	base := filepath.Dir(exe)
	gm := flag.Bool("gm-only", false, "只轉 GM 曲目；預設完整 FM／GM 曲庫")
	fm := flag.Bool("fm-only", false, "只轉原版 OPL 的 FM 曲目")
	source := flag.String("source", "", "玩家自備的原版遊戲目錄（唯讀）")
	output := flag.String("output", "", "輸出 OGG 曲庫目錄")
	tools := flag.String("tools", filepath.Join(base, "tools"), "隨包工具目錄；診斷可指定絕對目錄")
	font := flag.String("soundfont", filepath.Join(base, "soundfont", "FluidR3_GM.sf2"), "隨包的固定音色庫")
	tracks := flag.String("tracks", "", "只轉指定曲目，以逗號分隔；預設完整二十二首")
	flag.Parse()
	if flag.NArg() != 0 {
		fail(fmt.Errorf("不接受未具名的參數"))
	}
	var names []string
	if *tracks != "" {
		names = strings.Split(*tracks, ",")
	}
	gmNames, fmNames, err := selectTracks(names, *gm, *fm)
	if err != nil {
		fail(err)
	}
	o := options{source: *source, output: *output, tools: *tools, font: *font}
	if len(gmNames) > 0 {
		o.tracks = gmNames
		if err := run(o); err != nil {
			fail(err)
		}
	}
	if len(fmNames) > 0 {
		o.tracks = fmNames
		if err := runFM(o); err != nil {
			fail(err)
		}
	}
	fmt.Printf("本次匯入完成：GM %d 首、FM %d 首。\n", len(gmNames), len(fmNames))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "音樂匯入失敗：", err)
	os.Exit(1)
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func hashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// canonical 對尚未存在的輸出也解開父層 symlink，避免寫進原版目錄。
func canonical(path string) (string, error) {
	p, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	var tail []string
	for {
		if _, err := os.Lstat(p); err == nil {
			resolved, err := filepath.EvalSymlinks(p)
			if err != nil {
				return "", err
			}
			for i := len(tail) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, tail[i])
			}
			return resolved, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		tail = append(tail, filepath.Base(p))
		parent := filepath.Dir(p)
		if parent == p {
			return "", fmt.Errorf("無法解析路徑：%s", path)
		}
		p = parent
	}
}

func validatePaths(o *options) error {
	if o.source == "" || o.output == "" {
		return fmt.Errorf("必須提供 -source 與 -output")
	}
	var err error
	if o.source, err = canonical(o.source); err != nil {
		return err
	}
	st, err := os.Stat(o.source)
	if err != nil || !st.IsDir() {
		return fmt.Errorf("原版來源目錄不存在：%s", o.source)
	}
	if o.output, err = canonical(o.output); err != nil {
		return err
	}
	rel, err := filepath.Rel(o.source, o.output)
	if err != nil {
		// Windows 不同磁碟是分離的來源／輸出。
		if filepath.VolumeName(o.source) == filepath.VolumeName(o.output) {
			return err
		}
	} else if rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
		return fmt.Errorf("輸出不可位於原版來源目錄內")
	}
	if !filepath.IsAbs(o.tools) || !filepath.IsAbs(o.font) {
		return fmt.Errorf("工具目錄與音色庫須為絕對路徑")
	}
	return nil
}

func gmTracks(selected []string) ([]string, error) {
	allowed := map[string]bool{}
	var all []string
	for _, n := range audio.Tracks {
		if strings.HasSuffix(n, "M") {
			allowed[n] = true
			all = append(all, n)
		}
	}
	if selected == nil {
		return all, nil
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("曲目清單不可為空")
	}
	seen := map[string]bool{}
	for i, n := range selected {
		n = strings.ToUpper(strings.TrimSpace(n))
		if !allowed[n] || seen[n] {
			return nil, fmt.Errorf("不是唯一合法 GM 曲目：%q；FM 須走原版 OPL 管線", n)
		}
		seen[n], selected[i] = true, n
	}
	return selected, nil
}

func command(args []string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...) // 無 shell、PATH 或 Docker。
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("工具 %s 失敗：%w；%s", filepath.Base(args[0]), err, out)
	}
	return string(out), nil
}

func oggStats(path string) (seconds, peak float64, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	r, err := oggvorbis.NewReader(f)
	if err != nil {
		return 0, 0, err
	}
	if r.SampleRate() != audio.SampleRate || r.Channels() != 2 {
		return 0, 0, fmt.Errorf("OGG 格式不是 44100 Hz 立體聲")
	}
	buf, total := make([]float32, 8192), 0
	for {
		n, readErr := r.Read(buf)
		total += n
		for _, sample := range buf[:n] {
			v := math.Abs(float64(sample))
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return 0, 0, fmt.Errorf("OGG 含無效 PCM 取樣")
			}
			peak = math.Max(peak, v)
		}
		if total > audio.SampleRate*2*1800 {
			return 0, 0, fmt.Errorf("OGG 超過 30 分鐘護欄")
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return 0, 0, readErr
		}
	}
	if total == 0 || peak < 0.000001 {
		return 0, 0, fmt.Errorf("OGG 無取樣或完全靜音")
	}
	return float64(total) / float64(audio.SampleRate*2), peak, nil
}

func writeReceipt(output string, r *receipt) error {
	return writeJSONReceipt(output, "GM-SOURCES.json", r)
}

func writeJSONReceipt(output, name string, r any) error {
	d, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(output, ".gm-receipt-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	_, err = f.Write(append(d, '\n'))
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(f.Name(), filepath.Join(output, name))
}

func run(o options) (result error) {
	if err := validatePaths(&o); err != nil {
		return err
	}
	names, err := gmTracks(o.tracks)
	if err != nil {
		return err
	}
	st, err := os.Stat(o.font)
	if err != nil || !st.Mode().IsRegular() || st.Size() != 148398306 {
		return fmt.Errorf("缺固定 FluidR3_GM 音色庫或長度不符")
	}
	fh, err := hashFile(o.font)
	if err != nil || fh != fontHash {
		return fmt.Errorf("FluidR3_GM SHA-256 不符，拒絕合成")
	}
	r := &receipt{Scope: "GM 合成收據；FM 另見 FM-SOURCES.json，平台封包未驗收", Status: "in-progress", Approx: "legacy approximation：FA 當 BPM；raw 時間沿用匯出器；FB／FC 不展開反覆；非原版 MPU 音色", Platform: runtime.GOOS + "/" + runtime.GOARCH, FontHash: fh, Tools: map[string]string{}, Versions: map[string]string{}}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if r.ProgramHash, err = hashFile(exe); err != nil {
		return err
	}
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	fs, ff := filepath.Join(o.tools, "fluidsynth"+suffix), filepath.Join(o.tools, "ffmpeg"+suffix)
	for _, p := range []string{fs, ff} {
		if r.Tools[p], err = hashFile(p); err != nil {
			return fmt.Errorf("缺隨包工具 %s：%w", p, err)
		}
		version := "--version"
		if p == ff {
			version = "-version"
		}
		if r.Versions[p], err = command([]string{p, version}, 15*time.Second); err != nil {
			return err
		}
	}
	entries, err := os.ReadDir(o.source)
	if err != nil {
		return err
	}
	files := map[string]string{}
	for _, e := range entries {
		key := strings.ToUpper(e.Name())
		if _, exists := files[key]; exists {
			return fmt.Errorf("來源有大小寫撞名：%s", key)
		}
		files[key] = filepath.Join(o.source, e.Name())
	}
	type input struct {
		source     string
		data, midi []byte
	}
	inputs := map[string]input{}
	for _, n := range names {
		p := files[n+".MDS"]
		st, err := os.Stat(p)
		if err != nil || !st.Mode().IsRegular() || st.Size() > 1<<20 {
			return fmt.Errorf("缺原版曲目或形態／長度不符：%s.MDS", n)
		}
		d, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		mid, err := asset.MDSMIDI(d)
		if err != nil {
			return fmt.Errorf("%s：%w", n, err)
		}
		inputs[n] = input{p, d, mid}
	}
	if err := os.MkdirAll(o.output, 0755); err != nil {
		return err
	}
	if err := writeReceipt(o.output, r); err != nil {
		return err
	}
	defer func() {
		if result != nil {
			r.Status, r.Error = "failed", result.Error()
			if err := writeReceipt(o.output, r); err != nil {
				result = fmt.Errorf("%w；失敗收據寫入亦失敗：%v", result, err)
			}
		}
	}()
	tmp, err := os.MkdirTemp(o.output, ".gm-import-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp) // 只移除本程序建立的獨立暫存。
	for _, n := range names {
		in := inputs[n]
		mid, wav, ogg := filepath.Join(tmp, n+".mid"), filepath.Join(tmp, n+".wav"), filepath.Join(tmp, n+".ogg")
		if err := os.WriteFile(mid, in.midi, 0600); err != nil {
			return err
		}
		commands := [][]string{
			{fs, "-ni", "-F", wav, "-r", fmt.Sprint(audio.SampleRate), "-g", "0.8", o.font, mid},
			{ff, "-v", "error", "-y", "-i", wav, "-ar", fmt.Sprint(audio.SampleRate), "-c:a", "libvorbis", "-q:a", "5", ogg},
		}
		for _, args := range commands {
			if _, err := command(args, 15*time.Minute); err != nil {
				return fmt.Errorf("曲目 %s：%w", n, err)
			}
		}
		seconds, peak, err := oggStats(ogg)
		if err != nil {
			return fmt.Errorf("曲目 %s 輸出驗證：%w", n, err)
		}
		oh, err := hashFile(ogg)
		if err != nil {
			return err
		}
		if err := os.Rename(ogg, filepath.Join(o.output, n+".ogg")); err != nil {
			return err
		}
		r.Tracks = append(r.Tracks, trackReceipt{n, in.source, hashBytes(in.data), hashBytes(in.midi), oh, seconds, peak, commands})
		if err := writeReceipt(o.output, r); err != nil {
			return err
		}
		fmt.Printf("%s：%.2f 秒，峰值 %.4f，已轉 OGG\n", n, seconds, peak)
		for _, p := range []string{mid, wav} {
			if err := os.Remove(p); err != nil {
				return err
			}
		}
	}
	r.Status = "complete-gm-stage"
	return writeReceipt(o.output, r)
}
