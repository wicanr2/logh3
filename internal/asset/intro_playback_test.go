package asset

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestIntroPlaybackSequenceMatchesOriginal 釘住 dosgolem `introfull` 量到的
// 玩家可見格序列：0–348、350–373、350–373、350–373。
func TestIntroPlaybackSequenceMatchesOriginal(t *testing.T) {
	root := filepath.Join("..", "..")
	data, err := os.ReadFile(filepath.Join(root, "org_game/extracted/IBM16.350"))
	if err != nil {
		t.Skipf("沒有原版素材，跳過：%v", err)
	}
	lengths, err := os.ReadFile(filepath.Join(root, "org_game/extracted/IBMFLS16.350"))
	if err != nil {
		t.Skipf("沒有原版素材，跳過：%v", err)
	}
	p, err := NewIntroPlaybackDecoder(data, lengths)
	if err != nil {
		t.Fatal(err)
	}
	if p.Frames() != 421 {
		t.Fatalf("播放格數 %d，預期 421", p.Frames())
	}

	// 先把前段來源解碼成每一格的累加結果，作為序列測試的獨立基準。
	raw, err := NewIntroDecoder(data, lengths)
	if err != nil {
		t.Fatal(err)
	}
	want := make([][32]byte, raw.Frames())
	for i := 0; i < 349; i++ {
		if ok, err := raw.Next(); err != nil || !ok {
			t.Fatalf("來源第 %d 格解碼失敗：%v", i, err)
		}
		want[i] = sha256.Sum256(raw.Pixels().Pix)
	}
	// 原版進尾段前清 plane，所以尾段的預期不是來源解碼器從 0
	// 一路累加的結果；另開一個來源解碼器重現「吃掉 349、清零、從
	// 350 開始累加」的狀態。
	tail, err := NewIntroDecoder(data, lengths)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i <= 349; i++ {
		if ok, err := tail.Next(); err != nil || !ok {
			t.Fatalf("尾段前來源第 %d 格解碼失敗：%v", i, err)
		}
	}
	for p := range tail.Planes {
		for i := range tail.Planes[p] {
			tail.Planes[p][i] = 0
		}
	}
	for i := 350; i < 374; i++ {
		if ok, err := tail.Next(); err != nil || !ok {
			t.Fatalf("尾段來源第 %d 格解碼失敗：%v", i, err)
		}
		want[i] = sha256.Sum256(tail.Pixels().Pix)
	}
	sequence := make([]int, 0, 421)
	for i := 0; i < 349; i++ {
		sequence = append(sequence, i)
	}
	for pass := 0; pass < 3; pass++ {
		for i := 350; i < 374; i++ {
			sequence = append(sequence, i)
		}
	}
	for display, source := range sequence {
		if ok, err := p.Next(); err != nil || !ok {
			t.Fatalf("播放第 %d 格失敗：%v", display, err)
		}
		if got := sha256.Sum256(p.Pixels().Pix); got != want[source] {
			t.Fatalf("播放第 %d 格不是來源第 %d 格", display, source)
		}
	}
	if ok, err := p.Next(); err != nil || ok {
		t.Fatalf("播放完 421 格後仍有下一格：ok=%v err=%v", ok, err)
	}
}

// TestIntroPlaybackMatchesDosgolemReceipts 比對 `introframes` 收據裡的尾段
// 色號陣列。收據是本機 oracle 證據，換機沒有收據時只跳過，不用代用品。
func TestIntroPlaybackMatchesDosgolemReceipts(t *testing.T) {
	root := filepath.Join("..", "..")
	data, err := os.ReadFile(filepath.Join(root, "org_game/extracted/IBM16.350"))
	if err != nil {
		t.Skipf("沒有原版素材，跳過：%v", err)
	}
	lengths, err := os.ReadFile(filepath.Join(root, "org_game/extracted/IBMFLS16.350"))
	if err != nil {
		t.Skipf("沒有原版素材，跳過：%v", err)
	}
	for _, f := range []int{348, 350, 355, 360, 365, 370, 373} {
		path := filepath.Join(root, "workplace/oracle/introframes/frames",
			fmt.Sprintf("introfull-f%03d.bin", f))
		if _, err := os.Stat(path); err != nil {
			t.Skipf("缺 dosgolem 收據 %s，跳過收據對拍", path)
		}
	}
	p, err := NewIntroPlaybackDecoder(data, lengths)
	if err != nil {
		t.Fatal(err)
	}
	for display := 0; display < p.Frames(); display++ {
		if ok, err := p.Next(); err != nil || !ok {
			t.Fatalf("播放第 %d 格失敗：%v", display, err)
		}
		// 收據檔名是來源格號；尾段三趟畫面應逐位元組一樣，
		// 只在第一次出現時抽查，避免重複讀同一份 oracle。
		source := display
		switch {
		case display >= 349 && display < 373:
			source = 350 + display - 349
		case display >= 373 && display < 397:
			source = 350 + display - 373
		case display >= 397:
			source = 350 + display - 397
		}
		if display != 348 && (display < 349 || source != 350+display-349 || display > 372) {
			continue
		}
		if source != 348 && source != 350 && source != 355 && source != 360 && source != 365 && source != 370 && source != 373 {
			continue
		}
		want, err := os.ReadFile(filepath.Join(root, "workplace/oracle/introframes/frames",
			fmt.Sprintf("introfull-f%03d.bin", source)))
		if err != nil {
			t.Fatal(err)
		}
		got := p.Pixels()
		page := make([]byte, 640*350)
		for y := 0; y < got.H; y++ {
			copy(page[(61+y)*640:(62+y)*640], got.Pix[y*got.W:(y+1)*got.W])
		}
		if len(want) != len(page) {
			t.Fatalf("來源格 %d 收據 %d bytes，remake %d bytes", source, len(want), len(page))
		}
		for i := range want {
			if want[i] != page[i] {
				t.Fatalf("來源格 %d 在第 %d 個色號不同：原版 %d remake %d", source, i, want[i], page[i])
			}
		}
	}
}
