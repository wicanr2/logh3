package main

import (
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
