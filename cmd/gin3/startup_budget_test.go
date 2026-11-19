package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/wicanr2/logh3/internal/audio"
	"github.com/wicanr2/logh3/internal/bootstrap"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/presentation"
)

// TestStartupBudget 逐段量啟動載入（規格160）：主選單前擋住玩家的部分與背景預熱分開記：耗時與讀取位元組（Linux `/proc/self/io` 的 rchar）。
// 只在設了 LOGH3_STARTUP_BUDGET（發行版 usr/bin 目錄）時跑，量測用，不判對錯。
func TestStartupBudget(t *testing.T) {
	root := os.Getenv("LOGH3_STARTUP_BUDGET")
	if root == "" {
		t.Skip("沒有設 LOGH3_STARTUP_BUDGET，跳過")
	}
	orig, data, art, music := filepath.Join(root, "original"), filepath.Join(root, "data"),
		filepath.Join(root, "art"), filepath.Join(root, "music")
	rchar := func() int64 {
		b, _ := os.ReadFile("/proc/self/io")
		for _, l := range strings.Split(string(b), "\n") {
			if v, ok := strings.CutPrefix(l, "rchar: "); ok {
				n, _ := strconv.ParseInt(v, 10, 64)
				return n
			}
		}
		return -1
	}
	cpu := func() time.Duration {
		var ru syscall.Rusage
		_ = syscall.Getrusage(syscall.RUSAGE_SELF, &ru)
		return time.Duration(ru.Utime.Nano() + ru.Stime.Nano())
	}
	var res *game.Resources
	var wallSum, cpuSum time.Duration
	var bytesSum int64
	stage := func(name string, f func() error) {
		r0, t0, c0 := rchar(), time.Now(), cpu()
		if err := f(); err != nil {
			t.Fatalf("%s：%v", name, err)
		}
		w, c, b := time.Since(t0), cpu()-c0, rchar()-r0
		wallSum, cpuSum, bytesSum = wallSum+w, cpuSum+c, bytesSum+b
		t.Logf("%-12s %8.1f ms  cpu %8.1f ms %10d bytes", name, float64(w.Microseconds())/1000, float64(c.Microseconds())/1000, b)
	}
	total := func(label string) {
		t.Logf("BUDGET %s wall=%.1fms cpu=%.1fms bytes=%d", label, float64(wallSum.Microseconds())/1000, float64(cpuSum.Microseconds())/1000, bytesSum)
		wallSum, cpuSum, bytesSum = 0, 0, 0
	}
	stage("原版資料", func() (err error) { res, err = bootstrap.Load(orig, data); return })
	stage("開場動畫", func() error { return bootstrap.LoadIntro(res, orig) })
	stage("存檔", func() error { return bootstrap.UseSaveDir(res, t.TempDir()) })
	stage("初始局", func() error { _, err := game.New(res); return err })
	var hd *presentation.Renderer
	stage("HD索引", func() (err error) { hd, err = presentation.Load(art, res.Faces); return })
	effects := newEffectPlayer(nil, audio.NewSoundLibrary(music))
	var posters *posterArt
	stage("海報索引", func() (err error) { posters, err = loadPosterArt(art); return })
	total("blocking")
	// 以下是規格160移到背景的工作：不擋主選單，也不擋第一格。
	g, err := game.New(res)
	if err != nil {
		t.Fatal(err)
	}
	stage("背景:畫面規格", func() error { g.WarmResources(); return nil })
	stage("背景:音效解碼", effects.validate)
	stage("背景:投影地板", func() error { hd.Warm(); return nil })
	if posters != nil && posters.approved != nil {
		stage("背景:海報", func() error {
			posters.approved.prefetch("base-v1.png")
			posters.approved.prefetch("hulls.png")
			posters.approved.wait()
			return nil
		})
	}
	total("background")
	fmt.Fprintln(os.Stderr, "done")
}
