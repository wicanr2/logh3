package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFMImportLocksAndRoutes(t *testing.T) {
	d, err := os.ReadFile("../../docs/internals/data/fm-import-lock.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		ProbeHead string            `json:"probe_head"`
		EXE       map[string]string `json:"executables"`
		Tracks    map[string]struct {
			Source string `json:"source_sha256"`
			VGM    string `json:"vgm_sha256"`
			OGG    string `json:"ogg_sha256"`
			Serial uint32 `json:"serial"`
			Gain   string `json:"gain_db"`
			Route  string `json:"route"`
		} `json:"tracks"`
	}
	if err := json.Unmarshal(d, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.ProbeHead != fmProbeHead || !reflect.DeepEqual(doc.EXE, fmExecutableHashes) || len(doc.Tracks) != 11 || len(fmLocks) != 11 {
		t.Fatal("原生生成表格與權威 JSON 不一致")
	}
	for name, l := range doc.Tracks {
		if fmLocks[name] != (fmLock{l.Source, l.VGM, l.OGG, l.Serial, l.Gain, l.Route}) {
			t.Fatalf("生成表格與 JSON 不一致：%s", name)
		}
		args, err := fmProbeArgs(name, "/game", "/probe", "/output.vgm")
		if err != nil {
			t.Fatal(err)
		}
		value := func(flag string) string {
			for i, a := range args {
				if a == flag && i+1 < len(args) {
					return args[i+1]
				}
			}
			return ""
		}
		if value("-tick") != "20000" || value("-dump-vgm") != "/output.vgm" {
			t.Fatal("固定原版收據參數遺失")
		}
		switch name {
		case "OPEN_F":
			if value("-steps") != "1200000000" || value("-vgm-clear-on-open") != "OPEN_M.MDS" || value("-press") != "" {
				t.Fatal("OPEN 替換資料路徑／步數／輸入錯誤")
			}
		case "END_1F", "END_2F":
			if value("-exe") != "/game/ENDING.EXE" || value("-args") != " 9 "+string(name[4]) {
				t.Fatal("ENDING 原版參數錯誤")
			}
			hardware := false
			for _, a := range args {
				hardware = hardware || a == "-sound-blaster"
			}
			if !hardware {
				t.Fatal("ENDING 少了原版必需的 Sound Blaster 相容埠")
			}
		default:
			if value("-poke") != "lin:5DA64@117022200=0"+string(name[4]) || value("-steps") != "1600000000" {
				t.Fatal("BGM 原版固定狀態注入錯誤")
			}
		}
	}
	gm, fm, err := selectTracks(nil, false, false)
	if err != nil || len(gm) != 11 || len(fm) != 11 {
		t.Fatal("完整入口沒有兩種音源各十一首")
	}
	for _, flags := range [][2]bool{{true, false}, {false, true}} {
		gm, fm, err := selectTracks(nil, flags[0], flags[1])
		if err != nil || len(gm)+len(fm) != 11 {
			t.Fatal("單階段入口錯誤")
		}
	}
	for _, names := range [][]string{{}, {"../OPEN_F"}, {"OPEN_F", "open_f"}, {"BGM_8F"}} {
		if _, _, err := selectTracks(names, false, false); err == nil {
			t.Fatal("錯曲目未拒絕")
		}
	}
	if _, _, err := selectTracks(nil, true, true); err == nil {
		t.Fatal("音源階段旗標未互斥")
	}
	if _, _, err := selectTracks([]string{"OPEN_F"}, true, false); err == nil {
		t.Fatal("FM 被送去 GM")
	}
}

func TestFMProbeIdentity(t *testing.T) {
	p := filepath.Join("../../workplace/music69-fm/tools", "dosgolem-probe")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		t.Skip("缺本輪真實原生 dosgolem，不製造替代執行器")
	}
	if err := checkProbe(p); err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if checkProbe(exe) == nil {
		t.Fatal("其他原生 Go 程式被誤認為已驗 dosgolem")
	}
}
