package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGMRejectsReportedFailureDespiteSuccessfulExit(t *testing.T) {
	for _, output := range []string{
		"fluidsynth: error: Failed to load SoundFont\r\nRendering audio to file",
		"Parameter 'Z:\\??\\FluidR3_GM.sf2' not a SoundFont or MIDI file or error occurred identifying it.",
	} {
		if err := checkGMToolOutput(output); err == nil || !strings.Contains(err.Error(), output) {
			t.Fatalf("未保留並拒絕成功退出的載入失敗：%v", err)
		}
	}
	for _, output := range []string{"", "FluidSynth runtime version 2.3.1\r\nRendering audio to file", "fluidsynth: warning: Failed to set thread to high priority"} {
		if err := checkGMToolOutput(output); err != nil {
			t.Fatalf("正常輸出或非致命警告被誤拒絕：%v", err)
		}
	}
}

func TestMusicImportPreflight(t *testing.T) {
	source := t.TempDir()
	outside := t.TempDir()
	for _, output := range []string{source, filepath.Join(source, "music")} {
		o := options{source: source, output: output, tools: outside, font: filepath.Join(outside, "sf2")}
		if validatePaths(&o) == nil {
			t.Fatal("允許寫進原版目錄")
		}
	}
	link := filepath.Join(outside, "source-link")
	if err := os.Symlink(source, link); err == nil {
		o := options{source: source, output: filepath.Join(link, "new", "music"), tools: outside, font: filepath.Join(outside, "sf2")}
		if validatePaths(&o) == nil {
			t.Fatal("父目錄連結繞過原始資料保護")
		}
	}
	for _, names := range [][]string{{}, {"BGM_0F"}, {"OPEN_M", "open_m"}, {"../OPEN_M"}} {
		if _, err := gmTracks(names); err == nil {
			t.Fatalf("錯誤曲目未拒絕：%v", names)
		}
	}
	if names, err := gmTracks(nil); err != nil || len(names) != 11 {
		t.Fatalf("預設 GM 清單錯誤：%v，%v", names, err)
	}
	o := options{source: source, output: filepath.Join(outside, "output"), tools: outside, font: filepath.Join(outside, "bad.sf2"), tracks: []string{"OPEN_M"}}
	if err := os.WriteFile(o.font, []byte("wrong font"), 0600); err != nil {
		t.Fatal(err)
	}
	if run(o) == nil {
		t.Fatal("錯誤音色庫未拒絕")
	}
	f, err := os.OpenFile(o.font, os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(148398306); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()
	if err := run(o); err == nil {
		t.Fatal("長度正確但 SHA-256 不符的音色庫未拒絕")
	}
	if _, err := os.Stat(o.output); !os.IsNotExist(err) {
		t.Fatal("前置拒絕仍留下輸出")
	}
}
