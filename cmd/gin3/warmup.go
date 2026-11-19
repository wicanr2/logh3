package main

// 背景預熱（規格160）。啟動只載主選單前一定要用的東西，其餘首次使用才讀；
// 這裡在主選單出來之後，用一條背景 goroutine 把「之後很快會用到、第一次用
// 會卡一下」的資源先準備好。順序照玩家最先碰到的排。
//
// 每一項自己負責並行安全：畫面規格表與星圖表上鎖、投影地板用 sync.Once、
// 音效 PCM 快取上鎖、海報解碼只交出解好的圖。

// startPosterPrefetch 在載入尾段就開始解目前那一組海報（畫面最先看到的就是它）。
func (a *app) startPosterPrefetch() {
	if a.posters == nil || a.posters.approved == nil || !a.posterMode {
		return
	}
	art := a.posters.approved
	art.prefetch(posterBaseName(a.posters.selected))
	art.prefetch("hulls.png")
}

func posterBaseName(set int) string {
	return [...]string{"base-v1.png", "base-v2.png", "base-v3.png"}[set]
}

// startWarmUp 在畫面執行緒上先取好要預熱的對象，再交給背景 goroutine：
// app 的欄位（例如音效壞掉時 a.effects 會被清成 nil）只由畫面執行緒改，
// 背景不讀 app。回傳的 channel 在預熱結束時關閉。
func (a *app) startWarmUp() <-chan struct{} {
	g, hd := a.g, a.hd
	effects, _ := a.effects.(*effectPlayer)
	done := make(chan struct{})
	go func() {
		defer close(done)
		g.WarmResources()
		if effects != nil {
			effects.warm()
		}
		if hd != nil {
			hd.Warm()
		}
	}()
	return done
}
