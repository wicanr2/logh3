package game

import (
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// 情報收集畫面（`screen-info`，#63、`docs/re/342`）：三個分頁（治安／勢力／敵艦）
// 與「參謀」「確認」兩顆鈕。
//
//   - 分頁：把 `ds:8F2F[座位]` 設成動作碼（治安 0、敵艦 1、勢力 2），整塊重畫核取
//     方塊、標題與五行圖例，並照那一頁的規則替 32 個星域重新著色（三頁共用同一張
//     色表 `ds:1780`）。
//   - 確認：熱區閃一下選中色，**不重繪內容**，直接關窗回戰略地圖。
//     **分頁狀態留在 `ds:8F2F`**，下次開還停在同一頁（收據 `fcmd-info-reopen`）。
//   - 參謀：開「待命中提督資料」重新指派情報參謀——那是 #62，還沒接。
//   - 游標經過不反白（原版沒有 hover 分支，只有命中之後的動作分派）。

// 情報收集的三個分頁，順序與畫面上的鈕相同（原版存進 `ds:8F2F` 的是動作碼，
// 那是另一組編號：治安 0、敵艦 1、勢力 2）。
const (
	infoTabSecurity = iota
	infoTabPower
	infoTabEnemy
	infoTabCount
)

const (
	infoButtonY    = 378
	infoButtonW    = 38
	infoButtonH    = 22
	infoStaffX     = 525
	infoConfirmX   = 573
	infoTitleLine  = 2
	infoTitleX     = 336
	infoLegendLine = 3
	infoLegendX    = 400
	infoLegendRows = 5
)

// infoScreenClick 是情報收集畫面的左鍵。
func (g *Game) infoScreenClick(spec *screen.MainGameSpec, x, y int) {
	if spec == nil {
		return
	}
	if k := infoTabAt(spec, x, y); k >= 0 {
		g.infoTab[g.seatIndex()] = k
		return
	}
	// 確認 (573,378) 38×22 關窗回戰略地圖；參謀 (525,378) 是 #62，還沒接。
	if b := infoButtonAt(spec, x, y); b == 1 {
		g.state = StateCommandPhase
	}
}

// infoTabAt 回 (x, y) 落在第幾個分頁鈕上；不在回 −1。
func infoTabAt(spec *screen.MainGameSpec, x, y int) int {
	for _, w := range spec.Windows {
		if w == nil || w.Tabs == nil {
			continue
		}
		for _, r := range w.Tabs.Rows {
			if y < r.Y || y >= r.Y+w.Tabs.H {
				continue
			}
			for i, it := range r.Items {
				if x >= it.X && x < it.X+it.W {
					return i
				}
			}
		}
	}
	return -1
}

// infoButtonAt 回 (x, y) 落在哪一顆鈕：0 ＝ 參謀、1 ＝ 確認、−1 ＝ 都不是。
// 兩顆的框在 `screen-info.json` 的 `boxes` 最後兩筆（525／573，y 378，38×22）。
func infoButtonAt(spec *screen.MainGameSpec, x, y int) int {
	for _, w := range spec.Windows {
		if w == nil {
			continue
		}
		for _, b := range w.Boxes {
			if b.Y != infoButtonY || b.W != infoButtonW || b.H != infoButtonH {
				continue
			}
			if x < b.X || x >= b.X+b.W || y < b.Y || y >= b.Y+b.H {
				continue
			}
			switch b.X {
			case infoStaffX:
				return 0
			case infoConfirmX:
				return 1
			}
		}
	}
	return -1
}

// seatIndex 回現在這個座位在「一個座位一格」的陣列（`ds:8F1E` 當索引的那一批）
// 裡的位置。單人局只有座位 0；對戰兩個座位各記各的。
func (g *Game) seatIndex() int {
	if g.versus && g.faction == 1 {
		return 1
	}
	return 0
}

// infoScreenTab 回現在停在哪一頁。
func (g *Game) infoScreenTab() int {
	k := g.infoTab[g.seatIndex()]
	if k < 0 || k >= infoTabCount {
		return infoTabSecurity
	}
	return k
}

// applyInfoScreen 每一幀把核取方塊與那六行字換成目前這一頁的
// （標題在 y 202 x 336，五行圖例在 y 242 起 x 400，行距 20）。
func (g *Game) applyInfoScreen(spec *screen.MainGameSpec) {
	p := g.res.InfoPanel
	if spec == nil || p == nil {
		return
	}
	tab := g.infoScreenTab()
	setTabsOn(spec, tab)
	w := spec.Windows["overlay"]
	if w == nil {
		return
	}
	w.SetLineText(infoTitleLine, infoTitleX, p.TitleCodes(tab))
	for k := 0; k < infoLegendRows; k++ {
		w.SetLineText(infoLegendLine+k, infoLegendX, p.LegendCodes(tab, k))
	}
}

// infoStarMapColours 回這一頁 32 格各自的色號：每一格算一個 0–4 的等級，
// 查同一張色表（`ds:1780`）。`securityStarMap` 拿它畫圖。
func (g *Game) infoStarMapColours() []uint8 {
	n := len(g.res.Stellar) / rules.RegionRecord
	if n == 0 || g.res.Planets == nil {
		return nil
	}
	// 色表沒載到就用規則層那一份（兩邊是同一張 `ds:1780`）。
	colour := rules.InfoMapColour
	if p := g.res.InfoPanel; p != nil {
		colour = p.Colour
	}
	tab := g.infoScreenTab()
	div := rules.SecurityDivisor(g.res.Admiral,
		rules.StaffAdmiral(g.workRecord(), rules.PostIntel, g.faction))
	cols := make([]uint8, n)
	for i := range cols {
		rec := g.res.Stellar[i*rules.RegionRecord : (i+1)*rules.RegionRecord]
		band := 4
		switch tab {
		case infoTabPower:
			band = rules.OwnPowerBand(g.regionFleetGroups(i, g.faction))
		case infoTabEnemy:
			// 敵艦那一頁的等級先過一道門檻（`0A02:0672` → `2B99:028F` 查表）：
			// 表在 `ds:0D062`／`ds:0D066` 的遠指標後面，指到 **0xD0000 一帶的 UMB 區**，
			// 內容看起來像位元圖樣而不是距離值——**那張表是什麼還沒解**
			// （`docs/re/342` 附錄 B）。門檻決定要走「掃艦隊表拿確定答案」還是
			// 「只看 `+0x13` bit4 猜」，兩條路的算式都已經解出來（同一份附錄），
			// 但沒有門檻就選不了，所以整張先塗「情報資料不足」。
			band = 4
		default:
			own := g.res.Planets.RegionOwner(i) == g.faction
			band = rules.SecurityBand(int(rec[rules.RegionUnrest]), own, div)
		}
		cols[i] = colour(band)
	}
	return cols
}

// regionFleetGroups 回這一格上某一方所有艦隊的編制組數總和（艦隊記錄 `+0x0F` 加總）。
func (g *Game) regionFleetGroups(r, faction int) int {
	n := 0
	for _, f := range g.fleets {
		if f.Region == r && f.Faction == faction {
			n += f.Groups
		}
	}
	return n
}
