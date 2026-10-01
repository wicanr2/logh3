package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	intaudio "github.com/wicanr2/logh3/internal/audio"
	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
)

// Windows 原生產物由正式播放層讀完；這只驗解碼，不冒稱 Windows 音效裝置或聆聽。
func TestWindowsGMImportPlayerDecode(t *testing.T) {
	testWindowsImportPlayerDecode(t, "GM", "M")
}

func TestWindowsUnicodeGMImportPlayerDecode(t *testing.T) {
	testWindowsImportPlayerDecodeAt(t, "GM", "M", filepath.Join("..", "..", "workplace", "music69-windows", "gm-utf8-full", "生成 曲庫"))
}

func TestWindowsFMImportPlayerDecode(t *testing.T) {
	testWindowsImportPlayerDecode(t, "FM", "F")
}

func testWindowsImportPlayerDecode(t *testing.T, stage, suffix string) {
	t.Helper()
	dir := filepath.Join("..", "..", "workplace", "music69-windows", strings.ToLower(stage)+"-full")
	testWindowsImportPlayerDecodeAt(t, stage, suffix, dir)
}

func testWindowsImportPlayerDecodeAt(t *testing.T, stage, suffix, dir string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, stage+"-SOURCES.json"))
	if os.IsNotExist(err) {
		t.Skip("缺 Windows 原生 " + stage + " 驗證批次")
	}
	if err != nil {
		t.Fatal(err)
	}
	var receipt struct {
		Status   string `json:"status"`
		Platform string `json:"platform"`
		Tracks   []struct {
			Track string `json:"track"`
			Hash  string `json:"ogg_sha256"`
			VGM   string `json:"vgm_sha256"`
			WAV   string `json:"wav_sha256"`
			PCM   string `json:"pcm_sha256"`
			Gain  string `json:"gain_db"`
			Rate  int    `json:"sample_rate"`
		} `json:"tracks"`
	}
	if err := json.Unmarshal(raw, &receipt); err != nil || receipt.Status != "complete-"+strings.ToLower(stage)+"-stage" || receipt.Platform != "windows/amd64" {
		t.Fatalf("Windows %s 批次未完成或平台錯誤：%s／%s，%v", stage, receipt.Status, receipt.Platform, err)
	}
	var lock struct {
		Tracks map[string]struct {
			VGM string `json:"vgm_sha256"`
			WAV string `json:"wav_sha256"`
			PCM string `json:"pcm_sha256"`
		} `json:"tracks"`
	}
	if stage == "FM" {
		data, err := os.ReadFile("../../docs/internals/data/fm-import-lock.json")
		if err != nil || json.Unmarshal(data, &lock) != nil {
			t.Fatal("缺或損壞無損 FM 基準")
		}
	}
	expected := map[string]bool{}
	for _, name := range intaudio.Tracks {
		if strings.HasSuffix(name, suffix) {
			expected[name] = true
		}
	}
	if len(receipt.Tracks) != len(expected) {
		t.Fatalf("Windows %s 曲數 %d，預期 %d", stage, len(receipt.Tracks), len(expected))
	}
	p := newMusicPlayer(nil, intaudio.NewLibrary(dir))
	for _, track := range receipt.Tracks {
		if !expected[track.Track] {
			t.Fatalf("Windows %s 重複或非合法曲目：%q", stage, track.Track)
		}
		delete(expected, track.Track)
		t.Run(track.Track, func(t *testing.T) {
			if stage == "FM" {
				want, ok := lock.Tracks[track.Track]
				if !ok || len(want.WAV) != 64 || len(want.PCM) != 64 || track.VGM != want.VGM || track.WAV != want.WAV || track.PCM != want.PCM || track.Gain != "-1.00" || track.Rate != intaudio.SampleRate {
					t.Fatal("Windows FM 收據沒有通過完整無損閘門")
				}
				vgm, err := os.ReadFile(filepath.Join(dir, track.Track+".vgm"))
				if err != nil {
					t.Fatal(err)
				}
				h := sha256.Sum256(vgm)
				if hex.EncodeToString(h[:]) != want.VGM {
					t.Fatal("Windows 原版 VGM 與基準不同")
				}
			}
			data, err := os.ReadFile(filepath.Join(dir, track.Track+".ogg"))
			if err != nil {
				t.Fatal(err)
			}
			h := sha256.Sum256(data)
			if hex.EncodeToString(h[:]) != track.Hash {
				t.Fatal("Windows " + stage + " 產物與收據雜湊不同")
			}
			st, ok, err := p.decode(track.Track)
			if err != nil || !ok || st == nil || st.Length() <= 0 {
				t.Fatalf("Windows %s 正式解碼失敗：%v", stage, err)
			}
			if n, err := io.Copy(io.Discard, st); err != nil || n == 0 {
				t.Fatalf("Windows %s PCM 串流錯誤：%d，%v", stage, n, err)
			}
		})
	}
}

// 隨包 Linux 工具在乾淨容器產生的實物，仍須交給正式播放層完整解碼。
func TestBundledLinuxMusicImportPlayerDecode(t *testing.T) {
	base := filepath.Join("..", "..", "workplace", "music69-linux")
	raw, err := os.ReadFile(filepath.Join(base, "CHECK.json"))
	if os.IsNotExist(err) {
		t.Skip("缺 Linux 隨包工具驗收批次")
	}
	if err != nil {
		t.Fatal(err)
	}
	var receipt struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(raw, &receipt); err != nil || receipt.Status != "complete-linux-tools-sampling" {
		t.Fatalf("Linux 隨包工具批次尚未成功：%s，%v", receipt.Status, err)
	}
	p := newMusicPlayer(nil, intaudio.NewLibrary(filepath.Join(base, "imported")))
	for _, name := range []string{"BGM_0F", "BGM_0M", "END_2F", "OPEN_F"} {
		t.Run(name, func(t *testing.T) {
			st, ok, err := p.decode(name)
			if err != nil || !ok || st == nil || st.Length() <= 0 {
				t.Fatalf("隨包曲庫正式解碼失敗：%v", err)
			}
			if n, err := io.Copy(io.Discard, st); err != nil || n == 0 {
				t.Fatalf("隨包曲庫 PCM 串流錯誤：%d，%v", n, err)
			}
		})
	}
}

// 完整原生匯入器的真實曲庫經正式播放層；進行中的本機批次不冒稱驗收。
func TestNativeFullMusicImportPlayerDecode(t *testing.T) {
	dir := filepath.Join("..", "..", "workplace", "music69-fm", "full")
	raw, err := os.ReadFile(filepath.Join(dir, "FM-SOURCES.json"))
	if os.IsNotExist(err) {
		t.Skip("缺完整原生曲庫驗收批次")
	}
	if err != nil {
		t.Fatal(err)
	}
	var receipt struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Status == "in-progress" {
		t.Skip("完整原生曲庫批次尚未完成，需完成後再跑正式解碼驗收")
	}
	if receipt.Status != "complete-fm-stage" {
		t.Fatalf("完整原生 FM 批次非成功：%s", receipt.Status)
	}
	lib := intaudio.NewLibrary(dir)
	if lib.Have() != len(intaudio.Tracks) {
		t.Fatalf("完整曲庫缺曲：%v", lib.Missing())
	}
	p := newMusicPlayer(nil, lib)
	for _, name := range intaudio.Tracks {
		t.Run(name, func(t *testing.T) {
			st, ok, err := p.decode(name)
			if err != nil || !ok || st == nil || st.Length() <= 0 {
				t.Fatalf("完整曲庫正式解碼失敗：%v", err)
			}
			if n, err := io.Copy(io.Discard, st); err != nil || n == 0 {
				t.Fatalf("正式 PCM 串流錯誤：%d，%v", n, err)
			}
		})
	}
}

// 原生匯入器的實際產物須經正式播放器解碼；缺本機批次時 skip，不代造 OGG。
func TestNativeGMImportPlayerDecode(t *testing.T) {
	dir := filepath.Join("..", "..", "workplace", "music69-native", "曲庫 測試")
	if _, err := os.Stat(filepath.Join(dir, "GM-SOURCES.json")); os.IsNotExist(err) {
		t.Skip("缺原生 GM 驗收批次")
	} else if err != nil {
		t.Fatal(err)
	}
	p := newMusicPlayer(nil, intaudio.NewLibrary(dir))
	for _, name := range intaudio.Tracks {
		if !strings.HasSuffix(name, "M") {
			continue
		}
		t.Run(name, func(t *testing.T) {
			st, ok, err := p.decode(name)
			if err != nil || !ok || st == nil || st.Length() <= 0 {
				t.Fatalf("原生曲庫正式解碼失敗：%v", err)
			}
			n, err := io.Copy(io.Discard, st)
			if err != nil || n <= 0 {
				t.Fatalf("正式 PCM 串流沒有內容：%d，%v", n, err)
			}
			t.Logf("正式播放器 PCM：%d bytes", n)
		})
	}
}

type recordingMusic struct {
	playing     string
	transitions []string
}

func (p *recordingMusic) Play(track string) error {
	if p.playing == track {
		return nil
	}
	p.playing = track
	p.transitions = append(p.transitions, track)
	return nil
}

func (p *recordingMusic) Stop() {
	if p.playing == "" {
		return
	}
	p.playing = ""
	p.transitions = append(p.transitions, "<stop>")
}

func (p *recordingMusic) Playing() string { return p.playing }

func playerPathGame(t *testing.T) *game.Game {
	t.Helper()
	orig := filepath.Join("..", "..", "org_game", "extracted")
	data := filepath.Join("..", "..", "docs", "internals", "data")
	if _, err := os.Stat(filepath.Join(orig, "GIN3.COM")); err != nil {
		t.Skip("沒有唯讀原版資料，跳過正常玩家路徑音樂整合")
	}
	res, err := bootstrap.Load(orig, data)
	if err != nil {
		t.Fatal(err)
	}
	g, err := game.New(res)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

// TestPlayerPathMusicSwitches 從正式 Game 狀態走，不直接餵 TuneFor：主選單與
// 防拷保持安靜、進遊戲選平時曲、設定畫面換曲會切換、改戰鬥曲不試聽、
// 關閉音源會停止。主觀音量與音色不在這支機器測試的裁決範圍。
func TestPlayerPathMusicSwitches(t *testing.T) {
	g := playerPathGame(t)
	p := &recordingMusic{}
	a := &app{g: g, music: p}

	mustSync := func() {
		t.Helper()
		if err := a.syncMusic(); err != nil {
			t.Fatal(err)
		}
	}
	mustSync() // 主選單
	g.Click(300, 210)
	mustSync() // 防拷問答
	for _, key := range []byte{'1', '3', '8', '\r'} {
		g.Key(key)
	}
	mustSync() // 遊戲本體：BGM_0F

	g.RightClick(300, 240)
	g.RightClick(300, 240)
	g.Click(544, 125)
	g.Click(560, 185) // 設定
	mustSync()
	g.Click(420, 150) // 平時配樂 1
	mustSync()
	g.Click(420, 270) // 戰鬥配樂 1；原版不試聽
	mustSync()
	g.Click(410, 105) // 音源關閉
	mustSync()
	g.Click(360, 105) // 音源回到 FM
	mustSync()

	want := []string{"BGM_0F", "BGM_1F", "<stop>", "BGM_1F"}
	if !reflect.DeepEqual(p.transitions, want) {
		t.Fatalf("播放器切換 = %v，預期 %v", p.transitions, want)
	}
	if got := p.Playing(); got != "BGM_1F" {
		t.Fatalf("最後播放 %q，預期 BGM_1F", got)
	}
}

// 缺檔與壞檔在建立音效裝置之前就能裁決：缺檔靜默略過，壞 OGG 明確回錯。
func TestMusicPlayerMissingAndInvalidFile(t *testing.T) {
	dir := t.TempDir()
	p := newMusicPlayer(nil, intaudio.NewLibrary(dir))
	if err := p.Play("BGM_0F"); err != nil || p.Playing() != "" {
		t.Fatalf("缺檔：錯誤 %v、Playing %q", err, p.Playing())
	}

	if err := os.WriteFile(filepath.Join(dir, "BGM_0F.ogg"), []byte("不是 OGG"), 0o644); err != nil {
		t.Fatal(err)
	}
	p = newMusicPlayer(nil, intaudio.NewLibrary(dir))
	if err := p.Play("BGM_0F"); err == nil || !strings.Contains(err.Error(), "解 BGM_0F 的 OGG") {
		t.Fatalf("壞 OGG 沒有明確錯誤：%v", err)
	}
}

// 玩家自備曲庫若存在，逐首走正式播放器的 decode 與 44.1 kHz 閘門；素材本身
// 留在 gitignore 的 workplace，乾淨 clone 沒有素材時依專案契約 skip。
func TestAvailableMusicLibraryDecodes(t *testing.T) {
	dir := filepath.Join("..", "..", "workplace", "assets", "music", "ogg")
	lib := intaudio.NewLibrary(dir)
	if lib.Have() == 0 {
		t.Skip("沒有玩家自備 OGG，跳過曲庫解碼")
	}
	p := newMusicPlayer(nil, lib)
	for _, track := range intaudio.Tracks {
		if _, found := lib.Path(track); !found {
			continue
		}
		st, ok, err := p.decode(track)
		if err != nil {
			t.Errorf("%s：%v", track, err)
			continue
		}
		if !ok || st == nil || st.Length() <= 0 {
			t.Errorf("%s：解碼後沒有音訊內容", track)
		}
	}
}
