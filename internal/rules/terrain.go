package rules

import (
	"encoding/json"
	"fmt"
	"os"
)

// 戰場地形的通行規則與集結點（`docs/re/100`）。
//
// 地圖一格 16×20，判定用的半格是 8×10。每個圖塊有一個四位元遮罩，
// 位元對到格子的四個象限；倒出來的遮罩只有 `0x00` 與 `0x0F` 兩種值，
// 所以「半格」這一層在原版的資料下沒有真的生效——照原版的形狀實作，
// 換一張遮罩表就會跟著變。

// RallyTile 是地圖上代表集結點的圖塊（`docs/re/100`）。
//
// 行星用的是 `PlanetTile`，取格子中央；集結點取格子的左上角。
const RallyTile = 1

// BattleTerrain 是通行判定用的兩張表，倒自執行期記憶體。
type BattleTerrain struct {
	BlockMask    []uint8 `json:"block_mask"`    // ds:9BB4[圖塊]
	QuadrantBits []uint8 `json:"quadrant_bits"` // ds:3C3E[象限]
	HalfCellW    int     `json:"half_cell_w"`
	HalfCellH    int     `json:"half_cell_h"`
	Footprint    int     `json:"footprint"` // 佔位框邊長
	Step         int     `json:"step"`      // 直線判定一次走多遠
}

// LoadBattleTerrain 讀 `docs/internals/data/battle-terrain.json`。
func LoadBattleTerrain(path string) (*BattleTerrain, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var t BattleTerrain
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, fmt.Errorf("%s：%w", path, err)
	}
	if len(t.BlockMask) == 0 || len(t.QuadrantBits) != 4 {
		return nil, fmt.Errorf("%s 的地形表不完整", path)
	}
	return &t, nil
}

// RallySite 是地圖上埋的一個集結點。
type RallySite struct {
	Col int `json:"col"`
	Row int `json:"row"`
	X   int `json:"x"`
	Y   int `json:"y"`
}

// RallySites 掃一張星域地圖，回上面的集結點（依列序）。
//
// 32 張圖每張都有 4–13 個，是作者一格一格埋進去的。
func RallySites(cells []uint8, tileW, tileH int) []RallySite {
	if tileW <= 0 || tileH <= 0 {
		return nil
	}
	var out []RallySite
	for i, c := range cells {
		if c != RallyTile {
			continue
		}
		col, row := i%planetMapCol, i/planetMapCol
		out = append(out, RallySite{Col: col, Row: row, X: col * tileW, Y: row * tileH})
	}
	return out
}

// blocked 回半格 (c,r) 擋不擋（原版 `1588:621`）。
func (t *BattleTerrain) blocked(cells []uint8, c, r int) bool {
	col, row := c>>1, r>>1
	if col < 0 || row < 0 || col >= planetMapCol {
		return true
	}
	i := row*planetMapCol + col
	if i >= len(cells) {
		return true
	}
	tile := int(cells[i])
	if tile >= len(t.BlockMask) {
		return true
	}
	quad := (r&1)*2 + (c & 1)
	return t.BlockMask[tile]&t.QuadrantBits[quad] != 0
}

// Passable 回 (x,y) 那個佔位框壓到的半格有沒有被擋住（原版 `1588:4EA`）。
func (t *BattleTerrain) Passable(cells []uint8, x, y int) bool {
	if t.HalfCellW <= 0 || t.HalfCellH <= 0 {
		return true
	}
	n := t.Footprint - 1
	for c := x / t.HalfCellW; c <= (x+n)/t.HalfCellW; c++ {
		for r := y / t.HalfCellH; r <= (y+n)/t.HalfCellH; r++ {
			if t.blocked(cells, c, r) {
				return false
			}
		}
	}
	return true
}

// LineOpen 回兩點之間沿直線走得通不通（原版 `1588:891`）。
//
// 每次沿方位前進 Step 像素試一格，剩下不足 Step 就直接到終點。方位取
// 一整圈 256 的 atan2，與原版同一個定義；原版真正走一步的
// `2BE1:026F` 還沒逆向，所以這裡用 `cos256`／`sin256` 自己算，
// 落點可能差一兩個像素。
func (t *BattleTerrain) LineOpen(cells []uint8, x1, y1, x2, y2 int) bool {
	step := t.Step
	if step <= 0 {
		step = 8
	}
	for i := 0; i < maxLineSteps; i++ {
		if x1 == x2 && y1 == y2 {
			return true
		}
		dx, dy := x2-x1, y2-y1
		if isqrt(dx*dx+dy*dy) < step {
			x1, y1 = x2, y2
			continue
		}
		a := atan256(dx, dy)
		x1 += cos256(a) * step / 1024
		y1 += sin256(a) * step / 1024
		if !t.Passable(cells, x1, y1) {
			return false
		}
	}
	return false
}

// maxLineSteps 擋住走不完的情形。整張戰場 2048×2040，對角線約 2900，
// 一步 8 像素——360 步走得完，這裡留一倍餘裕。
const maxLineSteps = 720

// RallyBox 回挑集結點用的方框（原版 `31DB:01B6` 的開頭）。
//
// 以艦隊與參考點為對角線，短的那一邊撐開成長的那一邊——**方框補成
// 正方形**，中心留在兩點的中點。
func RallyBox(fx, fy, refX, refY int) (loX, loY, hiX, hiY int) {
	w, h := abs32(fx-refX), abs32(fy-refY)
	loX, hiX = minInt(fx, refX), maxInt(fx, refX)
	loY, hiY = minInt(fy, refY), maxInt(fy, refY)
	if w >= h {
		mid := (fy + refY) / 2
		hiY = mid + w/2
		loY = hiY - w
	} else {
		mid := (fx + refX) / 2
		hiX = mid + h/2
		loX = hiX - h
	}
	return loX, loY, hiX, hiY
}

// PickRally 照原版 `31DB:01B6` 挑一支艦隊這一回合要去的點。
//
// 先收框內的候選，再隨機抽；抽到「第 n 個」（機率 1/(n+1)）就直接壓向
// 參考點。抽中的候選要通得過兩道直線判定——從艦隊到它、從它到參考點,
// 通不過就把它刪掉再抽。
func PickRally(t *BattleTerrain, cells []uint8, sites []RallySite,
	fx, fy, refX, refY int, rnd *Rand) (int, int) {
	loX, loY, hiX, hiY := RallyBox(fx, fy, refX, refY)
	var cand []RallySite
	for _, s := range sites {
		if s.X == fx && s.Y == fy {
			continue
		}
		if s.X <= loX || s.X >= hiX || s.Y <= loY || s.Y >= hiY {
			continue
		}
		cand = append(cand, s)
	}
	for {
		n := len(cand)
		i := 0
		if rnd != nil {
			i = rnd.Next() % (n + 1)
		}
		if i == n {
			return refX, refY
		}
		s := cand[i]
		if t == nil || cells == nil ||
			(t.LineOpen(cells, fx, fy, s.X, s.Y) && t.LineOpen(cells, s.X, s.Y, refX, refY)) {
			return s.X, s.Y
		}
		// 通不過就把它從候選裡刪掉再抽；抽到空的時候 n ＝ 0，
		// 下一輪一定落在「第 n 個」那一支，也就是壓向參考點。
		cand = append(cand[:i], cand[i+1:]...)
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
