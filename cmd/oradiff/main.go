// Command oradiff 把 remake 的畫面跟 oracle 收據逐位元組比。
//
// **這是唯一算數的驗收。** remake 的單元測試全綠只證明它內部自洽，
// 不證明它跟原版一樣；目視相似更不算數。這支比的是色號陣列，
// 顏色錯與圖形錯因此分得開。
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

const W, H = 640, 480

func main() {
	a := flag.String("oracle", "", "oracle 的色號陣列（dosgolem -dump-ega 的 .bin）")
	b := flag.String("remake", "", "remake 的色號陣列")
	pa := flag.String("oracle-pal", "", "oracle 的調色盤（256×3）")
	pb := flag.String("remake-pal", "", "remake 的調色盤")
	diff := flag.String("diff", "", "差分圖輸出路徑（PNG）")
	flag.Parse()
	if *a == "" || *b == "" {
		fmt.Fprintln(os.Stderr, "需要 -oracle 與 -remake")
		os.Exit(2)
	}
	x, y := mustRead(*a, W*H), mustRead(*b, W*H)

	if *pa != "" && *pb != "" {
		q, r := mustRead(*pa, 256*3), mustRead(*pb, 256*3)
		var bad int
		for i := 0; i < 16*3; i++ {
			if q[i] != r[i] {
				bad++
			}
		}
		if bad == 0 {
			fmt.Println("調色盤（前 16 色）：一致")
		} else {
			fmt.Printf("調色盤（前 16 色）：%d/48 個分量不同\n", bad)
			for i := 0; i < 16; i++ {
				if q[i*3] != r[i*3] || q[i*3+1] != r[i*3+1] || q[i*3+2] != r[i*3+2] {
					fmt.Printf("  色號 %2d oracle=(%3d,%3d,%3d) remake=(%3d,%3d,%3d)\n",
						i, q[i*3], q[i*3+1], q[i*3+2], r[i*3], r[i*3+1], r[i*3+2])
				}
			}
		}
	}

	var n int
	minX, minY, maxX, maxY := W, H, -1, -1
	byPair := map[[2]uint8]int{}
	for i := range x {
		if x[i] != y[i] {
			n++
			px, py := i%W, i/W
			if px < minX {
				minX = px
			}
			if px > maxX {
				maxX = px
			}
			if py < minY {
				minY = py
			}
			if py > maxY {
				maxY = py
			}
			byPair[[2]uint8{x[i], y[i]}]++
		}
	}
	fmt.Printf("像素：%d 個，相同 %d，不同 %d（%.4f%%）\n", W*H, W*H-n, n, float64(n)*100/float64(W*H))
	if n == 0 {
		fmt.Println("**逐位元組一致**")
		return
	}
	fmt.Printf("差異 bounding box：(%d,%d)-(%d,%d)  %d×%d\n",
		minX, minY, maxX, maxY, maxX-minX+1, maxY-minY+1)
	fmt.Println("色號錯配（oracle → remake：次數）：")
	for k, v := range byPair {
		fmt.Printf("  %2d → %2d : %d\n", k[0], k[1], v)
	}
	if *diff != "" {
		img := image.NewRGBA(image.Rect(0, 0, W, H))
		for i := range x {
			c := color.RGBA{0, 0, 0, 255}
			if x[i] != y[i] {
				c = color.RGBA{255, 0, 0, 255}
			} else if x[i] != 0 {
				c = color.RGBA{40, 40, 40, 255}
			}
			img.Set(i%W, i/W, c)
		}
		f, err := os.Create(*diff)
		if err == nil {
			png.Encode(f, img)
			f.Close()
			fmt.Println("差分圖：", *diff)
		}
	}
	os.Exit(1)
}

func mustRead(p string, want int) []byte {
	b, err := os.ReadFile(p)
	if err != nil {
		fmt.Fprintln(os.Stderr, "讀不到", p, err)
		os.Exit(2)
	}
	if len(b) != want {
		fmt.Fprintf(os.Stderr, "%s 長度 %d，預期 %d\n", p, len(b), want)
		os.Exit(2)
	}
	return b
}
