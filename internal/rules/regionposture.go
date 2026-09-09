package rules

// 星域態勢重算（原版 `2703:0933`，`docs/re/244`、`docs/spec/032`）。
//
// 原版每個回合結束、逐格重算擁有者與處理易主之後，先重建路徑表
// （`2B99:0003`，`BuildTables`）再對 32 個星域各跑一次這一支。它先把
// `+0x13` 的 bit0｜bit2｜bit3 清掉（`and 0F2h`）再重填，所以 bit1 與
// bit4–bit7 是別處維護的，這裡不動。

// RegionOwnerNone 是星域記錄 `+0x12` 的「無主」。
const RegionOwnerNone = 2

// FarCost 是 `+0x13` bit0 的門檻：到自己首都的路徑成本 `>= 0x400`，
// 也就是路上至少有一條邊的兩端不全是自己的（`k >= 1`）。
//
// **不要跟 `2B99:0392` 的 `0x800` 搞混**——那個是「從首都走得到嗎」用的
// （`docs/re/185`）。
const FarCost = 0x400

// RegionPosture 是一個星域跑完 `2703:0933` 之後的四個可觀測量。
type RegionPosture struct {
	// CapitalCost 是 `+0x24`：到自己陣營首都的路徑成本。
	// **-1 ＝ 這一格無主，原版不寫**，呼叫端要留著上一次的值。
	CapitalCost int
	Far         bool // `+0x13` bit0：CapitalCost >= FarCost
	Contested   bool // `+0x13` bit2：這一格雙方都有艦隊
	Frontier    bool // `+0x13` bit3：至少一個鄰居的擁有者跟自己不同
}

// Bits 回這一格要寫進 `+0x13` 的三個位元（bit0｜bit2｜bit3）。
func (p RegionPosture) Bits() byte {
	var b byte
	if p.Far {
		b |= 0x01
	}
	if p.Contested {
		b |= 0x04
	}
	if p.Frontier {
		b |= 0x08
	}
	return b
}

// ApplyPosture 把三個位元填進既有的 `+0x13`，其餘位元原樣保留
// （原版的 `and 0F2h` ＋ `or`）。
func ApplyPosture(flags byte, p RegionPosture) byte {
	return flags&0xF2 | p.Bits()
}

// Postures 跑一輪 32 格的 `2703:0933`。
//
// `owner` 是 32 格的星域擁有者（`+0x12`）。`fleets` 只用得到每一支的
// `Region`／`Faction`／`Flags`——與原版 `297B:14F8` 一樣，**`Flags` 的
// bit0 有設的不計入 `Contested`**。
//
// 距離取自 `BuildTables(owner, 擁有者)`：路徑表是擁有者的函數，所以這裡
// 就地重建，不吃外面傳進來的舊表。
func (g *StarGraph) Postures(owner []int, fleets []Fleet) [32]RegionPosture {
	var dist [2][][]int
	for f := 0; f < 2; f++ {
		dist[f], _ = g.BuildTables(owner, f)
	}

	// 每一格雙方各有幾支艦隊（`297B:14F8`）。
	var count [32][2]int
	for _, f := range fleets {
		if f.Flags&1 != 0 || f.Region < 0 || f.Region >= 32 {
			continue
		}
		if f.Faction == 0 || f.Faction == 1 {
			count[f.Region][f.Faction]++
		}
	}

	var out [32]RegionPosture
	for i := 0; i < 32; i++ {
		p := RegionPosture{CapitalCost: -1}
		p.Contested = count[i][0] > 0 && count[i][1] > 0
		for _, j := range g.Neighbours(i) {
			if owner[j] != owner[i] {
				p.Frontier = true
				break
			}
		}
		if o := owner[i]; o == 0 || o == 1 {
			p.CapitalCost = dist[o][i][CapitalRegion(o)]
			p.Far = p.CapitalCost >= FarCost
		}
		out[i] = p
	}
	return out
}
