package main

import (
	"path/filepath"
	"testing"
)

// 規格160：核定海報底圖在背景解碼。畫面執行緒拿到一次就交出所有權；
// 換組回來（貼圖已釋放）要能再解一次。用 -race 跑檢查並行。
func TestApprovedPosterPrefetch(t *testing.T) {
	p, err := loadPosterArt(filepath.Join("..", "..", "assets", "hd"))
	if err != nil || p == nil || p.approved == nil {
		t.Skip("缺已入庫定案海報素材", err)
	}
	art := p.approved
	name := posterBaseName(0)
	art.prefetch(name)
	art.prefetch(name) // 重複呼叫不重解
	art.wait()
	im, ready := art.take(name)
	if !ready || im == nil {
		t.Fatal("背景解碼完成後拿不到底圖")
	}
	want := art.files[name]
	if b := im.Bounds(); b.Dx() != want.w || b.Dy() != want.h {
		t.Fatalf("底圖尺寸 %v，清單是 %dx%d", b, want.w, want.h)
	}
	// 再拿一次：所有權已交出，回到「重新解碼中」。
	if im2, ready2 := art.take(name); ready2 && im2 != nil {
		t.Fatal("同一張底圖被交出兩次")
	}
	art.wait()
	if im3, ready3 := art.take(name); !ready3 || im3 == nil {
		t.Fatal("換組回來之後沒有重新解碼")
	}
}
