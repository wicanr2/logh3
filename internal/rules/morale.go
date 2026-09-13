package rules

// 民心（星域記錄 `+0x7C`）每回合的變化（`docs/spec/040`、`docs/re/295`）。
//
// 民心只有三個寫入者：新遊戲初始化照擁有者設成 0／200／100（`Morale`，
// 原版 `2703:0000`）、每回合往擁有者的目標值走一格（這一支，原版
// `056A:0AA5`）、以及費沙事件把「費沙擁有者（限人類座位）的其他星域」
// 一律設成 100（`1213:03F5`，見 `RunPhezzanEvent`）。讀存檔時整張
// 星域表照檔案的用，民心不重算。

// moraleTarget 是 `056A:0AA5` 每回合收斂的目標：擁有者 0 → 0、2 → 100、
// 其餘 → 200。分支順序照原版（先比 0、再比 2、落到最後的是 200），與
// `Morale` 在 0／1／2 三個合法值上相同。
func moraleTarget(owner int) int {
	switch owner {
	case 0:
		return 0
	case 2:
		return 100
	}
	return 200
}

// DriftRegionMorale 把每一格的民心往那一格擁有者的目標值走一格
// （原版 `056A:0AA5`，每個回合在 `056A:000C` 那一串裡跑一次，排在逐支扣
// 維持費 `056A:09EB` 之後、艦隊移動 `056A:03E0` 之前）。
//
//	d = 民心 − 目標
//	d < 0 → 民心 + 1；d > 0 → 民心 − 1；d == 0 → 不動
//
// 擁有者讀的是**星域記錄的 `+0x12`**，不是行星現算——收據 `mrx295d`
// 把四格的 `+0x12` 改掉、行星不動，那四格的民心往改過的擁有者走。
// 民心是無號 byte，原版先零延伸再減（`mov al,es:[bx] / xor ah,ah /
// sub ax,dx`），所以 255 也是往下走，不會繞回。
func DriftRegionMorale(records []byte) {
	for i := 0; (i+1)*RegionRecord <= len(records); i++ {
		r := records[i*RegionRecord : (i+1)*RegionRecord]
		d := int(r[regionMorale]) - moraleTarget(int(r[regionOwner]))
		switch {
		case d < 0:
			r[regionMorale]++
		case d > 0:
			r[regionMorale]--
		}
	}
}
