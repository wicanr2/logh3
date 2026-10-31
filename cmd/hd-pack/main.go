// hd-pack產生正式顯示尺寸的人物圖複本；來源HD保持原樣。
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/presentation"
)

func main() {
	orig := flag.String("orig", "", "原版資料目錄")
	art := flag.String("art", "", "HD原始圖目錄")
	out := flag.String("out", "", "尚不存在的發行人物圖目錄")
	check := flag.Bool("check", false, "逐張核對既有發行圖與原runtime像素")
	flag.Parse()
	if *orig == "" || *art == "" || *out == "" {
		flag.Usage()
		os.Exit(2)
	}
	b, err := os.ReadFile(filepath.Join(*orig, "FACE.AMG"))
	if err == nil {
		var faces *asset.AMG
		faces, err = asset.OpenAMG(b)
		if err == nil {
			if *check {
				err = presentation.CheckReleasePortraits(*art, *out, faces)
			} else {
				err = presentation.WriteReleasePortraits(*art, *out, faces)
			}
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
