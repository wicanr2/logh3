package rules

// Fleet 是一支在戰略地圖上的艦隊。
//
// 對應原版執行期艦隊表（`ds:C074`，32 筆 × 122）已解出來的那幾個欄位
// （`docs/internals/33`）：
//
//	+0x01  Faction   陣營
//	+0x02  Number    艦隊編號
//	+0x04  0xFF ＝ 這一筆沒在用（remake 用切片長度表示）
//	+0x05  Adjutant  副官
//	+0x06  Region    所在星域
//	+0x07  Target    要去的星域；`NoTarget` ＝ 沒有命令
//
// 其餘 100 多個 byte 還沒解，remake 不假裝知道。
type Fleet struct {
	Number   int `json:"number"`   // 艦隊編號（1 起）
	Faction  int `json:"faction"`  // 0 帝國、1 同盟
	Admiral  int `json:"admiral"`  // 帶隊提督的索引
	Adjutant int `json:"adjutant"` // 副官的提督索引；-1 ＝ 沒有
	Region   int `json:"region"`   // 現在在哪一個星域
	Target   int `json:"target"`   // 要去哪一個星域
}

// NoTarget ＝ 沒有移動命令。原版在那一格填 `0xFF`（`docs/re/77`）。
const NoTarget = -1

// Moving 回這一支有沒有在移動。
func (f Fleet) Moving() bool { return f.Target != NoTarget && f.Target != f.Region }

// Advance 把一支艦隊往目的地推進一格，回有沒有動。
//
// 走的是**原版那張路徑表**（`StarGraph.Route`，每個陣營各一張），
// 不是相鄰圖上的最短路——兩者不一定相同（`docs/internals/26`）。
// 到了就把命令清掉，與原版把 `+0x07` 填回 `0xFF` 一致。
//
// ⚠ **原版走一格之後就停了**（`docs/re/77`，收據 `s2move`）：載入 S2
// 之後按六次「實行」，第 0 支只在**第一次**動了一格（星域 2 → 6），
// 之後五次一個 byte 都沒動，而命令 `+0x07` 還留在那裡沒被清掉。
// 為什麼停下來未解——可能是一次實行只走一格且要重新下令，可能要消耗
// 某種資源，也可能是存檔裡只剩一格的移動量。
//
// 這裡做成「呼叫一次走一格」，路線照原版的路徑表；**多久走一格由呼叫端
// 決定**。目前 `advanceTurn` 每回合叫一次，會一路走到目的地，
// 與原版不同。這是**已知的偏差**，在那個「為什麼停」解出來之前，
// 不要拿它宣稱與原版一致。
func (g *StarGraph) Advance(f *Fleet) bool {
	if f == nil || !f.Moving() {
		return false
	}
	route := g.Route(f.Region, f.Target, f.Faction)
	if len(route) < 2 {
		// 走不出來就把命令清掉，不要卡在原地每回合重算一次。
		f.Target = NoTarget
		return false
	}
	f.Region = route[1]
	if f.Region == f.Target {
		f.Target = NoTarget
	}
	return true
}

// MoveFleets 把所有艦隊各推進一格，回實際動了幾支。
func (g *StarGraph) MoveFleets(fleets []Fleet) int {
	n := 0
	for i := range fleets {
		if g.Advance(&fleets[i]) {
			n++
		}
	}
	return n
}

// CountFleets 數某一個星域裡某一個陣營有幾支艦隊。
//
// 原版星域記錄的 `+0x28`／`+0x29` 就是這兩個數，而且**每次都是重數
// 出來的**，不是累加維護的（`docs/re/40`）——remake 照做，
// 這樣狀態就不可能對不起來。
func CountFleets(fleets []Fleet, region, faction int) int {
	n := 0
	for _, f := range fleets {
		if f.Region == region && f.Faction == faction {
			n++
		}
	}
	return n
}
